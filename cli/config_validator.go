// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/alcomist/go-portfolio/internal/config"
	"strings"
)

func main() {

	var prefix string
	flag.StringVar(&prefix, "p", "", "(optional) Section Prefix")
	flag.Parse()

	for _, s := range config.Sections(prefix) {
		fmt.Println("[" + s.Name() + "]")
		for _, k := range s.Keys() {
			values := k.ValueWithShadows()
			if len(values) == 1 {
				fmt.Printf("=> %s=%s\n", k.Name(), values[0])
			} else {
				fmt.Printf("=> %s:\n%s\n", k.Name(), strings.Join(k.ValueWithShadows(), "\n"))
			}
		}
	}
}
