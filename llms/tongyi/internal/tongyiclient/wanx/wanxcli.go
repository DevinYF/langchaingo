package wanxclient

import (
	"context"
	"errors"
	"time"

	httpclient "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/httpclient"
)

var (
	ErrEmptyResponse = errors.New("empty response")
	ErrEmptyTaskID   = errors.New("task id is empty")
	ErrTaskUnsuccess = errors.New("task is not success")
	ErrModelNotSet   = errors.New("model is not set")
)

//nolint:lll
func CreateImageGeneration(ctx context.Context, payload *WanxImageSynthesisRequest, httpcli httpclient.IHttpClient, token string) ([]*WanxImgBlob, error) {
	// fmt.Println("debug...token: ", token)
	tokenOpt := httpclient.WithTokenHeaderOption(token)
	resp, err := SyncCall(ctx, payload, httpcli, tokenOpt)
	if err != nil {
		return nil, err
	}

	blobList := make([]*WanxImgBlob, 0, len(resp.Results))
	for _, img := range resp.Results {
		imgByte, err := httpcli.GetImage(ctx, img.URL, tokenOpt)
		if err != nil {
			return nil, err
		}

		blobList = append(blobList, &WanxImgBlob{Data: imgByte, ImgType: "image/png"})
	}

	return blobList, nil
}

// tongyi-wanx-api only support AsyncCall, so we need to warp it to be Sync.
func SyncCall(ctx context.Context, req *WanxImageSynthesisRequest, httpcli httpclient.IHttpClient, options ...httpclient.HTTPOption) (*WanxOutput, error) {
	rsp, err := AsyncCall(ctx, req, httpcli, options...)
	if err != nil {
		return nil, err
	}

	currentTaskStatus := TaskStatus(rsp.Output.TaskStatus)

	taskID := rsp.Output.TaskID
	if taskID == "" {
		return nil, ErrEmptyTaskID
	}

	taskReq := WanxTaskRequest{TaskID: taskID}
	taskResp := &WanxTaskResponse{}

	for currentTaskStatus == TaskPending ||
		currentTaskStatus == TaskRunning ||
		currentTaskStatus == TaskSuspended {
		delayDurationToCheckStatus := 500
		time.Sleep(time.Duration(delayDurationToCheckStatus) * time.Millisecond)

		// log.Println("TaskStatus: ", currentTaskStatus)
		taskResp, err = CheckTaskStatus(ctx, &taskReq, httpcli, options...)
		if err != nil {
			return nil, err
		}
		currentTaskStatus = TaskStatus(taskResp.Output.TaskStatus)
	}

	if currentTaskStatus == TaskFailed ||
		currentTaskStatus == TaskCanceled {
		return nil, ErrTaskUnsuccess
	}

	if len(taskResp.Output.Results) == 0 {
		return nil, ErrEmptyResponse
	}

	return &taskResp.Output, nil
}

// calling tongyi-wanx-api to get image-generation async task id.
func AsyncCall(ctx context.Context, req *WanxImageSynthesisRequest, httpcli httpclient.IHttpClient, options ...httpclient.HTTPOption) (*WanxImageResponse, error) {
	header := map[string]string{"X-DashScope-Async": "enable"}
	headerOpt := httpclient.WithHeader(header)
	options = append(options, headerOpt)

	if req.Model == "" {
		return nil, ErrModelNotSet
	}

	resp := WanxImageResponse{}
	err := httpcli.Post(ctx, WanxImageSynthesisURL(), req, &resp, options...)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func CheckTaskStatus(ctx context.Context, req *WanxTaskRequest, httpcli httpclient.IHttpClient, options ...httpclient.HTTPOption) (*WanxTaskResponse, error) {
	resp := WanxTaskResponse{}
	err := httpcli.Get(ctx, WanxTaskURL(req.TaskID), &resp, options...)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
