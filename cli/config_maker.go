// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/alcomist/go-portfolio/internal/config"
	"github.com/alcomist/go-portfolio/internal/constant"
	"github.com/alcomist/go-portfolio/internal/database"
	"log"
	"strings"
)

func main() {

	f, err := config.Load(config.DefaultIniFile())
	if err != nil {
		log.Fatalln(err)
	}

	db := database.MustGet(constant.CKDBMain)

	// db configs
	dbConfigs := db.DBConfigs()

	for _, c := range dbConfigs {

		if f.HasSection(c.Name) {
			continue
		}

		f.Section(c.Name).Key("adapter").SetValue(c.Adapter)
		f.Section(c.Name).Key("host").SetValue(c.Host)
		f.Section(c.Name).Key("port").SetValue(fmt.Sprintf("%d", c.Port))
		f.Section(c.Name).Key("username").SetValue(c.Username)
		f.Section(c.Name).Key("password").SetValue(c.Password)
		f.Section(c.Name).Key("dbname").SetValue(c.DBName)
		f.Section(c.Name).Key("charset").SetValue(c.Charset)
	}

	// es configs (host has shadow values)
	names := db.EsConfigClusterNames()

	for _, name := range names {

		if f.HasSection(name) {
			log.Printf("%s section already exists", name)
			continue
		}

		ips := db.EsConfigInternalIPs(name)

		for _, ip := range ips {

			if !strings.HasPrefix(ip, "http://") {
				ip = "http://" + ip
			}

			err := f.Section(name).Key("host").AddShadow(ip)
			if err != nil {
				log.Println(err)
			}
		}
	}

	// slack configs
	provider := "slack"
	hooks := db.Webhooks(provider)
	for _, hook := range hooks {
		f.Section(provider).Key(hook.Channel).SetValue(hook.URL)
	}

	err = f.SaveTo(config.DefaultIniFile())
	if err != nil {
		log.Fatalln(err)
	}

}
