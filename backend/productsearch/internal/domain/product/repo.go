package product

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/ringbrew/newaim/productsearch/internal/domain"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const productIndex = "newaim_product_sku_index"

type ESResponse struct {
	Took         int                    `json:"took"`
	TimedOut     bool                   `json:"timed_out"`
	Shards       ESShardResponse        `json:"_shards"`
	Hits         ESHitResponse          `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations"`
}

type ESShardResponse struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

type Hit struct {
	Index  string                 `json:"_index"`
	Type   string                 `json:"_type"`
	Id     string                 `json:"_id"`
	Source map[string]interface{} `json:"_source"`
	Fields map[string]interface{} `json:"fields"`
	Sort   []int                  `json:"sort"`
	Score  float64                `json:"_score"`
}

type ESHitResponse struct {
	Total struct {
		Value    int64  `json:"value"`
		Relation string `json:"relation"`
	} `json:"total"`
	Hits []Hit
}

type repo struct {
	es        *elasticsearch.Client
	bulkIndex map[string]esutil.BulkIndexer
}

func newRepo(ctx *domain.UseCaseContext) *repo {
	r := &repo{
		es:        ctx.ElasticSearch,
		bulkIndex: make(map[string]esutil.BulkIndexer),
	}

	if exist, err := r.CheckIndexExist(productIndex); err != nil {
		log.Fatal(err.Error())
	} else if !exist {
		if err := r.CreateIndexES(productIndex, productMapping); err != nil {
			log.Fatal(err.Error())
		}
	}

	if err := r.BulkIndex(productIndex); err != nil {
		log.Fatal(err.Error())
	}

	return r
}

func (r *repo) BulkIndex(indexName string) error {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         indexName,        // The default index name
		Client:        r.es,             // The Elasticsearch client
		NumWorkers:    1,                // The number of worker goroutines
		FlushBytes:    int(5e+6),        // The flush threshold in bytes
		FlushInterval: 30 * time.Second, // The periodic flush interval
	})
	if err != nil {
		return err
	}

	r.bulkIndex[indexName] = bi

	return nil
}

