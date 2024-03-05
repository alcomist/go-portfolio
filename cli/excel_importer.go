// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/alcomist/go-portfolio/task/excel_importer"
)

func main() {

	ef := flag.String("f", "", "(required) Excel File Name")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	task := excel_importer.NewExcelImporter(*ef)
	task.Execute()
}
