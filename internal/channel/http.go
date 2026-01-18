package channel

import (
	"encoding/json"
	"fmt"

	"github.com/yoyo3287258/home-gateway/internal/model"
)

// HTTPParser 通用HTTP请求解析器
type HTTPParser struct{}

func (p *HTTPParser) Name() string {
	return "http"
}

// HTTPRequest 通用HTTP请求结构
type HTTPRequest struct {
	Content string                 `json:"content"`
	UserID  string                 `json:"user_id"`
	RawData map[string]interface{} `json:"raw_data"`
}

func (p *HTTPParser) Parse(rawData []byte) (*model.UnifiedMessage, error) {
	var req HTTPRequest
	if err := json.Unmarshal(rawData, &req); err != nil {
		return nil, fmt.Errorf("解析HTTP请求失败: %w", err)
	}

	if req.Content == "" {
		return nil, fmt.Errorf("消息内容不能为空")
	}

	// 如果没有指定UserID，默认使用 anonymous
	if req.UserID == "" {
		req.UserID = "anonymous"
	}

	return model.NewUnifiedMessage(req.Content, model.ChannelHTTP, req.UserID, "", req.RawData), nil
}
