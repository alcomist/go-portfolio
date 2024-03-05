// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package index_creator

import (
	"github.com/alcomist/go-portfolio/internal/es"
	"log"
)

type IndexCreator struct {
	id, index              string
	importCluster, cluster string
}

func New(id, index, importCluster, cluster string) *IndexCreator {

	return &IndexCreator{id: id, index: index, importCluster: importCluster, cluster: cluster}
}

func (task *IndexCreator) GetIndexTemplate() map[string]any {

	importEsInst := es.MustGet(task.importCluster)

	pattern := "index_template"
	template := importEsInst.Template(pattern)

	types := make(map[string]any)
	types["bool"] = map[string]any{"type": "boolean"}
	types["long"] = map[string]any{"type": "long"}
	types["keyword"] = map[string]any{"type": "keyword"}
	types["text"] = map[string]any{
		"type": "text",
		"fields": map[string]any{
			"keyword": map[string]any{
				"type":         "keyword",
				"ignore_above": 256}}}
	types["date"] = map[string]any{
		"type":   "date",
		"format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis",
	}

	// common template
	template["id"] = types["keyword"]
	template["text"] = types["text"]
	template["registered_time"] = types["date"]
	template["long_id"] = types["long"]
	template["true_or_false"] = types["bool"]

	return template
}

func (task *IndexCreator) Execute() bool {

	template := task.GetIndexTemplate()
	//template := nil

	esInst := es.MustGet(task.cluster)

	if !esInst.IndexExist(task.index) {
		if esInst.CreateIndex(task.index, template) {
			log.Printf("index('%s') created successfully", task.index)
			return true
		} else {
			log.Printf("index('%s') creation failed", task.index)
		}
	} else {
		log.Printf("index('%s') already exists", task.index)
	}

	return false
}
