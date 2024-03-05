package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/alcomist/go-portfolio/internal/util"
	"github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"io"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func (e *ElasticInstance) Search(req *Request) *Response {

	var sizePtr *int = nil
	if req.Size > 0 {
		sizePtr = &req.Size
	}

	response := NewResponse()
	response.Reserve(req.Size)
	response.Result.scroll = req.Scroll

	res, err := esapi.SearchRequest{
		Index:  []string{req.Index},
		Scroll: req.Scroll,
		Size:   sizePtr,
		Body:   strings.NewReader(req.Query)}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error search documents", res.Status())
		log.Printf("[%s] Error string search documents", res.String())
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

		response.Result.SetTotalCount(e.TotalCount(r))
		response.Result.SetScrollId(e.ScrollId(r))

		//log.Printf("total: %d\n", response.Total)

		val, err := NestedMapLookup(r, "hits", "hits")
		if err != nil {
			log.Println(err)
			return nil
		}

		hits := val.([]any)

		response.Result.SetDocCount(len(hits))
		response.Result.AddProcessedCount(int64(response.Result.DocCount()))

		for _, hit := range hits {

			h := e.Header(hit)
			_source := hit.(map[string]any)["_source"]

			var s util.Interface
			s.From(_source.(map[string]interface{}))

			response.Result.Docs = append(response.Result.Docs, &Doc{Id: h.Id, Index: h.Index, Source: s})
		}

		return response
	}

	return nil
}

func (e *ElasticInstance) SearchAggs(req *Request) *Response {

	response := NewResponse()
	response.Reserve(req.Size)
	response.Result.scroll = time.Duration(0)

	res, err := esapi.SearchRequest{
		Index:  []string{req.Index},
		Scroll: response.Result.scroll,
		Size:   &req.Size,
		Body:   strings.NewReader(req.Query)}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error search documents", res.Status())
		return nil
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, res.Body)
	if err != nil {
		log.Println(err)
		return nil
	}

	if n > 0 {

		//os.WriteFile("./dat1", buf.Bytes(), 0644)

		var r map[string]any
		if err := json.NewDecoder(&buf).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
			return nil
		}

		response.Result.SetTotalCount(e.TotalCount(r))

		ks := []string{"aggregations", "aggs_hashcode", "buckets"}

		buckets, err := NestedMapLookup(r, ks...)
		if err != nil {
			log.Println(err)
			return nil
		}

		hitCount := 0

		for _, bucket := range buckets.([]any) {

			ks = []string{"aggs_top_hits", "hits", "hits"}

			hits, err := NestedMapLookup(bucket.(map[string]any), ks...)
			if err != nil {
				log.Println(err)
				continue
			}

			for _, hit := range hits.([]any) {

				hitCount++

				h := e.Header(hit)
				_source := hit.(map[string]any)["_source"]

				var s util.Interface
				s.From(_source.(map[string]interface{}))

				response.Result.Docs = append(response.Result.Docs, &Doc{Id: h.Id, Index: h.Index, Source: s})
			}
		}

		response.Result.SetDocCount(hitCount)
		response.Result.AddProcessedCount(int64(hitCount))

		return response
	}

	return nil
}

func (e *ElasticInstance) Scroll(response *Response) bool {

	response.Result.Docs = Documents{}

	res, err := esapi.ScrollRequest{
		ScrollID: response.Result.scrollID,
		Scroll:   response.Result.scroll}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	var buf bytes.Buffer
	n, err := io.Copy(&buf, res.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	if n > 0 {

		var r map[string]any
		if err := json.NewDecoder(&buf).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		}

		sid := e.ScrollId(r)
		if response.Result.scrollID != sid {
			response.Result.scrollID = sid
			log.Printf("scroll id has been changed: (%s => %s)\n", response.Result.scrollID, sid)
		}

		response.Result.SetTotalCount(e.TotalCount(r))

		val, err := NestedMapLookup(r, "hits", "hits")
		if err != nil {
			log.Println(err)
			return false
		}

		hits := val.([]any)

		hitCount := len(hits)

		response.Result.SetDocCount(hitCount)
		if response.Result.DocCount() == 0 {
			return false
		}

		response.Result.AddProcessedCount(int64(hitCount))

		for _, hit := range hits {

			h := e.Header(hit)
			_source := hit.(map[string]any)["_source"]

			var s util.Interface
			s.From(_source.(map[string]interface{}))

			response.Result.Docs = append(response.Result.Docs, &Doc{Id: h.Id, Index: h.Index, Source: s})
		}

		return true
	}

	return false
}

