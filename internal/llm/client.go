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

// Client LLM API客户端
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
	maxRetries int
}

// NewClient 创建LLM客户端
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

// ChatMessage 对话消息
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // 消息内容
}

// ChatRequest OpenAI兼容的对话请求
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ChatResponse OpenAI兼容的对话响应
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

// Chat 发送对话请求
func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.3, // 较低的温度，使输出更确定
	}

	var lastErr error
	for i := 0; i <= c.maxRetries; i++ {
		if i > 0 {
			// 重试前等待
			time.Sleep(time.Duration(i) * 500 * time.Millisecond)
		}

		result, err := c.doRequest(ctx, req)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("LLM请求失败（已重试%d次）: %w", c.maxRetries, lastErr)
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(ctx context.Context, req ChatRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w, 原始响应: %s", err, string(respBody))
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("LLM API错误: %s (type: %s, code: %s)",
			chatResp.Error.Message, chatResp.Error.Type, chatResp.Error.Code)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM返回空响应")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ChatWithJSON 发送对话请求并解析JSON响应
func (c *Client) ChatWithJSON(ctx context.Context, messages []ChatMessage, result interface{}) error {
	content, err := c.Chat(ctx, messages)
	if err != nil {
		return err
	}

	// 尝试提取JSON（LLM可能会在JSON前后添加额外文本）
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return fmt.Errorf("LLM响应中未找到有效JSON: %s", content)
	}

	if err := json.Unmarshal([]byte(jsonStr), result); err != nil {
		return fmt.Errorf("解析LLM JSON响应失败: %w, 内容: %s", err, jsonStr)
	}

	return nil
}

// extractJSON 从文本中提取JSON
func extractJSON(s string) string {
	// 尝试找到JSON对象
	start := strings.Index(s, "{")
	if start == -1 {
		// 尝试找JSON数组
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
