package api

import (
	"inference-engine/internal/config"
	"inference-engine/internal/engine"
	"inference-engine/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	cfg     *config.Config
	engine  *engine.InferenceEngine
	db      *gorm.DB
	router  *gin.Engine
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
	c.HTML(200, "inference.html", gin.H{
		"title": "创建推理任务",
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
	if req.Model == "" {
		req.Model = "minimax"
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
		"data":    task,
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
