package product

import (
	"github.com/mholt/binding"
	"github.com/ringbrew/gsv/service"
	"github.com/ringbrew/newaim/productsearch/internal/delivery/common"
	"github.com/ringbrew/newaim/productsearch/internal/domain"
	"github.com/ringbrew/newaim/productsearch/internal/domain/product"
	"net/http"
	"strings"
)

type Handler struct {
	ctx *domain.UseCaseContext
	uc  *product.UseCase
}

func NewHandler(ctx *domain.UseCaseContext, uc *product.UseCase) *Handler {
	return &Handler{
		ctx: ctx,
		uc:  uc,
	}
}

type SearchParam struct {
	From    int64  `json:"from"`
	Size    int64  `json:"size"`
	Keyword string `json:"keyword"`
}

func (sp *SearchParam) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&sp.From: "from",
		&sp.Size: "size",
		&sp.Keyword: binding.Field{
			Form: "keyword",
		},
	}
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-Newaim-Api-Key")
	if apiKey == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("auth fail"))
		return
	}

	if err := NewLimiter(h.ctx).Check(r.Context(), CheckLimitInput{
		Aspect: AspectApiKeyAccess,
		ApiKey: apiKey,
	}); err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(err.Error()))
		return
	}

	sp := SearchParam{}
	if err := binding.Bind(r, &sp); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	sp.Keyword = strings.TrimSpace(sp.Keyword)

	if err := NewLimiter(h.ctx).Check(r.Context(), CheckLimitInput{
		Aspect: AspectApiKeyInput,
		ApiKey: apiKey,
		Input:  sp,
	}); err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(err.Error()))
		return
	}

	data, total, err := h.uc.Query(r.Context(), sp.Keyword, sp.From, sp.Size)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if err := NewLimiter(h.ctx).Check(r.Context(), CheckLimitInput{
		Aspect: AspectApiKeyOutput,
		ApiKey: apiKey,
		Output: data,
	}); err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(err.Error()))
		return
	}

	common.Render().JSON(w, http.StatusOK, map[string]interface{}{
		"total": total,
		"data":  data,
	})
}

func (h *Handler) HttpRoute() []service.HttpRoute {
	result := []service.HttpRoute{
		service.NewHttpRoute(http.MethodGet, "/product", h.Query, service.HttpMeta{
			Remark: "查询产品",
		}),
	}
	return result
}
