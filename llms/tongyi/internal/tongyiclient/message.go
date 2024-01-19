package tongyiclient

import (
	embedding "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/embedding"
	qwen "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/qwen"
)

// qwen
type TextMessage = qwen.Message[*qwen.TextContent]
type TextInput = qwen.Input[*qwen.TextContent]
type TextQwenRequest = qwen.QwenRequest[*qwen.TextContent]

type TextQwenOutputMessage = qwen.QwenOutputMessage[*qwen.TextContent]
type TextQwenOutput = qwen.QwenOutput[*qwen.TextContent]

// qwen-vl
// type VLContentList = *qwen.VLContentList
type VLMessage = qwen.Message[*qwen.VLContentList]
type VLInput = qwen.Input[*qwen.VLContentList]
type VLQwenRequest = qwen.QwenRequest[*qwen.VLContentList]

type VLQwenOutputMessage = qwen.QwenOutputMessage[*qwen.VLContentList]

type EmbeddingRequest = embedding.EmbeddingRequest

// type TextQwenOutputMessage = qwen.QwenOutputMessage[*qwen.TextContent]

// qwen message Decorator in order to support llms.MessageContent.
type TongyiMessageContent struct {
	TextMessage *TextMessage

	// TODO: intergrate tongyi-wanx and qwen-vl api to support image type
	// ImageMessage *qwen_client.ImageData
}

func (q *TongyiMessageContent) GetTextMessage() *TextMessage {
	return q.TextMessage
}
