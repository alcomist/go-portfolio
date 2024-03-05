package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"io"
	"log"
	"sort"
	"strings"
)

func (e *ElasticInstance) Indices(p string) []string {

	indices := make([]string, 0)

	pattern := "*"

	if len(strings.Trim(p, " ")) > 0 {
		pattern = fmt.Sprintf("%s*", p)
	}

	res, err := esapi.CatIndicesRequest{Index: []string{pattern}, Format: "json"}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error getting indices", res.Status())
		return indices
	}

	var r []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s", err)
	}

	for _, info := range r {

		v, ok := info["index"].(string)
		if ok {
			if !strings.HasPrefix(v, ".") {
				indices = append(indices, v)
			}
		}
	}

	sort.Strings(indices)
	return indices
}

func (e *ElasticInstance) IndexExist(p string) bool {

	res, err := esapi.IndicesExistsRequest{Index: []string{p}}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("indices exists response status: %s\n", res.Status())
		return false
	}

	return true
}

func (e *ElasticInstance) CreateIndex(p string, t map[string]any) bool {

	body := make(map[string]any)
	body["settings"] = map[string]any{
		"number_of_shards":   3,
		"number_of_replicas": 1,
	}

	var includeTypeNamePtr *bool = nil

	if t != nil {
		if e.IsMajorVersion(6) {
			body["mappings"] = map[string]any{
				p: map[string]any{
					"properties": t,
				},
			}
		} else {
			body["mappings"] = map[string]any{
				"properties": t,
			}
			includeTypeName := true
			includeTypeNamePtr = &includeTypeName
		}
	}

	buf, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
		return false
	}

	res, err := esapi.IndicesCreateRequest{Index: p, Body: bytes.NewReader(buf), IncludeTypeName: includeTypeNamePtr}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error create index", res.Status())
		fmt.Println(res.String())
		return false
	}

	var b bytes.Buffer
	n, err := io.Copy(&b, res.Body)
	if err != nil {
		log.Println(err)
	}

	if n > 0 {
		var r map[string]any
		if err := json.NewDecoder(&b).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		}
	}

	return true
}

func (e *ElasticInstance) DeleteIndex(p string) bool {

	if strings.Contains(p, "*") {
		log.Println("'*' character not allowed in deleting index")
		return false
	}

	res, err := esapi.IndicesDeleteRequest{Index: []string{p}}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error delete index", res.Status())
		return false
	}

	return true
}

func (e *ElasticInstance) Mapping(p string) map[string]any {

	res, err := esapi.IndicesGetMappingRequest{Index: []string{p}}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	log.Printf("indices exists response status: %s\n", res.Status())
	if res.IsError() {
		log.Printf("[%s] Error get mapping", res.Status())
		return nil
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, res.Body)
	if err != nil {
		log.Println(err)
		return nil
	}

	if n > 0 {

		var r map[string]any
		if err := json.NewDecoder(&buf).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s\n", err)
		}

		properties, err := NestedMapLookup(r, p, "mappings", "_doc", "properties")
		if err != nil {
			log.Printf("Error looking up nested map : %s\n", err)
			return nil
		}

		return properties.(map[string]any)
	}

	return nil
}

func (e *ElasticInstance) Template(p string) map[string]any {

	res, err := esapi.IndicesGetTemplateRequest{Name: []string{p}}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error get indices template", res.Status())
		return nil
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, res.Body)
	if err != nil {
		log.Println(err)
		return nil
	}

	if n > 0 {

		var r map[string]any
		if err := json.NewDecoder(&buf).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
			return nil
		}

		ks := []string{p, "mappings", "properties"}
		if e.IsMajorVersion(6) {
			ks = []string{p, "mappings", p, "properties"}
		}

		properties, err := NestedMapLookup(r, ks...)
		if err != nil {
			log.Printf("error getting nested map lookup: %s\n", err)
			return nil
		}

		//fmt.Println(reflect.TypeOf(properties))
		return properties.(map[string]any)
	}

	return nil
}
