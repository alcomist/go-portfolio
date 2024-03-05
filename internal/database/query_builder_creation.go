// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package database

import (
	"bytes"
	"fmt"
	"strings"
)

type createBuilder struct {
	table string
	stmt  []string
	arg   map[string]any
}

func newCreateBuilder() *createBuilder {

	return &createBuilder{
		stmt: make([]string, 0),
		arg:  make(map[string]any),
	}
}

func (b *createBuilder) Table(t string) {
	b.table = t
}

func (b *createBuilder) AddCond(k string, op string, v any) {

}

func (b *createBuilder) AddSet(k string, v any) {

}

func (b *createBuilder) AddUpdate(k string, v any) {

}

func (b *createBuilder) Build() (string, map[string]any) {

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("CREATE TABLE `%s` (", b.table))
	buf.WriteString(strings.Join(b.stmt, ", "))
	buf.WriteString(") ")
	buf.WriteString("ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_unicode_ci ;")

	return string(buf.Bytes()), b.arg
}
