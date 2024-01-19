package tongyi

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"

	"github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient"
	"github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/qwen"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingToken             = errors.New("missing the Dashscope API key, set it in the DASHSCOPE_API_KEY environment variable")
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrIncompleteEmbedding      = errors.New("no all input got emmbedded")
	ErrNotSupportImsgePart      = errors.New("not support Image parts yet")
	ErrMultipleTextParts        = errors.New("found multiple text parts in message")
	ErrEmptyMessageContent      = errors.New("tongyi MessageContent is empty")
	ErrNotImplemented           = errors.New("not implemented yet")
)

type UnSupportedRoleError struct {
	Role schema.ChatMessageType
}

func (e *UnSupportedRoleError) Error() string {
	return fmt.Sprintf("qwen role %s not supported", e.Role)
}

type LLM struct {
	CallbackHandler callbacks.Handler
	client  *tongyiclient.TongyiClient
	options options
}

var _ llms.Model = (*LLM)(nil)

func New(opts ...Option) (*LLM, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	client := tongyiclient.NewTongyiClient(o.model, o.token)

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

	log.Println("debug... model: ", opts.Model)

	choices, err := q.doCompletionRequest(ctx, messages, opts)
	if err != nil {
		return nil, err
	}

	return &llms.ContentResponse{Choices: choices}, nil
}

func (q *LLM) doCompletionRequest(
	ctx context.Context,
	messages []llms.MessageContent,
	opts llms.CallOptions,
) ([]*llms.ContentChoice, error) {
	switch opts.Model {
	// pure text-gen model.
	case "qwen-turbo", "qwen-plus", "qwen-max", "qwen-max-1201", "qwen-max-longcontext":

		tongyiContents := messagesCntentToQwenMessages(messages)

		rsp, err := q.doTextCompletionRequest(ctx, tongyiContents, opts)
		if err != nil {
			return nil, err
		}
		if len(rsp.Output.Choices) == 0 {
			return nil, ErrEmptyResponse
		}

		converter := TextRespConverter(*rsp)

		return convertTongyiResultToContentChoice(&converter)
	// multi-modal: text / image.
	case "qwen-vl-plus":

		tongyiContents := messageConventToQwenVLMessage(messages)
		rsp, err := q.doVLCompletionRequest(ctx, tongyiContents, opts)
		if err != nil {
			return nil, err
		}
		if len(rsp.Output.Choices) == 0 {
			return nil, ErrEmptyResponse
		}

		converter := VLRespConverter(*rsp)

		return convertTongyiResultToContentChoice(&converter)
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
		&tongyiclient.EmbeddingRequest{
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

func messagesCntentToQwenMessages(messagesContent []llms.MessageContent) []tongyiclient.TextMessage {
	qwenMessages := make([]tongyiclient.TextMessage, len(messagesContent))

	for i, mc := range messagesContent {
		foundText := false

		qmsg := tongyiclient.NewTextMessage(typeToQwenRole(mc.Role))
		for _, p := range mc.Parts {
			switch pt := p.(type) {
			case llms.TextContent:

				qmsg.Content.SetText(pt.Text)

				if foundText {
					panic(ErrMultipleTextParts)
				}
				foundText = true
			default:
				panic("only support Text parts right now")
			}
		}

		qwenMessages[i] = *qmsg
	}
	return qwenMessages
}

func messageConventToQwenVLMessage(messagesContent []llms.MessageContent) []tongyiclient.VLMessage {
	qwenMessages := make([]tongyiclient.VLMessage, len(messagesContent))
	for i, mc := range messagesContent {
		qmsg := tongyiclient.NewVLMessage(typeToQwenRole(mc.Role))

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
				qmsg.Content.SetText(pt.URL)
			default:
				panic("only support Text Image parts")

			}
		}
		qwenMessages[i] = *qmsg
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

func (q *LLM) doTextCompletionRequest(ctx context.Context, qwenTextMessages []tongyiclient.TextMessage, opts llms.CallOptions) (*tongyiclient.TextQwenResponse, error) {
	if len(qwenTextMessages) == 0 {
		return nil, ErrEmptyMessageContent
	}

	input := tongyiclient.TextInput{
		Messages: qwenTextMessages,
	}

	params := tongyiclient.DefaultQwenParameters()
	params.
		SetMaxTokens(opts.MaxTokens).
		SetTemperature(opts.Temperature).
		SetTopP(opts.TopP).
		SetTopK(opts.TopK).
		SetSeed(opts.Seed)

	req := &tongyiclient.TextQwenRequest{}
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

func (q *LLM) doVLCompletionRequest(ctx context.Context, qwenVLMessages []tongyiclient.VLMessage, opts llms.CallOptions) (*tongyiclient.VLQwenResponse, error) {
	if len(qwenVLMessages) == 0 {
		return nil, ErrEmptyMessageContent
	}

	input := tongyiclient.VLInput{
		Messages: qwenVLMessages,
	}
	// return foo(ctx, input, opts)
	params := tongyiclient.DefaultQwenParameters()
	params.
		SetMaxTokens(opts.MaxTokens).
		SetTemperature(opts.Temperature).
		SetTopP(opts.TopP).
		SetTopK(opts.TopK).
		SetSeed(opts.Seed)

	req := &tongyiclient.VLQwenRequest{}
	req.
		SetModel(opts.Model).
		SetInput(input).
		SetParameters(params).
		SetStreamingFunc(opts.StreamingFunc)

	rsp, err := q.client.CreateVLCompletion(ctx, req)
	if err != nil {
		if q.CallbackHandler != nil {
			q.CallbackHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	return rsp, nil
}

func convertTongyiResultToContentChoice(resp IQwenResponseConverter) ([]*llms.ContentChoice, error) {
	return resp.convertToContentChoice(), nil
}

type IQwenResponseConverter interface {
	convertToContentChoice() []*llms.ContentChoice
}

var _ IQwenResponseConverter = (*TextRespConverter)(nil)
var _ IQwenResponseConverter = (*VLRespConverter)(nil)

type TextRespConverter tongyiclient.TextQwenResponse

func (t *TextRespConverter) convertToContentChoice() []*llms.ContentChoice {
	return convertToContentChoice(t.Output.Choices, t.Usage)
}

type VLRespConverter tongyiclient.VLQwenResponse

func (t *VLRespConverter) convertToContentChoice() []*llms.ContentChoice {
	vlresp := tongyiclient.VLQwenResponse(*t)
	return convertToContentChoice(vlresp.Output.Choices, vlresp.Usage)
}

func convertToContentChoice[T tongyiclient.ITongyiCntent](rawchoices []qwen.Choice[T], usage qwen.Usage) []*llms.ContentChoice {
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
