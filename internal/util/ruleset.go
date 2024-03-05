// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"reflect"
	"strings"
)

func ConvertToCatalogKeywords(rules []any) ([]string, []string, []string) {

	includes := make([]string, 0)
	excludes := make([]string, 0)
	replaces := make([]string, 0)

	for _, rule := range rules {
		val := reflect.ValueOf(rule)
		if val.Kind() == reflect.Slice {
			rval := val.Interface().([]any)

			if val.Len() == 2 {

				t := rval[0].(string)
				v := rval[1].(string)

				switch t {
				case "include":
					includes = append(includes, v)
				case "exclude":
					excludes = append(excludes, v)
				case "replace":
					replaces = append(replaces, v)
				}
			} else if val.Len() == 3 {

				t := rval[0].(string)
				if t == "replace" {
					v1 := rval[1].(string)
					v2 := rval[2].(string)

					replaces = append(replaces, strings.Join([]string{v1, v2}, " "))
				}
			}
		}
	}

	return includes, excludes, replaces
}
