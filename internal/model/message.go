package model

// MessageChannel 消息来源渠道
type MessageChannel string

const (
	ChannelHTTP        MessageChannel = "http"
	ChannelTelegram    MessageChannel = "telegram"
	ChannelWeChatWork  MessageChannel = "wechat_work"
	ChannelWebSocket   MessageChannel = "websocket"
)

// UnifiedMessage 统一消息格式
type UnifiedMessage struct {
	// Content 原始内容（用户输入的文本）
	Content string `json:"content"`

	// Channel 来源渠道
	Channel MessageChannel `json:"channel"`

	// UserID 用户ID（在渠道中的唯一标识）
	UserID string `json:"user_id"`

	// ChatID 会话ID（如Telegram的ChatID）
	ChatID string `json:"chat_id"`

	// RawData 原始数据（可选，保存特定渠道的原始payload）
	RawData map[string]interface{} `json:"raw_data,omitempty"`
}

// NewUnifiedMessage 创建新的统一消息
func NewUnifiedMessage(content string, channel MessageChannel, userID, chatID string, rawData map[string]interface{}) *UnifiedMessage {
	return &UnifiedMessage{
		Content: content,
		Channel: channel,
		UserID:  userID,
		ChatID:  chatID,
		RawData: rawData,
	}
}
