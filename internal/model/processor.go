package model

// Processor 澶勭悊鍣ㄥ畾涔?
type Processor struct {
	// ID 澶勭悊鍣ㄥ敮涓€鏍囪瘑
	ID string `yaml:"id" json:"id"`

	// Name 澶勭悊鍣ㄥ悕绉帮紙鐢ㄤ簬鏄剧ず锛?
	Name string `yaml:"name" json:"name"`

	// Group 澶勭悊鍣ㄥ垎缁勶紙濡傦細lighting, climate, network锛?
	Group string `yaml:"group" json:"group"`

	// Description 澶勭悊鍣ㄦ弿杩帮紙鐢ㄤ簬LLM鍖归厤锛?
	Description string `yaml:"description" json:"description"`

	// Keywords 鍏抽敭璇嶅垪琛紙杈呭姪鍖归厤锛?
	Keywords []string `yaml:"keywords" json:"keywords"`

	// Parameters 鍙傛暟瀹氫箟鍒楄〃
	Parameters []Parameter `yaml:"parameters" json:"parameters"`

	// Enabled 鏄惁鍚敤
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// Parameter 澶勭悊鍣ㄥ弬鏁板畾涔?
type Parameter struct {
	// Name 鍙傛暟鍚嶇О
	Name string `yaml:"name" json:"name"`

	// Type 鍙傛暟绫诲瀷: string, int, float, bool, enum
	Type string `yaml:"type" json:"type"`

	// Required 鏄惁蹇呭～
	Required bool `yaml:"required" json:"required"`

	// Description 鍙傛暟鎻忚堪锛堢敤浜嶭LM鎻愬彇锛?
	Description string `yaml:"description" json:"description"`

	// Values 鏋氫妇鍊煎垪琛紙浠呭綋Type涓篹num鏃舵湁鏁堬級
	Values []string `yaml:"values,omitempty" json:"values,omitempty"`

	// Range 鏁板€艰寖鍥?[min, max]锛堜粎褰揟ype涓篿nt鎴杅loat鏃舵湁鏁堬級
	Range []float64 `yaml:"range,omitempty" json:"range,omitempty"`

	// Default 榛樿鍊?
	Default interface{} `yaml:"default,omitempty" json:"default,omitempty"`
}

// ProcessorMatchResult LLM澶勭悊鍣ㄥ尮閰嶇粨鏋?
type ProcessorMatchResult struct {
	// Matches 鍖归厤鐨勫鐞嗗櫒鍒楄〃锛屾寜缃俊搴﹂檷搴忔帓鍒?
	Matches []ProcessorMatch `json:"matches"`
}

// ProcessorMatch 鍗曚釜澶勭悊鍣ㄥ尮閰嶇粨鏋?
type ProcessorMatch struct {
	// ProcessorID 澶勭悊鍣↖D
	ProcessorID string `json:"processor_id"`

	// Confidence 缃俊搴?0-1
	Confidence float64 `json:"confidence"`

	// Reason 鍖归厤鍘熷洜锛堢敤浜庤皟璇曪級
	Reason string `json:"reason,omitempty"`
}

// ParameterExtractionResult LLM鍙傛暟鎻愬彇缁撴灉
type ParameterExtractionResult struct {
	// Success 鏄惁鎴愬姛鎻愬彇
	Success bool `json:"success"`

	// Parameters 鎻愬彇鐨勫弬鏁伴敭鍊煎
	Parameters map[string]interface{} `json:"parameters"`

	// MissingRequired 缂哄け鐨勫繀濉弬鏁?
	MissingRequired []string `json:"missing_required,omitempty"`

	// Message 鎻愬彇澶辫触鏃剁殑閿欒淇℃伅
	Message string `json:"message,omitempty"`
}
