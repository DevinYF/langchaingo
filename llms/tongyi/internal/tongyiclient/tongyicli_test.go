package tongyiclient

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	httpclient "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/httpclient"
	wanx "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/wanx"
	"go.uber.org/mock/gomock"
)

func newTongyiClient(t *testing.T, model string) *TongyiClient {
	t.Helper()
	token := os.Getenv("DASHSCOPE_API_KEY")

	cli := NewTongyiClient(model, token)
	if cli.token == "" {
		t.Skip("token is empty")
	}
	return cli
}

func newMockClient(t *testing.T, model string, ctrl *gomock.Controller, f mockFn) *TongyiClient {
	t.Helper()

	mockHTTPCli := httpclient.NewMockIHttpClient(ctrl)
	fackToken := ""

	f(mockHTTPCli)

	qwenCli := newTongyiCLientWithHttpCli(model, fackToken, mockHTTPCli)
	return qwenCli
}

type mockFn func(mockHTTPCli *httpclient.MockIHttpClient)

func TestBasic(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	cli := newTongyiClient(t, "")

	input := TextInput{
		Messages: []TextMessage{
			{Role: "user", Content: "Hello!"},
		},
	}

	req := &TextQwenRequest{
		Model: "qwen-turbo",
		Input: input,
	}

	resp, err := cli.CreateCompletion(ctx, req)

	require.NoError(t, err)
	assert.Regexp(t, "hello|hi|how|assist", resp.Output.Choices[0].Message.Content)
}

func TestStreamingChunk(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	cli := newTongyiClient(t, "qwen-turbo")

	output := ""

	input := TextInput{
		Messages: []TextMessage{
			{Role: "user", Content: "Hello!"},
		},
	}

	req := &TextQwenRequest{
		// Model: "qwen-turbo",
		Input: input,
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			output += string(chunk)
			return nil
		},
	}
	resp, err := cli.CreateCompletion(ctx, req)

	require.NoError(t, err)
	assert.Regexp(t, "hello|hi|how|assist", resp.Output.Choices[0].Message.Content)
	assert.Regexp(t, "hello|hi|how|assist", output)
}

func TestImageGeneration(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	cli := newTongyiClient(t, "wanx-v1")

	req := &wanx.WanxImageSynthesisRequest{
		Model: "wanx-v1",
		Input: wanx.WanxImageSynthesisInput{
			Prompt: "A beautiful sunset",
		},
	}

	imgBlobs, err := cli.CreateImageGeneration(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, imgBlobs)

	for _, blob := range imgBlobs {
		assert.NotEmpty(t, blob.Data)
		assert.Equal(t, "image/png", blob.ImgType)

	}

}

func TestMockStreamingChunk(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cli := newMockClient(t, "qwen-turbo", ctrl, _mockAsyncFunc)

	output := ""

	input := TextInput{
		Messages: []TextMessage{
			{Role: "user", Content: "Hello!"},
		},
	}

	req := &TextQwenRequest{
		Input: input,
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			output += string(chunk)
			return nil
		},
	}
	resp, err := cli.CreateCompletion(ctx, req)

	require.NoError(t, err)

	assert.Equal(t, "Hello! How can I assist you today?", resp.Output.Choices[0].Message.Content)
	assert.Equal(t, "Hello! How can I assist you today?", output)
}

func TestMockBasic(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cli := newMockClient(t, "qwen-turbo", ctrl, _mockSyncFunc)
	input := TextInput{
		Messages: []TextMessage{
			{Role: "user", Content: "Hello!"},
		},
	}

	req := &TextQwenRequest{
		Input: input,
	}

	resp, err := cli.CreateCompletion(ctx, req)

	require.NoError(t, err)

	assert.Equal(t, "Hello! This is a mock message.", resp.Output.Choices[0].Message.Content)
	assert.Equal(t, "mock-ac55-9fd3-8326-8415cbdf5683", resp.RequestID)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
}

func _mockAsyncFunc(mockHTTPCli *httpclient.MockIHttpClient) {
	MockStreamData := []string{
		`id:1`,
		`event:result`,
		`:HTTP_STATUS/200`,
		`data:{
			"output": {
				"choices": [{
					"message": {
						"content": "Hello! How",
						"role": "assistant"
					},
					"finish_reason": "null"
				}]
			},
			"usage": {
				"total_tokens": 9,
				"input_tokens": 6,
				"output_tokens": 3
			},
			"request_id": "95bea986-ac55-9fd3-8326-8415cbdf5683"
		}`,
		`    `,
		`id:2`,
		`event:result`,
		`:HTTP_STATUS/200`,
		`data:{
			"output": {
				"choices": [{
					"message": {
						"content": " can I assist you today?",
						"role": "assistant"
					},
					"finish_reason": "null"
				}]
			},
			"usage": {
				"total_tokens": 15,
				"input_tokens": 6,
				"output_tokens": 9
			},
			"request_id": "95bea986-ac55-9fd3-8326-8415cbdf5683"
		}`,
		`    `,
		`id:3`,
		`event:result`,
		`:HTTP_STATUS/200`,
		`data:{
			"output": {
				"choices": [{
					"message": {
						"content": "",
						"role": "assistant"
					},
					"finish_reason": "stop"
				}]
			},
			"usage": {
				"total_tokens": 15,
				"input_tokens": 6,
				"output_tokens": 9
			},
			"request_id": "95bea986-ac55-9fd3-8326-8415cbdf5683"
		}`,
	}

	ctx := context.TODO()

	_rawStreamOutChannel := make(chan string, 500)

	mockHTTPCli.EXPECT().
		PostSSE(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(_rawStreamOutChannel, nil)
	go func() {
		for _, line := range MockStreamData {
			_rawStreamOutChannel <- line
		}
		close(_rawStreamOutChannel)
	}()
}

func _mockSyncFunc(mockHTTPCli *httpclient.MockIHttpClient) {
	ctx := context.TODO()

	mockResp := TextQwenOutputMessage{
		Output: TextQwenOutput{
			Choices: []struct {
				Message      TextMessage `json:"message"`
				FinishReason string      `json:"finish_reason"`
			}{
				{
					Message: TextMessage{
						Content: "Hello! This is a mock message.",
						Role:    "assistant",
					},
					FinishReason: "stop",
				},
			},
		},
		Usage: struct {
			TotalTokens  int `json:"total_tokens"`
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{
			TotalTokens:  15,
			InputTokens:  6,
			OutputTokens: 9,
		},
		RequestID: "mock-ac55-9fd3-8326-8415cbdf5683",
	}
	mockHTTPCli.EXPECT().
		Post(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		SetArg(3, mockResp).
		Return(nil)
}
