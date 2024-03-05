// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflects

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func AsString(a any) string {

	t := reflect.TypeOf(a)
	v := reflect.ValueOf(a)

	switch t.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Float64:
		return fmt.Sprintf("%d", int(v.Float()))
	case reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	default:
		panic(fmt.Sprintf("unexpected type (%s)", t))
	}

	return ""
}

func AsStrings(a any) []string {

	v := AsString(a)
	if len(v) > 0 {
		return strings.Split(v, ",")
	}

	return []string{""}
}

func AsInts(a any) []int {

	vs := AsStrings(a)

	r := make([]int, 0, len(vs))
	for _, v := range vs {
		i, err := strconv.Atoi(v)
		if err == nil && i > 0 {
			r = append(r, i)
		}
	}

	return r
}

func AsInt(a any) int {

	t := reflect.TypeOf(a)
	v := reflect.ValueOf(a)

	switch t.Kind() {
	case reflect.String:
		i, err := strconv.Atoi(v.String())
		if err != nil {
			return 0
		}
		return i
	case reflect.Float64:
		return int(v.Float())
	case reflect.Int64:
		return int(v.Int())
	default:
		panic(fmt.Sprintf("unexpected type (%s)", t))
	}
}
