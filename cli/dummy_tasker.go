// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/alcomist/go-portfolio/internal/glog"
	"github.com/alcomist/go-portfolio/task/dummy"
	"os"
)

func main() {

	defer glog.Set(os.Args[0])()

	dum := flag.String("d", "d", "(optional) dummy")

	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	tasker := dummy.NewDummyTasker(*dum)
	tasker.Execute()
}
