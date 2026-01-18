package channel

import (
	"time"

	"github.com/yoyo3287258/home-gateway/internal/model"
)

// Parser 娓犻亾娑堟伅瑙ｆ瀽鍣ㄦ帴鍙?
// 姣忎釜娓犻亾瀹炵幇姝ゆ帴鍙ｅ皢鍘熷娑堟伅杞崲涓虹粺涓€娑堟伅鏍煎紡
type Parser interface {
	// Name 杩斿洖娓犻亾鍚嶇О
	Name() string

	// Parse 瑙ｆ瀽鍘熷娑堟伅涓虹粺涓€鏍煎紡
	Parse(rawData []byte) (*model.UnifiedMessage, error)

	// Validate 楠岃瘉webhook璇锋眰鐨勭鍚嶏紙濡傛灉闇€瑕侊級
	Validate(headers map[string]string, body []byte) bool
}

// Registry 娓犻亾瑙ｆ瀽鍣ㄦ敞鍐岃〃
type Registry struct {
	parsers map[string]Parser
}

// NewRegistry 鍒涘缓瑙ｆ瀽鍣ㄦ敞鍐岃〃
func NewRegistry() *Registry {
	return &Registry{
		parsers: make(map[string]Parser),
	}
}

// Register 娉ㄥ唽瑙ｆ瀽鍣?
func (r *Registry) Register(parser Parser) {
	r.parsers[parser.Name()] = parser
}

// Get 鑾峰彇瑙ｆ瀽鍣?
func (r *Registry) Get(name string) Parser {
	return r.parsers[name]
}

// List 鍒楀嚭鎵€鏈夊凡娉ㄥ唽鐨勮В鏋愬櫒
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.parsers))
	for name := range r.parsers {
		names = append(names, name)
	}
	return names
}

// DefaultRegistry 榛樿鐨勮В鏋愬櫒娉ㄥ唽琛?
var DefaultRegistry = NewRegistry()

// init 鍒濆鍖栭粯璁よВ鏋愬櫒
func init() {
	DefaultRegistry.Register(&HTTPParser{})
	DefaultRegistry.Register(&TelegramParser{})
}

// BaseMessage 鍩虹娑堟伅缁撴瀯锛堜緵鍚勮В鏋愬櫒澶嶇敤锛?
func NewUnifiedMessage(text, channel, userID, chatID string, rawData map[string]interface{}) *model.UnifiedMessage {
	return &model.UnifiedMessage{
		Text:          text,
		Channel:       channel,
		ChannelUserID: userID,
		ChannelChatID: chatID,
		Timestamp:     time.Now(),
		RawData:       rawData,
	}
}
