package qwenclient

type WanxImageSynthesisParams struct {
	// The style of the output image, currently supports the following style values:
	// "<auto>" default,
	// "<3d cartoon>" 3D cartoon,
	// "<anime>" animation,
	// "<oil painting>" oil painting,
	// "<watercolor>" watercolor,
	// "<sketch>" sketch,
	// "<chinese painting>" Chinese painting,
	// "<flat illustration>" flat illustration,
	Style string `json:"style"`
	// The number of images generated, currently supports 1~4, default is 1.
	Size string `json:"size"`
	// The resolution of the generated image,
	// currently only supports '1024*1024', '720*1280', '1280*720' three resolutions,
	// default is 1024*1024 pixels.
	N int `json:"n"`
	// seed.
	Seed int `json:"seed"`
}

type WanxImageSynthesisInput struct {
	Prompt        string `json:"prompt"`
	NegativePromp string `json:"negative_promp"`
}

type WanxImageSynthesisRequest struct {
	Model  string                   `json:"model"`
	Input  WanxImageSynthesisInput  `json:"input"`
	Params WanxImageSynthesisParams `json:"parameters"`
}

type ImageData struct {
	Data []byte
	// ImgURL string
}

type WanxImageResponse struct {
	StatusCode int    `json:"status_code"`
	RequestID  string `json:"request_id"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Output     struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
		Results    []struct {
			URL string `json:"url"`
		} `json:"results"`
		TaskMetrics struct {
			Total     int `json:"TOTAL"`
			Succeeded int `json:"SUCCEEDED"`
			Failed    int `json:"FAILED"`
		} `json:"task_metrics"`
	} `json:"output"`
	Usage struct {
		ImageCount int `json:"image_count"`
	} `json:"usage"`
}
