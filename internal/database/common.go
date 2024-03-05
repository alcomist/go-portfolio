// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package database

import (
	"fmt"
	"log"
)

type CreateTableStatement struct {
	Table string `db:"Table"`
	DDL   string `db:"Create Table"`
}

func (db *DB) Exist(t string) bool {

	if len(t) == 0 {
		return false
	}

	q := fmt.Sprintf("SHOW TABLES LIKE '%s';", t)

	r := make([]string, 0)
	err := db.Select(&r, q)
	if err != nil {
		log.Println(err)
		return false
	}

	if len(r) > 0 {
		return true
	}

	return false
}

func (db *DB) Run(q string, arg map[string]any) int64 {

	result, err := db.NamedExec(q, arg)

	if err != nil {
		log.Println(err)
		return -1
	}

	count, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		return -1
	}

	return count
}
func (db *DB) Tables(p string) []string {

	q := "SHOW TABLES "

	if len(p) > 0 {
		q += fmt.Sprintf("LIKE '%s%%'", p)
	}

	tables := make([]string, 0)
	err := db.Select(&tables, q)
	if err != nil {
		log.Println(err)
	}

	return tables
}

func (db *DB) Count(t string) int {

	count := 0

	err := db.Get(&count, fmt.Sprintf("SELECT COUNT(*) AS count FROM %s", t))
	if err != nil {
		log.Println(err)
	}

	return count
}

func (db *DB) CreateStatement(t string) CreateTableStatement {

	stmt := CreateTableStatement{}

	err := db.Get(&stmt, fmt.Sprintf("SHOW CREATE TABLE %s;", t))
	if err != nil {
		log.Println(err)
	}

	return stmt
}

func (db *DB) Rename(s, t string) bool {

	if s == t {
		return false
	}

	q := fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`;", s, t)

	if db.Run(q, nil) != -1 {
		return true
	}
	return false
}

func (db *DB) Drop(t string) bool {

	if len(t) == 0 {
		return false
	}

	q := fmt.Sprintf("DROP TABLE `%s`;", t)

	if db.Run(q, nil) != -1 {
		return true
	}
	return false
}

type DBColumn struct {
	Field   string  `db:"Field"`
	Type    string  `db:"Type"`
	Null    string  `db:"Null"`
	Key     string  `db:"Key"`
	Default *string `db:"Default"`
	Extra   string  `db:"Extra"`
}

func (db *DB) Columns(t string) []string {

	q := fmt.Sprintf("SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '%s'", t)

	columns := make([]string, 0)
	err := db.Select(&columns, q)
	if err != nil {
		log.Println(err)
	}

	return columns
}
