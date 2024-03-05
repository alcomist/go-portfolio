// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"strings"
)

// grammars
// ab[1/2/3] => ab1, ab2, ab3
// ab(1/2/3) => ab1, ab2, ab3, 1, 2, 3
// 한[a/b/c][1/2/3] => 한a1, 한a2, 한a3, 한b1, 한b2, 한b3, 한c1, 한c2, 한c3
// abc[1/2/3], def[4/5/6] => abc1, abc2, abc3, def4, def5, def6
// sa => stand alone

func parseTag(s string) ([][]string, []string) {

	res := make([][]string, 0)
	sa := make([]string, 0)

	stack := make([]rune, 0)

	for _, r := range []rune(s) {

		if r == '[' || r == '(' || r == ']' || r == ')' {

			if len(stack) > 0 {

				ss := strings.Split(string(stack), "/")

				if r == ')' {
					sa = append(sa, ss...)
				}

				res = append(res, ss)
				stack = stack[0:0]
			}
			continue
		} else {
			stack = append(stack, r)
		}
	}

	if len(stack) > 0 {
		res = append(res, strings.Split(string(stack), "/"))
	}

	return res, sa
}

func hasValidParen(s string) bool {

	parens := make([]rune, 0)

	for _, r := range []rune(s) {

		if r == '(' || r == '[' {
			parens = append(parens, r)
		} else if r == ')' || r == ']' {

			top := parens[len(parens)-1]

			if (top == '(' && r == ')') || top == '[' && r == ']' {
				parens = parens[0 : len(parens)-1]
			}
		}
	}

	return len(parens) == 0
}

func expand(ss [][]string, i, n int, temp []string, res *[]string) {

	if len(temp) >= n {
		*res = append(*res, strings.ToLower(strings.Join(temp, "")))
		return
	}

	for _, s := range ss[i] {
		expand(ss, i+1, n, append(temp, RemoveSpace(s)), res)
	}
}

func ExpandTag(t string) []string {

	if !hasValidParen(t) {
		return []string{}
	}

	res := make([]string, 0)

	for _, s := range strings.Split(t, ",") {

		ss, sa := parseTag(s)
		if len(sa) > 0 {
			res = append(res, sa...)
		}

		temp := make([]string, 0)

		expand(ss, 0, len(ss), temp, &res)
	}

	return res
}
