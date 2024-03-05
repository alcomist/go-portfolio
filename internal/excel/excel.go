// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package excel

import (
	"fmt"
	"github.com/alcomist/go-portfolio/internal/util"
	"github.com/xuri/excelize/v2"
	"strings"
)

type Excel struct {
	f      *excelize.File
	sheets map[string]string
}

func New() *Excel {

	return &Excel{}
}

func (e *Excel) Open(filename string) bool {

	f, err := excelize.OpenFile(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}

	e.f = f
	e.sheets = make(map[string]string)

	for _, sheet := range e.f.GetSheetList() {
		e.sheets[util.RemoveSpace(strings.ToLower(sheet))] = sheet
	}

	return true
}

func (e *Excel) Close() {
	e.f.Close()
}

func (e *Excel) Rows(s string) [][]string {

	val, ok := e.sheets[strings.ToLower(s)]
	if ok {
		rows, err := e.f.GetRows(val)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		return rows
	}

	return nil
}
