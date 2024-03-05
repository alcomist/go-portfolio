// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"log"
	"os"
	"path"
)

func ExecutableDir() string {

	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
		return "."
	}

	return path.Dir(ex)
}
