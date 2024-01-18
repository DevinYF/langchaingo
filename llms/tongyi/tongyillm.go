package tongyi

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	qwen_client "github.com/tmc/langchaingo/llms/tongyi/internal/qwenclient"
	tongyi_client "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingToken             = errors.New("missing the Dashscope API key, set it in the DASHSCOPE_API_KEY environment variable")
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrIncompleteEmbedding      = errors.New("no all input got emmbedded")
	ErrNotSupportImsgePart      = errors.New("not support Image parts yet")
	ErrMultipleTextParts        = errors.New("found multiple text parts in message")
	ErrEmptyMessageContent      = errors.New("TongyiMessageContent is empty")
)

type UnSupportedRoleError struct {
	Role schema.ChatMessageType
}

func (e *UnSupportedRoleError) Error() string {
	return fmt.Sprintf("qwen role %s not supported", e.Role)
}

type LLM struct {
	CallbackHandler callbacks.Handler
	client          *qwen_client.QwenClient
	options         options
}

var _ llms.Model = (*LLM)(nil)

func New(opts ...Option) (*LLM, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	client := qwen_client.NewQwenClient(o.model, o.token, qwen_client.NewHTTPClient())

	return &LLM{client: client, options: o}, nil
}

// Call implements llms.LLM.
func (q *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.CallLLM(ctx, q, prompt, options...)
}

// nolint:lll
// GenerateContent implements llms.Model.
func (q *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	var model string
	if opts.Model != "" {
		model = string(qwen_client.ChoseQwenModel(opts.Model))
	} else {
		model = string(q.client.Model)
	}

	tongyiContents := messagesCntentToQwenMessages(messages)
	qwenTextMessages := make([]qwen_client.Message, len(tongyiContents))
	for i, tc := range tongyiContents {
		qwenTextMessages[i] = *tc.GetTextMessage()
	}
	rsp, err := q.doCompletionRequest(ctx, model, qwenTextMessages, opts)
	if err != nil {
		return nil, err
	}

	choices := make([]*llms.ContentChoice, len(rsp.Output.Choices))
	for i, c := range rsp.Output.Choices {
		choices[i] = &llms.ContentChoice{
			Content:    c.Message.Content,
			StopReason: c.FinishReason,
			GenerationInfo: map[string]any{
				"PromptTokens":     rsp.Usage.InputTokens,
				"CompletionTokens": rsp.Usage.OutputTokens,
				"TotalTokens":      rsp.Usage.TotalTokens,
			},
		}
	}

	return &llms.ContentResponse{Choices: choices}, nil
}

func (q *LLM) doCompletionRequest(
	ctx context.Context,
	model string,
	qwenTextMessages []qwen_client.Message,
	opts llms.CallOptions,
) (*qwen_client.QwenOutputMessage, error) {
	input := qwen_client.Input{
		Messages: qwenTextMessages,
	}
	params := qwen_client.DefaultParameters()
	params.
		SetMaxTokens(opts.MaxTokens).
		SetTemperature(opts.Temperature).
		SetTopP(opts.TopP).
		SetTopK(opts.TopK).
		SetSeed(opts.Seed)

	req := &qwen_client.QwenRequest{}
	req.
		SetModel(model).
		SetInput(input).
		SetParameters(params).
		SetStreamingFunc(opts.StreamingFunc)

	rsp, err := q.client.CreateCompletion(ctx, req)
	if err != nil {
		if q.CallbackHandler != nil {
			q.CallbackHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}
	if len(rsp.Output.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	return rsp, nil
}

func (q *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	input := struct {
		Texts []string `json:"texts"`
	}{
		Texts: inputTexts,
	}
	embeddings, err := q.client.CreateEmbedding(ctx,
		&qwen_client.EmbeddingRequest{
			Input: input,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(embeddings) != len(inputTexts) {
		return nil, ErrIncompleteEmbedding
	}

	return embeddings, nil
}

func messagesCntentToQwenMessages(messagesContent []llms.MessageContent) []tongyi_client.TongyiMessageContent {
	qwenMessages := make([]tongyi_client.TongyiMessageContent, len(messagesContent))

	for i, mc := range messagesContent {
		foundText := false
		qmsg := tongyi_client.TongyiMessageContent{}
		for _, p := range mc.Parts {
			switch pt := p.(type) {
			case llms.TextContent:
				qmsg.TextMessage = &qwen_client.Message{
					Role:    typeToQwenRole(mc.Role),
					Content: pt.Text,
				}
				if foundText {
					panic(ErrMultipleTextParts)
				}
				foundText = true
			case llms.BinaryContent:
				// imageData := qwen_client.ImageData{Data: pt.Data}
				// qmsg.ImageMessage = &imageData
				panic(ErrNotSupportImsgePart)
			default:
				panic("only support Text parts right now")
			}
		}

		qwenMessages[i] = qmsg
	}
	return qwenMessages
}

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
