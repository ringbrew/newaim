package delivery

import (
	"github.com/ringbrew/gsv/server"
	"github.com/ringbrew/gsv/service"
	"github.com/ringbrew/newaim/productsearch/internal/delivery/product"
	"github.com/ringbrew/newaim/productsearch/internal/domain"
	"github.com/rs/cors"
)

func NewServer(ctx *domain.UseCaseContext) server.Server {
	opt := server.Classic()
	// set the server port
	opt.Name = "productsearch"
	opt.Host = ctx.Config.Host
	opt.Port = ctx.Config.Port
	opt.HttpMiddleware = append(opt.HttpMiddleware, cors.AllowAll())

	return server.NewServer(server.HTTP, &opt)
}

func ServiceList(ctx *domain.UseCaseContext) []service.Service {
	return []service.Service{product.NewService(ctx)}
}
