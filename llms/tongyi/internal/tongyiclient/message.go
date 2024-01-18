package tongyiclient

import (
	embedding "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/embedding"
	qwen "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/qwen"
)

type TextMessage = qwen.Message
type TextInput = qwen.Input
type TextQwenRequest = qwen.QwenRequest

type TextQwenOutputMessage = qwen.QwenOutputMessage
type TextQwenOutput = qwen.QwenOutput

type EmbeddingRequest = embedding.EmbeddingRequest

// qwen message Decorator in order to support llms.MessageContent.
type TongyiMessageContent struct {
	TextMessage *TextMessage

	// TODO: intergrate tongyi-wanx and qwen-vl api to support image type
	// ImageMessage *qwen_client.ImageData
}

func (q *TongyiMessageContent) GetTextMessage() *TextMessage {
	return q.TextMessage
}
