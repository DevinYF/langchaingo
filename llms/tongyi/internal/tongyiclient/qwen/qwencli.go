package qwenclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	httpclient "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/httpclient"
)

func AsyncParseStreamingChatResponse(ctx context.Context, payload *QwenRequest, cli httpclient.IHttpClient, token string) (*QwenOutputMessage, error) {
	if payload.Model == "" {
		return nil, ErrModelNotSet
	}
	responseChan := asyncChatStreaming(ctx, payload, cli, token)
	outputMessage := QwenOutputMessage{}
	for rspData := range responseChan {
		if rspData.Err != nil {
			return nil, &httpclient.HTTPRequestError{Message: "parseStreamingChatResponse failed", Cause: rspData.Err}
		}
		if len(rspData.Output.Output.Choices) == 0 {
			return nil, ErrEmptyResponse
		}

		chunk := []byte(rspData.Output.Output.Choices[0].Message.Content)

		if payload.StreamingFunc != nil {
			err := payload.StreamingFunc(ctx, chunk)
			if err != nil {
				return nil, &httpclient.HTTPRequestError{Message: "parseStreamingChatResponse failed", Cause: err}
			}
		}

		outputMessage.RequestID = rspData.Output.RequestID
		outputMessage.Usage = rspData.Output.Usage
		if outputMessage.Output.Choices == nil {
			outputMessage.Output.Choices = rspData.Output.Output.Choices
		} else {
			outputMessage.Output.Choices[0].Message.Role = rspData.Output.Output.Choices[0].Message.Role
			outputMessage.Output.Choices[0].Message.Content += rspData.Output.Output.Choices[0].Message.Content
			outputMessage.Output.Choices[0].FinishReason = rspData.Output.Output.Choices[0].FinishReason
		}
	}

	return &outputMessage, nil
}

func SyncCall(ctx context.Context, payload *QwenRequest, cli httpclient.IHttpClient, token string) (*QwenOutputMessage, error) {
	if payload.Model == "" {
		return nil, ErrModelNotSet
	}

	resp := QwenOutputMessage{}
	tokenOpt := httpclient.WithTokenHeaderOption(token)
	err := cli.Post(ctx, QwenURL(), payload, &resp, tokenOpt)
	if err != nil {
		return nil, err
	}
	if len(resp.Output.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	return &resp, nil
}

func asyncChatStreaming(ctx context.Context, r *QwenRequest, cli httpclient.IHttpClient, token string) <-chan QwenResponse {
	chanBuffer := 100
	_respChunkChannel := make(chan QwenResponse, chanBuffer)

	go func() {
		withHeader := map[string]string{
			"Accept": "text/event-stream",
		}

		_combineStreamingChunk(ctx, r, withHeader, _respChunkChannel, cli, token)
	}()
	return _respChunkChannel
}

/*
 * combine SSE streaming lines to be a structed response data
 * id: xxxx
 * event: xxxxx
 * ......
 */
func _combineStreamingChunk(
	ctx context.Context,
	reqBody *QwenRequest,
	withHeader map[string]string,
	_respChunkChannel chan QwenResponse,
	cli httpclient.IHttpClient,
	token string,
) {
	defer close(_respChunkChannel)
	var _rawStreamOutChannel chan string

	var err error
	headerOpt := httpclient.WithHeader(withHeader)
	tokenOpt := httpclient.WithTokenHeaderOption(token)

	_rawStreamOutChannel, err = cli.PostSSE(ctx, QwenURL(), reqBody, headerOpt, tokenOpt)
	if err != nil {
		_respChunkChannel <- QwenResponse{Err: err}
		return
	}

	rsp := QwenResponse{}

	for v := range _rawStreamOutChannel {
		if strings.TrimSpace(v) == "" {
			// streaming out combined response
			_respChunkChannel <- rsp
			rsp = QwenResponse{}
			continue
		}

		err = fillInRespData(v, &rsp)
		if err != nil {
			rsp.Err = err
			_respChunkChannel <- rsp
			break
		}
	}
}

// filled in response data line by line.
func fillInRespData(line string, output *QwenResponse) error {
	if strings.TrimSpace(line) == "" {
		return nil
	}

	switch {
	case strings.HasPrefix(line, "id:"):
		output.ID = strings.TrimPrefix(line, "id:")
	case strings.HasPrefix(line, "event:"):
		output.Event = strings.TrimPrefix(line, "event:")
	case strings.HasPrefix(line, ":HTTP_STATUS/"):
		code, err := strconv.Atoi(strings.TrimPrefix(line, ":HTTP_STATUS/"))
		if err != nil {
			output.Err = fmt.Errorf("http_status err: strconv.Atoi  %w", err)
		}
		output.HTTPStatus = code
	case strings.HasPrefix(line, "data:"):
		dataJSON := strings.TrimPrefix(line, "data:")
		if output.Event == "error" {
			output.Err = &WrapMessageError{Message: dataJSON}
			return nil
		}
		outputData := QwenOutputMessage{}
		err := json.Unmarshal([]byte(dataJSON), &outputData)
		if err != nil {
			return &WrapMessageError{Message: "unmarshal OutputData Err", Cause: err}
		}

		output.Output = outputData
	default:
		data := bytes.TrimSpace([]byte(line))
		log.Printf("unknown line: %s", data)
	}

	return nil
}
