package api

import (
	"context"
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

// Version 绋嬪簭鐗堟湰鍙凤紙鍦ㄧ紪璇戞椂娉ㄥ叆锛?
var Version = "dev"

// Handler API璇锋眰澶勭悊鍣?
type Handler struct {
	configMgr   *config.Manager
	llmClient   *llm.Client
	kafkaClient *kafka.Client
	startTime   time.Time
}

// NewHandler 鍒涘缓璇锋眰澶勭悊鍣?
func NewHandler(configMgr *config.Manager, llmClient *llm.Client, kafkaClient *kafka.Client) *Handler {
	return &Handler{
		configMgr:   configMgr,
		llmClient:   llmClient,
		kafkaClient: kafkaClient,
		startTime:   time.Now(),
	}
}

// Health 鍋ュ悍妫€鏌ユ帴鍙?
func (h *Handler) Health(c *gin.Context) {
	components := make(map[string]string)
	
	// 妫€鏌LM
	if h.llmClient != nil {
		components["llm"] = "ok"
	} else {
		components["llm"] = "not_configured"
	}

	// 妫€鏌afka
	if h.kafkaClient != nil {
		components["kafka"] = "ok"
	} else {
		components["kafka"] = "not_configured"
	}

	c.JSON(http.StatusOK, model.HealthResponse{
		Status:     "ok",
		Version:    Version,
		Uptime:     int64(time.Since(h.startTime).Seconds()),
		Components: components,
	})
}

// ListProcessors 鑾峰彇澶勭悊鍣ㄥ垪琛?
func (h *Handler) ListProcessors(c *gin.Context) {
	processors := h.configMgr.GetProcessors()
	
	// 鍙繑鍥炲惎鐢ㄧ殑澶勭悊鍣紝涓斾笉鏆撮湶鍐呴儴缁嗚妭
	var result []map[string]interface{}
	for _, p := range processors {
		if p.Enabled {
			result = append(result, map[string]interface{}{
				"id":          p.ID,
				"name":        p.Name,
				"description": p.Description,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"processors": result,
		"count":      len(result),
	})
}

// Command 閫氱敤鍛戒护澶勭悊鎺ュ彛
func (h *Handler) Command(c *gin.Context) {
	var req model.CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.CommandResponse{
			Success: false,
			Message: "璇锋眰鏍煎紡閿欒: " + err.Error(),
		})
		return
	}

	// 璁剧疆榛樿娓犻亾
	if req.Channel == "" {
		req.Channel = "http"
	}

	// 鍒涘缓缁熶竴娑堟伅
	msg := &model.UnifiedMessage{
		Text:      req.Text,
		Channel:   req.Channel,
		Timestamp: time.Now(),
	}

	// 澶勭悊娑堟伅
	resp := h.processMessage(c.Request.Context(), msg)
	
	if resp.Success {
		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusOK, resp) // 涓氬姟閿欒浠嶇劧杩斿洖200
	}
}

// ReloadConfig 閲嶆柊鍔犺浇閰嶇疆
func (h *Handler) ReloadConfig(c *gin.Context) {
	if err := h.configMgr.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, model.ReloadResponse{
			Success: false,
			Message: "閰嶇疆閲嶈浇澶辫触: " + err.Error(),
		})
		return
	}

	processors := h.configMgr.GetProcessors()
	enabledCount := 0
	for _, p := range processors {
		if p.Enabled {
			enabledCount++
		}
	}

	c.JSON(http.StatusOK, model.ReloadResponse{
		Success:        true,
		Message:        "閰嶇疆閲嶈浇鎴愬姛",
		ProcessorCount: enabledCount,
	})
}

// TelegramWebhook Telegram Webhook澶勭悊
func (h *Handler) TelegramWebhook(c *gin.Context) {
	cfg := h.configMgr.Get()
	if !cfg.Channels.Telegram.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Telegram娓犻亾鏈惎鐢?})
		return
	}

	// 璇诲彇璇锋眰浣?
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "璇诲彇璇锋眰澶辫触"})
		return
	}

	// 鑾峰彇瑙ｆ瀽鍣ㄥ苟楠岃瘉
	parser := &channel.TelegramParser{
		WebhookSecret: cfg.Channels.Telegram.WebhookSecret,
	}

	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	if !parser.Validate(headers, body) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "楠岃瘉澶辫触"})
		return
	}

	// 瑙ｆ瀽娑堟伅
	msg, err := parser.Parse(body)
	if err != nil {
		// Telegram闇€瑕佽繑鍥?00锛屽惁鍒欎細閲嶈瘯
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// 澶勭悊娑堟伅锛堝紓姝ワ紝涓嶉樆濉濿ebhook鍝嶅簲锛?
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		resp := h.processMessage(ctx, msg)
		
		// TODO: 閫氳繃Telegram API鍙戦€佸洖澶?
		_ = resp
	}()

	// Telegram Webhook闇€瑕佸揩閫熻繑鍥?00
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// WeChatWorkWebhook 浼佷笟寰俊Webhook澶勭悊锛堥鐣欙級
func (h *Handler) WeChatWorkWebhook(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "浼佷笟寰俊娓犻亾鏆傛湭瀹炵幇"})
}