func (r *repo) CountIndex(indexName string) (int64, error) {
	req := esapi.CountRequest{
		Index: []string{indexName},
	}

	resp, err := req.Do(context.Background(), r.es)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var count struct {
		Count int64 `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&count); err != nil {
		return 0, err
	}

	return count.Count, nil
}

func (r *repo) CreateMany(ctx context.Context, product []*Product) error {
	for _, v := range product {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}

		if err := r.bulkIndex[productIndex].Add(
			ctx,
			esutil.BulkIndexerItem{
				//index, create, delete, update
				Action:     "index",
				DocumentID: v.Id,
				Body:       bytes.NewReader(data),
				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("ERROR: %s", err)
					} else {
						log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *repo) SearchById(ctx context.Context, id []string) ([]Product, error) {
	query := map[string]interface{}{
		"sort": []interface{}{
			map[string]interface{}{
				"_score": "desc",
			},
		},

		"fields": []string{"id", "createTime", "updateTime", "sku", "title", "description"},
	}

	query["query"] = map[string]interface{}{
		"term": map[string]interface{}{
			"id": id,
		},
	}

	result, _, err := r.searchProductByQuery(0, int64(len(id)), query)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *repo) Search(ctx context.Context, keyword string, from, size int64, isSku ...bool) ([]Product, int64, error) {
	query := map[string]interface{}{
		"sort": []interface{}{
			map[string]interface{}{
				"_score": "desc",
			},
		},

		"fields": []string{"id", "createTime", "updateTime", "sku", "title", "description"},
	}

	if len(isSku) > 0 && isSku[0] {
		query["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"sku": keyword,
						},
					},
				},
			},
		}
	} else {
		query["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"sku": keyword,
						},
					},
					{
						"match": map[string]interface{}{
							"title": keyword,
						},
					},
					{
						"match": map[string]interface{}{
							"description": keyword,
						},
					},
				},
			},
		}
	}

	return r.searchProductByQuery(from, size, query)
}

func (r *repo) searchProductByQuery(from, size int64, query map[string]interface{}) ([]Product, int64, error) {
	data, err := r.searchFromES(productIndex, from, size, query)
	if err != nil {
		return nil, 0, err
	}

	parseString := func(input interface{}) string {
		if val, ok := input.([]interface{}); ok {
			if len(val) > 0 {
				if sVal, ok := val[0].(string); ok {
					return sVal
				}
			}
		}
		return ""
	}

	parseDate := func(input interface{}) time.Time {
		if val, ok := input.([]interface{}); ok {
			if len(val) > 0 {
				if sVal, ok := val[0].(string); ok {
					df := strings.Fields(sVal)
					if len(df) != 2 {
						return time.Time{}
					}

					if result, err := time.Parse("2006-01-02 15:04:05", strings.Join([]string{df[0], strings.Split(df[1], ".")[0]}, " ")); err == nil {
						return result
					}
				}
			}
		}
		return time.Time{}
	}

	hitsToProduct := func(hit Hit) Product {
		return Product{
			Id:          hit.Id,
			CreateTime:  parseDate(hit.Fields["createTime"]),
			UpdateTime:  parseDate(hit.Fields["updateTime"]),
			Title:       parseString(hit.Fields["title"]),
			SKU:         parseString(hit.Fields["sku"]),
			Description: parseString(hit.Fields["description"]),
			Score:       hit.Score,
		}
	}

	result := make([]Product, 0, len(data.Hits.Hits))
	for _, v := range data.Hits.Hits {
		result = append(result, hitsToProduct(v))
	}

	return result, data.Hits.Total.Value, nil
}

func (r *repo) searchFromES(index string, from, size int64, query map[string]interface{}) (ESResponse, error) {
	var buf bytes.Buffer
	var result ESResponse
	query["from"] = from
	query["size"] = size
	//query["search_after"] = []int64{from}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return result, err
	}

	res, err := r.es.Search(
		r.es.Search.WithContext(context.Background()),
		r.es.Search.WithIndex(index),
		r.es.Search.WithBody(&buf),
		r.es.Search.WithTrackTotalHits(true),
		r.es.Search.WithPretty(),
		r.es.Search.WithExplain(true),
	)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return result, fmt.Errorf("error parsing the error response body: %s", err.Error())
		} else {
			if errInfo, exist := e["error"]; exist {
				if ei, ok := errInfo.(map[string]interface{}); ok {
					return result, fmt.Errorf("error query from es,status[%s] type[%v],reason[%v]", res.Status(), ei["type"], ei["reason"])
				}
			}

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return result, fmt.Errorf("error query from es, unknown error[%s]", string(data))
			}
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("error parsing the response body: %s", err)
	}

	return result, nil
}

/*
@desc: product mapping
*/
var productMapping = map[string]interface{}{
	"mappings": map[string]interface{}{
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type": "keyword",
			},
			"createTime": map[string]interface{}{
				"type":   "date",
				"format": "yyyy-MM-dd HH:mm:ss.S||strict_date_optional_time||epoch_millis",
			},
			"updateTime": map[string]interface{}{
				"type":   "date",
				"format": "yyyy-MM-dd HH:mm:ss.S||strict_date_optional_time||epoch_millis",
			},
			"sku": map[string]interface{}{
				"type": "keyword",
			},
			"title": map[string]interface{}{
				"type": "text",
			},
			"description": map[string]interface{}{
				"type": "text",
			},
		},
	},
}

func (r *repo) CheckIndexExist(idx string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{idx},
	}

	resp, err := req.Do(context.Background(), r.es)
	if err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else {
		return true, nil
	}
}

func (r *repo) CreateIndexES(idx string, mapping map[string]interface{}) error {
	b, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	req := esapi.IndicesCreateRequest{
		Index: idx,
		Body:  bytes.NewReader(b),
	}

	resp, err := req.Do(context.Background(), r.es)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Println(string(data))

	return nil
}

func (r *repo) DeleteIndexES(index string) error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{index},
	}

	resp, err := req.Do(context.Background(), r.es)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Println(string(data))

	return nil
}
