// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package es

import (
	"fmt"
	"github.com/alcomist/go-portfolio/internal/hash"
	"github.com/alcomist/go-portfolio/internal/util"
	"log"
	"strconv"
	"strings"
)

type Doc struct {
	Index string
	Id    string

	Meta   Header
	Source util.Interface
}

type Documents []*Doc

func (d *Doc) KeyExist(k string) bool {

	return d.Source.KeyExist(k)
}

func (d *Doc) Int64(k string) int64 {

	return d.Source.Int64(k)
}

func (d *Doc) String(k string) string {

	return d.Source.String(k)
}

func (d *Doc) Float64(k string) float64 {

	return d.Source.Float64(k)
}

func (d *Doc) Strings(k string) []string {

	return d.Source.Strings(k)
}

func (d *Doc) List(k string) []any {

	return d.Source.List(k)
}

func (d *Doc) NestedString(ks []string) []string {

	val, err := NestedMapLookup(d.Source.Map(), ks...)
	if err != nil {
		log.Println(err)
		return []string{""}
	}

	r, ok := val.(string)
	if ok {
		return []string{r}
	}

	rs, ok := val.([]string)
	if ok {
		return rs
	}

	fmt.Printf("%T %v\n", val, val)
	return []string{""}
}

func (d *Doc) Date(k string) string {

	return d.Source.Date(k)
}

func (d *Doc) Hashcode(ks []string) string {

	ts := make([]string, 0, len(ks))
	for _, k := range ks {
		ts = append(ts, d.Source.String(k))
	}

	if len(ts) != len(ks) {
		log.Fatalf("hashcode elements size mismatch : %d / %d\n", len(ks), len(ts))
	}

	return hash.GenerateHashcode(ts)
}

func (d *Doc) SetValue(k string, v any) {

	d.Source.Set(k, v)
}

func (d *Doc) RemoveKey(k string) {

	d.Source.Remove(k)
}

func (d *Doc) SetMeta(h Header) {
	d.Meta = h
}

func (d *Doc) RawPkId() string {

	pk := "primary_id"

	id1 := strconv.FormatInt(d.Int64(pk), 10)
	pv := id1

	id2 := d.String("secondary_id")
	if len(id2) > 0 {
		pv = strings.Join([]string{id1, id2}, "|")
	}

	return pv
}

func (d *Doc) AddCombinedString(key string, ks []string) {

	if len(key) == 0 {
		key = strings.Join(ks, "_")
	}

	ss := make([]string, 0, len(ks))
	for _, k := range ks {
		ss = append(ss, d.String(k))
	}

	d.SetValue(key, strings.Join(ss, " "))
}

func (d *Doc) AddTime() {

	d.SetValue("registered_date", d.Date("registered_time"))
}

func (d *Doc) AddHashcode(k string) {

	nk := fmt.Sprintf("%s_hashcode", k)
	d.Source.Set(nk, d.Hashcode([]string{k}))
}

func (d *Doc) AddCombinedHashcode(key string, ks []string) {

	if len(key) == 0 {
		key = fmt.Sprintf("%s_hashcode", strings.Join(ks, "_"))
	}

	d.SetValue(key, d.Hashcode(ks))
}

func (d *Doc) AddNgram(k string) {

	nk := fmt.Sprintf("%s_ngram", k)
	d.SetValue(nk, seed.Ngram(d.String(k)))
}

func (d *Doc) AddCombinedNgram(key string, ks []string) {

	if len(key) == 0 {
		key = fmt.Sprintf("%s_ngram", strings.Join(ks, "_"))
	}

	ss := make([]string, 0, len(ks))
	for _, k := range ks {
		ss = append(ss, d.String(k))
	}

	d.SetValue(key, seed.Ngram(strings.Join(ss, " ")))
}

func (d *Doc) RemoveByKeys(ks []string) {

	for _, k := range ks {
		d.RemoveKey(k)
	}
}

func (d *Doc) DeepCopy() *Doc {

	return &Doc{Index: d.Index, Id: d.Id, Meta: d.Meta, Source: d.Source.DeepCopy()}
}

func (docs Documents) LastId(k string) int64 {

	if len(docs) > 0 {
		return docs[0].Int64(k)
	}
	return 0
}
