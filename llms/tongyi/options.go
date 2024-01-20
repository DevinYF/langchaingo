package tongyi

import (
	"log"
	"net/url"
)

const (
	dashscopeTokenEnvName = "DASHSCOPE_API_KEY" //nolint:gosec
)

type options struct {
	token        string
	dashscopeURL *url.URL
	model        string
}

type Option func(*options)

func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

func WithDashscopeURL(rawURL string) Option {
	return func(opts *options) {
		var err error
		opts.dashscopeURL, err = url.Parse(rawURL)
		if err != nil {
			log.Fatal(err)
		}
	}
}
