package tongyi

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"

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
	// client          *qwen_client.QwenClient
	client  *tongyi_client.TongyiClient
	options options
}

var _ llms.Model = (*LLM)(nil)

func New(opts ...Option) (*LLM, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	client := tongyi_client.NewTongyiClient(o.model, o.token)

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

	if opts.Model == "" {
		opts.Model = q.client.Model
	}

	tongyiContents := messagesCntentToQwenMessages(messages)

	choices, err := q.doCompletionRequest(ctx, tongyiContents, opts)
	if err != nil {
		return nil, err
	}

	return &llms.ContentResponse{Choices: choices}, nil
}

func (q *LLM) doCompletionRequest(
	ctx context.Context,
	tongyiMessage []tongyi_client.TongyiMessageContent,
	opts llms.CallOptions,
) ([]*llms.ContentChoice, error) {
	switch opts.Model {
	// pure text generation
	case "qwen-turbo", "qwen-plus", "qwen-max", "qwen-max-1201", "qwen-max-longcontext":
		rsp, err := q.doTextCompletionRequest(ctx, tongyiMessage, opts)
		if err != nil {
			return nil, err
		}
		if len(rsp.Output.Choices) == 0 {
			return nil, ErrEmptyResponse
		}

		return q.createTextResult(rsp)
	// multimodal text and image
	case "qwen-vl-plus":
		panic("do Completion Error: not implemented qwen-vl-plus yet")
	// image generation
	case "wanx-v1":
		panic("do Completion Error: not implemented wanx-v1 yet")
	default:
		panic("do Completion Error: not support model")
	}
}

func (q *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	input := struct {
		Texts []string `json:"texts"`
	}{
		Texts: inputTexts,
	}
	embeddings, err := q.client.CreateEmbedding(ctx,
		&tongyi_client.EmbeddingRequest{
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
				qmsg.TextMessage = &tongyi_client.TextMessage{
					Role: typeToQwenRole(mc.Role),
					// Content: pt.Text,
				}
				qmsg.TextMessage.Content.SetText(pt.Text)
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

func (q *LLM) doTextCompletionRequest(ctx context.Context, tongyiMessage []tongyi_client.TongyiMessageContent, opts llms.CallOptions) (*tongyi_client.TextQwenOutputMessage, error) {
	qwenTextMessages := make([]tongyi_client.TextMessage, len(tongyiMessage))

	for i, tc := range tongyiMessage {
		qwenTextMessages[i] = *tc.GetTextMessage()
	}

	if len(qwenTextMessages) == 0 {
		return nil, ErrEmptyMessageContent
	}

	input := tongyi_client.TextInput{
		Messages: qwenTextMessages,
	}
	params := tongyi_client.DefaultQwenParameters()
	params.
		SetMaxTokens(opts.MaxTokens).
		SetTemperature(opts.Temperature).
		SetTopP(opts.TopP).
		SetTopK(opts.TopK).
		SetSeed(opts.Seed)

	req := &tongyi_client.TextQwenRequest{}
	req.
		SetModel(opts.Model).
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

	return rsp, nil
}

func (q *LLM) createTextResult(textResponse *tongyi_client.TextQwenOutputMessage) ([]*llms.ContentChoice, error) {
	choices := make([]*llms.ContentChoice, len(textResponse.Output.Choices))
	for i, c := range textResponse.Output.Choices {
		choices[i] = &llms.ContentChoice{
			Content:    c.Message.Content.ToString(),
			StopReason: c.FinishReason,
			GenerationInfo: map[string]any{
				"PromptTokens":     textResponse.Usage.InputTokens,
				"CompletionTokens": textResponse.Usage.OutputTokens,
				"TotalTokens":      textResponse.Usage.TotalTokens,
			},
		}
	}

	return choices, nil
}
