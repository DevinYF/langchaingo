package dashscope

import (
	"context"

	"github.com/tmc/langchaingo/callbacks"
	wanx "github.com/devinyf/dashscopego/wanx"
)

type TongyiWanx struct {
	CallbacksHandler callbacks.Handler
	model            wanx.ModelWanx
}

func NewTongyiWanx() *TongyiWanx {
	return &TongyiWanx{}
}

func (t *TongyiWanx) Name() string {
	return "TongyiWanx"
}

func (t *TongyiWanx) Description() string {
	return `
	Useful for when you need to generate images from text,
	Input should be the textual description that you want to put into a picture.`
}

func (t *TongyiWanx) Call(ctx context.Context, input string) ([]byte, error) {
	panic("not implemented")
}
