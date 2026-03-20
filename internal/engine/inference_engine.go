package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"inference-engine/internal/models"
	"regexp"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// InferenceEngine 推理引擎
type InferenceEngine struct {
	modelManager *ModelManager
	db           *gorm.DB
}

// NewInferenceEngine 创建推理引擎
func NewInferenceEngine(mm *ModelManager, db *gorm.DB) *InferenceEngine {
	return &InferenceEngine{
		modelManager: mm,
		db:           db,
	}
}

// GetAvailableModels 获取可用模型列表
func (e *InferenceEngine) GetAvailableModels() []string {
	return e.modelManager.ListModels()
}

// RunInference 执行推理
func (e *InferenceEngine) RunInference(ctx context.Context, req *models.InferenceRequest) (*models.InferenceResult, error) {
	// 1. 获取模型
	provider, err := e.modelManager.GetProvider(req.Model)
	if err != nil {
		return nil, err
	}

	// 2. 创建任务记录
	// 1. 创建任务记录
	task := &models.InferenceTask{
		Title:       req.Title,
		Domain:      req.Domain,
		Subject:     req.Subject,
		ChangePoint: req.ChangePoint,
		TimeFrame:   fmt.Sprintf("%s - %s", req.TimeFrame.Start, req.TimeFrame.End),
		StepsCount:  req.StepsCount,
		ModelUsed:   req.Model,
		Status:      "processing",
	}

	if req.Variables != nil {
		varBytes, _ := json.Marshal(req.Variables)
		task.Variables = string(varBytes)
	}

	if err := e.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	// 3. 构建场景
	scenario := &models.Scenario{
		Domain:      req.Domain,
		Subject:     req.Subject,
		ChangePoint: req.ChangePoint,
		TimeFrame:   req.TimeFrame,
		Variables:   req.Variables,
	}

	// 4. 多步推理
	result := &models.InferenceResult{
		TaskID:    task.ID,
		Title:     req.Title,
		Steps:     make([]models.StepResult, req.StepsCount),
		CreatedAt: task.CreatedAt,
	}

	currentState := scenario.Variables

	for i := 0; i < req.StepsCount; i++ {
		// 构建步骤提示词
		prompt := buildStepPrompt(scenario, i+1, currentState)

		// 调用模型
		messages := []Message{
			{Role: "user", Content: prompt},
		}

		response, err := provider.Chat(ctx, messages)
		if err != nil {
			e.db.Model(task).Update("status", "failed")
			return nil, fmt.Errorf("步骤 %d 推理失败: %w", i+1, err)
		}

		// 解析响应
		stepResult := parseStepResponse(response, i+1)
		result.Steps[i] = *stepResult

		// 保存步骤到数据库
		stepRecord := &models.InferenceStep{
			TaskID:      task.ID,
			StepNumber:  i + 1,
			Title:       stepResult.Title,
			Description: stepResult.Description,
			Reasoning:   stepResult.Reasoning,
			Confidence:  stepResult.Confidence,
			ModelUsed:   provider.GetName(),
		}

		if inputState, err := json.Marshal(currentState); err == nil {
			stepRecord.InputState = string(inputState)
		}
		if outputState, err := json.Marshal(stepResult.State); err == nil {
			stepRecord.OutputState = string(outputState)
		}

		e.db.Create(stepRecord)

		// 更新当前状态
		currentState = stepResult.State
	}

	// 5. 生成总结
	summaryPrompt := buildSummaryPrompt(result)
	summaryMessages := []Message{{Role: "user", Content: summaryPrompt}}
	summary, _ := provider.Chat(ctx, summaryMessages)
	if summaryData := parseSummaryResponse(summary); summaryData != nil {
		result.Summary = summaryData.Summary
		result.SummaryData = summaryData
		if summaryJSON, err := json.Marshal(summaryData); err == nil {
			task.SummaryData = string(summaryJSON)
		}
	} else {
		result.Summary = summary
	}
	task.Summary = result.Summary
	task.GraphStatus = "pending"

	graph, err := generateInferenceGraph(ctx, provider, req, result)
	if err != nil {
		result.GraphStatus = "failed"
		result.GraphError = err.Error()
		task.GraphStatus = "failed"
		task.GraphError = truncateGraphError(err)
	} else {
		result.GraphData = graph
		result.GraphStatus = "completed"
		if graphJSON, err := json.Marshal(graph); err == nil {
			task.GraphData = string(graphJSON)
		}
		task.GraphStatus = "completed"
		task.GraphError = ""
	}

	// 6. 更新任务状态
	e.db.Model(task).Updates(map[string]interface{}{
		"summary":      task.Summary,
		"summary_data": task.SummaryData,
		"graph_data":   task.GraphData,
		"graph_status": task.GraphStatus,
		"graph_error":  task.GraphError,
		"status":       "completed",
	})

	return result, nil
}

