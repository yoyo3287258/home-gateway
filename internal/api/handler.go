package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yoyo3287258/home-gateway/internal/channel"
	"github.com/yoyo3287258/home-gateway/internal/config"
	"github.com/yoyo3287258/home-gateway/internal/kafka"
	"github.com/yoyo3287258/home-gateway/internal/llm"
	"github.com/yoyo3287258/home-gateway/internal/model"
)

// Handler API处理器
type Handler struct {
	configMgr   *config.Manager
	llmClient   *llm.Client
	kafkaClient *kafka.Client
	parsers     map[string]channel.Parser
}

// NewHandler 创建API处理器
func NewHandler(configMgr *config.Manager, llmClient *llm.Client, kafkaClient *kafka.Client) *Handler {
	h := &Handler{
		configMgr:   configMgr,
		llmClient:   llmClient,
		kafkaClient: kafkaClient,
		parsers:     make(map[string]channel.Parser),
	}
	
	// 初始化解析器
	h.registerParsers()
	
	return h
}

// registerParsers 注册所有支持的渠道解析器
func (h *Handler) registerParsers() {
	cfg := h.configMgr.Get()
	
	// Generic HTTP Parser
	httpParser := &channel.HTTPParser{}
	h.parsers[httpParser.Name()] = httpParser
	
	// Telegram Parser
	if cfg.Channels.Telegram.Enabled {
		telegramParser := &channel.TelegramParser{
			WebhookSecret: cfg.Channels.Telegram.WebhookSecret,
		}
		h.parsers[telegramParser.Name()] = telegramParser
	}
	
	// WeChat Work Parser (TODO)
}

// Health 健康检查
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "up",
		"time":   time.Now(),
		"version": "dev", // TODO: inject version
	})
}

// ListProcessors 获取处理器列表
func (h *Handler) ListProcessors(c *gin.Context) {
	processors := h.configMgr.GetProcessors()
	c.JSON(http.StatusOK, gin.H{
		"data": processors,
	})
}

// Command 处理通用命令请求
func (h *Handler) Command(c *gin.Context) {
	// 1. 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取请求体"})
		return
	}

	// 2. 解析消息（默认使用HTTP parser）
	parser := h.parsers["http"]
	if parser == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "HTTP解析器未初始化"})
		return
	}

	msg, err := parser.Parse(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("解析请求失败: %v", err)})
		return
	}

	// 3. 处理消息
	h.processMessage(c, msg)
}

// TelegramWebhook 处理Telegram Webhook请求
func (h *Handler) TelegramWebhook(c *gin.Context) {
	// 1. 验证请求
	parser, ok := h.parsers["telegram"]
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Telegram未启用"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取请求体"})
		return
	}

	// 转换header map
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// 尝试作为TelegramParser进行验证
	if tgParser, ok := parser.(*channel.TelegramParser); ok {
		if !tgParser.Validate(headers, body) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Webhook验证失败"})
			return
		}
	} else {
		// 理论上不会走到这里，因为上面已经根据 key 取到了 parser
		// 但为了保险，如果取到的不是TelegramParser，应该通过还是拒绝？
		// 既然是telegram路由，应该必须是TelegramParser
		fmt.Printf("获取到的解析器类型错误: %T\n", parser)
	}

	// 2. 解析消息
	msg, err := parser.Parse(body)
	if err != nil {
		// Telegram可能会重试，如果解析失败记录日志并返回200避免重试轰炸?
		// 但为了调试，先返回错误
		fmt.Printf("Telegram解析失败: %v\n", err)
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": err.Error()})
		return
	}

	// 3. 处理消息
	h.processMessage(c, msg)
}

// WeChatWorkWebhook 处理企业微信Webhook请求
func (h *Handler) WeChatWorkWebhook(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "尚未实现"})
}

// ReloadConfig 重载配置
func (h *Handler) ReloadConfig(c *gin.Context) {
	if err := h.configMgr.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重载配置失败: %v", err)})
		return
	}
	
	// 重新注册解析器（配置可能改变）
	h.registerParsers()
	
	c.JSON(http.StatusOK, gin.H{"message": "配置已重载"})
}

// processMessage 处理统一消息的核心逻辑
func (h *Handler) processMessage(c *gin.Context, msg *model.UnifiedMessage) {
	ctx := c.Request.Context()
	traceID := c.GetString("trace_id")
	if traceID == "" {
		traceID = uuid.New().String()
	}

	fmt.Printf("[%s] 收到消息: %s (来自: %s)\n", traceID, msg.Content, msg.Channel)

	// 1. LLM 意图识别 (匹配处理器)
	processors := h.configMgr.GetProcessors()
	matchResult, err := h.llmClient.MatchProcessors(ctx, msg.Content, processors)
	if err != nil {
		fmt.Printf("[%s] LLM匹配失败: %v\n", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "意图识别服务异常"})
		return
	}

	if len(matchResult.Matches) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "抱歉，我没有理解您的指令，或者没有找到对应的功能。",
			"trace_id": traceID,
		})
		return
	}

	// 取置信度最高的匹配
	bestMatch := matchResult.Matches[0]
	fmt.Printf("[%s] 匹配处理器: %s (置信度: %.2f)\n", traceID, bestMatch.ProcessorID, bestMatch.Confidence)

	// 获取处理器详情
	processor := h.configMgr.GetProcessor(bestMatch.ProcessorID)
	if processor == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理器配置不存在"})
		return
	}

	// 2. LLM 参数提取
	paramResult, err := h.llmClient.ExtractParameters(ctx, msg.Content, *processor)
	if err != nil {
		fmt.Printf("[%s] 参数提取失败: %v\n", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "参数解析服务异常"})
		return
	}

	if !paramResult.Success {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("指令不完整: %s", paramResult.Message),
			"missing_params": paramResult.MissingRequired,
			"trace_id": traceID,
		})
		return
	}

	fmt.Printf("[%s] 提取参数: %v\n", traceID, paramResult.Parameters)

	// 3. 发送请求到Kafka (如果有Kafka客户端)
	if h.kafkaClient == nil {
		// 无Kafka模式，直接返回模拟成功
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("已识别指令：使用 [%s] 执行操作，参数：%v (演示模式，未发送到后端)", 
				processor.Name, paramResult.Parameters),
			"processor": processor.Name,
			"parameters": paramResult.Parameters,
			"trace_id": traceID,
		})
		return
	}

	kafkaReq := &model.KafkaRequest{
		TraceID:     traceID,
		ProcessorID: processor.ID,
		Parameters:  paramResult.Parameters,
		RawMessage:  *msg,
		CreatedAt:   time.Now(),
	}

	resp, err := h.kafkaClient.SendAndWait(kafkaReq)
	if err != nil {
		fmt.Printf("[%s] 后端处理超时或失败: %v\n", traceID, err)
		c.JSON(http.StatusGatewayTimeout, gin.H{
			"error": "后端服务响应超时",
			"trace_id": traceID,
		})
		return
	}

	// 4. 返回结果
	if !resp.Success {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("执行失败: %s", resp.Error),
			"trace_id": traceID,
		})
		return
	}

	// 如果Result是字符串，直接显示，如果是结构体，序列化
	msgResult := "操作成功"
	if resultStr, ok := resp.Result.(string); ok {
		msgResult = resultStr
	} else {
		bytes, _ := json.Marshal(resp.Result)
		msgResult = string(bytes)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": msgResult,
		"data": resp.Result,
		"trace_id": traceID,
	})
}
