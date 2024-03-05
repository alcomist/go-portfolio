// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package es

import (
	"time"
)

type SearchResult struct {
	totalCount     int64
	processedCount int64
	docCount       int
	scrollID       string
	scroll         time.Duration
	Docs           Documents
}

func (r *SearchResult) ScrollId() string {
	return r.scrollID
}

func (r *SearchResult) SetScrollId(id string) {
	r.scrollID = id
}

func (r *SearchResult) TotalCount() int64 {
	return r.totalCount
}

func (r *SearchResult) SetTotalCount(c int64) {
	r.totalCount = c
}

func (r *SearchResult) DocCount() int {
	return r.docCount
}

func (r *SearchResult) SetDocCount(c int) {
	r.docCount = c
}

func (r *SearchResult) ProcessedCount() int64 {
	return r.processedCount
}

func (r *SearchResult) AddProcessedCount(c int64) {
	r.processedCount += c
}
