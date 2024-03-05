// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"github.com/alcomist/go-portfolio/internal/constant"
	"log"
	"strconv"
	"time"
)

func ToInt(x any) int {
	return int(ToInt64(x))
}

func ToFloat64(x any) float64 {

	switch x := x.(type) {
	case bool:
		if x {
			return 1
		} else {
			return 0
		}
	case int:
		return float64(x)
	case uint:
		return float64(x)
	case int64:
		return float64(x)
	case uint64:
		return float64(x)
	case float32:
		return float64(x)
	case float64:
		return x
	case string:
		v, e := strconv.ParseFloat(x, 64)
		if e != nil {
			return 0
		}
		return v
	default:
		log.Printf("unnexpected type %T: %v", x, x)
		return 0
	}
}

func ToInt64(x any) int64 {

	switch x := x.(type) {
	case bool:
		if x {
			return 1
		} else {
			return 0
		}
	case int:
		return int64(x)
	case uint:
		return int64(x)
	case int64:
		return x
	case uint64:
		return int64(x)
	case float32:
		return int64(x)
	case float64:
		return int64(x)
	case string:
		v, e := strconv.ParseInt(x, 10, 64)
		if e != nil {
			return 0
		}
		return v
	default:
		log.Printf("unnexpected type %T: %v", x, x)
		return 0
	}
}

func ToString(x any) string {

	switch x := x.(type) {
	case nil, []int, []string:
		return ""
	case bool:
		return ""
	case int:
		return strconv.Itoa(x)
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	case float32:
		return strconv.FormatFloat(float64(x), 'f', 0, 64)
	case float64:
		return strconv.FormatFloat(x, 'f', 0, 64)
	case string:
		return x
	default:
		log.Printf("unnexpected type %T: %v", x, x)
		return ""
	}
}

func QuoteID(x any) string {

	switch x := x.(type) {
	case nil:
		return ""
	case []int:
		return ""
	case []string:
		return ""
	case bool:
		return ""
	case float32:
		return fmt.Sprintf("%d", int(x))
	case float64:
		return fmt.Sprintf("%d", int(x))
	case int, uint:
		return fmt.Sprintf("%d", x)
	case string:
		return x
	default:
		log.Panicf("unnexpected type %T: %v", x, x)
		return ""
	}
}

type Interface struct {
	val map[string]any
}

func (i *Interface) New() {
	i.val = make(map[string]any)
}

func (i *Interface) From(v map[string]any) {
	i.val = v
}

func (i *Interface) Remove(k string) {
	delete(i.val, k)
}

func (i *Interface) Set(k string, v any) {
	i.val[k] = v
}

func (i *Interface) KeyExist(k string) bool {

	_, ok := i.val[k]
	return ok
}

func (i *Interface) ID(k string) string {

	return i.String(k)
}

func (i *Interface) Int(k string) int {

	val, ok := i.val[k]
	if ok {
		return int(ToInt64(val))
	}

	return 0
}

func (i *Interface) Uint(k string) uint {

	val, ok := i.val[k]
	if ok {
		return uint(ToInt64(val))
	}

	return 0
}

func (i *Interface) Float64(k string) float64 {

	val, ok := i.val[k]
	if ok {
		return ToFloat64(val)
	}

	return 0
}

func (i *Interface) Int64(k string) int64 {

	val, ok := i.val[k]
	if ok {
		return ToInt64(val)
	}

	return 0
}

func (i *Interface) String(k string) string {

	val, ok := i.val[k]
	if ok {
		return ToString(val)
	}

	return ""
}

func (i *Interface) Strings(k string) []string {

	ss := make([]string, 0)

	list := i.List(k)
	if len(list) > 0 {
		for _, m := range list {
			ss = append(ss, ToString(m))
		}
	}

	return ss
}

func (i *Interface) List(k string) []any {

	val, ok := i.val[k]
	if ok {
		return val.([]any)
	}

	return []any{}
}

func (i *Interface) Date(k string) string {

	val, ok := i.val[k]
	if ok {
		r, ok := val.(float64)
		if ok {
			return time.UnixMilli(int64(r)).Format(constant.TimeFormat)
		}
	}

	return ""
}

func (i *Interface) Raw(k string) any {

	val, ok := i.val[k]
	if ok {
		return val
	}

	return nil
}

func (i *Interface) Map() map[string]any {
	return i.val
}

func (i *Interface) DeepCopy() Interface {

	val2 := CopyableMap(i.val).DeepCopy()

	result := Interface{}
	result.From(val2)
	return result
}

func MergeMaps(maps ...Interface) (result Interface) {

	result = Interface{}
	result.New()

	for _, m := range maps {
		for k, v := range m.Map() {
			result.Set(k, v)
		}
	}

	return result
}
