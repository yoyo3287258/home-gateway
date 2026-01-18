package channel

import "github.com/yoyo3287258/home-gateway/internal/model"

// Parser 消息解析器接口
type Parser interface {
	// Name 返回解析器名称（对应MessageChannel）
	Name() string

	// Parse 解析原始数据为统一消息格式
	Parse(rawData []byte) (*model.UnifiedMessage, error)
}
