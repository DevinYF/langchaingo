package tongyi

import (
	"log"
	"net/url"
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

// TODO: This configuration has not taken effect.
func WithDashscopeURL(rawURL string) Option {
	return func(opts *options) {
		var err error
		opts.dashscopeURL, err = url.Parse(rawURL)
		if err != nil {
			log.Fatal(err)
		}
	}
}
