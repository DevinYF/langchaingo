package tongyi

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

func newQwenLlm(t *testing.T) *LLM {
	t.Helper()
	dashscopeKey := os.Getenv(dashscopeTokenEnvName)
	if dashscopeKey == "" {
		t.Skip("DASHSCOPE_API_KEY not set")
		return nil
	}
	modelOption := WithModel("qwen-turbo")
	tokenOption := WithToken(dashscopeKey)

	llm, err := New(modelOption, tokenOption)
	require.NoError(t, err)
	return llm
}

func TestLLmBasic(t *testing.T) {
	t.Parallel()
	llm := newQwenLlm(t)

	ctx := context.TODO()

	resp, err := llm.Call(ctx, "greet me in English.")
	require.NoError(t, err)

	assert.Regexp(t, "hello|hi|how|moring|good|today|assist", strings.ToLower(resp)) //nolint:all
}

func TestLLmStream(t *testing.T) {
	t.Parallel()
	llm := newQwenLlm(t)

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
	llm := newQwenLlm(t)

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

	resp, err := llm.GenerateContent(ctx, mc, llms.WithStreamingFunc(
		func(ctx context.Context, chunk []byte) error {
			return nil
		},
	))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]

	c1.Content = strings.ToLower(c1.Content)

	assert.Regexp(t, "hello|hi|how|moring|good|today|assist", strings.ToLower(c1.Content))
}

func TestGenerateContentStream(t *testing.T) {
	t.Parallel()
	llm := newQwenLlm(t)

	ctx := context.TODO()
	var sb strings.Builder

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

	resp, err := llm.GenerateContent(ctx, mc, llms.WithStreamingFunc(
		func(ctx context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		},
	))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]

	c1.Content = strings.ToLower(c1.Content)

	assert.Regexp(t, "hello|hi|how|moring|good|today|assist", strings.ToLower(c1.Content))
	assert.Regexp(t, "hello|hi|how|moring|good|today|assist", strings.ToLower(sb.String()))
}

// func TestGenerateContentImsge(t *testing.T) {
// 	t.Parallel()
// 	// llm := newQwenLlm(t)
// 	// ctx := context.TODO()

// 	// userContent := llms.TextContent{
// 	// 	Text: "greet me in english.",
// 	// }
// }

func TestEMbedding(t *testing.T) {
	t.Parallel()
	llm := newQwenLlm(t)

	ctx := context.TODO()

	embeddingText := []string{"风急天高猿啸哀", "渚清沙白鸟飞回", "无边落木萧萧下", "不尽长江滚滚来"}

	resp, err := llm.CreateEmbedding(ctx, embeddingText)

	require.NoError(t, err)
	assert.NotEmpty(t, resp)
	assert.Len(t, resp, len(embeddingText))
}
