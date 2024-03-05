// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package es

import (
	"fmt"
	"math"
	"sort"
)

func NestedMapLookup(m map[string]any, ks ...string) (rval any, err error) {

	var ok bool

	if len(ks) == 0 { // degenerate input
		return nil, fmt.Errorf("NestedMapLookup needs at least one key")
	}
	if rval, ok = m[ks[0]]; !ok {
		return nil, fmt.Errorf("key not found; remaining keys: %v", ks)
	} else if len(ks) == 1 { // we've reached the final key
		return rval, nil
	} else if m, ok = rval.(map[string]any); !ok {
		return nil, fmt.Errorf("malformed structure at %#v", rval)
	} else { // 1+ more keys
		return NestedMapLookup(m, ks[1:]...)
	}
}

func CategoryPainlessScript() string {

	script := `
		if ( params.categories != null && params.categories.size() != 0 ) {
			if ( ctx._source.categories == null ) {
				ctx._source.categories = [];
			}
			
			for ( def category : params.categories ) {
				if ( ctx._source.categories.indexOf(category) == -1 ) {
					ctx._source.categories.add(category);
				}
			}
		}
		params.categories = null;
	`

	return script
}

func HashKeys() []string {

	return []string{"primary", "secondary"}
}

func RawIndexRemoveKeys() []string {

	return []string{
		"will_be_removed",
	}
}

func RemoveSlices(d Documents, is []int) Documents {

	sort.Slice(is, func(a, b int) bool {
		return is[b] < is[a]
	})

	for _, i := range is {
		d[i] = d[len(d)-1]
		d = d[:len(d)-1]
	}

	return d
}

func HashIndexRemoveKeys() []string {

	return []string{
		"will_be_removed_1",
		"will_be_removed_2",
		"will_be_removed_3",
	}
}

func Aggs(f string, size int) map[string]any {

	aggs := make(map[string]any)
	aggs[f] = map[string]any{
		"terms": map[string]any{
			"field": f,
			"size":  size},
	}

	return aggs
}

func PartitionAggs(p, np, size int) map[string]any {

	saggs := make(map[string]any)
	saggs["aggs_key1"] = map[string]any{
		"max": map[string]any{
			"field": "key1"}}
	saggs["aggs_key2"] = map[string]any{
		"terms": map[string]any{
			"field": "key2", "size": 20}}
	saggs["aggs_key3"] = map[string]any{
		"terms": map[string]any{
			"field": "key3", "size": 20}}
	saggs["aggs_key4"] = map[string]any{
		"top_hits": map[string]any{
			"size": 1}}

	terms := make(map[string]any)
	terms["field"] = "hashcode"
	terms["size"] = size
	terms["include"] = map[string]any{
		"partition":      p,
		"num_partitions": np,
	}
	terms["order"] = map[string]any{
		"aggs_registered_time": "asc",
	}

	aggs := make(map[string]any)
	aggs["aggs_hashcode"] = map[string]any{
		"terms": terms,
		"aggs":  saggs,
	}

	return aggs
}

func PartitionCount(cdr, size int) int {

	size = int(float64(size) * 1.2)
	return int(math.Ceil(float64(cdr / size)))
}

func CardinalityAggregationTerm(field string) map[string]any {

	field = sanitizeQueryField(field)

	return map[string]any{
		field: map[string]any{
			"cardinality": map[string]any{
				"field": field,
			},
		},
	}
}
