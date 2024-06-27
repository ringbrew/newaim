package domain

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis/v8"
	"github.com/ringbrew/newaim/productsearch/internal/conf"
	"log"
	"sync"
)

type UseCaseContext struct {
	Config        conf.Config
	ElasticSearch *elasticsearch.Client
	Redis         *redis.Client
	Signal        context.Context
	cancel        context.CancelFunc
	WaitGroup     sync.WaitGroup
}

func (ctx *UseCaseContext) Watch() {
	ctx.WaitGroup.Add(1)
}

func (ctx *UseCaseContext) Close() {
	if ctx.cancel != nil {
		ctx.cancel()
	}
	ctx.WaitGroup.Wait()
}

var dsc *UseCaseContext

func NewUseCaseContext(c conf.Config) *UseCaseContext {
	if dsc == nil {
		dsc = &UseCaseContext{
			Config: c,
		}

		dsc.Redis = redis.NewClient(&redis.Options{
			Addr:     c.Redis.Host,
			DB:       c.Redis.DB,
			Password: c.Redis.Password,
		})

		//if err := dsc.Redis.Ping(context.Background()).Err(); err != nil {
		//	log.Fatal(err.Error())
		//}

		esClient, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: c.ElasticSearch.Address,
			Username:  c.ElasticSearch.UserName,
			Password:  c.ElasticSearch.Password,
		})
		if err != nil {
			log.Fatal(err)
		}

		dsc.ElasticSearch = esClient

		dsc.Signal, dsc.cancel = context.WithCancel(context.Background())
	}

	return dsc
}
