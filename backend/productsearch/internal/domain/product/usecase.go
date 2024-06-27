package product

import (
	"context"
	"github.com/ringbrew/newaim/productsearch/internal/domain"
	"github.com/ringbrew/newaim/productsearch/internal/domain/embedding"
	"log"
	"strings"
	"time"
	"unicode"
)

type UseCase struct {
	ctx  *domain.UseCaseContext
	repo *repo
	ms   *MilvusStore
}

func NewUseCase(ctx *domain.UseCaseContext) *UseCase {
	uc := &UseCase{
		ctx:  ctx,
		repo: newRepo(ctx),
	}

	if ctx.Config.Miluvs.Endpoint != "" {
		ms, err := newMilvusStore(ctx)
		if err != nil {
			log.Fatal(err.Error())
		}
		uc.ms = ms
	}

	return uc
}

func (uc *UseCase) Count(ctx context.Context) (int64, error) {
	return uc.repo.CountIndex(productIndex)
}

func (uc *UseCase) Rebuild(ctx context.Context) error {
	if err := uc.repo.DeleteIndexES(productIndex); err != nil {
		return err
	}
	if err := uc.repo.CreateIndexES(productIndex, productMapping); err != nil {
		return err
	}
	return nil
}

func (uc *UseCase) BatchCreate(ctx context.Context, product []*Product) error {
	emDoc := make([]string, len(product))

	for i, v := range product {
		v.SetId(NewIdGenerator().NewId())
		v.CreateTime = time.Now()
		v.UpdateTime = time.Now()
		emDoc[i] = v.Description
	}

	if err := uc.repo.CreateMany(ctx, product); err != nil {
		return err
	}

	if uc.ctx.Config.OpenAI.Token != "" {
		em, err := embedding.NewEmbedding(uc.ctx, "AdaEmbeddingV2")
		if err != nil {
			return err
		}

		embeddingResult, err := em.EmbedDocument(ctx, embedding.DocumentRequest{Documents: emDoc})
		if err != nil {
			return err
		}

		er := make(map[int]embedding.Vector)
		for _, v := range embeddingResult.Data {
			er[v.Index] = v.Vector
		}

		for i := range product {
			product[i].Vector = er[i]
		}

		if err := uc.ms.BatchCreate(ctx, product); err != nil {
			return err
		}
	}

	return nil
}

func (uc *UseCase) Query(ctx context.Context, keyword string, from, size int64) ([]Product, int64, error) {
	isSku := func(s string) bool {
		for _, r := range s {
			if !unicode.IsUpper(r) && unicode.IsLetter(r) && r != '-' {
				return false
			}
		}
		return true
	}

	result, total, err := uc.repo.Search(ctx, strings.Join(strings.Fields(keyword), " AND "), from, size, isSku(keyword))
	if err != nil {
		return nil, 0, err
	}

	if total == 0 && uc.ctx.Config.OpenAI.Token != "" {
		em, err := embedding.NewEmbedding(uc.ctx, "AdaEmbeddingV2")
		if err != nil {
			return nil, 0, err
		}

		qv, err := em.EmbedSingle(ctx, embedding.SingleRequest{Content: keyword})
		if err != nil {
			return nil, 0, err
		}

		qvr := QueryVectorRequest{}
		qvr.Input = qv.Data.Vector
		qvr.Top = int(size)

		vr, err := uc.ms.Query(ctx, qvr)
		if err != nil {
			return nil, 0, err
		}

		idList := make([]string, 0)
		for _, v := range vr.Data {
			idList = append(idList, v.Id)
		}

		result, err = uc.repo.SearchById(ctx, idList)
		if err != nil {
			return nil, 0, err
		}

		total = int64(len(result))
	}

	return result, total, nil
}
