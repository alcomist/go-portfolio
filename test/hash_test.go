// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"github.com/alcomist/go-portfolio/internal/hash"
	"testing"
)

func TestGenerateHashcode(t *testing.T) {

	var tests = []struct {
		arg  []string
		want string
	}{
		{[]string{"[도착보장] 올바르고 반듯한 수비드 닭가슴살 130g 오리지널 3+허브 3+블랙페퍼 3 (총 9개)", "올바르고 반듯한 닭가슴살", ""}, "0438b37c1a662fbc52ad4029f84e0c6a"},
		{[]string{"오뚜기 컵누들 마라탕 컵 44.7g 15개 외 6종", "컵누들 마라탕 15개"}, "280a8dc720a8fbeb0afff60ccd771140"},
	}

	for _, test := range tests {
		if got := hash.GenerateHashcode(test.arg); got != test.want {
			t.Errorf("core.GenerateHashcode\n(%q) = %v\n(WANT:%v)",
				test.arg, got, test.want)
		}
	}

}

func TestNgram(t *testing.T) {

	//s = "Glenfiddich 18 Years 700ml (with box)"
	//ngram = core.Ngram(s)
	//fmt.Println(ngram)

	var tests = []struct {
		input string
		want  string
	}{
		{"Glenfiddich 18 Years 700ml (with box)",
			"y ) m l b o x w ( c d f n i t h g s r a e id ml ch ic di th it ye ea ar rs dd 18 fi nf en le wi gl ox bo dic idd 700 ich ddi ars ear yea box wit ith gle len enf nfi fid dich year with ears glen lenf enfi nfid fidd iddi ddic years glenf lenfi enfid nfidd fiddi iddic ddich glenfi lenfid enfidd nfiddi fiddic iddich glenfid lenfidd enfiddi nfiddic fiddich glenfidd lenfiddi enfiddic nfiddich glenfiddi lenfiddic enfiddich glenfiddic lenfiddich glenfiddich"},
	}

	for _, test := range tests {
		if got := hash.Ngram(test.input); got != test.want {
			t.Errorf("core.Ngram(%q) = %v (WANT:%v)", test.input, got, test.want)
		}
	}
}
