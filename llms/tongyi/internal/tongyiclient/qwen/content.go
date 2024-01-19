package qwenclient

type IQwenContent interface {
	*TextContent | *VLContentList
	ToBytes() []byte
	ToString() string
	SetText(string)
	AppendText(string)

	// TODO: 临时解决方案，后续需要重新设计
	TargetURL() string
}

type TextContent string

func (t *TextContent) ToBytes() []byte {
	str := *t
	return []byte(str)
}

func (t *TextContent) ToString() string {
	str := *t
	return string(str)
}

func (t *TextContent) SetText(s string) {
	*t = TextContent(s)
}

func (t *TextContent) AppendText(s string) {
	str := *t
	*t = TextContent(string(str) + s)
}

func (t *TextContent) TargetURL() string {
	return QwenURL()
}

type VLContent struct {
	Image string `json:"image,omitempty"`
	Text  string `json:"text,omitempty"`
}

type VLContentList []VLContent

func (vlist *VLContentList) ToBytes() []byte {
	if vlist == nil || len(*vlist) == 0 {
		return []byte("")
	}
	// TODO: handle multiple items later
	return []byte((*vlist)[0].Text)
}

func (vlist *VLContentList) ToString() string {
	if vlist == nil || len(*vlist) == 0 {
		return ""
	}
	// TODO: handle multiple items later
	return (*vlist)[0].Text
}

func (vlist *VLContentList) SetText(s string) {
	if vlist == nil || len(*vlist) == 0 {
		panic("VLContentList is nil or empty")
	}
	(*vlist)[0].Text = s
}
func (vlist *VLContentList) AppendText(s string) {
	if vlist == nil || len(*vlist) == 0 {
		panic("VLContentList is nil or empty")
	}
	(*vlist)[0].Text = (*vlist)[0].Text + s
}

func (vlist *VLContentList) TargetURL() string {
	return QwenVLURL()
}
