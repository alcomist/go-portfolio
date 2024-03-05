// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package excel_importer

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

type ExcelImporter struct {
	fileName string
}

func NewExcelImporter(f string) *ExcelImporter {

	return &ExcelImporter{fileName: f}
}

func (task *ExcelImporter) Execute() {

	f, err := excelize.OpenFile(task.fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}

	columns := make(map[int]string)

	for _, row := range rows {

		// 컬럼 값이 비어있을 경우 컬럼 값을 먼저 넣어준다.
		if len(columns) == 0 {
			for index, col := range row {
				columns[index] = col
			}
		} else {
			// 여기 부분은 실제 excel 데이터
		}
	}

}
