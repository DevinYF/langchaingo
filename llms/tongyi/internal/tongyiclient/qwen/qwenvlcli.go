package qwenclient



type VLContent struct {
	Text string `json:"text"`
	Image string `json:"image"`
}

type MessageVL struct {
	Role    string `json:"role"`
	Content []VLContent `json:"content"`
}