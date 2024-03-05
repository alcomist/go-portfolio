// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tunneler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/alcomist/go-portfolio/internal/config"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
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
		ID       string
		Password string
		Bind     string

		name          string
		hostAddr      string
		bindAddr      string
		mode          byte // '>' for forward, '<' for reverse
		dialAddr      string
		retryInterval time.Duration
		keepAlive     KeepAliveConfig
	}
)

func (tc TunnelConfig) String() string {

	var b bytes.Buffer

	for _, v := range tc.Tunnel {

		fmt.Fprintf(&b, "\nNAME = %v\n", v.name)
		fmt.Fprintf(&b, "\tSERVER = %v\n", v.Server)
		fmt.Fprintf(&b, "\tPORT = %v\n", v.Port)
		fmt.Fprintf(&b, "\tID = %v\n", v.ID)
		fmt.Fprintf(&b, "\tPASSWORD = %v\n", v.Password)
		fmt.Fprintf(&b, "\tBIND = %v\n", v.Bind)

		fmt.Fprintf(&b, "\tHOST ADDR = %v\n", v.hostAddr)
		fmt.Fprintf(&b, "\tBIND-DIAL = %v -%c %v\n", v.bindAddr, v.mode, v.dialAddr)
		fmt.Fprintf(&b, "\tRETRY INTERVAL = %v\n", v.retryInterval)
		fmt.Fprintf(&b, "\tKEEP ALIVE INTERVAL/COUNT = %v/%v\n", v.keepAlive.Interval, v.keepAlive.CountMax)
	}

	return b.String()
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

// Get default location of a private key
func (t tunnel) privateKeyPath() string {

	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	path := ""
	if runtime.GOOS == "windows" {
		path = dirname + `\.ssh\id_rsa`
	} else {
		path = dirname + "/.ssh/id_rsa"
	}

	return path
}

// Get private key for ssh authentication
func (t tunnel) parsePrivateKey(keyPath string) (ssh.Signer, error) {
	buff, _ := os.ReadFile(keyPath)
	return ssh.ParsePrivateKey(buff)
}

func (t tunnel) host() string {

	if t.Port == 0 {
		t.Port = 22
	}

	hostAddr := fmt.Sprintf("%s:%d", t.Server, t.Port)
	if len(hostAddr) == 1 {
		log.Fatalf("invalid server: %s", hostAddr)
	}

	return hostAddr
}

func (t tunnel) monitor(once *sync.Once, wg *sync.WaitGroup, client *ssh.Client) {

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

func (t tunnel) dial(ctx context.Context, wg *sync.WaitGroup, client *ssh.Client, cn1 net.Conn) {

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

func (t tunnel) bind(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	sshConfig, err := t.SSHConfig()
	if err != nil {
		log.Println(err)
		return
	}

	for {

		var once sync.Once // Only print errors once per session

		func() {

			client, err := ssh.Dial("tcp", t.hostAddr, sshConfig)
			if err != nil {
				once.Do(func() { log.Printf("%v - SSH dial error: %v", t, err) })
				return
			}

			wg.Add(1)

			go t.monitor(&once, wg, client)

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
					once.Do(func() { log.Printf("%v - accept error: %v", t, err) })
					return
				}
				wg.Add(1)
				go t.dial(bindCtx, wg, client, cn1)
			}
		}()

		select {
		case <-ctx.Done():
			return
		case <-time.After(t.retryInterval):
			log.Printf("%v - retrying...", t)
		}
	}
}

type Tunneler struct {
	name    string
	logfile string
	mode    string
	config  TunnelConfig
}

func New(f, m string) *Tunneler {

	return &Tunneler{name: "Tunneler", logfile: f, mode: m}
}

func (task *Tunneler) loadToml() {

	tomlFile := ""
	if len(task.logfile) > 0 {

		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}

		filename := cwd + "/" + task.logfile
		_, err = os.Stat(filename)
		if err != nil {
			log.Fatalln(err)
		}

		tomlFile = filename
	}

	if len(tomlFile) == 0 {
		tomlFile = config.DefaultTomlFile()
	}

	var tunnelConfig TunnelConfig
	_, err := toml.DecodeFile(tomlFile, &tunnelConfig)
	if err != nil {
		log.Fatalln(err)
	}

	task.config = tunnelConfig
}

func overrideDefault(def, t tunnel) tunnel {

	if len(t.Server) == 0 && len(def.Server) > 0 {
		t.Server = def.Server
		t.Port = def.Port
		t.ID = def.ID
		t.Password = def.Password
	}

	return t
}

func (task *Tunneler) loadConfig() {

	task.loadToml()

	config := task.config

	def, ok := config.Tunnel["default"]

	for k, tunnel := range config.Tunnel {

		if ok {
			tunnel = overrideDefault(def, tunnel)
		}

		tunnel.name = k
		tunnel.hostAddr = tunnel.host()
		tunnel.retryInterval = 30 * time.Second
		tunnel.keepAlive = KeepAliveConfig{Interval: 30, CountMax: 2}

		if tunnel.name == "default" {
			config.Tunnel[k] = tunnel
			continue
		}

		tt := strings.Fields(tunnel.Bind)
		if len(tt) != 3 {
			log.Fatalf("[%s] invalid tunnel syntax: %s", tunnel.name, tunnel.Bind)
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

		config.Tunnel[k] = tunnel
	}

	delete(task.config.Tunnel, "default")

	task.config = config
}

func (t tunnel) SSHConfig() (*ssh.ClientConfig, error) {

	key, err := t.parsePrivateKey(t.privateKeyPath())
	if err != nil {
		return nil, err
	}

	clientConfig := ssh.ClientConfig{
		User: t.ID,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
			ssh.Password(t.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	return &clientConfig, nil
}

func (task *Tunneler) Execute() {

	task.loadConfig()

	if task.mode == "l" {
		task.printConfig()
	} else if task.mode == "t" {
		task.tunnel()
	}
}

func (task *Tunneler) printConfig() {

	log.Println(task.config)
}

func (task *Tunneler) tunnel() {

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("received %v - initiating shutdown", <-ch)
		cancel()
	}()

	log.Printf("%s starting", task.name)
	defer log.Printf("%s shutdown", task.name)

	var wg sync.WaitGroup

	for _, t := range task.config.Tunnel {
		wg.Add(1)
		go t.bind(ctx, &wg)
	}

	wg.Wait()

}
