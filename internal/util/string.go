// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"bytes"
	"regexp"
	"sort"
	"strings"
)

func NarrowSpace(s string) string {

	s = strings.TrimSpace(s)

	re := regexp.MustCompile(`(\s+)`)
	r := re.ReplaceAll([]byte(s), []byte(" "))
	r = bytes.TrimSpace(r)
	return string(r)
}

func RemoveSpace(s string) string {

	re := regexp.MustCompile(`(\s+)`)
	r := re.ReplaceAll([]byte(s), []byte(""))
	return string(r)
}

func RemoveUnit(s string) string {

	re := regexp.MustCompile(`(\b[0-9]+[.]?[0-9]+[a-zA-Z]+\b)|(\b[0-9]+[xX][0-9]+\b)`)
	r := re.ReplaceAll([]byte(s), []byte(""))
	return string(r)
}

func AlphaNumericWithKorean(s string) string {

	re := regexp.MustCompile(`[^a-zㄱ-힣A-Z0-9]+`)
	r := re.ReplaceAll([]byte(s), []byte(""))
	return string(r)
}

func SpaceKorean(s string) string {

	re := regexp.MustCompile(`([ㄱ-힣]+)`)
	r := re.ReplaceAll([]byte(s), []byte(" ${1} "))
	r = bytes.TrimSpace(r)
	return string(r)
}

func SpaceNumeric(s string) string {

	re := regexp.MustCompile(`(\d+)`)
	r := re.ReplaceAll([]byte(s), []byte(" ${1} "))
	r = bytes.TrimSpace(r)
	return NarrowSpace(string(r))
}

func SpaceOperators(s string) string {

	re := regexp.MustCompile(`([+\-*\/]+)`)
	r := re.ReplaceAll([]byte(s), []byte(" ${1} "))
	r = bytes.TrimSpace(r)
	return string(r)
}

func SpaceAlphabet(s string) string {

	re := regexp.MustCompile(`(\w+)`)
	result := re.ReplaceAll([]byte(s), []byte(" ${1} "))
	result = bytes.TrimSpace(result)
	return string(result)
}

func Space(s string) string {

	if len(s) == 0 {
		return s
	}

	s = SpaceKorean(s)
	s = SpaceNumeric(s)
	s = SpaceAlphabet(s)
	s = SpaceSpecialCharacter(s)

	return NarrowSpace(s)
}

func Comma(s string) string {

	if len(s) < 4 {
		return s
	}

	return Comma(s[0:len(s)-3]) + "," + s[len(s)-3:]
}

func SpaceSpecialCharacter(s string) string {

	re := regexp.MustCompile(`([\W_]+)`)
	r := re.ReplaceAll([]byte(s), []byte(" ${1} "))
	r = bytes.TrimSpace(r)
	return string(r)
}

func GreedyPattern(s string) string {

	s = RemoveSpace(s)
	ss := strings.Split(s, "")

	for i, _s := range ss {
		ss[i] = regexp.QuoteMeta(_s)
	}

	s = strings.Join(ss, `\s*`)
	//$replace = '\s*';

	re := regexp.MustCompile(`\s+`)
	res := re.ReplaceAll([]byte(s), []byte(`\s*`))

	return string(res)
}

func PatternToRegex(p string) string {

	p = GreedyPattern(p)

	floatPattern := "(" + strings.Join([]string{`[0-9]+(\.[0-9]+)?`, `\.[0-9]+`}, "|") + ")"

	var converts = []struct {
		pattern string
		replace string
	}{
		{`\s+`, `\s*`},
		{`A`, `(\w+)`},
		{`D`, `(\d+)`},
		{`F`, floatPattern},
		{`X`, `(.*?)`},
		{`S`, `(^)`},
		{`E`, `($)`},
		{`B`, `(\b)`},
	}

	for _, convert := range converts {
		re, err := regexp.Compile(convert.pattern)
		if err == nil {
			p = string(re.ReplaceAll([]byte(p), []byte(convert.replace)))
		}
	}

	// for case insensitive regex search or replace
	return `(?i)` + p
}

func NormalizeModelName(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "-", "")
}

func ModelNames(t string) []string {

	// pattern for electronic appliance model name extraction
	//p := `[0-9a-zA-Z\-]{3,}`

	t = RemoveUnit(t)

	p := `([0-9]+[a-zA-Z\-][a-zA-Z0-9\-]*)|([a-zA-Z]+[0-9\-][a-zA-Z0-9\-]*)`

	r := regexp.MustCompile(p)
	ss := r.FindAllString(t, -1)

	for i, s := range ss {

		s = NormalizeModelName(s)
		if len(s) > 2 {
			ss[i] = s
		} else {
			ss[i] = ""
		}
	}

	return Unique(ss)
}

func StartWithNumber(s string) bool {

	s = strings.Trim(s, " ")

	re := regexp.MustCompile("^[0-9]")
	return re.Match([]byte(s))
}

func SplitModelName(s string) []string {

	s = strings.ToLower(s)
	fields := strings.Fields(s)

	re := regexp.MustCompile(`([0-9]+)|([a-zA-Z]+)`)

	ss := make([]string, 0)

	for _, field := range fields {
		ns := re.FindAllString(field, -1)
		for _, n := range ns {
			ss = append(ss, n)
			field = strings.ReplaceAll(field, n, "")
		}

		for _, r := range field {
			ss = append(ss, string(r))
		}
	}

	return ss
}

func longestCommonPrefix(ss []string) string {

	longestPrefix := ""
	endPrefix := false

	if len(ss) > 0 {

		sort.Strings(ss)
		first := ss[0]
		last := ss[len(ss)-1]

		for i := 0; i < len(first); i++ {

			if !endPrefix && string(last[i]) == string(first[i]) {
				longestPrefix += string(last[i])
			} else {
				endPrefix = true
			}
		}
	}
	return longestPrefix
}
