package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/yoyo3287258/home-gateway/internal/model"
)

// MatchProcessors 使用LLM匹配用户输入到合适的处理器
// 返回按置信度排序的处理器匹配结果
func (c *Client) MatchProcessors(ctx context.Context, userInput string, processors []model.Processor) (*model.ProcessorMatchResult, error) {
	// 构建处理器描述
	var processorDescriptions []string
	for _, p := range processors {
		if !p.Enabled {
			continue
		}
		desc := fmt.Sprintf("- ID: %s, 名称: %s, 描述: %s, 关键词: %s",
			p.ID, p.Name, p.Description, strings.Join(p.Keywords, "、"))
		processorDescriptions = append(processorDescriptions, desc)
	}

	systemPrompt := `你是一个智能家居控制意图识别助手。你的任务是分析用户的输入，判断用户想要使用哪个处理器来完成操作。

可用的处理器列表：
` + strings.Join(processorDescriptions, "\n") + `

请根据用户输入，返回最匹配的处理器列表。每个匹配项包含处理器ID和置信度（0-1的小数）。
如果用户输入模糊或可能对应多个处理器，请返回多个结果。
如果用户输入与所有处理器都不匹配，返回空的matches数组。

请以JSON格式返回，格式如下：
{
  "matches": [
    {"processor_id": "xxx", "confidence": 0.95, "reason": "匹配原因"},
    {"processor_id": "yyy", "confidence": 0.75, "reason": "匹配原因"}
  ]
}

只返回JSON，不要有其他内容。`

	userPrompt := fmt.Sprintf("用户输入：%s", userInput)

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	var result model.ProcessorMatchResult
	if err := c.ChatWithJSON(ctx, messages, &result); err != nil {
		return nil, fmt.Errorf("处理器匹配失败: %w", err)
	}

	return &result, nil
}

// ExtractParameters 使用LLM从用户输入中提取处理器所需的参数
func (c *Client) ExtractParameters(ctx context.Context, userInput string, processor model.Processor) (*model.ParameterExtractionResult, error) {
	// 构建参数描述
	var paramDescriptions []string
	var requiredParams []string
	for _, p := range processor.Parameters {
		desc := fmt.Sprintf("- %s (%s): %s", p.Name, p.Type, p.Description)
		if p.Required {
			desc += " [必填]"
			requiredParams = append(requiredParams, p.Name)
		}
		if len(p.Values) > 0 {
			desc += fmt.Sprintf(" 可选值: %s", strings.Join(p.Values, ", "))
		}
		if len(p.Range) == 2 {
			desc += fmt.Sprintf(" 范围: %v-%v", p.Range[0], p.Range[1])
		}
		if p.Default != nil {
			desc += fmt.Sprintf(" 默认值: %v", p.Default)
		}
		paramDescriptions = append(paramDescriptions, desc)
	}

	systemPrompt := fmt.Sprintf(`你是一个智能家居参数提取助手。你的任务是从用户输入中提取【%s】处理器所需的参数。

处理器描述：%s

需要提取的参数：
%s

请分析用户输入，提取所需参数值。
- 如果用户没有明确指定某个可选参数，不要在parameters中包含该参数
- 如果用户没有明确指定某个必填参数，在missing_required中列出
- 对于enum类型的参数，请将用户的自然语言转换为对应的值（如"打开"转换为"on"）
- 对于数值类型，请确保值在有效范围内

请以JSON格式返回，格式如下：
{
  "success": true,
  "parameters": {
    "param1": "value1",
    "param2": 123
  },
  "missing_required": [],
  "message": ""
}

如果无法提取必填参数，设置success为false，并在message中说明原因。

只返回JSON，不要有其他内容。`, processor.Name, processor.Description, strings.Join(paramDescriptions, "\n"))

	userPrompt := fmt.Sprintf("用户输入：%s", userInput)

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	var result model.ParameterExtractionResult
	if err := c.ChatWithJSON(ctx, messages, &result); err != nil {
		return nil, fmt.Errorf("参数提取失败: %w", err)
	}

	// 验证必填参数
	if result.Success && len(result.MissingRequired) > 0 {
		result.Success = false
		result.Message = fmt.Sprintf("缺少必填参数: %s", strings.Join(result.MissingRequired, ", "))
	}

	return &result, nil
}
