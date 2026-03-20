package engine

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"inference-engine/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newEngineTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "engine-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&models.InferenceTask{}, &models.InferenceStep{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}

type scriptedProvider struct {
	name      string
	responses []string
	errs      []error
	calls     int
}

func (p *scriptedProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	if p.calls < len(p.errs) && p.errs[p.calls] != nil {
		err := p.errs[p.calls]
		p.calls++
		return "", err
	}
	if p.calls >= len(p.responses) {
		p.calls++
		return "", errors.New("unexpected provider call")
	}
	response := p.responses[p.calls]
	p.calls++
	return response, nil
}

func (p *scriptedProvider) StreamChat(ctx context.Context, messages []Message) (<-chan string, error) {
	ch := make(chan string)
	close(ch)
	return ch, nil
}

func (p *scriptedProvider) GetName() string {
	if p.name != "" {
		return p.name
	}
	return "scripted"
}

func TestRunInferenceDoesNotCreateTaskWhenProviderUnavailable(t *testing.T) {
	db := newEngineTestDB(t)
	mm := &ModelManager{
		providers:    map[string]ModelProvider{},
		defaultModel: "deepseek",
	}
	eng := NewInferenceEngine(mm, db)

	_, err := eng.RunInference(context.Background(), &models.InferenceRequest{
		Title:       "测试任务",
		Domain:      "历史",
		Subject:     "秦朝",
		ChangePoint: "扶苏继位",
		TimeFrame: models.TimeFrame{
			Start: "前210年",
			End:   "前180年",
		},
		StepsCount: 1,
		Model:      "missing-model",
	})
	if err == nil {
		t.Fatal("expected missing provider error")
	}

	var count int64
	if err := db.Model(&models.InferenceTask{}).Count(&count).Error; err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no task records to be created, got %d", count)
	}
}

func TestParseSummaryResponseParsesStructuredSummary(t *testing.T) {
	summary := parseSummaryResponse("```json\n{\n  \"summary\": \"秦朝政局会更稳定\",\n  \"key_findings\": [\"中央集权更平稳\"],\n  \"recommendations\": [\"继续观察地方治理\"]\n}\n```")
	if summary == nil {
		t.Fatal("expected structured summary to be parsed")
	}
	if summary.Summary != "秦朝政局会更稳定" {
		t.Fatalf("unexpected summary: %q", summary.Summary)
	}
	if len(summary.KeyFindings) != 1 || summary.KeyFindings[0] != "中央集权更平稳" {
		t.Fatalf("unexpected key findings: %#v", summary.KeyFindings)
	}
	if len(summary.Recommendations) != 1 || summary.Recommendations[0] != "继续观察地方治理" {
		t.Fatalf("unexpected recommendations: %#v", summary.Recommendations)
	}
}

func TestRunInferencePersistsSummaryAndGraphData(t *testing.T) {
	db := newEngineTestDB(t)
	provider := &scriptedProvider{
		name: "graph-provider",
		responses: []string{
			`{"title":"第1步","description":"局势开始变化","reasoning":"继承秩序保持稳定","confidence":0.91,"state":{"phase":"stabilized"}}`,
			`{"summary":"总体趋势更稳定","key_findings":["政治风险下降"],"recommendations":["继续观察后续改革"]}`,
			`{"nodes":[{"id":"fact-1","label":"既有秩序稳定","type":"fact"},{"id":"reasoning-1","label":"局势逐步稳定","type":"reasoning"}],"edges":[{"source":"fact-1","target":"reasoning-1","label":"基于"}]}`,
		},
	}
	mm := &ModelManager{
		providers: map[string]ModelProvider{
			"graph-provider": provider,
		},
		defaultModel: "graph-provider",
	}
	eng := NewInferenceEngine(mm, db)

	result, err := eng.RunInference(context.Background(), &models.InferenceRequest{
		Title:       "图谱持久化测试",
		Domain:      "历史",
		Subject:     "秦朝",
		ChangePoint: "扶苏继位",
		TimeFrame: models.TimeFrame{
			Start: "前210年",
			End:   "前180年",
		},
		StepsCount: 1,
		Model:      "graph-provider",
	})
	if err != nil {
		t.Fatalf("run inference: %v", err)
	}

	var stored struct {
		Summary     string
		SummaryData string
		GraphData   string
		GraphStatus string
	}
	if err := db.Raw(
		"SELECT summary, summary_data, graph_data, graph_status FROM inference_tasks WHERE id = ?",
		result.TaskID,
	).Scan(&stored).Error; err != nil {
		t.Fatalf("load persisted fields: %v", err)
	}
	if stored.Summary != "总体趋势更稳定" {
		t.Fatalf("unexpected persisted summary: %q", stored.Summary)
	}
	if stored.SummaryData == "" {
		t.Fatal("expected persisted summary_data")
	}
	if stored.GraphData == "" {
		t.Fatal("expected persisted graph_data")
	}
	if stored.GraphStatus != "completed" {
		t.Fatalf("expected graph status completed, got %q", stored.GraphStatus)
	}
}

