package tongyiclient

import (
	"github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/qwen"
)

type (
	TextInput        = qwen.Input[*qwen.TextContent]
	TextRequest      = qwen.Request[*qwen.TextContent]
	TextQwenResponse = qwen.OutputResponse[*qwen.TextContent]
	VLInput          = qwen.Input[*qwen.VLContentList]
	VLRequest        = qwen.Request[*qwen.VLContentList]
	VLQwenResponse   = qwen.OutputResponse[*qwen.VLContentList]
)

// qwen.
type TextMessage = qwen.Message[*qwen.TextContent]

func NewTextMessage(role string) *TextMessage {
	content := qwen.NewTextContent()

	return &TextMessage{
		Role:    role,
		Content: content,
	}
}

// qwen-vl.
type VLMessage = qwen.Message[*qwen.VLContentList]

func NewVLMessage(role string) *VLMessage {
	content := qwen.NewVLContentList()

	return &qwen.Message[*qwen.VLContentList]{
		Role:    role,
		Content: content,
	}
}
