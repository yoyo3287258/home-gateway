package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/yoyo3287258/home-gateway/internal/model"
)

// MatchProcessors 浣跨敤LLM鍖归厤鐢ㄦ埛杈撳叆鍒板悎閫傜殑澶勭悊鍣?
// 杩斿洖鎸夌疆淇″害鎺掑簭鐨勫鐞嗗櫒鍖归厤缁撴灉
func (c *Client) MatchProcessors(ctx context.Context, userInput string, processors []model.Processor) (*model.ProcessorMatchResult, error) {
	// 鏋勫缓澶勭悊鍣ㄦ弿杩?
	var processorDescriptions []string
	for _, p := range processors {
		if !p.Enabled {
			continue
		}
		desc := fmt.Sprintf("- ID: %s, 鍚嶇О: %s, 鎻忚堪: %s, 鍏抽敭璇? %s",
			p.ID, p.Name, p.Description, strings.Join(p.Keywords, "銆?))
		processorDescriptions = append(processorDescriptions, desc)
	}

	systemPrompt := `浣犳槸涓€涓櫤鑳藉灞呮帶鍒舵剰鍥捐瘑鍒姪鎵嬨€備綘鐨勪换鍔℃槸鍒嗘瀽鐢ㄦ埛鐨勮緭鍏ワ紝鍒ゆ柇鐢ㄦ埛鎯宠浣跨敤鍝釜澶勭悊鍣ㄦ潵瀹屾垚鎿嶄綔銆?

鍙敤鐨勫鐞嗗櫒鍒楄〃锛?
` + strings.Join(processorDescriptions, "\n") + `

璇锋牴鎹敤鎴疯緭鍏ワ紝杩斿洖鏈€鍖归厤鐨勫鐞嗗櫒鍒楄〃銆傛瘡涓尮閰嶉」鍖呭惈澶勭悊鍣↖D鍜岀疆淇″害锛?-1鐨勫皬鏁帮級銆?
濡傛灉鐢ㄦ埛杈撳叆妯＄硦鎴栧彲鑳藉搴斿涓鐞嗗櫒锛岃杩斿洖澶氫釜缁撴灉銆?
濡傛灉鐢ㄦ埛杈撳叆涓庢墍鏈夊鐞嗗櫒閮戒笉鍖归厤锛岃繑鍥炵┖鐨刴atches鏁扮粍銆?

璇蜂互JSON鏍煎紡杩斿洖锛屾牸寮忓涓嬶細
{
  "matches": [
    {"processor_id": "xxx", "confidence": 0.95, "reason": "鍖归厤鍘熷洜"},
    {"processor_id": "yyy", "confidence": 0.75, "reason": "鍖归厤鍘熷洜"}
  ]
}

鍙繑鍥濲SON锛屼笉瑕佹湁鍏朵粬鍐呭銆俙

	userPrompt := fmt.Sprintf("鐢ㄦ埛杈撳叆锛?s", userInput)

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	var result model.ProcessorMatchResult
	if err := c.ChatWithJSON(ctx, messages, &result); err != nil {
		return nil, fmt.Errorf("澶勭悊鍣ㄥ尮閰嶅け璐? %w", err)
	}

	return &result, nil
}

// ExtractParameters 浣跨敤LLM浠庣敤鎴疯緭鍏ヤ腑鎻愬彇澶勭悊鍣ㄦ墍闇€鐨勫弬鏁?
func (c *Client) ExtractParameters(ctx context.Context, userInput string, processor model.Processor) (*model.ParameterExtractionResult, error) {
	// 鏋勫缓鍙傛暟鎻忚堪
	var paramDescriptions []string
	var requiredParams []string
	for _, p := range processor.Parameters {
		desc := fmt.Sprintf("- %s (%s): %s", p.Name, p.Type, p.Description)
		if p.Required {
			desc += " [蹇呭～]"
			requiredParams = append(requiredParams, p.Name)
		}
		if len(p.Values) > 0 {
			desc += fmt.Sprintf(" 鍙€夊€? %s", strings.Join(p.Values, ", "))
		}
		if len(p.Range) == 2 {
			desc += fmt.Sprintf(" 鑼冨洿: %v-%v", p.Range[0], p.Range[1])
		}
		if p.Default != nil {
			desc += fmt.Sprintf(" 榛樿鍊? %v", p.Default)
		}
		paramDescriptions = append(paramDescriptions, desc)
	}

	systemPrompt := fmt.Sprintf(`浣犳槸涓€涓櫤鑳藉灞呭弬鏁版彁鍙栧姪鎵嬨€備綘鐨勪换鍔℃槸浠庣敤鎴疯緭鍏ヤ腑鎻愬彇銆?s銆戝鐞嗗櫒鎵€闇€鐨勫弬鏁般€?

澶勭悊鍣ㄦ弿杩帮細%s

闇€瑕佹彁鍙栫殑鍙傛暟锛?
%s

璇峰垎鏋愮敤鎴疯緭鍏ワ紝鎻愬彇鎵€闇€鍙傛暟鍊笺€?
- 濡傛灉鐢ㄦ埛娌℃湁鏄庣‘鎸囧畾鏌愪釜鍙€夊弬鏁帮紝涓嶈鍦╬arameters涓寘鍚鍙傛暟
- 濡傛灉鐢ㄦ埛娌℃湁鏄庣‘鎸囧畾鏌愪釜蹇呭～鍙傛暟锛屽湪missing_required涓垪鍑?
- 瀵逛簬enum绫诲瀷鐨勫弬鏁帮紝璇峰皢鐢ㄦ埛鐨勮嚜鐒惰瑷€杞崲涓哄搴旂殑鍊硷紙濡?鎵撳紑"杞崲涓?on"锛?
- 瀵逛簬鏁板€肩被鍨嬶紝璇风‘淇濆€煎湪鏈夋晥鑼冨洿鍐?

璇蜂互JSON鏍煎紡杩斿洖锛屾牸寮忓涓嬶細
{
  "success": true,
  "parameters": {
    "param1": "value1",
    "param2": 123
  },
  "missing_required": [],
  "message": ""
}

濡傛灉鏃犳硶鎻愬彇蹇呭～鍙傛暟锛岃缃畇uccess涓篺alse锛屽苟鍦╩essage涓鏄庡師鍥犮€?

鍙繑鍥濲SON锛屼笉瑕佹湁鍏朵粬鍐呭銆俙, processor.Name, processor.Description, strings.Join(paramDescriptions, "\n"))

	userPrompt := fmt.Sprintf("鐢ㄦ埛杈撳叆锛?s", userInput)

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	var result model.ParameterExtractionResult
	if err := c.ChatWithJSON(ctx, messages, &result); err != nil {
		return nil, fmt.Errorf("鍙傛暟鎻愬彇澶辫触: %w", err)
	}

	// 楠岃瘉蹇呭～鍙傛暟
	if result.Success && len(result.MissingRequired) > 0 {
		result.Success = false
		result.Message = fmt.Sprintf("缂哄皯蹇呭～鍙傛暟: %s", strings.Join(result.MissingRequired, ", "))
	}

	return &result, nil
}
