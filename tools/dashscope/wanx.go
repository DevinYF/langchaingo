package dashscope

import (
	"context"
	"os"
	"strings"

	"github.com/devinyf/dashscopego"
	"github.com/devinyf/dashscopego/wanx"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
)

type TongyiWanx struct {
	CallbacksHandler callbacks.Handler
	client           *dashscopego.TongyiClient
	description      string
	separator        string
}

const (
	_descriptionDefault = "A wrapper around Dashscope Wanx API. Useful for when you need to generate images from a text description. Input should be the textual description."
	// TongyiWanxDescriptionCn is the Chinese description for TongyiWanx.
	WanxDescriptionCn = "Dashscope TongyiWanx API 的封装, 当你需要根据文本描述生成图像时使用。输入应该是文本描述。" //nolint:gosmopolitan
)

type options struct {
	description string
}

type Option func(*options)

func WithDescription(description string) func(*options) {
	return func(o *options) {
		o.description = description
	}
}

var _ tools.Tool = TongyiWanx{}

func NewTongyiWanx(opts ...Option) *TongyiWanx {
	o := options{
		description: _descriptionDefault,
	}

	for _, opt := range opts {
		opt(&o)
	}

	model := wanx.WanxV1
	token := os.Getenv(dashscopego.DashscopeTokenEnvName)
	cli := dashscopego.NewTongyiClient(model, token)
	toolWanx := &TongyiWanx{
		client:      cli,
		separator:   "\n",
		description: o.description,
	}

	return toolWanx
}

func (t TongyiWanx) Name() string {
	return "TongyiWanx-Image-Generator"
}

func (t TongyiWanx) Description() string {
	return t.description
}

func (t TongyiWanx) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	model := wanx.WanxV1

	req := &wanx.ImageSynthesisRequest{
		Model: model,
		Input: wanx.ImageSynthesisInput{
			Prompt: input,
		},
		Params: wanx.ImageSynthesisParams{
			// TODO: current only generate one image.
			N: 1,
		},
	}

	imbBlob, err := t.client.CreateImageGeneration(ctx, req)
	if err != nil {
		if t.CallbacksHandler != nil {
			t.CallbacksHandler.HandleToolError(ctx, err)
		}
		return "", err
	}

	imgURLList := []string{}

	for _, img := range imbBlob {
		imgURLList = append(imgURLList, img.ImgURL)
	}

	urls := strings.Join(imgURLList, t.separator)

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, urls)
	}

	return urls, nil
}
