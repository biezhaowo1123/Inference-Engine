package api

import (
	"inference-engine/internal/config"
	"inference-engine/internal/engine"
	"inference-engine/internal/models"
	"sort"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	cfg    *config.Config
	engine *engine.InferenceEngine
	db     *gorm.DB
	router *gin.Engine
}

type modelOption struct {
	Value    string
	Label    string
	Selected bool
}

type modelOptionGroup struct {
	Label   string
	Options []modelOption
}

type modelOptionMeta struct {
	Value string
	Label string
	Group string
}

var modelOptionCatalog = []modelOptionMeta{
	{Value: "minimax", Label: "MiniMax（推荐）", Group: "国内创业"},
	{Value: "moonshot", Label: "Moonshot（长文本）", Group: "国内创业"},
	{Value: "baichuan", Label: "百川", Group: "国内创业"},
	{Value: "yi", Label: "零一万物", Group: "国内创业"},
	{Value: "deepseek", Label: "DeepSeek（便宜）", Group: "国际主流"},
	{Value: "claude", Label: "Claude（推理强）", Group: "国际主流"},
	{Value: "gpt", Label: "GPT-4（综合强）", Group: "国际主流"},
	{Value: "gemini", Label: "Gemini（免费）", Group: "国际主流"},
	{Value: "grok", Label: "Grok", Group: "国际主流"},
	{Value: "llama", Label: "Llama", Group: "国际主流"},
	{Value: "mistral", Label: "Mistral", Group: "国际主流"},
	{Value: "qwen", Label: "通义千问", Group: "国内大厂"},
	{Value: "glm", Label: "智谱GLM", Group: "国内大厂"},
	{Value: "wenxin", Label: "文心一言", Group: "国内大厂"},
	{Value: "hunyuan", Label: "混元", Group: "国内大厂"},
	{Value: "spark", Label: "星火", Group: "国内大厂"},
	{Value: "doubao", Label: "豆包", Group: "国内大厂"},
	{Value: "perplexity", Label: "Perplexity", Group: "其他平台"},
	{Value: "cohere", Label: "Cohere", Group: "其他平台"},
	{Value: "together", Label: "Together AI", Group: "其他平台"},
	{Value: "openrouter", Label: "OpenRouter", Group: "其他平台"},
	{Value: "ollama", Label: "Ollama", Group: "本地模型（免费）"},
	{Value: "vllm", Label: "vLLM", Group: "本地模型（免费）"},
	{Value: "localai", Label: "LocalAI", Group: "本地模型（免费）"},
}

func NewServer(cfg *config.Config, engine *engine.InferenceEngine, db *gorm.DB) *gin.Engine {
	s := &Server{
		cfg:    cfg,
		engine: engine,
		db:     db,
		router: gin.Default(),
	}

	// 设置路由
	s.setupRoutes()

	return s.router
}

func (s *Server) setupRoutes() {
	// 静态文件
	s.router.Static("/static", "./web/static")
	s.router.LoadHTMLGlob("./web/templates/*")

	// 页面路由
	s.router.GET("/", s.indexPage)
	s.router.GET("/inference", s.inferencePage)
	s.router.GET("/history", s.historyPage)

	// API路由
	api := s.router.Group("/api")
	{
		api.POST("/inference", s.runInference)
		api.GET("/inference/:id", s.getInference)
		api.GET("/history", s.getHistory)
		api.DELETE("/history/:id", s.deleteHistory)
		api.DELETE("/history", s.clearHistory)
		api.GET("/models", s.getModels)
	}
}

// ==================== 页面处理 ====================

func (s *Server) indexPage(c *gin.Context) {
	c.HTML(200, "index.html", gin.H{
		"title": "进程推理引擎",
	})
}

func (s *Server) inferencePage(c *gin.Context) {
	modelGroups := buildModelOptionGroups(s.engine.GetAvailableModels(), s.cfg.Models.Default)
	c.HTML(200, "inference.html", gin.H{
		"title":       "创建推理任务",
		"ModelGroups": modelGroups,
		"HasModels":   len(modelGroups) > 0,
	})
}