func TestRunInferenceKeepsTaskCompletedWhenGraphGenerationFails(t *testing.T) {
	db := newEngineTestDB(t)
	provider := &scriptedProvider{
		name: "graph-provider",
		responses: []string{
			`{"title":"第1步","description":"局势开始变化","reasoning":"继承秩序保持稳定","confidence":0.91,"state":{"phase":"stabilized"}}`,
			`{"summary":"总体趋势更稳定","key_findings":["政治风险下降"],"recommendations":["继续观察后续改革"]}`,
			`{"nodes":[{"id":"fact-1","label":"坏图谱","type":"fact"}],"edges":[{"source":"fact-1","target":"missing-node","label":"基于"}]}`,
		},
	}
	mm := &ModelManager{
		providers: map[string]ModelProvider{
			"graph-provider": provider,
		},
		defaultModel: "graph-provider",
	}
	eng := NewInferenceEngine(mm, db)

	result, err := eng.RunInference(context.Background(), &models.InferenceRequest{
		Title:       "图谱失败测试",
		Domain:      "历史",
		Subject:     "秦朝",
		ChangePoint: "扶苏继位",
		TimeFrame: models.TimeFrame{
			Start: "前210年",
			End:   "前180年",
		},
		StepsCount: 1,
		Model:      "graph-provider",
	})
	if err != nil {
		t.Fatalf("expected inference to succeed when graph generation fails, got %v", err)
	}
	if result == nil {
		t.Fatal("expected inference result")
	}

	var stored struct {
		Status      string
		Summary     string
		GraphStatus string
		GraphError  string
	}
	if err := db.Raw(
		"SELECT status, summary, graph_status, graph_error FROM inference_tasks WHERE id = ?",
		result.TaskID,
	).Scan(&stored).Error; err != nil {
		t.Fatalf("load persisted graph failure fields: %v", err)
	}
	if stored.Status != "completed" {
		t.Fatalf("expected task status completed, got %q", stored.Status)
	}
	if stored.Summary != "总体趋势更稳定" {
		t.Fatalf("expected persisted summary on graph failure, got %q", stored.Summary)
	}
	if stored.GraphStatus != "failed" {
		t.Fatalf("expected graph status failed, got %q", stored.GraphStatus)
	}
	if stored.GraphError == "" {
		t.Fatal("expected graph error to be persisted")
	}
}

func TestParseGraphResponseSkipsPreambleJSONExample(t *testing.T) {
	raw := `<think>
我会先给出一个示例 {"nodes":[{"id":"example","label":"示例","type":"fact"}],"edges":[]}
然后再输出最终结果。
</think>

{"nodes":[{"id":"fact-1","label":"既有秩序稳定","type":"fact"},{"id":"reasoning-1","label":"局势逐步稳定","type":"reasoning"}],"edges":[{"source":"fact-1","target":"reasoning-1","label":"基于"}]}`

	graph, err := parseGraphResponse(raw)
	if err != nil {
		t.Fatalf("expected graph response to parse, got %v", err)
	}
	if graph == nil {
		t.Fatal("expected graph")
	}
	if len(graph.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(graph.Nodes))
	}
	if graph.Nodes[0].ID != "fact-1" {
		t.Fatalf("expected first parsed node to be fact-1, got %q", graph.Nodes[0].ID)
	}
}