// processMessage 澶勭悊缁熶竴娑堟伅
func (h *Handler) processMessage(ctx context.Context, msg *model.UnifiedMessage) *model.CommandResponse {
	traceID := uuid.New().String()

	// 1. 浣跨敤LLM鍖归厤澶勭悊鍣?
	processors := h.configMgr.GetProcessors()
	if len(processors) == 0 {
		return &model.CommandResponse{
			Success: false,
			Message: "娌℃湁鍙敤鐨勫鐞嗗櫒锛岃妫€鏌ラ厤缃?,
			TraceID: traceID,
		}
	}

	matchResult, err := h.llmClient.MatchProcessors(ctx, msg.Text, processors)
	if err != nil {
		return &model.CommandResponse{
			Success: false,
			Message: "澶勭悊鍣ㄥ尮閰嶅け璐? " + err.Error(),
			TraceID: traceID,
		}
	}

	if len(matchResult.Matches) == 0 {
		return &model.CommandResponse{
			Success: false,
			Message: "鏃犳硶璇嗗埆鎮ㄧ殑鎸囦护锛岃灏濊瘯鏇存槑纭殑鎻忚堪",
			TraceID: traceID,
		}
	}

	// 2. 妫€鏌ユ槸鍚︽湁澶氫釜楂樼疆淇″害鍖归厤
	const confidenceThreshold = 0.8
	const ambiguityGap = 0.15
	
	topMatch := matchResult.Matches[0]
	if len(matchResult.Matches) > 1 {
		secondMatch := matchResult.Matches[1]
		// 濡傛灉绗竴鍜岀浜屽尮閰嶇殑缃俊搴﹀樊璺濆皬浜庨槇鍊硷紝璁╃敤鎴烽€夋嫨
		if topMatch.Confidence < confidenceThreshold || 
		   (topMatch.Confidence - secondMatch.Confidence) < ambiguityGap {
			var candidates []model.ProcessorCandidate
			for _, m := range matchResult.Matches {
				if m.Confidence > 0.5 { // 鍙繑鍥炵疆淇″害澶т簬0.5鐨?
					p := h.configMgr.GetProcessor(m.ProcessorID)
					if p != nil {
						candidates = append(candidates, model.ProcessorCandidate{
							ID:         m.ProcessorID,
							Name:       p.Name,
							Confidence: m.Confidence,
						})
					}
				}
			}
			return &model.CommandResponse{
				Success:    false,
				Message:    "鎮ㄧ殑鎸囦护鍙兘瀵瑰簲澶氫釜鍔熻兘锛岃閫夋嫨鎴栨洿鏄庣‘鍦版弿杩?,
				TraceID:    traceID,
				Candidates: candidates,
			}
		}
	}

	// 3. 鑾峰彇鍖归厤鐨勫鐞嗗櫒
	processor := h.configMgr.GetProcessor(topMatch.ProcessorID)
	if processor == nil {
		return &model.CommandResponse{
			Success: false,
			Message: "澶勭悊鍣ㄤ笉瀛樺湪: " + topMatch.ProcessorID,
			TraceID: traceID,
		}
	}

	// 4. 浣跨敤LLM鎻愬彇鍙傛暟
	paramResult, err := h.llmClient.ExtractParameters(ctx, msg.Text, *processor)
	if err != nil {
		return &model.CommandResponse{
			Success:     false,
			Message:     "鍙傛暟鎻愬彇澶辫触: " + err.Error(),
			TraceID:     traceID,
			ProcessorID: processor.ID,
		}
	}

	if !paramResult.Success {
		return &model.CommandResponse{
			Success:     false,
			Message:     paramResult.Message,
			TraceID:     traceID,
			ProcessorID: processor.ID,
		}
	}

	// 5. 鏋勯€燢afka璇锋眰
	kafkaReq := &model.KafkaRequest{
		TraceID:       traceID,
		Timestamp:     time.Now(),
		ProcessorID:   processor.ID,
		Parameters:    paramResult.Parameters,
		OriginalText:  msg.Text,
		Channel:       msg.Channel,
		ChannelUserID: msg.ChannelUserID,
	}

	// 6. 鍙戦€佸埌Kafka骞剁瓑寰呭搷搴?
	if h.kafkaClient == nil {
		// Kafka鏈厤缃紝杩斿洖妯℃嫙鍝嶅簲锛堢敤浜庢祴璇曪級
		return &model.CommandResponse{
			Success:     true,
			Message:     "鎸囦护宸茶瘑鍒紙Kafka鏈厤缃紝鏃犳硶鍙戦€佸埌澶勭悊鍣級",
			TraceID:     traceID,
			ProcessorID: processor.ID,
			Data: map[string]interface{}{
				"parameters": paramResult.Parameters,
			},
		}
	}

	kafkaResp, err := h.kafkaClient.SendAndWait(kafkaReq)
	if err != nil {
		return &model.CommandResponse{
			Success:     false,
			Message:     "绛夊緟澶勭悊鍣ㄥ搷搴旇秴鏃舵垨澶辫触: " + err.Error(),
			TraceID:     traceID,
			ProcessorID: processor.ID,
		}
	}

	// 7. 杩斿洖澶勭悊缁撴灉
	return &model.CommandResponse{
		Success:     kafkaResp.Success,
		Message:     kafkaResp.Message,
		TraceID:     traceID,
		ProcessorID: processor.ID,
		Data:        kafkaResp.Data,
	}
}
