// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package es

import (
	"github.com/alcomist/go-portfolio/internal/util"
)

type Response struct {
	Success bool
	Result  *SearchResult
}

func NewResponse() *Response {
	return &Response{}
}

func (r *Response) Reserve(s int) {
	r.Result = &SearchResult{Docs: make([]*Doc, 0, s)}
}

func (r *Response) Expand(k string) (int, int) {

	is := make([]int, 0)

	newDocs := make(Documents, 0)
	for i, doc := range r.Result.Docs {

		list := doc.List(k)
		if len(list) > 0 {

			for _, m := range list {
				var n util.Interface
				n.From(m.(map[string]any))
				doc.Source = util.MergeMaps(doc.Source, n)
				newDocs = append(newDocs, doc)
			}

			is = append(is, i)
		}
	}

	r.Result.Docs = RemoveSlices(r.Result.Docs, is)
	r.Result.Docs = append(r.Result.Docs, newDocs...)

	return len(newDocs), len(is)
}
