package qwenclient

const WanxImageSynthesis = "/api/v1/services/aigc/text2image/image-synthesis"

const (
	WanxV1             = "wanx-v1"
	WanxStyleRepaintV1 = "wanx-style-repaint-v1"
	WanxBgGenV2        = "wanx-background-generation-v2"
)

func WanxImageSynthesisURL() string {
	return DashScopeBaseURL + WanxImageSynthesis
}
