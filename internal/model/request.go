package model

import "time"

// KafkaRequest 鍙戦€佺粰澶勭悊鍣ㄧ殑Kafka璇锋眰娑堟伅
type KafkaRequest struct {
	// TraceID 璋冪敤娴佹按鍙凤紝鐢ㄤ簬杩借釜璇锋眰-鍝嶅簲
	TraceID string `json:"trace_id"`

	// Timestamp 璇锋眰鏃堕棿鎴?
	Timestamp time.Time `json:"timestamp"`

	// ProcessorID 鐩爣澶勭悊鍣↖D
	ProcessorID string `json:"processor_id"`

	// Parameters 鎻愬彇鐨勫弬鏁?
	Parameters map[string]interface{} `json:"parameters"`

	// OriginalText 鐢ㄦ埛鍘熷杈撳叆鏂囨湰
	OriginalText string `json:"original_text"`

	// Channel 娑堟伅鏉ユ簮娓犻亾
	Channel string `json:"channel"`

	// ChannelUserID 娓犻亾鐢ㄦ埛ID锛堝彲閫夛級
	ChannelUserID string `json:"channel_user_id,omitempty"`
}

// KafkaResponse 澶勭悊鍣ㄨ繑鍥炵殑Kafka鍝嶅簲娑堟伅
type KafkaResponse struct {
	// TraceID 璋冪敤娴佹按鍙凤紝涓庤姹傚搴?
	TraceID string `json:"trace_id"`

	// Timestamp 鍝嶅簲鏃堕棿鎴?
	Timestamp time.Time `json:"timestamp"`

	// ProcessorID 澶勭悊璇ヨ姹傜殑澶勭悊鍣↖D
	ProcessorID string `json:"processor_id"`

	// Success 澶勭悊鏄惁鎴愬姛
	Success bool `json:"success"`

	// Message 杩斿洖缁欑敤鎴风殑娑堟伅
	Message string `json:"message"`

	// Data 棰濆鏁版嵁锛堝彲閫夛級
	Data map[string]interface{} `json:"data,omitempty"`

	// Error 閿欒淇℃伅锛堜粎褰揝uccess涓篺alse鏃讹級
	Error string `json:"error,omitempty"`
}
