package tongyiclient

import (
	qwen "github.com/tmc/langchaingo/llms/tongyi/internal/qwenclient"
)

type TongyiClient struct {
	qwen.QwenClient
}

func NewTongyiClient(model string, token string, httpCli qwen.IHttpClient) *qwen.QwenClient {
	// qwenModel := ChoseQwenModel(model)

	return qwen.NewQwenClient(model, token, httpCli)

}
