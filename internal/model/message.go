package model

import "time"

// UnifiedMessage 缁熶竴娑堟伅鏍煎紡
// 鎵€鏈夋笭閬撶殑娑堟伅閮戒細琚В鏋愪负杩欎釜鏍煎紡
type UnifiedMessage struct {
	// Text 鐢ㄦ埛杈撳叆鐨勬枃鏈唴瀹?
	Text string `json:"text"`

	// Channel 娑堟伅鏉ユ簮娓犻亾: http, telegram, wechat_work, wechat_mp
	Channel string `json:"channel"`

	// ChannelUserID 娓犻亾鐢ㄦ埛ID锛堝彲閫夛紝鐢ㄤ簬鍥炲锛?
	ChannelUserID string `json:"channel_user_id,omitempty"`

	// ChannelChatID 娓犻亾浼氳瘽ID锛堝彲閫夛紝鐢ㄤ簬缇ょ粍娑堟伅锛?
	ChannelChatID string `json:"channel_chat_id,omitempty"`

	// Timestamp 娑堟伅鏃堕棿鎴?
	Timestamp time.Time `json:"timestamp"`

	// RawData 鍘熷娑堟伅鏁版嵁锛堢敤浜庤皟璇曪級
	RawData map[string]interface{} `json:"raw_data,omitempty"`
}

// CommandRequest 閫氱敤HTTP鎺ュ彛鐨勮姹傛牸寮?
type CommandRequest struct {
	// Text 鐢ㄦ埛杈撳叆鐨勬枃鏈?
	Text string `json:"text" binding:"required"`

	// Channel 娓犻亾鏍囪瘑锛堝彲閫夛紝榛樿涓?http锛?
	Channel string `json:"channel,omitempty"`
}

// CommandResponse 閫氱敤HTTP鎺ュ彛鐨勫搷搴旀牸寮?
type CommandResponse struct {
	// Success 鏄惁鎴愬姛
	Success bool `json:"success"`

	// Message 杩斿洖缁欑敤鎴风殑娑堟伅
	Message string `json:"message"`

	// TraceID 璋冪敤娴佹按鍙凤紝鐢ㄤ簬杩借釜
	TraceID string `json:"trace_id"`

	// ProcessorID 澶勭悊璇ヨ姹傜殑澶勭悊鍣↖D
	ProcessorID string `json:"processor_id,omitempty"`

	// Data 棰濆鏁版嵁锛堝彲閫夛級
	Data map[string]interface{} `json:"data,omitempty"`

	// Candidates 褰撴湁澶氫釜鍖归厤澶勭悊鍣ㄦ椂锛岃繑鍥炲€欓€夊垪琛ㄨ鐢ㄦ埛閫夋嫨
	Candidates []ProcessorCandidate `json:"candidates,omitempty"`
}

// ProcessorCandidate 澶勭悊鍣ㄥ€欓€夐」
type ProcessorCandidate struct {
	// ID 澶勭悊鍣↖D
	ID string `json:"id"`

	// Name 澶勭悊鍣ㄥ悕绉?
	Name string `json:"name"`

	// Confidence 鍖归厤缃俊搴?0-1
	Confidence float64 `json:"confidence"`
}

// ReloadResponse 閰嶇疆閲嶈浇鎺ュ彛鐨勫搷搴?
type ReloadResponse struct {
	// Success 鏄惁鎴愬姛
	Success bool `json:"success"`

	// Message 缁撴灉娑堟伅
	Message string `json:"message"`

	// ProcessorCount 閲嶈浇鍚庣殑澶勭悊鍣ㄦ暟閲?
	ProcessorCount int `json:"processor_count,omitempty"`
}

// HealthResponse 鍋ュ悍妫€鏌ュ搷搴?
type HealthResponse struct {
	// Status 鐘舵€? ok, degraded, error
	Status string `json:"status"`

	// Version 鐗堟湰鍙?
	Version string `json:"version,omitempty"`

	// Uptime 杩愯鏃堕棿锛堢锛?
	Uptime int64 `json:"uptime,omitempty"`

	// Components 鍚勭粍浠剁姸鎬?
	Components map[string]string `json:"components,omitempty"`
}
