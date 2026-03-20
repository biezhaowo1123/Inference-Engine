package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"inference-engine/internal/models"
	"strings"
)

func buildGraphPrompt(req *models.InferenceRequest, result *models.InferenceResult) string {
	stepsJSON, _ := json.Marshal(result.Steps)
	summaryJSON, _ := json.Marshal(result.SummaryData)

	return fmt.Sprintf(`
请基于以下整次推理结果，生成一张“事实节点”和“推理节点”共同组成的拓扑图。

【场景信息】
- 标题: %s
- 领域: %s
- 主体: %s
- 关键变化: %s

【步骤结果】
%s

【总结信息】
%s

输出要求:
1. 只输出 JSON
2. 节点类型只允许 "fact" 或 "reasoning"
3. 每个节点必须包含 id、label、type
4. 每条边必须包含 source、target、label
5. 不要输出 markdown 代码块

输出格式:
{
  "nodes": [
    {"id": "fact-1", "label": "事实节点", "type": "fact"},
    {"id": "reasoning-1", "label": "推理节点", "type": "reasoning"}
  ],
  "edges": [
    {"source": "fact-1", "target": "reasoning-1", "label": "基于"}
  ]
}
`,
		req.Title,
		req.Domain,
		req.Subject,
		req.ChangePoint,
		string(stepsJSON),
		string(summaryJSON),
	)
}

func generateInferenceGraph(ctx context.Context, provider ModelProvider, req *models.InferenceRequest, result *models.InferenceResult) (*models.InferenceGraph, error) {
	response, err := provider.Chat(ctx, []Message{
		{Role: "user", Content: buildGraphPrompt(req, result)},
	})
	if err != nil {
		return nil, err
	}

	return parseGraphResponse(response)
}

func parseGraphResponse(raw string) (*models.InferenceGraph, error) {
	var graph models.InferenceGraph

	jsonStr := extractJSON(raw)
	if err := json.Unmarshal([]byte(jsonStr), &graph); err != nil {
		return nil, fmt.Errorf("解析图谱失败: %w", err)
	}
	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("图谱缺少节点")
	}

	nodeIDs := make(map[string]struct{}, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if strings.TrimSpace(node.ID) == "" {
			return nil, fmt.Errorf("图谱节点缺少 id")
		}
		if strings.TrimSpace(node.Label) == "" {
			return nil, fmt.Errorf("图谱节点缺少 label")
		}
		if node.Type != "fact" && node.Type != "reasoning" {
			return nil, fmt.Errorf("图谱节点类型非法: %s", node.Type)
		}
		nodeIDs[node.ID] = struct{}{}
	}

	for _, edge := range graph.Edges {
		if strings.TrimSpace(edge.Source) == "" || strings.TrimSpace(edge.Target) == "" {
			return nil, fmt.Errorf("图谱边缺少 source 或 target")
		}
		if _, ok := nodeIDs[edge.Source]; !ok {
			return nil, fmt.Errorf("图谱边 source 不存在: %s", edge.Source)
		}
		if _, ok := nodeIDs[edge.Target]; !ok {
			return nil, fmt.Errorf("图谱边 target 不存在: %s", edge.Target)
		}
	}

	return &graph, nil
}

func truncateGraphError(err error) string {
	if err == nil {
		return ""
	}
	text := err.Error()
	if len(text) > 200 {
		return text[:200]
	}
	return text
}
