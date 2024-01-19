package qwen

import (
	"log"
)

const (
	DashScopeBaseURL = "https://dashscope.aliyuncs.com"
	QwenSubURL       = "/api/v1/services/aigc/text-generation/generation"
	QwenVLSubURL     = "/api/v1/services/aigc/multimodal-generation/generation"
)

type QwenModel string

const (
	QwenTurbo          QwenModel = "qwen-turbo"
	QwenPlus           QwenModel = "qwen-plus"
	QwenMax            QwenModel = "qwen-max"
	QwenMax1201        QwenModel = "qwen-max-1201"
	QwenMaxLongContext QwenModel = "qwen-max-longcontext"
)

type Model struct{}

// text-generation only.
func QwenURL() string {
	return DashScopeBaseURL + QwenSubURL
}

// multimodal.
func QwenVLURL() string {
	return DashScopeBaseURL + QwenVLSubURL
}

func ChoseQwenModel(model string) QwenModel {
	m := Model{}
	switch model {
	case "qwen-turbo":
		return m.QwenTurbo()
	case "qwen-plus":
		return m.QwenPlus()
	case "qwen-max":
		return m.QwenMax()
	case "qwen-max-1201":
		return m.QwenMax1201()
	case "qwen-max-longcontext":
		return m.QwenMaxLongContext()
	default:
		log.Println("target model not found, use default model: qwen-turbo")
		return m.QwenTurbo()
	}
}

func (m *Model) QwenTurbo() QwenModel {
	return QwenTurbo
}

func (m *Model) QwenPlus() QwenModel {
	return QwenPlus
}

func (m *Model) QwenMax() QwenModel {
	return QwenMax
}

func (m *Model) QwenMax1201() QwenModel {
	return QwenMax1201
}

func (m *Model) QwenMaxLongContext() QwenModel {
	return QwenMaxLongContext
}
