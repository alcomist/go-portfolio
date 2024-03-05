// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package database

import (
	"bytes"
	"fmt"
	"github.com/alcomist/go-portfolio/internal/constant"
	"reflect"
	"strings"
)

type updateBuilder struct {
	table  string
	sets   []string
	wheres []string
	arg    map[string]any
}

func newUpdateBuilder() *updateBuilder {

	return &updateBuilder{
		sets:   make([]string, 0),
		wheres: make([]string, 0),
		arg:    make(map[string]any),
	}
}

func (b *updateBuilder) Table(t string) {
	b.table = t
}

func (b *updateBuilder) AddCond(k string, op string, v any) {

	if len(k) > 0 {

		if op == constant.IN {
			val := reflect.ValueOf(v)
			if val.Kind() == reflect.Slice {
				v = quoteSlice(v)
			}
		}

		if op == constant.IN {
			b.wheres = append(b.wheres, fmt.Sprintf("`%s` %s (%v)", k, op, v))
		} else {
			b.arg[k] = v
			b.wheres = append(b.wheres, fmt.Sprintf("`%s`%s:%s", k, op, k))
		}
	}
}

func (b *updateBuilder) AddSet(k string, v any) {

	if len(k) > 0 {
		nk := fmt.Sprintf("set_%s", k)

		b.arg[nk] = v
		b.sets = append(b.sets, fmt.Sprintf("`%s`=:%s", k, nk))
	}
}

func (b *updateBuilder) AddUpdate(k string, v any) {

}

func (b *updateBuilder) Build() (string, map[string]any) {

	if len(b.wheres) == 0 {
		panic("not allowed in update query (no where statements)")
	}

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("UPDATE `%s` ", b.table))
	buf.WriteString(fmt.Sprintf("SET %s ", strings.Join(b.sets, ", ")))

	buf.WriteString(fmt.Sprintf("WHERE %s ", strings.Join(b.wheres, " AND ")))

	return buf.String(), b.arg
}
