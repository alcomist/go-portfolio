// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hash

import (
	"crypto/md5"
	"fmt"
	"github.com/alcomist/go-portfolio/internal/constant"
	"github.com/alcomist/go-portfolio/internal/util"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

func CalculateHash(s string) string {

	h := md5.New()
	_, err := io.WriteString(h, s)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func prepareHashcode(s string) string {

	s = util.RemoveSpace(s)
	return strings.ToUpper(s)
}

func GenerateHashcode(ss []string) string {

	var bases []string

	for _, s := range ss {
		bases = append(bases, prepareHashcode(s))
	}

	return CalculateHash(strings.Join(bases, "|"))
}

func ngram(s string, n int) []string {

	r := make([]string, 0)

	runes := []rune(s)

	for len(runes) >= n {

		r = append(r, string(runes[0:n]))
		runes = runes[1:]
	}

	return r
}

func Ngram(s string) string {

	s = strings.ToLower(s)
	s = util.Space(s)

	r := make([]string, 0)

	ss := strings.Split(s, " ")

	sort.Slice(ss, func(i, j int) bool {
		return utf8.RuneCountInString(ss[i]) < utf8.RuneCountInString(ss[j])
	})

	for _, p := range ss {

		if util.IsNumeric(p) {
			r = append(r, p)
			continue
		}

		if utf8.RuneCountInString(p) > constant.NgramMaxTokenSize*2 {
			r = append(r, p)
			continue
		}

		for i := 1; i <= utf8.RuneCountInString(p); i++ {
			r = append(r, ngram(p, i)...)
		}
	}

	r = util.Unique(r)

	sort.Slice(r, func(i, j int) bool {
		return utf8.RuneCountInString(r[i]) < utf8.RuneCountInString(r[j])
	})

	rs := strings.Join(r, " ")
	if utf8.RuneCountInString(rs) > constant.NgramMaxStringLength {
		rs = rs[:constant.NgramMaxStringLength]
	}
	return rs
}
