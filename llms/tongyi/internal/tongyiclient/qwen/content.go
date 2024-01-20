package qwen

type IQwenContent interface {
	*TextContent | *VLContentList
	IQwenContentMethods
}

type IQwenContentMethods interface {
	ToBytes() []byte
	ToString() string
	SetText(text string)
	AppendText(text string)

	// TODO: 临时解决方案，后续需要重新设计
	TargetURL() string
}

type TextContent string

func NewTextContent() *TextContent {
	t := TextContent("")
	return &t
}

func (t *TextContent) ToBytes() []byte {
	str := *t
	return []byte(str)
}

func (t *TextContent) ToString() string {
	str := *t
	return string(str)
}

func (t *TextContent) SetText(text string) {
	*t = TextContent(text)
}

func (t *TextContent) AppendText(text string) {
	str := *t
	*t = TextContent(string(str) + text)
}

func (t *TextContent) TargetURL() string {
	return URLQwen()
}

type VLContent struct {
	Image string `json:"image,omitempty"`
	Text  string `json:"text,omitempty"`
}

type VLContentList []VLContent

func NewVLContentList() *VLContentList {
	vl := VLContentList(make([]VLContent, 0))
	return &vl
}

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
	if vlist == nil {
		panic("VLContentList is nil")
	}
	*vlist = append(*vlist, VLContent{Text: s})
}

func (vlist *VLContentList) SetImage(url string) {
	if vlist == nil {
		panic("VLContentList is nil or empty")
	}
	*vlist = append(*vlist, VLContent{Image: url})
}

func (vlist *VLContentList) AppendText(s string) {
	if vlist == nil || len(*vlist) == 0 {
		panic("VLContentList is nil or empty")
	}
	(*vlist)[0].Text += s
}

func (vlist *VLContentList) TargetURL() string {
	return URLQwenVL()
}
