// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"strings"
)

func Abs(i int) int {

	if i < 0 {
		return i * -1
	}
	return i
}

func IsNumeric(s string) bool {

	if len(s) == 0 {
		return false
	}
	return strings.IndexFunc(s, func(c rune) bool { return c < '0' || c > '9' }) == -1
}

func Progress(i, j int) string {

	if j == 0 {
		return "0%"
	}

	if i >= j {
		return "100%"
	}

	p := int((float64(i) / float64(j)) * 100)

	return fmt.Sprintf("%v%%", p)
}
