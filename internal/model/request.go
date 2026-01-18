package model

import "time"

// KafkaRequest 发送到后端处理器的请求
type KafkaRequest struct {
	// TraceID 请求追踪ID
	TraceID string `json:"trace_id"`

	// ProcessorID 目标处理器ID
	ProcessorID string `json:"processor_id"`

	// Parameters 提取的参数
	Parameters map[string]interface{} `json:"parameters"`

	// RawMessage 原始消息上下文
	RawMessage UnifiedMessage `json:"raw_message"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// KafkaResponse 后端处理器返回的响应
type KafkaResponse struct {
	// TraceID 请求追踪ID
	TraceID string `json:"trace_id"`

	// ProcessorID 来源处理器ID
	ProcessorID string `json:"processor_id"`

	// Success 是否执行成功
	Success bool `json:"success"`

	// Result 执行结果（可以是文本或JSON对象）
	Result interface{} `json:"result,omitempty"`

	// Error 错误信息（如果Success为false）
	Error string `json:"error,omitempty"`

	// ProcessedAt 处理完成时间
	ProcessedAt time.Time `json:"processed_at"`
}
