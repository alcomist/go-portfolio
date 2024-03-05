// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package database

import (
	"bytes"
	"fmt"
	"strings"
)

type insertBuilder struct {
	table   string
	keys    []string
	values  []string
	updates []string

	arg map[string]any
}

func newInsertBuilder() *insertBuilder {

	return &insertBuilder{
		keys:   make([]string, 0),
		values: make([]string, 0),
		arg:    make(map[string]any),
	}
}

func (b *insertBuilder) Table(t string) {
	b.table = t
}

func (b *insertBuilder) AddCond(k string, op string, v any) {

}

func (b *insertBuilder) AddSet(k string, v any) {

	if len(k) > 0 {
		b.arg[k] = v
		b.keys = append(b.keys, fmt.Sprintf("`%s`", k))
		b.values = append(b.values, fmt.Sprintf(":%s", k))
	}
}

func (b *insertBuilder) AddUpdate(k string, v any) {

	if len(k) > 0 {
		nk := fmt.Sprintf("update_%s", k)
		b.arg[nk] = v
		b.updates = append(b.updates, fmt.Sprintf("`%s`=:%s", k, nk))
	}
}

func (b *insertBuilder) Build() (string, map[string]any) {

	// INSERT INTO table_name (columns) VALUES(values) WHERE condition = value;
	// ON DUPLICATE KEY UPDATE `marker_value`=:marker_value, `rtime`=NOW()";

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("INSERT INTO `%s` ", b.table))
	buf.WriteString(fmt.Sprintf("(%s) ", strings.Join(b.keys, ", ")))
	buf.WriteString(fmt.Sprintf("VALUES(%s) ", strings.Join(b.values, ", ")))

	if len(b.updates) > 0 {
		buf.WriteString(fmt.Sprintf("ON DUPLICATE KEY UPDATE %s;", strings.Join(b.updates, ", ")))
	}

	return buf.String(), b.arg
}
