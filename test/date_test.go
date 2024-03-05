// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"github.com/alcomist/go-portfolio/internal/util"
	"testing"
)

func TestDate(t *testing.T) {

	var tests = []struct {
		start string
		end   string
		want  int
	}{
		{"20230101", "20230131", 30},
		{"20230101", "20230227", 57},
	}

	for _, test := range tests {
		if got := util.DiffDays(test.start, test.end); got != test.want {
			t.Errorf("util.DiffDays(%q,%q) = %v (WANT:%v)", test.start, test.end, got, test.want)
		}
	}

}

func TestDateAfter(t *testing.T) {

	var tests = []struct {
		input1 string
		input2 string
		want   bool
	}{
		{"20230101", "20230101", true},
		{"20230101", "20221231", false},
		{"20230101", "20231211", true},
	}

	for _, test := range tests {
		if got := util.DateAfter(test.input1, test.input2); got != test.want {
			t.Errorf("util.DateAfter(%q, %q) = %v (WANT:%v)", test.input1, test.input2, got, test.want)
		}
	}
}
