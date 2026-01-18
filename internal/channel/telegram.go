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

// TelegramParser Telegram娑堟伅瑙ｆ瀽鍣?
type TelegramParser struct {
	// WebhookSecret 鐢ㄤ簬楠岃瘉Webhook璇锋眰鐨勫瘑閽?
	WebhookSecret string
}

// Name 杩斿洖娓犻亾鍚嶇О
func (p *TelegramParser) Name() string {
	return "telegram"
}

// TelegramUpdate Telegram Webhook鏇存柊娑堟伅
type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

// TelegramMessage Telegram娑堟伅
type TelegramMessage struct {
	MessageID int           `json:"message_id"`
	From      *TelegramUser `json:"from,omitempty"`
	Chat      *TelegramChat `json:"chat"`
	Date      int64         `json:"date"`
	Text      string        `json:"text,omitempty"`
}

// TelegramUser Telegram鐢ㄦ埛
type TelegramUser struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// TelegramChat Telegram浼氳瘽
type TelegramChat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"` // private, group, supergroup, channel
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Parse 瑙ｆ瀽Telegram娑堟伅
func (p *TelegramParser) Parse(rawData []byte) (*model.UnifiedMessage, error) {
	var update TelegramUpdate
	if err := json.Unmarshal(rawData, &update); err != nil {
		return nil, fmt.Errorf("瑙ｆ瀽Telegram娑堟伅澶辫触: %w", err)
	}

	if update.Message == nil {
		return nil, fmt.Errorf("涓嶆敮鎸佺殑Telegram鏇存柊绫诲瀷锛堥潪娑堟伅锛?)
	}

	if update.Message.Text == "" {
		return nil, fmt.Errorf("绌烘秷鎭垨闈炴枃鏈秷鎭?)
	}

	msg := update.Message
	
	// 鎻愬彇鐢ㄦ埛ID
	var userID string
	if msg.From != nil {
		userID = strconv.FormatInt(msg.From.ID, 10)
	}

	// 鎻愬彇浼氳瘽ID
	chatID := strconv.FormatInt(msg.Chat.ID, 10)

	// 鏋勫缓鍘熷鏁版嵁
	rawMap := map[string]interface{}{
		"update_id":  update.UpdateID,
		"message_id": msg.MessageID,
		"chat_type":  msg.Chat.Type,
	}
	if msg.From != nil {
		rawMap["from_username"] = msg.From.Username
		rawMap["from_name"] = msg.From.FirstName + " " + msg.From.LastName
	}

	return NewUnifiedMessage(msg.Text, "telegram", userID, chatID, rawMap), nil
}

// Validate 楠岃瘉Telegram Webhook璇锋眰
// 浣跨敤 X-Telegram-Bot-Api-Secret-Token 澶磋繘琛岄獙璇?
func (p *TelegramParser) Validate(headers map[string]string, body []byte) bool {
	if p.WebhookSecret == "" {
		// 鏈厤缃瘑閽ワ紝璺宠繃楠岃瘉
		return true
	}

	// Telegram 浣跨敤 X-Telegram-Bot-Api-Secret-Token 澶?
	secretToken := headers["X-Telegram-Bot-Api-Secret-Token"]
	if secretToken == "" {
		secretToken = headers["x-telegram-bot-api-secret-token"]
	}

	return secretToken == p.WebhookSecret
}

// CalculateTelegramSecretTokenHash 璁＄畻Telegram瀵嗛挜Hash锛堢敤浜庤缃甒ebhook鏃讹級
func CalculateTelegramSecretTokenHash(botToken, data string) string {
	h := hmac.New(sha256.New, []byte(botToken))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