// GenerateScenarios 生成多个场景方案
func (e *InferenceEngine) GenerateScenarios(ctx context.Context, req *models.InferenceRequest, count int) ([]*models.InferenceResult, error) {
	results := make([]*models.InferenceResult, count)

	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// 稍微调整参数以产生多样性
			variantReq := *req
			result, err := e.RunInference(ctx, &variantReq)
			if err != nil {
				errChan <- err
				return
			}

			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	if err := <-errChan; err != nil {
		return nil, err
	}

	// 按置信度排序
	// sort.Slice(results, ...)

	return results, nil
}

// ==================== Prompt构建 ====================

func buildStepPrompt(scenario *models.Scenario, stepNumber int, currentState map[string]interface{}) string {
	return fmt.Sprintf(`
你是一个专业的进程发展推理专家。现在需要进行第 %d 步推理。

【场景信息】
- 领域: %s
- 主体: %s
- 关键变化: %s
- 时间范围: %s 到 %s

【当前状态】
%s

请基于以上信息,推理第 %d 步的发展情况。

请以JSON格式输出:
{
    "title": "步骤标题",
    "description": "步骤描述",
    "reasoning": "详细推理过程",
    "confidence": 0.85,
    "state": {
        "updated_variable": "新值"
    }
}

只输出JSON,不要其他内容。
`,
		stepNumber,
		scenario.Domain,
		scenario.Subject,
		scenario.ChangePoint,
		scenario.TimeFrame.Start,
		scenario.TimeFrame.End,
		formatState(currentState),
		stepNumber,
	)
}

func buildSummaryPrompt(result *models.InferenceResult) string {
	stepsJSON, _ := json.Marshal(result.Steps)
	return fmt.Sprintf(`
请对以下推理过程进行总结:

%s

输出格式:
{
    "summary": "整体推理总结",
    "key_findings": ["发现1", "发现2"],
    "recommendations": ["建议1", "建议2"]
}
`, string(stepsJSON))
}

func parseSummaryResponse(response string) *models.InferenceSummary {
	var summary models.InferenceSummary

	jsonStr := extractJSON(response)
	if err := json.Unmarshal([]byte(jsonStr), &summary); err != nil {
		return nil
	}

	if summary.Summary == "" && len(summary.KeyFindings) == 0 && len(summary.Recommendations) == 0 {
		return nil
	}

	return &summary
}

func parseStepResponse(response string, stepNumber int) *models.StepResult {
	var result struct {
		Title       string                 `json:"title"`
		Description string                 `json:"description"`
		Reasoning   string                 `json:"reasoning"`
		Confidence  float64                `json:"confidence"`
		State       map[string]interface{} `json:"state"`
	}

	// 清理响应内容，提取JSON
	jsonStr := extractJSON(response)

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// 解析失败,返回默认值
		return &models.StepResult{
			StepNumber:  stepNumber,
			Title:       fmt.Sprintf("第%d步", stepNumber),
			Description: response,
			Reasoning:   response,
			Confidence:  0.7,
			State:       make(map[string]interface{}),
		}
	}

	return &models.StepResult{
		StepNumber:  stepNumber,
		Title:       result.Title,
		Description: result.Description,
		Reasoning:   result.Reasoning,
		Confidence:  result.Confidence,
		State:       result.State,
	}
}

// extractJSON 从响应中提取JSON内容
func extractJSON(response string) string {
	response = strings.TrimSpace(response)

	// 尝试匹配 ```json ... ``` 或 ``` ... ``` 代码块
	re := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(.*?)\\n?```")
	matches := re.FindStringSubmatch(response)
	if len(matches) > 1 {
		if extracted := extractMostRelevantJSONObject(matches[1]); extracted != "" {
			return extracted
		}
		return strings.TrimSpace(matches[1])
	}

	if extracted := extractMostRelevantJSONObject(response); extracted != "" {
		return extracted
	}

	return response
}

func extractMostRelevantJSONObject(input string) string {
	objects := extractTopLevelJSONObjects(input)
	if len(objects) == 0 {
		return ""
	}
	return objects[len(objects)-1]
}

func extractTopLevelJSONObjects(input string) []string {
	var objects []string
	for start := 0; start < len(input); start++ {
		if input[start] != '{' {
			continue
		}

		end := findJSONObjectEnd(input, start)
		if end == -1 {
			continue
		}

		candidate := strings.TrimSpace(input[start : end+1])
		if json.Valid([]byte(candidate)) {
			objects = append(objects, candidate)
			start = end
		}
	}

	return objects
}

func findJSONObjectEnd(input string, start int) int {
	depth := 0
	inString := false
	escaped := false

	for end := start; end < len(input); end++ {
		ch := input[end]

		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return end
			}
		}
	}

	return -1
}

func formatState(state map[string]interface{}) string {
	if state == nil || len(state) == 0 {
		return "初始状态"
	}
	bytes, _ := json.MarshalIndent(state, "", "  ")
	return string(bytes)
}
