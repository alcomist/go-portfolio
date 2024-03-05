// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package database

import (
	"bytes"
	"fmt"
	"strings"
)

type deleteBuilder struct {
	table  string
	wheres []string
	arg    map[string]any
}

func newDeleteBuilder() *deleteBuilder {

	return &deleteBuilder{
		wheres: make([]string, 0),
		arg:    make(map[string]any),
	}
}

func (b *deleteBuilder) Table(t string) {
	b.table = t
}

func (b *deleteBuilder) AddCond(k string, op string, v any) {

	if len(k) > 0 {
		b.arg[k] = v
		b.wheres = append(b.wheres, fmt.Sprintf("`%s`%s:%s", k, op, k))
	}
}

func (b *deleteBuilder) AddSet(k string, v any) {

}

func (b *deleteBuilder) AddUpdate(k string, v any) {

}

func (b *deleteBuilder) Build() (string, map[string]any) {

	if len(b.wheres) == 0 {
		panic("not allowed in delete query (no where statements)")
	}

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("DELETE FROM `%s` ", b.table))
	buf.WriteString(fmt.Sprintf("WHERE %s", strings.Join(b.wheres, " AND ")))

	return buf.String(), b.arg
}
