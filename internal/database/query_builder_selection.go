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

type DqlBuilderArg struct {
	arg  map[string]any
	args []any
}

func (a *DqlBuilderArg) new() {
	a.arg = make(map[string]any)
	a.args = make([]any, 0)
}

type DqlBuilder struct {
	named   bool
	table   string
	joins   []string
	columns []string
	wheres  []string
	orders  []string
	groups  []string
	having  []string
	limit   int
	offset  int
	arg     DqlBuilderArg
}

func NewDqlBuilder() *DqlBuilder {

	b := DqlBuilder{
		joins:   make([]string, 0),
		columns: make([]string, 0),
		wheres:  make([]string, 0),
		orders:  make([]string, 0),
		groups:  make([]string, 0),
		having:  make([]string, 0),
		offset:  -1,
	}
	b.arg.new()
	return &b
}

func quoteSlice(x any) string {

	switch x := x.(type) {
	case []int:
		res := make([]string, 0, len(x))
		for _, xs := range x {
			res = append(res, fmt.Sprintf("%d", xs))
		}
		return strings.Join(res, ", ")
	case []uint64:
		res := make([]string, 0, len(x))
		for _, xs := range x {
			res = append(res, fmt.Sprintf("%d", xs))
		}
		return strings.Join(res, ", ")
	case []string:
		res := make([]string, 0, len(x))
		for _, xs := range x {
			if len(xs) > 0 {
				res = append(res, fmt.Sprintf(`"%s"`, xs))
			}
		}
		return strings.Join(res, ", ")
	default:
		return ""
	}
}

func (b *DqlBuilder) Named(n bool) {
	b.named = n
}

func (b *DqlBuilder) Table(t string) {
	b.table = t
}

func (b *DqlBuilder) AddJoin(j string) {

	b.joins = append(b.joins, j)
}

func (b *DqlBuilder) AddColumn(c ...string) {

	b.columns = append(b.columns, c...)
}

func (b *DqlBuilder) AddCond(k string, op string, v any) {

	if len(k) > 0 {

		if strings.Index(k, ".") == -1 {
			k = fmt.Sprintf("`%s`", k)
		}

		if op == constant.IN {
			val := reflect.ValueOf(v)
			if val.Kind() == reflect.Slice {
				v = quoteSlice(v)
			}
		}

		if op == constant.IN {
			b.wheres = append(b.wheres, fmt.Sprintf("%s %s (%v)", k, op, v))
		} else {

			if b.named {
				b.arg.arg[k] = v
				b.wheres = append(b.wheres, fmt.Sprintf("%s%s:%s", k, op, k))
			} else {
				b.arg.args = append(b.arg.args, v)
				b.wheres = append(b.wheres, fmt.Sprintf("%s%s?", k, op))
			}
		}

	}
}

func (b *DqlBuilder) AddGroup(g ...string) {

	if len(g) > 0 {
		b.groups = append(b.groups, g...)
	}
}

func (b *DqlBuilder) AddHaving(h ...string) {

	if len(h) > 0 {
		b.having = append(b.having, h...)
	}
}

func (b *DqlBuilder) AddOrder(k string, o string) {

	if len(o) > 0 {
		b.orders = append(b.orders, fmt.Sprintf("%s %s", k, o))
	} else {
		b.orders = append(b.orders, fmt.Sprintf("%s", k))
	}
}

func (b *DqlBuilder) Limit(l int) {
	b.limit = l
}

func (b *DqlBuilder) Offset(o int) {
	b.offset = o
}

func (b *DqlBuilder) build() string {

	if len(b.table) == 0 {
		panic("no table name in dql builder")
	}

	var buf bytes.Buffer

	column := strings.Join(b.columns, " , ")
	if len(column) == 0 {
		column = "*"
	}

	fmt.Fprintf(&buf, "SELECT %s FROM %s ", column, b.table)

	if len(b.joins) > 0 {
		fmt.Fprintf(&buf, "%s ", strings.Join(b.joins, " "))
	}
	if len(b.wheres) > 0 {
		fmt.Fprintf(&buf, "WHERE %s ", strings.Join(b.wheres, " AND "))
	}
	if len(b.groups) > 0 {
		fmt.Fprintf(&buf, "GROUP BY %s ", strings.Join(b.groups, " , "))
	}
	if len(b.having) > 0 {
		fmt.Fprintf(&buf, "HAVING %s ", b.having)
	}
	if len(b.orders) > 0 {
		fmt.Fprintf(&buf, "ORDER BY %s ", strings.Join(b.orders, " , "))
	}

	if b.limit > 0 {
		if b.offset != -1 {
			fmt.Fprintf(&buf, "LIMIT %d, %d ", b.offset, b.limit)
		} else {
			fmt.Fprintf(&buf, "LIMIT %d", b.limit)
		}
	}
	return buf.String()
}

func (b *DqlBuilder) Build() (string, DqlBuilderArg) {

	return b.build(), b.arg
}