func (s *Server) historyPage(c *gin.Context) {
	c.HTML(200, "history.html", gin.H{
		"title": "推理历史",
	})
}

// ==================== API处理 ====================

// runInference 执行推理
func (s *Server) runInference(c *gin.Context) {
	var req models.InferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 默认值
	if req.StepsCount == 0 {
		req.StepsCount = 5
	}

	result, err := s.engine.RunInference(c.Request.Context(), &req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    result,
	})
}

// getInference 获取推理详情
func (s *Server) getInference(c *gin.Context) {
	taskID := c.Param("id")

	var task models.InferenceTask
	if err := s.db.Preload("Steps").First(&task, taskID).Error; err != nil {
		c.JSON(404, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    newInferenceTaskDetailResponse(task),
	})
}

// getHistory 获取历史记录
func (s *Server) getHistory(c *gin.Context) {
	var tasks []models.InferenceTask
	s.db.Order("created_at desc").Limit(50).Find(&tasks)

	c.JSON(200, gin.H{
		"success": true,
		"data":    tasks,
	})
}

// getModels 获取可用模型
func (s *Server) getModels(c *gin.Context) {
	models := s.engine.GetAvailableModels()
	c.JSON(200, gin.H{
		"success": true,
		"data":    models,
	})
}

// deleteHistory 删除单条历史记录
func (s *Server) deleteHistory(c *gin.Context) {
	taskID := c.Param("id")

	// 先删除关联的步骤
	s.db.Where("task_id = ?", taskID).Delete(&models.InferenceStep{})

	// 再删除任务
	result := s.db.Delete(&models.InferenceTask{}, taskID)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

// clearHistory 清空所有历史记录
func (s *Server) clearHistory(c *gin.Context) {
	// 先删除所有步骤
	s.db.Where("1 = 1").Delete(&models.InferenceStep{})

	// 再删除所有任务
	s.db.Where("1 = 1").Delete(&models.InferenceTask{})

	c.JSON(200, gin.H{
		"success": true,
		"message": "已清空所有历史记录",
	})
}

func buildModelOptionGroups(available []string, defaultModel string) []modelOptionGroup {
	availableSet := make(map[string]struct{}, len(available))
	for _, name := range available {
		availableSet[name] = struct{}{}
	}

	selectedModel := defaultModel
	if _, ok := availableSet[selectedModel]; !ok {
		selectedModel = ""
		for _, meta := range modelOptionCatalog {
			if _, ok := availableSet[meta.Value]; ok {
				selectedModel = meta.Value
				break
			}
		}
		if selectedModel == "" && len(available) > 0 {
			fallback := append([]string(nil), available...)
			sort.Strings(fallback)
			selectedModel = fallback[0]
		}
	}

	groups := make([]modelOptionGroup, 0)
	groupIndexes := make(map[string]int)
	seen := make(map[string]bool, len(available))

	for _, meta := range modelOptionCatalog {
		if _, ok := availableSet[meta.Value]; !ok {
			continue
		}

		groupIndex, ok := groupIndexes[meta.Group]
		if !ok {
			groupIndex = len(groups)
			groupIndexes[meta.Group] = groupIndex
			groups = append(groups, modelOptionGroup{Label: meta.Group})
		}

		groups[groupIndex].Options = append(groups[groupIndex].Options, modelOption{
			Value:    meta.Value,
			Label:    meta.Label,
			Selected: meta.Value == selectedModel,
		})
		seen[meta.Value] = true
	}

	unknown := make([]string, 0)
	for _, name := range available {
		if !seen[name] {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(unknown)
	if len(unknown) > 0 {
		group := modelOptionGroup{Label: "其他"}
		for _, name := range unknown {
			group.Options = append(group.Options, modelOption{
				Value:    name,
				Label:    name,
				Selected: name == selectedModel,
			})
		}
		groups = append(groups, group)
	}

	return groups
}
