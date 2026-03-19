package main

import (
	"log"
	"inference-engine/internal/api"
	"inference-engine/internal/config"
	"inference-engine/internal/engine"
	"inference-engine/internal/models"
	"inference-engine/internal/storage"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	db, err := storage.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 自动迁移
	db.AutoMigrate(&models.User{}, &models.InferenceTask{}, &models.InferenceStep{}, &models.ModelConfig{})

	// 初始化模型管理器
	modelManager := engine.NewModelManager(cfg.Models)

	// 初始化推理引擎
	inferenceEngine := engine.NewInferenceEngine(modelManager, db)

	// 启动API服务
	server := api.NewServer(cfg, inferenceEngine, db)
	
	log.Printf("🚀 服务启动在 http://localhost:%s", cfg.Server.Port)
	if err := server.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
