// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"github.com/alcomist/go-portfolio/internal/constant"
	"log"
	"strings"
	"time"
)

func IndexDate(s string) string {

	ss := strings.Split(s, "_")

	candidate := ss[len(ss)-1]
	if len(candidate) == 8 {
		return candidate
	}

	return ""
}

func Today() string {

	return Date(0, constant.TimeFormat)
}

func Yesterday() string {

	return Date(-1, constant.TimeFormat)
}

func Date(days int, format string) string {

	t := time.Now()
	t = t.AddDate(0, 0, days)
	return t.Format(format)
}

func FullTime() string {

	t := time.Now()
	return t.Format("2006-01-02 15:04:05")
}

func DiffDays(start, end string) int {

	ts, err := time.Parse(constant.TimeFormat, start)
	if err != nil {
		log.Println(err)
		return -1
	}

	te, err := time.Parse(constant.TimeFormat, end)
	if err != nil {
		log.Println(err)
		return -1
	}

	diff := te.Sub(ts)
	return int(diff.Hours() / 24)
}

func Dates(start, end string) []string {

	dates := make([]string, 0)

	ts, err := time.Parse(constant.TimeFormat, start)
	if err != nil {
		log.Println(err)
		return dates
	}

	te, err := time.Parse(constant.TimeFormat, end)
	if err != nil {
		log.Println(err)
		return dates
	}

	for d := ts; d.After(te) == false; d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(constant.TimeFormat))
	}

	return dates
}

func Milliseconds() int64 {

	t := time.Now()
	return t.UnixMilli()
}

func Microseconds() int64 {

	t := time.Now()
	return t.UnixMicro()
}

func TimestampToString(ts int) string {

	t := time.Unix(int64(ts), 0)
	return t.Format("2006-01-02 15:04:05")
}

func Ymd(d string) string {

	ts, err := time.Parse(constant.TimeFormatDash, d)
	if err != nil {
		log.Println(err)
		return ""
	}

	return ts.Format(constant.TimeFormat)
}

func KibanaDate(t int64) string {

	tm := time.Unix(t, 0)
	return tm.Format(constant.TimeFormat)
}

func DateAfter(ld, id string) bool {

	if len(ld) == 0 {
		return true
	}

	start, _ := time.Parse(constant.TimeFormat, ld)
	end, _ := time.Parse(constant.TimeFormat, id)

	return end.Equal(start) || end.After(start)
}

func DateBefore(ld, id string) bool {

	if len(ld) == 0 {
		return false
	}

	start, _ := time.Parse(constant.TimeFormat, ld)
	end, _ := time.Parse(constant.TimeFormat, id)

	return end.Before(start)
}
