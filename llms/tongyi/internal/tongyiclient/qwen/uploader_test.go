package qwen

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func getApiKey(t *testing.T) string {
	apikey := os.Getenv("DASHSCOPE_API_KEY")
	if apikey == "" {
		t.Skip("token is empty")
	}

	return apikey
}

func TestGetUploadCertificate(t *testing.T) {
	t.Parallel()
	apiKey := getApiKey(t)
	ctx := context.TODO()

	resp, err := getUploadCertificate(ctx, "qwen-vl-plus", apiKey)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

// 	//  this local file is not exist for other user
// func TestUploadingLocalImg(t *testing.T) {
// 	t.Parallel()
// 	ctx := context.TODO()

// 	homePath := os.Getenv("HOME")
// 	ossFilePath, err := UploadLocalImg(ctx, homePath+"/Downloads/dog_and_girl.jpeg", "qwen-vl-plus", os.Getenv("DASHSCOPE_API_KEY"))

// 	fmt.Println(ossFilePath)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, ossFilePath)
// }

func TestUploadingImageFromURL(t *testing.T) {
	t.Parallel()
	apiKey := getApiKey(t)

	// network problem...
	// var testImgURL = "https://github.com/tmc/langchaingo/blob/main/docs/static/img/parrot-icon.png"
	var testImgURL = "https://pic.ntimg.cn/20140113/8800276_184351657000_2.jpg"

	ctx := context.TODO()

	ossFilePath, err := UploadImgFromURL(ctx, testImgURL, "qwen-vl-plus", apiKey)

	// fmt.Println(ossFilePath)
	require.NoError(t, err)
	require.NotEmpty(t, ossFilePath)
}
