// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package adapter

import (
	"bytes"
	"encoding/json"
	"github.com/alcomist/go-portfolio/internal/es"
	"github.com/alcomist/go-portfolio/internal/util"
	"log"
)

type Adapter struct {
	*es.Response
}

func New(r *es.Response) *Adapter {
	return &Adapter{r}
}

func (a *Adapter) AdaptRawData(in, tn, id string) {

	ks := es.HashKeys()

	for _, doc := range a.Result.Docs {

		var header es.Header
		header.Build(in, tn, doc.RawPkId())

		doc.SetMeta(header)
		doc.SetValue("id", id)
		doc.SetValue("option", doc.String("option_text"))

		doc.SetValue("hashcode", doc.Hashcode(ks))

		for _, k := range ks {
			doc.AddHashcode(k)
		}

		doc.AddCombinedString("", ks)
		doc.AddCombinedHashcode("", ks)
		doc.RemoveByKeys(es.RawIndexRemoveKeys())
	}
}

func (a *Adapter) AdaptHashData(in, tn string) {

	ks := es.HashKeys()

	for _, doc := range a.Result.Docs {

		var header es.Header
		header.Build(in, tn, doc.String("hashcode"))

		doc.SetMeta(header)

		for _, k := range ks {
			doc.AddNgram(k)
		}

		doc.AddCombinedString("", ks)
		doc.AddCombinedNgram("", ks)
		doc.RemoveByKeys(es.HashIndexRemoveKeys())
	}
}

func (a *Adapter) DocUpsertData() []byte {

	var b bytes.Buffer

	for _, doc := range a.Result.Docs {

		header := make(map[string]any)

		header["update"] = doc.Meta

		h, err := json.Marshal(header)
		if err != nil {
			log.Println(err)
			continue
		}

		data := make(map[string]any)
		data["doc"] = doc.Source.Map()
		data["doc_as_upsert"] = true

		d, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
			continue
		}

		b.Write(h)
		b.WriteString("\n")
		b.Write(d)
		b.WriteString("\n")
	}

	return b.Bytes()
}

func (a *Adapter) ScriptedUpsertData() []byte {

	var b bytes.Buffer

	for _, doc := range a.Result.Docs {

		var bb bytes.Buffer

		nestedNode := doc.NestedString([]string{"root", "sub_node"})
		if len(nestedNode) > 0 {
			doc.SetValue("flattened_node", nestedNode)

			bb.WriteString(es.CategoryPainlessScript())
		}

		header := make(map[string]any)

		header["update"] = doc.Meta

		h, err := json.Marshal(header)
		if err != nil {
			log.Println(err)
			continue
		}

		doc := map[string]any{
			"scripted_upsert": true,
			"script": map[string]any{
				"lang":   "painless",
				"params": doc.Source.Map(),
				"source": bb.String(),
			},
			"upsert": doc.Source.Map(),
		}

		d, err := json.Marshal(doc)
		if err != nil {
			log.Println(err)
			continue
		}

		b.Write(h)
		b.WriteString("\n")
		b.Write(d)
		b.WriteString("\n")
	}

	return b.Bytes()
}

func (a *Adapter) ModelNamesBulkData() []byte {

	var b bytes.Buffer

	for _, doc := range a.Result.Docs {

		var hd es.Header
		hd.Build(doc.Index, doc.Index, doc.Id)

		header := make(map[string]any)
		header["update"] = hd

		h, err := json.Marshal(header)
		if err != nil {
			log.Println(err)
			continue
		}

		t := doc.String("primary")
		o := doc.String("secondary")

		primaries := util.ModelNames(t)
		secondaries := util.ModelNames(o)

		all := append(primaries, secondaries...)
		all = util.Unique(all)

		e := map[string]any{
			"primaries":   primaries,
			"secondaries": secondaries,
			"tertiary":    all,
		}

		doc := make(map[string]any)
		doc["doc"] = e
		doc["doc_as_upsert"] = true

		d, err := json.Marshal(doc)
		if err != nil {
			log.Println(err)
			continue
		}

		b.Write(h)
		b.WriteString("\n")
		b.Write(d)
		b.WriteString("\n")
	}

	return b.Bytes()
}
