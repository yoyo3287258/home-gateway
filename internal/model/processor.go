package model

// Processor 处理器定义
type Processor struct {
	// ID 处理器唯一标识
	ID string `yaml:"id" json:"id"`

	// Name 处理器名称（用于显示）
	Name string `yaml:"name" json:"name"`

	// Group 处理器分组（如：lighting, climate, network）
	Group string `yaml:"group" json:"group"`

	// Description 处理器描述（用于LLM匹配）
	Description string `yaml:"description" json:"description"`

	// Keywords 关键词列表（辅助匹配）
	Keywords []string `yaml:"keywords" json:"keywords"`

	// Parameters 参数定义列表
	Parameters []Parameter `yaml:"parameters" json:"parameters"`

	// Enabled 是否启用
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// Parameter 参数定义
type Parameter struct {
	// Name 参数名
	Name string `yaml:"name" json:"name"`

	// Type 参数类型: string, int, float, bool, enum
	Type string `yaml:"type" json:"type"`

	// Required 是否必填
	Required bool `yaml:"required" json:"required"`

	// Description 参数描述
	Description string `yaml:"description" json:"description"`

	// Default 默认值
	Default interface{} `yaml:"default,omitempty" json:"default,omitempty"`

	// Values 枚举值列表（当Type为enum时使用）
	Values []string `yaml:"values,omitempty" json:"values,omitempty"`

	// Range 数值范围 [min, max]（当Type为int/float时使用）
	Range []float64 `yaml:"range,omitempty" json:"range,omitempty"`
}

// ProcessorMatchResult 处理器匹配结果
type ProcessorMatchResult struct {
	Matches []struct {
		ProcessorID string  `json:"processor_id"`
		Confidence  float64 `json:"confidence"`
		Reason      string  `json:"reason"`
	} `json:"matches"`
}

// ParameterExtractionResult 参数提取结果
type ParameterExtractionResult struct {
	Success         bool                   `json:"success"`
	Parameters      map[string]interface{} `json:"parameters"`
	MissingRequired []string               `json:"missing_required"`
	Message         string                 `json:"message"`
}
