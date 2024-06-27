package product

import (
	"context"
	"github.com/ringbrew/gsv/service"
	"github.com/ringbrew/newaim/productsearch/internal/domain"
	"github.com/ringbrew/newaim/productsearch/internal/domain/product"
	"log"
)

type Service struct {
	ctx *domain.UseCaseContext

	name   string
	remark string
	desc   service.Description
}

func NewService(ctx *domain.UseCaseContext) service.Service {
	s := &Service{
		ctx:    ctx,
		name:   "product",
		remark: "产品模块",
	}

	uc := product.NewUseCase(ctx)

	if count, err := uc.Count(context.Background()); err != nil {
		log.Fatal(err.Error())
	} else if ctx.Config.ForceRebuild || count == 0 {
		r := NewDataReader("data/sku_list.zip")
		data, err := r.Read()
		if err != nil {
			log.Fatal(err.Error())
		}

		if ctx.Config.ForceRebuild {
			if err := uc.Rebuild(context.Background()); err != nil {
				log.Fatal(err.Error())
			}
		}

		if err := uc.BatchCreate(context.Background(), data); err != nil {
			log.Fatal(err.Error())
		}
	}

	handler := NewHandler(ctx, uc)
	s.desc.HttpRoute = append(s.desc.HttpRoute, handler.HttpRoute()...)
	return s
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Remark() string {
	return s.remark
}

func (s *Service) Description() service.Description {
	return s.desc
}
