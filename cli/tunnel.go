// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/alcomist/go-portfolio/internal/constant"
	"github.com/alcomist/go-portfolio/internal/glog"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
)

type KeepAliveConfig struct {

	// Interval is the amount of time in seconds to wait before the
	// tunnel client will send a keep-alive message to ensure some minimum
	// traffic on the SSH connection.
	Interval uint

	// CountMax is the maximum number of consecutive failed responses to
	// keep-alive messages the client is willing to tolerate before considering
	// the SSH connection as dead.
	CountMax uint
}

type (
	TunnelConfig struct {
		Tunnel map[string]tunnel
	}

	tunnel struct {
		Server   string
		Port     int
		Id       string
		Password string
		Bind     string

		name          string
		hostAddr      string
		bindAddr      string
		mode          byte // '>' for forward, '<' for reverse
		dialAddr      string
		retryInterval time.Duration
		keepAlive     *KeepAliveConfig
	}
)

func (tc TunnelConfig) String() string {

	var b bytes.Buffer

	for _, v := range tc.Tunnel {

		fmt.Fprintf(&b, "\nSERVER=%v\n", v.Server)
		fmt.Fprintf(&b, "PORT=%v\n", v.Port)
		fmt.Fprintf(&b, "ID=%v\n", v.Id)
		fmt.Fprintf(&b, "PASSWORD=%v\n", v.Password)
		fmt.Fprintf(&b, "BIND=%v\n", v.Bind)

		fmt.Fprintf(&b, "NAME=%v\n", v.name)
		fmt.Fprintf(&b, "HOST ADDR=%v\n", v.hostAddr)
		fmt.Fprintf(&b, "BIND-DIAL=%v -%c %v\n", v.bindAddr, v.mode, v.dialAddr)
		fmt.Fprintf(&b, "RETRY INTERVAL=%v\n", v.retryInterval)
		fmt.Fprintf(&b, "KEEP ALIVE INTERVAL/COUNT=%v:%v\n", v.keepAlive.Interval, v.keepAlive.CountMax)
	}

	return b.String()
}

// Get default location of a private key
func privateKeyPath() string {
	return os.Getenv("HOME") + "/.ssh/id_rsa"
}

// Get private key for ssh authentication
func parsePrivateKey(keyPath string) (ssh.Signer, error) {
	buff, _ := os.ReadFile(keyPath)
	return ssh.ParsePrivateKey(buff)
}

func getConfig(f string) TunnelConfig {

	tomlFile := ""
	if len(f) > 0 {

		dir, exist := os.LookupEnv(constant.EnvKeyConfigFileDir)
		if exist == false {
			log.Fatalln("GON_CONFIG_DIR environment variable not set")
		}

		filename := dir + f
		_, err := os.Stat(filename)
		if err != nil {
			log.Fatalln(err)
		}

		tomlFile = filename
	}

	if len(tomlFile) == 0 {

		_tomlFile, exist := os.LookupEnv(constant.EnvKeyTOMLFile)
		if exist == false {
			log.Fatalln("GON_TOML environment variable not set")
		}
		tomlFile = _tomlFile
	}

	var config TunnelConfig
	_, err := toml.DecodeFile(tomlFile, &config)
	if err != nil {
		log.Fatalln(err)
	}

	for k, tunnel := range config.Tunnel {

		tt := strings.Fields(tunnel.Bind)
		if len(tt) != 3 {
			log.Fatalf("invalid tunnel syntax: %s", tunnel.Bind)
		}

		// Parse for the tunnel endpoints.
		switch tt[1] {
		case "->":
			tunnel.bindAddr, tunnel.mode, tunnel.dialAddr = tt[0], '>', tt[2]
		case "<-":
			tunnel.dialAddr, tunnel.mode, tunnel.bindAddr = tt[0], '<', tt[2]
		default:
			log.Fatalf("invalid tunnel syntax: %s", tunnel.Bind)
		}

		for _, addr := range []string{tunnel.bindAddr, tunnel.dialAddr} {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				log.Fatalf("invalid endpoint: %s", addr)
			}
		}

		// Parse for the SSH target host.
		tunnel.name = k
		tunnel.hostAddr = tunnel.Host()
		tunnel.retryInterval = 30 * time.Second
		tunnel.keepAlive = &KeepAliveConfig{Interval: 30, CountMax: 2}

		config.Tunnel[k] = tunnel
	}

	return config
}

func (t tunnel) String() string {

	var left, right string
	mode := "<?>"
	switch t.mode {
	case '>':
		left, mode, right = t.bindAddr, "->", t.dialAddr
	case '<':
		left, mode, right = t.dialAddr, "<-", t.bindAddr
	}
	return fmt.Sprintf("[%s] %s (%s %s %s)", t.name, t.hostAddr, left, mode, right)
}

func (t tunnel) Host() string {

	if t.Port == 0 {
		t.Port = 22
	}

	hostAddr := fmt.Sprintf("%s:%d", t.Server, t.Port)
	if len(hostAddr) == 1 {
		log.Fatalf("invalid server: %s", hostAddr)
	}

	return hostAddr
}

// Get ssh client config for our connection
// SSH config will use 2 authentication strategies: by key and by password

