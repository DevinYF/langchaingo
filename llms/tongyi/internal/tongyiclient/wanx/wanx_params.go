package wanx

import "fmt"

const (
	DashScopeBaseURL   = "https://dashscope.aliyuncs.com"
	WanxImageSynthesis = "/api/v1/services/aigc/text2image/image-synthesis"
	WanxTask           = "/api/v1/tasks/%s"
)

type WanxModel string

const (
	WanxV1             WanxModel = "wanx-v1"
	WanxStyleRepaintV1 WanxModel = "wanx-style-repaint-v1"
	WanxBgGenV2        WanxModel = "wanx-background-generation-v2"
)

func WanxImageSynthesisURL() string {
	return DashScopeBaseURL + WanxImageSynthesis
}

func WanxTaskURL(taskID string) string {
	return DashScopeBaseURL + fmt.Sprintf(WanxTask, taskID)
}