func (e *ElasticInstance) BulkInsert(index string, docs []*Doc) bool {

	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      index,
		Client:     e.client,
		NumWorkers: runtime.NumCPU(),
	})

	if err != nil {
		log.Fatalf("new bulk indexer error %s\n", err)
	}

	var countSuccessful uint64
	start := time.Now().UTC()

	retryOnConflict := 1

	ctx := context.Background()

	for _, doc := range docs {

		data, err := json.Marshal(doc.Source)
		if err != nil {
			log.Fatalf("Cannot encode item %s: %s\n", doc.Index, doc.Id)
		}

		id := strconv.FormatInt(doc.Int64("mall_product_id"), 10)

		err = indexer.Add(

			ctx,
			esutil.BulkIndexerItem{

				Index:           index,
				RetryOnConflict: &retryOnConflict,

				// Action field configures the operation to perform (index, create, delete, update)
				Action: "create",

				// DocumentID is the (optional) document ID
				DocumentID: id,

				// Body is an `io.Reader` with the payload
				///Body: bytes.NewReader(b),
				Body: bytes.NewReader(data),

				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(&countSuccessful, 1)
				},

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						fmt.Printf("ERROR: %s", err)
					} else {
						fmt.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)

		if err != nil {
			log.Fatalf("Unexpected error: %s", err)
		}
	}

	if err := indexer.Close(ctx); err != nil {
		log.Fatalf("Unexpected error: %s", err)
	}

	biStats := indexer.Stats()

	dur := time.Since(start)

	if biStats.NumFailed > 0 {
		log.Fatalf(
			"Indexed [%s] documents with [%s] errors in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			humanize.Comma(int64(biStats.NumFailed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	} else {
		log.Printf(
			"Sucessfuly indexed [%s] documents in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	}

	return true
}

func (e *ElasticInstance) Bulk(b []byte) bool {

	res, err := esapi.BulkRequest{
		Body: bytes.NewReader(b), Refresh: "true"}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error bulk insert", res.Status())
		return false
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, res.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	if n > 0 {

		var r map[string]any
		if err := json.NewDecoder(&buf).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
			return false
		}

	}

	return true
}

func (e *ElasticInstance) Aggregate(index, field, query string) []any {

	rs := make([]any, 0)

	size := 0

	res, err := esapi.SearchRequest{
		Index:  []string{index},
		Scroll: time.Duration(0),
		Size:   &size,
		Body:   strings.NewReader(query)}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error search documents", res.Status())
		return rs
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, res.Body)
	if err != nil {
		log.Println(err)
		return rs
	}

	if n > 0 {

		var r map[string]any
		if err := json.NewDecoder(&buf).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		}

		buckets, err := NestedMapLookup(r, "aggregations", field, "buckets")
		if err != nil {
			log.Printf("Error getting buckets: %s\n", err)
			return rs
		}

		for _, bucket := range buckets.([]any) {
			key := bucket.(map[string]any)["key"]
			rs = append(rs, key)
		}

		return rs
	}

	return rs
}

func (e *ElasticInstance) ClearScroll(scrollID string) bool {

	res, err := esapi.ClearScrollRequest{ScrollID: []string{scrollID}}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error search documents", res.Status())
		return false
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, res.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	if n > 0 {

		var r map[string]any
		if err := json.NewDecoder(&buf).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
			return false
		}

		val, ok := r["succeeded"]
		if ok {
			return val.(bool)
		}

		return false
	}

	return false
}

func (e *ElasticInstance) Cardinality(index, field string, builder *QueryStringQueryBuilder) int {

	term := CardinalityAggregationTerm(field)
	builder.AddAggregation(term)
	query := builder.String()

	size := 0

	res, err := esapi.SearchRequest{
		Index:  []string{index},
		Scroll: time.Duration(0),
		Size:   &size,
		Body:   strings.NewReader(query)}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error search documents", res.Status())
		return 0
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, res.Body)
	if err != nil {
		log.Println(err)
		return 0
	}

	if n > 0 {

		var r map[string]any
		if err := json.NewDecoder(&buf).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
			return 0
		}

		ks := []string{"aggregations", field, "value"}
		val, err := NestedMapLookup(r, ks...)
		if err != nil {
			log.Println(err)
			return -1
		}

		return int(val.(float64))
	}

	return 0
}

func (e *ElasticInstance) DeleteByQuery(index, query string) int {

	res, err := esapi.DeleteByQueryRequest{
		Index: []string{index},
		Body:  strings.NewReader(query)}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error search documents", res.Status())
		return -1
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, res.Body)
	if err != nil {
		log.Println(err)
		return -1
	}

	if n > 0 {

		var r map[string]any
		if err := json.NewDecoder(&buf).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
			return 0
		}

		val, ok := r["deleted"]
		if ok {
			return int(val.(float64))
		}
	}

	return 0
}

func (e *ElasticInstance) Delete(index, id string) bool {

	res, err := esapi.DeleteRequest{
		Index:        index,
		DocumentType: index,
		DocumentID:   id}.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error delete document", res.Status())
		log.Println(res.String())
		return false
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, res.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	if n > 0 {

		var r map[string]any
		if err := json.NewDecoder(&buf).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
			return false
		}

		_, ok := r["deleted"]
		if ok {
			return true
		}
	}

	return false
}
