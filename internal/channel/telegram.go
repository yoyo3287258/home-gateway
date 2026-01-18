package channel

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/yoyo3287258/home-gateway/internal/model"
)

// TelegramParser Telegram消息解析器
type TelegramParser struct {
	// WebhookSecret 用于验证Webhook请求的密钥
	WebhookSecret string
}

// Name 返回渠道名称
func (p *TelegramParser) Name() string {
	return "telegram"
}

// TelegramUpdate Telegram Webhook更新消息
type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

// TelegramMessage Telegram消息
type TelegramMessage struct {
	MessageID int           `json:"message_id"`
	From      *TelegramUser `json:"from,omitempty"`
	Chat      *TelegramChat `json:"chat"`
	Date      int64         `json:"date"`
	Text      string        `json:"text,omitempty"`
}

// TelegramUser Telegram用户
type TelegramUser struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// TelegramChat Telegram会话
type TelegramChat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"` // private, group, supergroup, channel
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Parse 解析Telegram消息
func (p *TelegramParser) Parse(rawData []byte) (*model.UnifiedMessage, error) {
	var update TelegramUpdate
	if err := json.Unmarshal(rawData, &update); err != nil {
		return nil, fmt.Errorf("解析Telegram消息失败: %w", err)
	}

	if update.Message == nil {
		return nil, fmt.Errorf("不支持的Telegram更新类型（非消息）")
	}

	if update.Message.Text == "" {
		return nil, fmt.Errorf("空消息或非文本消息")
	}

	msg := update.Message
	
	// 提取用户ID
	var userID string
	if msg.From != nil {
		userID = strconv.FormatInt(msg.From.ID, 10)
	}

	// 提取会话ID
	chatID := strconv.FormatInt(msg.Chat.ID, 10)

	// 构建原始数据
	rawMap := map[string]interface{}{
		"update_id":  update.UpdateID,
		"message_id": msg.MessageID,
		"chat_type":  msg.Chat.Type,
	}
	if msg.From != nil {
		rawMap["from_username"] = msg.From.Username
		rawMap["from_name"] = msg.From.FirstName + " " + msg.From.LastName
	}

	return model.NewUnifiedMessage(msg.Text, model.ChannelTelegram, userID, chatID, rawMap), nil
}

// Validate 验证Telegram Webhook请求
// 使用 X-Telegram-Bot-Api-Secret-Token 头进行验证
func (p *TelegramParser) Validate(headers map[string]string, body []byte) bool {
	if p.WebhookSecret == "" {
		// 未配置密钥，跳过验证
		return true
	}

	// Telegram 使用 X-Telegram-Bot-Api-Secret-Token 头
	secretToken := headers["X-Telegram-Bot-Api-Secret-Token"]
	if secretToken == "" {
		secretToken = headers["x-telegram-bot-api-secret-token"]
	}

	return secretToken == p.WebhookSecret
}

// CalculateTelegramSecretTokenHash 计算Telegram密钥Hash（用于设置Webhook时）
func CalculateTelegramSecretTokenHash(botToken, data string) string {
	h := hmac.New(sha256.New, []byte(botToken))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