func (t tunnel) SSHConfig() (*ssh.ClientConfig, error) {

	key, err := parsePrivateKey(privateKeyPath())
	if err != nil {
		return nil, err
	}

	clientConfig := ssh.ClientConfig{
		User: t.Id,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
			ssh.Password(t.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	return &clientConfig, nil
}

// Handle local client connections and thunnel data to the remote server
// Will use io.Copy - http://golang.org/pkg/io/#Copy
func handleClient(client net.Conn, remote net.Conn) {

	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Println("error while copy remote->local:", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Println(err)
		}
		chDone <- true
	}()

	<-chDone
}

func (t tunnel) keepAliveMonitor(once *sync.Once, wg *sync.WaitGroup, client *ssh.Client) {

	defer wg.Done()
	if t.keepAlive.Interval == 0 || t.keepAlive.CountMax == 0 {
		return
	}

	// Detect when the SSH connection is closed.
	wait := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		wait <- client.Wait()
	}()

	// Repeatedly check if the remote server is still alive.
	var aliveCount int32
	ticker := time.NewTicker(time.Duration(t.keepAlive.Interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-wait:
			if err != nil && err != io.EOF {
				once.Do(func() { log.Printf("(%v) SSH error: %v", t, err) })
			}
			return
		case <-ticker.C:
			if n := atomic.AddInt32(&aliveCount, 1); n > int32(t.keepAlive.CountMax) {
				once.Do(func() { log.Printf("(%v) SSH keep-alive termination", t) })
				client.Close()
				return
			}
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
			if err == nil {
				atomic.StoreInt32(&aliveCount, 0)
			}
		}()
	}
}

func (t tunnel) dialTunnel(ctx context.Context, wg *sync.WaitGroup, client *ssh.Client, cn1 net.Conn) {

	defer wg.Done()

	// The inbound connection is established. Make sure we close it eventually.
	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-connCtx.Done()
		cn1.Close()
	}()

	// Establish the outbound connection.
	var cn2 net.Conn
	var err error
	switch t.mode {
	case '>':
		cn2, err = client.Dial("tcp", t.dialAddr)
	case '<':
		cn2, err = net.Dial("tcp", t.dialAddr)
	}
	if err != nil {
		log.Printf("(%v) dial error: %v", t.dialAddr, err)
		return
	}

	go func() {
		<-connCtx.Done()
		cn2.Close()
	}()

	log.Printf("%v - connection established", t)
	defer log.Printf("%v - connection closed", t)

	// Copy bytes from one connection to the other until one side closes.
	var once sync.Once
	var wg2 sync.WaitGroup
	wg2.Add(2)
	go func() {
		defer wg2.Done()
		defer cancel()
		if _, err := io.Copy(cn1, cn2); err != nil {
			once.Do(func() { log.Printf("%v - connection error: %v", t, err) })
		}
		once.Do(func() {}) // Suppress future errors
	}()
	go func() {
		defer wg2.Done()
		defer cancel()
		if _, err := io.Copy(cn2, cn1); err != nil {
			once.Do(func() { log.Printf("%v - connection error: %v", t, err) })
		}
		once.Do(func() {}) // Suppress future errors
	}()
	wg2.Wait()
}

func (t tunnel) bindTunnel(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	sshConfig, err := t.SSHConfig()
	if err != nil {
		log.Println(err)
		return
	}

	for {
		var once sync.Once // Only print errors once per session

		func() {

			// Connect to the server host via SSH.
			client, err := ssh.Dial("tcp", t.hostAddr, sshConfig)
			if err != nil {
				once.Do(func() { log.Printf("%v - SSH dial error: %v", t, err) })
				return
			}
			wg.Add(1)
			go t.keepAliveMonitor(&once, wg, client)
			defer client.Close()

			// Attempt to bind to the inbound socket.
			var listener net.Listener
			switch t.mode {
			case '>':
				listener, err = net.Listen("tcp", t.bindAddr)
			case '<':
				listener, err = client.Listen("tcp", t.bindAddr)
			}
			if err != nil {
				once.Do(func() { log.Printf("%v - bind error: %v", t.bindAddr, err) })
				return
			}

			// The socket is binded. Make sure we close it eventually.
			bindCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			go func() {
				client.Wait()
				cancel()
			}()
			go func() {
				<-bindCtx.Done()
				once.Do(func() {}) // Suppress future errors
				listener.Close()
			}()

			log.Printf("%v - binded tunnel", t)
			defer log.Printf("%v - collapsed tunnel", t)

			// Accept all incoming connections.
			for {
				cn1, err := listener.Accept()
				if err != nil {
					once.Do(func() { log.Printf("%v accept error: %v", t, err) })
					return
				}
				wg.Add(1)
				go t.dialTunnel(bindCtx, wg, client, cn1)
			}
		}()

		select {
		case <-ctx.Done():
			return
		case <-time.After(t.retryInterval):
			log.Printf("(%v) retrying...", t)
		}
	}
}

func main() {

	mode := flag.String("m", "", "(required) mode (print config, tunnel)")
	file := flag.String("f", "", "(optional) external config file")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	exec := path.Base(os.Args[0])

	defer glog.Set(os.Args[0])()

	tunnelConfig := getConfig(*file)
	// return

	if *mode == "l" {
		log.Println(tunnelConfig)
		return
	}

	if *mode == "t" {

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			log.Printf("received %v - initiating shutdown", <-sigChan)
			cancel()
		}()

		log.Printf("%s starting", exec)
		defer log.Printf("%s shutdown", exec)

		var wg sync.WaitGroup

		for _, t := range tunnelConfig.Tunnel {
			wg.Add(1)
			go t.bindTunnel(ctx, &wg)
		}
		wg.Wait()
	}

	log.Println("all done")
}
