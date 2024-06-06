package tongyi

import (
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingToken             = errors.New("missing the Dashscope API key, set it in the DASHSCOPE_API_KEY environment variable") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrIncompleteEmbedding      = errors.New("no all input got emmbedded")
	ErrNotSupportImsgePart      = errors.New("not support Image parts yet")
	ErrMultipleTextParts        = errors.New("found multiple text parts in message")
	ErrEmptyMessageContent      = errors.New("tongyi MessageContent is empty")
	ErrNotImplemented           = errors.New("not implemented yet")
)

type UnSupportedRoleError struct {
	Role llms.ChatMessageType
}

func (e *UnSupportedRoleError) Error() string {
	return fmt.Sprintf("langchain role %s not supported", e.Role)
}
