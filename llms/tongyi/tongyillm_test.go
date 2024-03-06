package tongyi

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/devinyf/dashscopego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

const (
	QwenTextModel  = "qwen-turbo"
	QwenVLModel    = "qwen-vl-plus"
	EmbeddingModel = "text-embedding-v1"
)

func newQwenLlm(t *testing.T, model string) *LLM {
	t.Helper()
	dashscopeKey := os.Getenv(dashscopego.DashscopeTokenEnvName)
	if dashscopeKey == "" {
		t.Skip("DASHSCOPE_API_KEY not set")
		return nil
	}
	modelOption := WithModel(model)
	tokenOption := WithToken(dashscopeKey)

	llm, err := New(modelOption, tokenOption)
	require.NoError(t, err)
	return llm
}

func TestVLBasic(t *testing.T) {
	t.Parallel()
	llm := newQwenLlm(t, QwenVLModel)

	ctx := context.TODO()

	parts := []llms.ContentPart{
		llms.ImageURLContent{URL: "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg"},
		llms.TextContent{Text: "briefly describe this image."},
	}

	mc := []llms.MessageContent{
		{Role: schema.ChatMessageTypeHuman, Parts: parts},
	}

	resp, err := llm.GenerateContent(ctx, mc)
	require.NoError(t, err)

	assert.Regexp(t, "dog|person|individual|woman|girl", strings.ToLower(resp.Choices[0].Content))
}

func TestVLStreamChund(t *testing.T) {
	t.Parallel()

	// set model to empty here to test modelOption below
	llm := newQwenLlm(t, "")

	modelOpt := llms.WithModel(QwenVLModel)

	ctx := context.TODO()

	parts := []llms.ContentPart{
		llms.ImageURLContent{URL: "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg"},
		llms.TextContent{Text: "briefly describe this image."},
	}

	mc := []llms.MessageContent{
		{Role: schema.ChatMessageTypeHuman, Parts: parts},
	}

	var output strings.Builder
	streamCallbackFnOption := llms.WithStreamingFunc(
		func(ctx context.Context, chunk []byte) error {
			output.Write(chunk)
			return nil
		},
	)

	resp, err := llm.GenerateContent(ctx, mc, streamCallbackFnOption, modelOpt)
	require.NoError(t, err)
	assert.Equal(t, output.String(), resp.Choices[0].Content)

	assert.Regexp(t, "dog|person|individual|woman|girl", strings.ToLower(resp.Choices[0].Content))
}

func TestLLmBasic(t *testing.T) {
	t.Parallel()
	llm := newQwenLlm(t, QwenTextModel)

	ctx := context.TODO()

	resp, err := llm.Call(ctx, "greet me in English.")
	require.NoError(t, err)

	assert.Regexp(t, "hello|hi|how|moring|good|today|assist", strings.ToLower(resp)) //nolint:all
}

func TestLLmStream(t *testing.T) {
	t.Parallel()
	llm := newQwenLlm(t, QwenTextModel)

	ctx := context.TODO()
	var sb strings.Builder

	resp, err := llm.Call(ctx, "greet me in English.", llms.WithStreamingFunc(
		func(ctx context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		},
	))
	require.NoError(t, err)

	assert.Regexp(t, "hello|hi|how|moring|good|today|assist", strings.ToLower(resp))        //nolint:all
	assert.Regexp(t, "hello|hi|how|moring|good|today|assist", strings.ToLower(sb.String())) //nolint:all
}

func TestGenerateContentText(t *testing.T) {
	t.Parallel()
	llm := newQwenLlm(t, QwenTextModel)

	ctx := context.TODO()

	sysContent := llms.TextContent{
		Text: "You are a helpful Ai assistant.",
	}

	userContent := llms.TextContent{
		Text: "greet me in english.",
	}

	mc := []llms.MessageContent{
		{Role: schema.ChatMessageTypeSystem, Parts: []llms.ContentPart{sysContent}},
		{Role: schema.ChatMessageTypeHuman, Parts: []llms.ContentPart{userContent}},
	}

	resp, err := llm.GenerateContent(ctx, mc)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]

	assert.Regexp(t, "hello|hi|how|moring|good|today|assist", strings.ToLower(c1.Content))
}

func TestGenerateContentStream(t *testing.T) {
	t.Parallel()
	llm := newQwenLlm(t, QwenTextModel)

	ctx := context.TODO()

	sysContent := llms.TextContent{
		Text: "You are a helpful Ai assistant.",
	}

	userContent := llms.TextContent{
		Text: "greet me in english.",
	}

	mc := []llms.MessageContent{
		{Role: schema.ChatMessageTypeSystem, Parts: []llms.ContentPart{sysContent}},
		{Role: schema.ChatMessageTypeHuman, Parts: []llms.ContentPart{userContent}},
	}

	var output strings.Builder
	streamCallbackFn := llms.WithStreamingFunc(
		func(ctx context.Context, chunk []byte) error {
			output.Write(chunk)
			return nil
		},
	)

	resp, err := llm.GenerateContent(ctx, mc, streamCallbackFn)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]

	assert.Regexp(t, "hello|hi|how|moring|good|today|assist", strings.ToLower(c1.Content))
	assert.Regexp(t, "hello|hi|how|moring|good|today|assist", strings.ToLower(output.String()))
}

func TestEmbedding(t *testing.T) {
	t.Parallel()
	llm := newQwenLlm(t, EmbeddingModel)

	ctx := context.TODO()

	embeddingText := []string{"风急天高猿啸哀", "渚清沙白鸟飞回", "无边落木萧萧下", "不尽长江滚滚来"}

	resp, err := llm.CreateEmbedding(ctx, embeddingText)

	require.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.Len(t, resp, len(embeddingText))
}

// func TestGenerateContentImsge(t *testing.T) {
// 	t.Parallel()
// 	// llm := newQwenLlm(t)
// 	// ctx := context.TODO()

// 	// userContent := llms.TextContent{
// 	// 	Text: "greet me in english.",
// 	// }
// }
