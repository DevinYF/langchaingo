package tongyiclient

import (
	qwen "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/qwen"
)

func DefaultQwenParameters() *qwen.Parameters {

	return qwen.DefaultParameters()
}
