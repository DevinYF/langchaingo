package tongyiclient

import (
	"github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/embedding"
	"github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/qwen"
)

type ITongyiCntent = qwen.IQwenContent

type Usage = qwen.Usage

// qwen
type TextMessage = qwen.Message[*qwen.TextContent]

func NewTextMessage(role string) *TextMessage {
	content := qwen.NewTextContent()

	return &TextMessage{
		Role:    "",
		Content: content,
	}
}

type TextInput = qwen.Input[*qwen.TextContent]
type TextQwenRequest = qwen.QwenRequest[*qwen.TextContent]

type TextQwenResponse = qwen.QwenOutputResponse[*qwen.TextContent]
type TextQwenOutput = qwen.QwenOutput[*qwen.TextContent]

type VLQwenOutput = qwen.QwenOutput[*qwen.VLContentList]

// qwen-vl
type VLMessage = qwen.Message[*qwen.VLContentList]

func NewVLMessage(role string) *VLMessage {
	content := qwen.NewVLContentList()

	return &VLMessage{
		Role:    role,
		Content: content,
	}
}

type VLInput = qwen.Input[*qwen.VLContentList]
type VLQwenRequest = qwen.QwenRequest[*qwen.VLContentList]

type VLQwenResponse = qwen.QwenOutputResponse[*qwen.VLContentList]

type EmbeddingRequest = embedding.EmbeddingRequest

type Choice[T ITongyiCntent] qwen.Choice[T]

type TmpTongyiOutput[T ITongyiCntent] interface {
	TextQwenOutput | VLQwenOutput
	GetChoices() []qwen.Choice[T]
	GetUsage() qwen.Usage
}
