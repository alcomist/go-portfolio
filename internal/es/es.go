package es

import (
	"encoding/json"
	"fmt"
	"github.com/alcomist/go-portfolio/internal/config"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"log"
	"strconv"
	"strings"
	"time"
)

type ElasticInstance struct {
	client  *elasticsearch7.Client
	cluster string

	MajorVersion int
}

type Header struct {
	Index string `json:"_index"`
	Type  string `json:"_type"`
	Id    string `json:"_id"`
	ROC   int    `json:"retry_on_conflict,omitempty"`
}

func (h *Header) Build(index, typ, id string) {
	h.Index = index
	h.Type = typ
	h.Id = id
	h.ROC = 10
}

type Request struct {
	Index  string
	Query  string
	Size   int
	Scroll time.Duration
}

func mustGetConfig(s string) elasticsearch7.Config {

	var cfg elasticsearch7.Config

	section := config.MustGet(s)

	hosts := section.Key("host").ValueWithShadows()
	if len(hosts) == 0 {
		log.Fatal(fmt.Errorf("section has no hosts : %v", section))
	}

	cfg.Addresses = hosts
	cfg.RetryOnStatus = []int{502, 503, 504, 429}
	return cfg
}

func MustGet(cluster string) ElasticInstance {

	cfg := mustGetConfig(cluster)

	es, err := elasticsearch7.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}

	var r map[string]any

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	ks := []string{"version", "number"}
	val, err := NestedMapLookup(r, ks...)
	if err != nil {
		log.Fatalln(err)
	}

	version := val.(string)

	vs := strings.Split(version, ".")
	vn, err := strconv.Atoi(vs[0])
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	log.Printf("[%s] elasticsearch server : %s", cluster, version)
	return ElasticInstance{es, cluster, vn}
}

func NewGenerator(e ElasticInstance, req *Request) func() *Response {

	var response *Response

	return func() *Response {

		if response == nil {
			response = e.Search(req)
			if response == nil {
				return nil
			}
		} else {
			if e.Scroll(response) == false {
				e.ClearScroll(response.Result.ScrollId())
				return nil
			}
		}

		return response
	}
}

func (e *ElasticInstance) IsMajorVersion(n int) bool {

	if e.MajorVersion == n {
		return true
	}
	return false
}

func (e *ElasticInstance) TotalCount(m map[string]any) int64 {

	ks := []string{"hits", "total"}
	if !e.IsMajorVersion(6) {
		ks = append(ks, "value")
	}

	val, err := NestedMapLookup(m, ks...)
	if err != nil {
		log.Println(err)
		return -1
	}

	return int64(val.(float64))
}

func (e *ElasticInstance) ScrollId(m map[string]any) string {

	sid, ok := m["_scroll_id"]
	if !ok {
		log.Println("no scroll id")
		return ""
	}

	return sid.(string)
}

func (e *ElasticInstance) Header(m any) Header {

	header := Header{}
	id, ok := m.(map[string]any)["_id"]
	if ok {
		header.Id = id.(string)
	}

	index, ok := m.(map[string]any)["_index"]
	if ok {
		header.Index = index.(string)
	}

	return header
}
