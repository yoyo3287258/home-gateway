package channel

import (
	"encoding/json"
	"fmt"

	"github.com/yoyo3287258/home-gateway/internal/model"
)

// HTTPParser 閫氱敤HTTP鎺ュ彛瑙ｆ瀽鍣?
// 鐢ㄤ簬鐩存帴閫氳繃HTTP API璋冪敤鐨勫満鏅?
type HTTPParser struct{}

// Name 杩斿洖娓犻亾鍚嶇О
func (p *HTTPParser) Name() string {
	return "http"
}

// HTTPRequest 閫氱敤HTTP璇锋眰鏍煎紡
type HTTPRequest struct {
	// Text 鐢ㄦ埛杈撳叆鐨勬枃鏈紙蹇呭～锛?
	Text string `json:"text"`

	// UserID 鐢ㄦ埛鏍囪瘑锛堝彲閫夛級
	UserID string `json:"user_id,omitempty"`

	// ChatID 浼氳瘽鏍囪瘑锛堝彲閫夛級
	ChatID string `json:"chat_id,omitempty"`

	// Extra 棰濆鏁版嵁锛堝彲閫夛級
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// Parse 瑙ｆ瀽HTTP璇锋眰
func (p *HTTPParser) Parse(rawData []byte) (*model.UnifiedMessage, error) {
	var req HTTPRequest
	if err := json.Unmarshal(rawData, &req); err != nil {
		return nil, fmt.Errorf("瑙ｆ瀽HTTP璇锋眰澶辫触: %w", err)
	}

	if req.Text == "" {
		return nil, fmt.Errorf("text瀛楁涓嶈兘涓虹┖")
	}

	rawMap := make(map[string]interface{})
	if req.Extra != nil {
		rawMap = req.Extra
	}

	return NewUnifiedMessage(req.Text, "http", req.UserID, req.ChatID, rawMap), nil
}

// Validate HTTP璇锋眰涓嶉渶瑕佺鍚嶉獙璇?
func (p *HTTPParser) Validate(headers map[string]string, body []byte) bool {
	return true
}
