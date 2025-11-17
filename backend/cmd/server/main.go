package main

import (
	"net/http"
	"os"
	"log"

	"github.com/gin-gonic/gin"

	"suxin/internal/pkg/config"
	"suxin/internal/pkg/database"
	"suxin/internal/appctx"
	"suxin/internal/api/v1"
)

func main() {
	// 加载配置
	env := config.AppEnv()
	cfg, err := config.Load(env)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// 初始化数据库并自动迁移
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connect db failed: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	app := appctx.New(db, cfg)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api/v1")
	v1.RegisterAuthRoutes(api, app)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	_ = r.Run(":" + port)
}
