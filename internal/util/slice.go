// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"sort"
)

func Contains(n string, hs []string) bool {

	for _, h := range hs {
		if h == n {
			return true
		}
	}

	return false
}

func Unique(ss []string) []string {

	m := make(map[string]bool)
	r := make([]string, 0, len(ss))

	for _, s := range ss {

		if len(s) == 0 {
			continue
		}

		exist := m[s]
		if exist {
			continue
		}

		m[s] = true
		r = append(r, s)
	}

	return r
}

func DeepEqual(s1, s2 []string) bool {

	if len(s1) != len(s2) {
		return false
	}

	sort.Strings(s1)
	sort.Strings(s2)

	for i, s := range s1 {
		if s2[i] != s {
			return false
		}
	}

	return true
}

func DiffInt(a, b []int) []int {

	temp := map[int]int{}
	for _, s := range a {
		temp[s]++
	}
	for _, s := range b {
		temp[s]--
	}

	var result []int
	for s, v := range temp {
		if v > 0 {
			result = append(result, s)
		}
	}
	return result
}
