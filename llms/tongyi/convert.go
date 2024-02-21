package tongyi

import (
	"fmt"

	tongyiclient "github.com/devinyf/dashscopego"
	"github.com/devinyf/dashscopego/qwen"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// IQwenResponseConverter is an interface for converting tongyi response to llms.ContentChoice.
type IQwenResponseConverter interface {
	convertToContentChoice() []*llms.ContentChoice
}

type TextRespConverter tongyiclient.TextQwenResponse

var _ IQwenResponseConverter = (*TextRespConverter)(nil)

func (t *TextRespConverter) convertToContentChoice() []*llms.ContentChoice {
	return convertToContentChoice(t.Output.Choices, t.Usage)
}

type VLRespConverter tongyiclient.VLQwenResponse

var _ IQwenResponseConverter = (*VLRespConverter)(nil)

func (t *VLRespConverter) convertToContentChoice() []*llms.ContentChoice {
	return convertToContentChoice(t.Output.Choices, t.Usage)
}

// nolint:lll
func convertToContentChoice[T qwen.IQwenContent](rawchoices []qwen.Choice[T], usage qwen.Usage) []*llms.ContentChoice {
	choices := make([]*llms.ContentChoice, len(rawchoices))
	for i, c := range rawchoices {
		choices[i] = &llms.ContentChoice{
			Content:    c.Message.Content.ToString(),
			StopReason: c.FinishReason,
			GenerationInfo: map[string]any{
				"PromptTokens":     usage.InputTokens,
				"CompletionTokens": usage.OutputTokens,
				"TotalTokens":      usage.TotalTokens,
			},
		}
	}
	return choices
}

// convert tongyi response to llms.ContentChoice.
func convertTongyiResultToContentChoice(resp IQwenResponseConverter) ([]*llms.ContentChoice, error) {
	return resp.convertToContentChoice(), nil
}

// convert llms.MessageContent to tongyi TextMessage.
func messagesCntentToQwenTextMessages(messagesContent []llms.MessageContent) []tongyiclient.TextMessage {
	return genericMessageToQwenMessage(messagesContent, qwen.NewTextContent())
}

// convert llms.MessageContent to tongyi VLMessage.
func messageConventToQwenVLMessage(messagesContent []llms.MessageContent) []tongyiclient.VLMessage {
	return genericMessageToQwenMessage(messagesContent, qwen.NewVLContentList())
}

// convert llms.MessageContent to tongyi Message.
// params content T is qwen.IQwenContent type, because there is no way to define new(T) behavior,
// we need to pass content T from outside which is non-nil.
// nolint:lll
func genericMessageToQwenMessage[T qwen.IQwenContent](messagesContent []llms.MessageContent, content T) []qwen.Message[T] {
	if content == nil {
		panic("content is nil")
	}

	qwenMessages := make([]qwen.Message[T], len(messagesContent))
	for i, mc := range messagesContent {
		qmsg := tongyiclient.NewQwenMessage[T](typeToQwenRole(mc.Role), content)

		foundText := false
		for _, p := range mc.Parts {
			switch pt := p.(type) {
			case llms.TextContent:
				qmsg.Content.SetText(pt.Text)
				if foundText {
					panic(ErrMultipleTextParts)
				}
				foundText = true
			case llms.ImageURLContent:
				qmsg.Content.SetImage(pt.URL)
			case llms.BinaryContent:
				panic("not implement BinaryContent yet")
			default:
				panic(fmt.Sprintf("not supported ContentPart type: %T", pt))
			}
		}
		qwenMessages[i] = *qmsg
	}

	return qwenMessages
}

// convert ChatMessageType to qwen role.
func typeToQwenRole(typ schema.ChatMessageType) string {
	switch typ {
	case schema.ChatMessageTypeSystem:
		return "system"
	case schema.ChatMessageTypeAI:
		return "assistant"
	case schema.ChatMessageTypeHuman:
		return "user"
	case schema.ChatMessageTypeGeneric:
		return "user"
	case schema.ChatMessageTypeFunction:
		fallthrough
	default:
		panic(&UnSupportedRoleError{Role: typ})
	}
}
