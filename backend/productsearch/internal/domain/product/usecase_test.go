package product

import (
	"context"
	"github.com/ringbrew/newaim/productsearch/internal/conf"
	"github.com/ringbrew/newaim/productsearch/internal/domain"
	"log"
	"testing"
)

func TestQuery(t *testing.T) {
	config := conf.Config{
		ElasticSearch: conf.ElasticSearch{},
	}

	ctx := domain.NewUseCaseContext(config)

	uc := NewUseCase(ctx)

	result, total, err := uc.Query(context.Background(), "V539-NIK-DD6337-661-L", 0, 10)
	if err != nil {
		t.Error(err.Error())
		return
	}

	for _, v := range result {
		log.Println(v.SKU)
		log.Println(v.Score)
		//log.Println(v.Description)
	}
	log.Println(len(result))
	log.Println(total)
}
