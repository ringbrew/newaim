package embedding

import (
	"context"
	"errors"
	"fmt"
	"github.com/ringbrew/newaim/productsearch/internal/domain"
	"github.com/sashabaranov/go-openai"
)

var modelAliasMap = map[string]openai.EmbeddingModel{
	"AdaEmbeddingV2": openai.AdaEmbeddingV2,
}

type OpenAI struct {
	ctx            *domain.UseCaseContext
	embeddingModel openai.EmbeddingModel
}

func newOpenAI(ctx *domain.UseCaseContext, embeddingModelAlias string) (*OpenAI, error) {
	if ctx.Config.OpenAI.Endpoint == "" || ctx.Config.OpenAI.Token == "" {
		return nil, errors.New("invalid open ai config")
	}

	em, found := modelAliasMap[embeddingModelAlias]
	if !found {
		return nil, fmt.Errorf("model-[%s] mapping not found", embeddingModelAlias)
	}

	return &OpenAI{
		ctx:            ctx,
		embeddingModel: em,
	}, nil
}

func (oa *OpenAI) EmbedDocument(ctx context.Context, req DocumentRequest) (DocumentResponse, error) {
	config := openai.DefaultConfig(oa.ctx.Config.OpenAI.Token)
	config.BaseURL = oa.ctx.Config.OpenAI.Endpoint
	client := openai.NewClientWithConfig(config)

	// Create an EmbeddingRequest for the user query
	embeddingReq := &openai.EmbeddingRequest{
		Input: req.Documents,
		Model: oa.embeddingModel,
	}

	// Create an embedding for the user query
	embeddingResp, err := client.CreateEmbeddings(ctx, embeddingReq)
	if err != nil {
		return DocumentResponse{}, err
	}

	result := DocumentResponse{
		Data: make([]Data, 0),
		Usage: Usage{
			PromptTokens:     embeddingResp.Usage.PromptTokens,
			CompletionTokens: embeddingResp.Usage.CompletionTokens,
			TotalTokens:      embeddingResp.Usage.TotalTokens,
		},
	}

	for _, v := range embeddingResp.Data {
		result.Data = append(result.Data, Data{
			Content: v.Object,
			Vector:  v.Embedding,
			Index:   v.Index,
		})
	}

	return result, nil
}

func (oa *OpenAI) EmbedSingle(ctx context.Context, req SingleRequest) (SingleResponse, error) {
	config := openai.DefaultConfig(oa.ctx.Config.OpenAI.Token)
	config.BaseURL = oa.ctx.Config.OpenAI.Endpoint
	client := openai.NewClientWithConfig(config)

	// Create an EmbeddingRequest for the user query
	embeddingReq := &openai.EmbeddingRequest{
		Input: []string{req.Content},
		Model: oa.embeddingModel,
	}

	// Create an embedding for the user query
	embeddingResp, err := client.CreateEmbeddings(ctx, embeddingReq)
	if err != nil {
		return SingleResponse{}, err
	}

	result := SingleResponse{
		Data: Data{},
		Usage: Usage{
			PromptTokens:     embeddingResp.Usage.PromptTokens,
			CompletionTokens: embeddingResp.Usage.CompletionTokens,
			TotalTokens:      embeddingResp.Usage.TotalTokens,
		},
	}

	if len(embeddingResp.Data) == 0 {
		return SingleResponse{}, errors.New("invalid embedding response")
	}

	result.Data = Data{
		Content: embeddingResp.Data[0].Object,
		Vector:  embeddingResp.Data[0].Embedding,
		Index:   embeddingResp.Data[0].Index,
	}

	return result, nil
}
