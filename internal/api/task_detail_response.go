package api

import (
	"encoding/json"
	"inference-engine/internal/models"
	"time"
)

type inferenceTaskDetailResponse struct {
	ID          uint                     `json:"id"`
	Title       string                   `json:"title"`
	Domain      string                   `json:"domain"`
	Subject     string                   `json:"subject"`
	ChangePoint string                   `json:"change_point"`
	TimeFrame   string                   `json:"time_frame"`
	ModelUsed   string                   `json:"model_used"`
	StepsCount  int                      `json:"steps_count"`
	Status      string                   `json:"status"`
	Summary     string                   `json:"summary"`
	SummaryData *models.InferenceSummary `json:"summary_data,omitempty"`
	GraphData   *models.InferenceGraph   `json:"graph_data,omitempty"`
	GraphStatus string                   `json:"graph_status"`
	GraphError  string                   `json:"graph_error,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	Steps       []models.InferenceStep   `json:"steps,omitempty"`
}

func newInferenceTaskDetailResponse(task models.InferenceTask) inferenceTaskDetailResponse {
	return inferenceTaskDetailResponse{
		ID:          task.ID,
		Title:       task.Title,
		Domain:      task.Domain,
		Subject:     task.Subject,
		ChangePoint: task.ChangePoint,
		TimeFrame:   task.TimeFrame,
		ModelUsed:   task.ModelUsed,
		StepsCount:  task.StepsCount,
		Status:      task.Status,
		Summary:     task.Summary,
		SummaryData: parsePersistedSummary(task.SummaryData),
		GraphData:   parsePersistedGraph(task.GraphData),
		GraphStatus: task.GraphStatus,
		GraphError:  task.GraphError,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		Steps:       task.Steps,
	}
}

func parsePersistedSummary(raw string) *models.InferenceSummary {
	if raw == "" {
		return nil
	}

	var summary models.InferenceSummary
	if err := json.Unmarshal([]byte(raw), &summary); err != nil {
		return nil
	}
	if summary.Summary == "" && len(summary.KeyFindings) == 0 && len(summary.Recommendations) == 0 {
		return nil
	}

	return &summary
}

func parsePersistedGraph(raw string) *models.InferenceGraph {
	if raw == "" {
		return nil
	}

	var graph models.InferenceGraph
	if err := json.Unmarshal([]byte(raw), &graph); err != nil {
		return nil
	}
	if len(graph.Nodes) == 0 {
		return nil
	}

	return &graph
}
