package tongyiclient

import (
	qwen "github.com/tmc/langchaingo/llms/tongyi/internal/qwenclient"
)

// qwen message Decorator in order to support llms.MessageContent.
type TongyiMessageContent struct {
	TextMessage *qwen.Message

	// TODO: intergrate tongyi-wanx and qwen-vl api to support image type
	// ImageMessage *qwen_client.ImageData
}

func (q *TongyiMessageContent) GetTextMessage() *qwen.Message {
	return q.TextMessage
}
