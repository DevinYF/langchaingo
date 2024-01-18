package wanxclient

import (
	"context"
	"errors"
	"os"
	"time"

	httpclient "github.com/tmc/langchaingo/llms/tongyi/internal/httpclient"
)

var (
	ErrEmptyResponse = errors.New("empty response")
	ErrEmptyTaskID   = errors.New("task id is empty")
	ErrTaskUnsuccess = errors.New("task is not success")
)

type WanxClient struct {
	Model           WanxModel
	imageGenURL     string
	taskStatusURLFn func(string) string
	token           string
	httpCli         httpclient.IHttpClient
}

func NewWanxClient(model WanxModel, httpCli httpclient.IHttpClient) *WanxClient {
	return &WanxClient{
		Model:           model,
		imageGenURL:     WanxImageSynthesisURL(),
		taskStatusURLFn: WanxTaskURL,
		token:           os.Getenv("DASHSCOPE_API_KEY"),
		httpCli:         httpCli,
	}
}

//nolint:lll
func (w *WanxClient) CreateImageGeneration(ctx context.Context, payload *WanxImageSynthesisRequest) ([]*WanxImgBlob, error) {
	resp, err := w.SyncCall(ctx, payload)
	if err != nil {
		return nil, err
	}

	blobList := make([]*WanxImgBlob, 0, len(resp.Results))
	for _, img := range resp.Results {
		imgByte, err := w.httpCli.GetImage(ctx, img.URL, w.TokenHeaderOption())
		if err != nil {
			return nil, err
		}

		blobList = append(blobList, &WanxImgBlob{Data: imgByte, ImgType: "image/png"})
	}

	return blobList, nil
}

// tongyi-wanx-api only support AsyncCall, so we need to warp it to be Sync.
func (w *WanxClient) SyncCall(ctx context.Context, req *WanxImageSynthesisRequest) (*WanxOutput, error) {
	rsp, err := w.AsyncCall(ctx, req)
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
		taskResp, err = w.CheckTaskStatus(ctx, &taskReq)
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
func (w *WanxClient) AsyncCall(ctx context.Context, req *WanxImageSynthesisRequest) (*WanxImageResponse, error) {
	header := map[string]string{"X-DashScope-Async": "enable"}
	headerOpt := httpclient.WithHeader(header)

	if req.Model == "" {
		req.Model = string(w.Model)
	}

	resp := WanxImageResponse{}
	err := w.httpCli.Post(ctx, w.imageGenURL, req, &resp, w.TokenHeaderOption(), headerOpt)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (w *WanxClient) CheckTaskStatus(ctx context.Context, req *WanxTaskRequest) (*WanxTaskResponse, error) {
	resp := WanxTaskResponse{}
	err := w.httpCli.Get(ctx, w.taskStatusURLFn(req.TaskID), &resp, w.TokenHeaderOption())
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (w *WanxClient) TokenHeaderOption() httpclient.HTTPOption {
	m := map[string]string{"Authorization": "Bearer " + w.token}
	return httpclient.WithHeader(m)
}
