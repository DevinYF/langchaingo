package tongyi

import (
	"context"

	tongyiclient "github.com/devinyf/dashscopego"
	"github.com/devinyf/dashscopego/embedding"
	"github.com/devinyf/dashscopego/qwen"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

type LLM struct {
	CallbackHandler callbacks.Handler
	client          *tongyiclient.TongyiClient
	options         options
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
	return llms.GenerateFromSinglePrompt(ctx, q, prompt, options...)
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

	choices, err := q.doCompletionRequest(ctx, messages, opts)
	if err != nil {
		return nil, err
	}

	return &llms.ContentResponse{Choices: choices}, nil
}

func (q *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	input := struct {
		Texts []string `json:"texts"`
	}{
		Texts: inputTexts,
	}
	embeddings, _, err := q.client.CreateEmbedding(ctx,
		&embedding.Request{
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

func (q *LLM) doCompletionRequest(
	ctx context.Context,
	messages []llms.MessageContent,
	opts llms.CallOptions,
) ([]*llms.ContentChoice, error) {
	var converter IQwenResponseConverter

	switch opts.Model {
	// pure text-gen model.
	case "qwen-turbo", "qwen-plus", "qwen-max", "qwen-max-1201", "qwen-max-longcontext":
		textMessages := messagesCntentToQwenTextMessages(messages)

		rsp, err := q.doTextCompletionRequest(ctx, textMessages, opts)
		if err != nil {
			return nil, err
		}

		temp := TextRespConverter(*rsp)
		converter = &temp
	// multi-modal: text / image.
	case "qwen-vl-plus":
		vlMessages := messageConventToQwenVLMessage(messages)

		rsp, err := q.doVLCompletionRequest(ctx, vlMessages, opts)
		if err != nil {
			return nil, err
		}
		temp := VLRespConverter(*rsp)
		converter = &temp
	// image generation
	case "wanx-v1":
		panic("do Completion Error: not implemented wanx-v1 yet")
	default:
		panic("do Completion Error: not support model")
	}

	return convertTongyiResultToContentChoice(converter)
}

//nolint:lll
func (q *LLM) doTextCompletionRequest(ctx context.Context, qwenTextMessages []tongyiclient.TextMessage, opts llms.CallOptions) (*tongyiclient.TextQwenResponse, error) {
	req, err := genericRequest(qwenTextMessages, opts)
	if err != nil {
		return nil, err
	}

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

//nolint:lll
func (q *LLM) doVLCompletionRequest(ctx context.Context, qwenVLMessages []tongyiclient.VLMessage, opts llms.CallOptions) (*tongyiclient.VLQwenResponse, error) {
	req, err := genericRequest(qwenVLMessages, opts)
	if err != nil {
		return nil, err
	}

	rsp, err := q.client.CreateVLCompletion(ctx, req)
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

// nolint:lll
func genericRequest[T qwen.IQwenContent](qwenMessages []qwen.Message[T], opts llms.CallOptions) (*qwen.Request[T], error) {
	if len(qwenMessages) == 0 {
		return nil, ErrEmptyMessageContent
	}

	input := qwen.Input[T]{
		Messages: qwenMessages,
	}

	params := qwen.DefaultParameters()
	params.
		SetMaxTokens(opts.MaxTokens).
		SetTemperature(opts.Temperature).
		SetTopP(opts.TopP).
		SetTopK(opts.TopK).
		SetSeed(opts.Seed)

	req := &qwen.Request[T]{}
	req.
		SetModel(opts.Model).
		SetInput(input).
		SetParameters(params).
		SetStreamingFunc(opts.StreamingFunc)

	return req, nil
}
