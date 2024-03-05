// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alcomist/go-portfolio/internal/constant"
	"log"
	"strings"
)

var esReserved = []string{"\\", "+", "=", "&&", "||", "!", "(", ")", "{", "}", "[", "]", "^", "\"", "~", "*", "?", ":", "/"}

type QueryRange [2]int64

type QueryItem struct {
	negate bool
	field  string
	value  any
}

type QueryDoc struct {
	Query                string `json:"query"`
	DefaultField         string `json:"default_field,omitempty"`          // (Optional, string)
	AllowLeadingWildcard bool   `json:"allow_leading_wildcard,omitempty"` // (Optional, Boolean) default true
	AnalyzeWildcard      bool   `json:"analyze_wildcard,omitempty"`       // (Optional, Boolean) Defaults to false.
	DefaultOperator      string `json:"default_operator,omitempty"`       // (Optional, string)
}

type QueryStringQuery struct {
	doc  QueryDoc
	sort []map[string]string
	size int
}

type QueryStringQueryBuilder struct {
	Query       QueryStringQuery
	Aggregation map[string]any
	Items       []QueryItem
}

func sanitizeQueryField(keyword string) string {

	sanitizedKeyword := keyword
	for _, char := range esReserved {
		if strings.Contains(sanitizedKeyword, char) {
			replaceWith := `\` + char
			sanitizedKeyword = strings.ReplaceAll(sanitizedKeyword, char, replaceWith)
		}
	}
	return sanitizedKeyword
}

func quoteSpace(s string) string {

	if strings.Contains(s, " ") {
		return fmt.Sprintf("\"%s\"", s)
	}
	return s
}

func quoteString(x any) string {

	sep := " OR "
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
		return fmt.Sprintf("%s", quoteSpace(sanitizeQueryField(x)))
	case []int:
		items := make([]string, 0, len(x))
		for _, xs := range x {
			items = append(items, fmt.Sprintf("%d", xs))
		}
		if len(items) == 0 {
			return ""
		}
		return fmt.Sprintf("(%s)", strings.Join(items, sep))
	case []string:
		items := make([]string, 0, len(x))
		for _, xs := range x {
			if len(xs) > 0 {
				items = append(items, fmt.Sprintf("%s", quoteSpace(sanitizeQueryField(xs))))
			}
		}
		if len(items) == 0 {
			return ""
		}
		return fmt.Sprintf("(%s)", strings.Join(items, sep))
	case QueryRange:
		q := fmt.Sprintf("[%d TO %d]", x[0], x[1])
		return strings.ReplaceAll(q, "-1", "*")
	default:
		log.Panicf("unnexpected type %T: %v", x, x)
		return ""
	}
}

func NewQueryBuilder() *QueryStringQueryBuilder {

	queryStringQuery := QueryStringQuery{}
	queryStringQuery.sort = make([]map[string]string, 0)
	queryStringQuery.size = 0

	return &QueryStringQueryBuilder{Query: queryStringQuery, Items: make([]QueryItem, 0)}
}

func (d *QueryStringQueryBuilder) AddAggregation(val map[string]any) {

	d.Aggregation = val
}

func (d *QueryStringQueryBuilder) AddSort(key, order string) {

	d.Query.sort = append(d.Query.sort, map[string]string{key: order})
}

func (d *QueryStringQueryBuilder) SetSize(s int) {

	d.Query.size = s
}

func (d *QueryStringQueryBuilder) Add(field string, val any) {

	if len(field) > 0 {
		d.Items = append(d.Items, QueryItem{false, field, val})
	}
}

func (d *QueryStringQueryBuilder) AddNot(field string, val any) {

	if len(field) > 0 {
		d.Items = append(d.Items, QueryItem{true, field, val})
	}
}

func (d *QueryStringQueryBuilder) AddExist(field string) {

	if len(field) > 0 {
		d.Items = append(d.Items, QueryItem{false, constant.ElasticKeyExists, field})
	}
}

func (d *QueryStringQueryBuilder) AddNotExist(field string) {

	if len(field) > 0 {
		d.Items = append(d.Items, QueryItem{true, constant.ElasticKeyExists, field})
	}
}

func (d *QueryStringQueryBuilder) SetDefaultField(f string) {
	d.Query.doc.DefaultField = f
}

func (d *QueryStringQueryBuilder) String() string {

	queryString := d.QueryString()

	d.Query.doc.Query = queryString

	query := map[string]any{}

	if len(d.Query.doc.Query) > 0 {
		query["query"] = map[string]any{
			"query_string": d.Query.doc,
		}
	}

	if len(d.Query.sort) > 0 {
		query["sort"] = d.Query.sort
	}
	if d.Query.size > 0 {
		query["size"] = d.Query.size
	}

	if d.Aggregation != nil {
		query["aggs"] = d.Aggregation
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Printf("Error encoding query: %s\n", err)
		return ""
	}

	return buf.String()
}

func (d *QueryStringQueryBuilder) QueryString() string {

	items := make([]string, 0, len(d.Items))
	for _, item := range d.Items {

		field := sanitizeQueryField(item.field)
		if item.negate {
			field = fmt.Sprintf("NOT %s", field)
		}

		q := fmt.Sprintf("%s:%s", field, quoteString(item.value))
		if len(q) > 0 {
			items = append(items, q)
		}
	}
	return strings.Join(items, " AND ")
}
