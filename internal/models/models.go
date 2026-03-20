package models

import (
	"gorm.io/gorm"
	"time"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;size:50"`
	Email     string         `json:"email" gorm:"uniqueIndex;size:100"`
	Password  string         `json:"-" gorm:"size:255"`
	APIKey    string         `json:"-" gorm:"size:100"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// InferenceTask 推理任务
type InferenceTask struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"index"`
	Title       string         `json:"title" gorm:"size:200"`
	Domain      string         `json:"domain" gorm:"size:50"`      // 历史、商业、技术、个人等
	Subject     string         `json:"subject" gorm:"size:200"`    // 推理主体
	ChangePoint string         `json:"change_point" gorm:"text"`   // 关键变化点
	TimeFrame   string         `json:"time_frame" gorm:"size:100"` // 时间范围
	Variables   string         `json:"variables" gorm:"type:json"` // JSON格式变量
	ModelUsed   string         `json:"model_used" gorm:"size:50"`
	StepsCount  int            `json:"steps_count"`
	Status      string         `json:"status" gorm:"size:20"` // pending, processing, completed, failed
	Summary     string         `json:"summary" gorm:"type:text"`
	SummaryData string         `json:"summary_data" gorm:"type:json"`
	GraphData   string         `json:"graph_data" gorm:"type:json"`
	GraphStatus string         `json:"graph_status" gorm:"size:20"`
	GraphError  string         `json:"graph_error" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Steps []InferenceStep `json:"steps,omitempty" gorm:"foreignKey:TaskID"`
}

// InferenceStep 推理步骤
type InferenceStep struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TaskID      uint      `json:"task_id" gorm:"index"`
	StepNumber  int       `json:"step_number"`
	Title       string    `json:"title" gorm:"size:200"`
	Description string    `json:"description" gorm:"text"`
	InputState  string    `json:"input_state" gorm:"type:json"`  // JSON
	OutputState string    `json:"output_state" gorm:"type:json"` // JSON
	Reasoning   string    `json:"reasoning" gorm:"type:text"`    // AI推理过程
	Confidence  float64   `json:"confidence"`
	ModelUsed   string    `json:"model_used" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
}

// ModelConfig 模型配置存储
type ModelConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex;size:50"`
	Provider  string    `json:"provider" gorm:"size:50"`
	APIKey    string    `json:"-" gorm:"size:255"`
	BaseURL   string    `json:"base_url" gorm:"size:255"`
	ModelName string    `json:"model_name" gorm:"size:100"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Scenario 场景定义
type Scenario struct {
	Domain      string                 `json:"domain"`
	Subject     string                 `json:"subject"`
	ChangePoint string                 `json:"change_point"`
	TimeFrame   TimeFrame              `json:"time_frame"`
	Variables   map[string]interface{} `json:"variables"`
	Constraints []string               `json:"constraints"`
}

// TimeFrame 时间范围
type TimeFrame struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// InferenceRequest 推理请求
type InferenceRequest struct {
	Title       string                 `json:"title"`
	Domain      string                 `json:"domain"`
	Subject     string                 `json:"subject"`
	ChangePoint string                 `json:"change_point"`
	TimeFrame   TimeFrame              `json:"time_frame"`
	Variables   map[string]interface{} `json:"variables"`
	StepsCount  int                    `json:"steps_count"`
	Model       string                 `json:"model"` // deepseek, claude, gpt
}

// InferenceResult 推理结果
type InferenceResult struct {
	TaskID      uint              `json:"task_id"`
	Title       string            `json:"title"`
	Steps       []StepResult      `json:"steps"`
	Summary     string            `json:"summary"`
	SummaryData *InferenceSummary `json:"summary_data,omitempty"`
	GraphData   *InferenceGraph   `json:"graph_data,omitempty"`
	GraphStatus string            `json:"graph_status,omitempty"`
	GraphError  string            `json:"graph_error,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// InferenceSummary 结构化总结
type InferenceSummary struct {
	Summary         string   `json:"summary"`
	KeyFindings     []string `json:"key_findings"`
	Recommendations []string `json:"recommendations"`
}

// InferenceGraph 整次推理的拓扑图
type InferenceGraph struct {
	Nodes []InferenceGraphNode `json:"nodes"`
	Edges []InferenceGraphEdge `json:"edges"`
}

// InferenceGraphNode 拓扑图节点
type InferenceGraphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// InferenceGraphEdge 拓扑图边
type InferenceGraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

// StepResult 步骤结果
type StepResult struct {
	StepNumber  int                    `json:"step_number"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Reasoning   string                 `json:"reasoning"`
	Confidence  float64                `json:"confidence"`
	State       map[string]interface{} `json:"state"`
}
