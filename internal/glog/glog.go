// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glog

import (
	"fmt"
	"github.com/alcomist/go-portfolio/internal/util"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

// trace()
// debug()
// info()
// warn()
// error()
// fatal()

const (
	LogPrefixDebug = "DEBUG : "
	LogPrefixInfo  = "INFO : "
	LogPrefixError = "ERROR : "
	LogPrefixFatal = "FATAL : "
)

func logFolder() string {

	dir := util.ExecutableDir() + "/logs"

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Println(err)
		return ""
	}

	return dir
}

func Set(fn string) func() {

	if len(fn) == 0 {
		return func() {}
	}

	folder := logFolder()

	base := filepath.Base(fn)
	fullPath := path.Clean(fmt.Sprintf("%s/%s.log", folder, base))

	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return func() {}
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(io.MultiWriter(f, os.Stdout))

	return func() {
		f.Close()
	}
}
