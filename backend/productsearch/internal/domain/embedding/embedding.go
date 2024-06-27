package embedding

import (
	"context"
	"errors"
	"github.com/ringbrew/newaim/productsearch/internal/domain"
)

type Embedding interface {
	EmbedDocument(ctx context.Context, req DocumentRequest) (DocumentResponse, error)
	EmbedSingle(ctx context.Context, req SingleRequest) (SingleResponse, error)
}

func NewEmbedding(ctx *domain.UseCaseContext, alias string) (Embedding, error) {
	return newOpenAI(ctx, alias)
}

type DocumentRequest struct {
	Documents []string
}

type DocumentResponse struct {
	Data  []Data
	Usage Usage
}

type SingleRequest struct {
	Content string
}

type SingleResponse struct {
	Data  Data
	Usage Usage
}

type Data struct {
	Content string
	Vector  Vector
	Index   int
}

var ErrVectorLengthMismatch = errors.New("vector length mismatch")

type Vector []float32

func (v Vector) DotProduct(other Vector) (float32, error) {
	if len(v) != len(other) {
		return 0, ErrVectorLengthMismatch
	}

	var dotProduct float32
	for i := range v {
		dotProduct += v[i] * other[i]
	}

	return dotProduct, nil
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
