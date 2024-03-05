// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package database

import (
	"fmt"
	"github.com/alcomist/go-portfolio/internal/constant"
	"log"
)

func quoteString(x any) string {

	switch x := x.(type) {
	case nil:
		return "null"
	case int, uint:
		return fmt.Sprintf("%d", x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case string:
		return fmt.Sprintf("%s", x)
	case []int:
		return ""
	case []string:
		return ""
	default:
		log.Panicf("unnexpected type %T: %v", x, x)
		return ""
	}
}

type TableSetter interface {
	Table(t string)
}

type CondAdder interface {
	AddCond(k string, op string, v any)
}

type Setter interface {
	AddSet(k string, v any)
}

type Updater interface {
	AddUpdate(k string, v any)
}

type Builder interface {
	Build() (string, map[string]any)
}

type QueryBuilder interface {
	TableSetter
	CondAdder
	Setter
	Updater
	Builder
}

func NewQueryBuilder(op string) QueryBuilder {

	if len(op) == 0 {
		panic("no op")
	}

	var builder QueryBuilder
	if op == constant.QueryTypeCreate {
		builder = newCreateBuilder()
	} else if op == constant.QueryTypeUpdate {
		builder = newUpdateBuilder()
	} else if op == constant.QueryTypeInsert {
		builder = newInsertBuilder()
	} else if op == constant.QueryTypeDelete {
		builder = newDeleteBuilder()
	} else {
		panic("not allowed query type")
	}

	return builder
}
