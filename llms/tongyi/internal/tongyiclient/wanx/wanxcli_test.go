package wanxclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	httpclient "github.com/tmc/langchaingo/llms/tongyi/internal/httpclient"
)

func newWanxClient(t *testing.T) *WanxClient {
	t.Helper()

	cli := NewWanxClient(WanxV1, httpclient.NewHTTPClient())
	if cli.token == "" {
		t.Skip("token is empty")
	}
	return cli
}

func TestCreateImageGeneration(t *testing.T) {
	t.Parallel()

	wanxCLi := newWanxClient(t)

	req := &WanxImageSynthesisRequest{
		Model: string(WanxV1),
		Input: WanxImageSynthesisInput{
			Prompt: "Draw a rabbit",
		},
		Params: WanxImageSynthesisParams{
			N: 1,
		},
	}

	blobList, err := wanxCLi.CreateImageGeneration(context.TODO(), req)
	require.NoError(t, err)
	require.NotEmpty(t, blobList)

	for _, blob := range blobList {
		require.NotEmpty(t, blob.Data)
		assert.Equal(t, "image/png", blob.ImgType)
	}
}
