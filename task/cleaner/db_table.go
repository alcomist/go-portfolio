// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cleaner

import (
	"fmt"
	"github.com/alcomist/go-portfolio/internal/constant"
	"github.com/alcomist/go-portfolio/internal/database"
)

type DbTableOptimizer struct {
}

func NewDbTableOptimizer() *DbTableOptimizer {

	return &DbTableOptimizer{}
}

func (task *DbTableOptimizer) optimize() {

	db := database.MustGet(constant.CKDBMain)

	tables := db.Tables("")

	for _, table := range tables {

		if len(table) == 0 {
			continue
		}

		count := db.Count(table)
		if count == 0 {
			stmt := db.CreateStatement(table)

			backupTable := "bak_" + table

			if db.Rename(table, backupTable) {
				if db.Run(stmt.DDL, nil) != -1 {
					if db.Drop(backupTable) {
						fmt.Printf("RENAME / CREATE / DROP SUCCESS : %s\n", table)
					}
				}
			}

			//fmt.Println(table)
			//fmt.Println(stmt)

			//break
		}
	}
	//fmt.Println(tables)
}

func (task *DbTableOptimizer) Execute() {

	task.optimize()
}
