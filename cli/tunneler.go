// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/alcomist/go-portfolio/internal/glog"
	"github.com/alcomist/go-portfolio/task/tunneler"
	"os"
)

func main() {

	mode := flag.String("mode", "", "(required) mode (t : start tunneling, l : print config)")
	file := flag.String("file", "", "(optional) external config file")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	defer glog.Set(os.Args[0])()

	task := tunneler.New(*file, *mode)
	task.Execute()
}
