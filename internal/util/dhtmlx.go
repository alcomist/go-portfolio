// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"reflect"
	"strings"
)

func DhtmlxMapResponse(ds []Interface, key string, cs string, ucs string) map[string]any {

	columns := strings.Split(cs, ",")
	udColumns := strings.Split(ucs, ",")

	rows := make([]any, 0)

	for _, d := range ds {

		data := make([]any, 0)

		for _, column := range columns {

			if strings.Index(column, "url") != -1 {
				cdata := d.String(column)
				data = append(data, fmt.Sprintf("LINK^%s", cdata))
			} else {

				splitColumns := strings.Split(column, "-")

				if len(splitColumns) > 1 {

					vals := make([]string, 0)
					for _, splitColumn := range splitColumns {
						val := d.String(splitColumn)
						if len(val) > 0 {
							vals = append(vals, val)
						}
					}

					if len(vals) > 1 {
						data = append(data, fmt.Sprintf("%s (%s)", vals[0], vals[1]))
					} else {
						data = append(data, strings.Join(vals, "-"))
					}
				} else {
					data = append(data, d.Raw(column))
				}
			}
		}

		var id any
		if d.KeyExist(key) {
			id = d.Raw(key)
		}

		if id == nil {
			id = uuid.New().String()
		}

		row := make(map[string]any)

		row["id"] = id
		row["data"] = data

		if len(udColumns) > 0 {

			userData := make(map[string]any)

			for _, column := range udColumns {
				if d.KeyExist(column) {
					userData[column] = d.Raw(column)
				}
			}

			row["userdata"] = userData
		}

		rows = append(rows, row)
	}

	response := map[string]any{
		"rows": rows,
	}

	return response
}

func DhtmlxResponse(ds []Interface, key string, cs string, ucs string) []byte {

	response := DhtmlxMapResponse(ds, key, cs, ucs)

	res, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
	}

	return res
}

func Response(in any) []Interface {

	v := reflect.ValueOf(in)
	rs := make([]Interface, 0, v.Len())

	if reflect.TypeOf(in).Kind() != reflect.Slice {
		return rs
	}

	for i := 0; i < v.Len(); i++ {

		j, err := StructToMap(v.Index(i).Interface())
		if err != nil {
			log.Println(err)
			continue
		}

		var r Interface
		r.From(j)

		rs = append(rs, r)
	}

	return rs
}
