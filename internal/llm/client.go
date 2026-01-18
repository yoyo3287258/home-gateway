package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yoyo3287258/home-gateway/internal/config"
)

// Client LLM API瀹㈡埛绔?
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
	maxRetries int
}

// NewClient 鍒涘缓LLM瀹㈡埛绔?
func NewClient(cfg *config.LLMConfig) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		maxRetries: cfg.MaxRetries,
	}
}

// ChatMessage 瀵硅瘽娑堟伅
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // 娑堟伅鍐呭
}

// ChatRequest OpenAI鍏煎鐨勫璇濊姹?
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ChatResponse OpenAI鍏煎鐨勫璇濆搷搴?
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Chat 鍙戦€佸璇濊姹?
func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.3, // 杈冧綆鐨勬俯搴︼紝浣胯緭鍑烘洿纭畾
	}

	var lastErr error
	for i := 0; i <= c.maxRetries; i++ {
		if i > 0 {
			// 閲嶈瘯鍓嶇瓑寰?
			time.Sleep(time.Duration(i) * 500 * time.Millisecond)
		}

		result, err := c.doRequest(ctx, req)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("LLM璇锋眰澶辫触锛堝凡閲嶈瘯%d娆★級: %w", c.maxRetries, lastErr)
}

// doRequest 鎵цHTTP璇锋眰
func (c *Client) doRequest(ctx context.Context, req ChatRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("搴忓垪鍖栬姹傚け璐? %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("鍒涘缓璇锋眰澶辫触: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("鍙戦€佽姹傚け璐? %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("璇诲彇鍝嶅簲澶辫触: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("瑙ｆ瀽鍝嶅簲澶辫触: %w, 鍘熷鍝嶅簲: %s", err, string(respBody))
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("LLM API閿欒: %s (type: %s, code: %s)",
			chatResp.Error.Message, chatResp.Error.Type, chatResp.Error.Code)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM杩斿洖绌哄搷搴?)
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ChatWithJSON 鍙戦€佸璇濊姹傚苟瑙ｆ瀽JSON鍝嶅簲
func (c *Client) ChatWithJSON(ctx context.Context, messages []ChatMessage, result interface{}) error {
	content, err := c.Chat(ctx, messages)
	if err != nil {
		return err
	}

	// 灏濊瘯鎻愬彇JSON锛圠LM鍙兘浼氬湪JSON鍓嶅悗娣诲姞棰濆鏂囨湰锛?
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return fmt.Errorf("LLM鍝嶅簲涓湭鎵惧埌鏈夋晥JSON: %s", content)
	}

	if err := json.Unmarshal([]byte(jsonStr), result); err != nil {
		return fmt.Errorf("瑙ｆ瀽LLM JSON鍝嶅簲澶辫触: %w, 鍐呭: %s", err, jsonStr)
	}

	return nil
}

// extractJSON 浠庢枃鏈腑鎻愬彇JSON
func extractJSON(s string) string {
	// 灏濊瘯鎵惧埌JSON瀵硅薄
	start := strings.Index(s, "{")
	if start == -1 {
		// 灏濊瘯鎵綣SON鏁扮粍
		start = strings.Index(s, "[")
		if start == -1 {
			return ""
		}
		end := strings.LastIndex(s, "]")
		if end == -1 || end <= start {
			return ""
		}
		return s[start : end+1]
	}

	end := strings.LastIndex(s, "}")
	if end == -1 || end <= start {
		return ""
	}

	return s[start : end+1]
}
