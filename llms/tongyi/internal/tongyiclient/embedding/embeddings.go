package qwenclient

import (
	"context"

	"github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/httpclient"
)

const (
	embeddingURL          = "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"
	defaultEmbeddingModel = "text-embedding-v1"
)

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input struct {
		Texts []string `json:"texts"`
	} `json:"input"`
	Params struct {
		TextType string `json:"text_type"` // query or document
	} `json:"parameters"`
}

type Embedding struct {
	TextIndex int       `json:"text_index"`
	Embedding []float32 `json:"embedding"`
}

type EmbeddingOutput struct {
	Embeddings []Embedding `json:"embeddings"`
	Usgae      struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
}

type EmbeddingResponse struct {
	Output EmbeddingOutput `json:"output"`
}

//nolilnt:lll
func CreateEmbedding(ctx context.Context, req *EmbeddingRequest, cli httpclient.IHttpClient, token string) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = defaultEmbeddingModel
	}
	if req.Params.TextType == "" {
		req.Params.TextType = "document"
	}

	resp := EmbeddingResponse{}
	tokenOption := httpclient.WithTokenHeaderOption(token)
	err := cli.Post(ctx, embeddingURL, req, &resp, tokenOption)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
