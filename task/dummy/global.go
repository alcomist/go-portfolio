// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dummy

import (
	"github.com/alcomist/go-portfolio/internal/config"
	"log"
)

type GlobalTasker struct {
	dummy string
}

func NewDummyTasker(d string) *GlobalTasker {

	return &GlobalTasker{dummy: d}
}

func (task *GlobalTasker) Execute() {

	log.Printf(config.DefaultIniFile())
	log.Println(task.dummy)
}
