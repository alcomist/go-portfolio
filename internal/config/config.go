// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"github.com/alcomist/go-portfolio/internal/util"
	"gopkg.in/ini.v1"
	"log"
	"strings"
)

func Load(f string) (*ini.File, error) {

	return ini.ShadowLoad(f)
}

func Sections(p string) []*ini.Section {

	sections := make([]*ini.Section, 0)

	f, err := Load(DefaultIniFile())
	if err != nil {
		log.Println(err)
		return sections
	}

	if len(p) == 0 {
		return f.Sections()
	}

	for _, s := range f.Sections() {
		sname := s.Name()
		if strings.Index(sname, p) != -1 {
			sections = append(sections, s)
		}
	}

	return sections
}

func DefaultIniFile() string {

	return util.ExecutableDir() + "/config/config.ini"
}

func DefaultTomlFile() string {

	return util.ExecutableDir() + "/config/config.toml"
}

func MustGet(s string) *ini.Section {

	file := DefaultIniFile()

	f, err := Load(file)
	if err != nil {
		log.Fatal(err)
	}

	if !f.HasSection(s) {
		log.Fatal(fmt.Errorf("ini file has no section : %v", s))
	}

	return f.Section(s)
}
