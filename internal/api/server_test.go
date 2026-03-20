package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"inference-engine/internal/config"
	"inference-engine/internal/engine"
	"inference-engine/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubProvider struct{}

func (stubProvider) Chat(ctx context.Context, messages []engine.Message) (string, error) {
	last := messages[len(messages)-1].Content
	if bytes.Contains([]byte(last), []byte("请对以下推理过程进行总结")) {
		return `{"summary":"总体趋势更稳定","key_findings":["政治风险下降"],"recommendations":["继续观察后续改革"]}`, nil
	}

	return `{"title":"第1步","description":"局势开始变化","reasoning":"继承秩序保持稳定","confidence":0.91,"state":{"phase":"stabilized"}}`, nil
}

func (stubProvider) StreamChat(ctx context.Context, messages []engine.Message) (<-chan string, error) {
	ch := make(chan string)
	close(ch)
	return ch, nil
}

func (stubProvider) GetName() string {
	return "stub"
}

func newAPITestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "api-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&models.InferenceTask{}, &models.InferenceStep{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}

func newInferenceRouter(t *testing.T, defaultModel string, providers map[string]engine.ModelProvider) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db := newAPITestDB(t)
	cfg := &config.Config{
		Models: config.ModelsConfig{
			Default: defaultModel,
		},
	}

	mm := engine.NewModelManager(cfg.Models)
	for name, provider := range providers {
		mm.Register(name, provider)
	}

	inferenceEngine := engine.NewInferenceEngine(mm, db)
	server := &Server{
		cfg:    cfg,
		engine: inferenceEngine,
		db:     db,
		router: gin.New(),
	}
	server.router.POST("/api/inference", server.runInference)
	server.router.GET("/api/inference/:id", server.getInference)

	return server.router
}

func TestRunInferenceUsesConfiguredDefaultModelWhenRequestModelEmpty(t *testing.T) {
	router := newInferenceRouter(t, "deepseek", map[string]engine.ModelProvider{
		"deepseek": stubProvider{},
	})

	body := map[string]interface{}{
		"title":        "默认模型测试",
		"domain":       "技术",
		"subject":      "编程行业",
		"change_point": "AI 继续提升",
		"time_frame": map[string]string{
			"start": "2024年",
			"end":   "2030年",
		},
		"steps_count": 1,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/inference", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Success bool                   `json:"success"`
		Error   string                 `json:"error"`
		Data    models.InferenceResult `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !response.Success {
		t.Fatalf("expected success response, got error: %s", response.Error)
	}
	if response.Data.Summary != "总体趋势更稳定" {
		t.Fatalf("expected parsed summary text, got %q", response.Data.Summary)
	}
	if response.Data.SummaryData == nil {
		t.Fatal("expected structured summary data in response")
	}
	if len(response.Data.SummaryData.KeyFindings) != 1 || response.Data.SummaryData.KeyFindings[0] != "政治风险下降" {
		t.Fatalf("unexpected structured summary: %#v", response.Data.SummaryData)
	}
	if response.Data.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be populated")
	}
	if response.Data.CreatedAt.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected a realistic created_at, got %s", response.Data.CreatedAt)
	}
}

func TestBuildModelOptionGroupsFiltersAndMarksDefault(t *testing.T) {
	groups := buildModelOptionGroups([]string{"deepseek", "qwen"}, "deepseek")

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	seen := map[string]bool{}
	selected := ""
	for _, group := range groups {
		for _, option := range group.Options {
			seen[option.Value] = true
			if option.Selected {
				selected = option.Value
			}
		}
	}

	if !seen["deepseek"] || !seen["qwen"] {
		t.Fatalf("expected deepseek and qwen options, got %#v", seen)
	}
	if seen["minimax"] || seen["claude"] {
		t.Fatalf("expected unavailable models to be filtered out, got %#v", seen)
	}
	if selected != "deepseek" {
		t.Fatalf("expected default model to be selected, got %q", selected)
	}
}

func TestGetInferenceReturnsSummaryAndGraphData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newAPITestDB(t)
	cfg := &config.Config{}
	inferenceEngine := engine.NewInferenceEngine(&engine.ModelManager{}, db)
	server := &Server{
		cfg:    cfg,
		engine: inferenceEngine,
		db:     db,
		router: gin.New(),
	}
	server.router.GET("/api/inference/:id", server.getInference)

	task := models.InferenceTask{
		Title:       "详情测试",
		Domain:      "历史",
		Subject:     "秦朝",
		ChangePoint: "扶苏继位",
		TimeFrame:   "前210年 - 前180年",
		ModelUsed:   "minimax",
		StepsCount:  1,
		Status:      "completed",
		Summary:     "总体趋势更稳定",
		SummaryData: `{"summary":"总体趋势更稳定","key_findings":["政治风险下降"],"recommendations":["继续观察后续改革"]}`,
		GraphData:   `{"nodes":[{"id":"fact-1","label":"既有秩序稳定","type":"fact"},{"id":"reasoning-1","label":"局势逐步稳定","type":"reasoning"}],"edges":[{"source":"fact-1","target":"reasoning-1","label":"基于"}]}`,
		GraphStatus: "completed",
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}
	step := models.InferenceStep{
		TaskID:      task.ID,
		StepNumber:  1,
		Title:       "第1步",
		Description: "局势开始变化",
		Reasoning:   "继承秩序保持稳定",
		Confidence:  0.91,
		ModelUsed:   "minimax",
	}
	if err := db.Create(&step).Error; err != nil {
		t.Fatalf("create step: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/inference/"+strconv.FormatUint(uint64(task.ID), 10), nil)
	recorder := httptest.NewRecorder()

	server.router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
		Data    struct {
			ID          uint                     `json:"id"`
			Summary     string                   `json:"summary"`
			SummaryData *models.InferenceSummary `json:"summary_data"`
			GraphData   *models.InferenceGraph   `json:"graph_data"`
			GraphStatus string                   `json:"graph_status"`
			Steps       []models.InferenceStep   `json:"steps"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode detail response: %v", err)
	}

	if !response.Success {
		t.Fatalf("expected success response, got error: %s", response.Error)
	}
	if response.Data.Summary != "总体趋势更稳定" {
		t.Fatalf("unexpected summary: %q", response.Data.Summary)
	}
	if response.Data.SummaryData == nil || response.Data.SummaryData.Summary != "总体趋势更稳定" {
		t.Fatalf("expected parsed summary_data, got %#v", response.Data.SummaryData)
	}
	if response.Data.GraphData == nil || len(response.Data.GraphData.Nodes) != 2 {
		t.Fatalf("expected parsed graph_data, got %#v", response.Data.GraphData)
	}
	if response.Data.GraphStatus != "completed" {
		t.Fatalf("expected graph status completed, got %q", response.Data.GraphStatus)
	}
	if len(response.Data.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(response.Data.Steps))
	}
}
