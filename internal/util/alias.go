// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"github.com/alcomist/go-portfolio/internal/constant"
	"strings"
)

func LocaleAlias(l string) string {

	switch strings.ToLower(l) {
	case constant.LocaleEnglish:
		return "ENGLISH (USA)"
	case constant.LocaleKorean:
		return "한글"
	default:
		return fmt.Sprintf("UNDEFINED LOCALE(%s)", l)
	}
}

func Locales() []string {
	return []string{constant.LocaleEnglish, constant.LocaleKorean}
}
