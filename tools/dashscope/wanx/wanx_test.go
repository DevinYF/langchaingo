package wanx

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/devinyf/dashscopego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/tongyi"
	"github.com/tmc/langchaingo/tools"
)

const (
	QwenTextModel  = "qwen-turbo"
	QwenVLModel    = "qwen-vl-plus"
	EmbeddingModel = "text-embedding-v1"
)

func newQwenLlm(t *testing.T, model string) *tongyi.LLM {
	t.Helper()
	dashscopeKey := os.Getenv(dashscopego.DashscopeTokenEnvName)
	if dashscopeKey == "" {
		t.Skip("DASHSCOPE_API_KEY not set")
		return nil
	}
	modelOption := tongyi.WithModel(model)
	tokenOption := tongyi.WithToken(dashscopeKey)
	llm, err := tongyi.New(modelOption, tokenOption)
	require.NoError(t, err)
	return llm
}

func TestGenerateImsgeByAgent(t *testing.T) {
	t.Parallel()
	// also support other llm model.
	llm := newQwenLlm(t, QwenTextModel)

	wanxDescOpt := WithDescription(WanxDescriptionCN) // 切换中文描述 默认是英文
	wanxTool := NewTongyiWanx(wanxDescOpt)

	agentTools := []tools.Tool{
		wanxTool,
	}

	callbackHandler := callbacks.NewFinalStreamHandler()

	agent := agents.NewOneShotAgent(
		llm,
		agentTools,
		agents.WithCallbacksHandler(callbackHandler))
	executor := agents.NewExecutor(agent)

	// set streaming result.
	var output strings.Builder
	outputFn := func(_ context.Context, chunk []byte) {
		output.Write(chunk)
	}
	// get streaming output.
	callbackHandler.ReadFromEgress(context.Background(), outputFn)

	question := "draw a girl and a dog"
	answer, err := chains.Run(context.Background(), executor, question)
	if err != nil {
		panic(err)
	}

	assert.Regexp(t, "https:", answer)
	assert.Regexp(t, "https:", output.String())
}
