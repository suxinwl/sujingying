package main

import (
	"net/http"
	"os"
	"log"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"suxin/internal/pkg/config"
	"suxin/internal/pkg/database"
	"suxin/internal/appctx"
	"suxin/internal/api/v1"
	"suxin/internal/middleware"
	"suxin/internal/scheduler"
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

	// 启动风控调度器（60秒间隔）
	riskScheduler := scheduler.NewRiskScheduler(app, 60)
	riskScheduler.Start()
	log.Println("[Main] ✅ 风控调度器已启动")

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 公开路由（无需认证）
	api := r.Group("/api/v1")
	v1.RegisterAuthRoutes(api, app)

	// 受保护路由（需要JWT认证）
	protected := api.Group("", middleware.AuthRequired(app))
	v1.RegisterOrderRoutes(protected, app)
	v1.RegisterRiskRoutes(protected, app)

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 启动HTTP服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 在独立协程中启动服务器
	go func() {
		log.Printf("[Main] 🚀 HTTP服务器启动在端口: %s", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("[Main] 启动服务器失败: %v", err)
		}
	}()

	// 优雅退出：监听系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Main] 🛑 收到退出信号，正在关闭服务...")
	riskScheduler.Stop()
	log.Println("[Main] ✅ 服务已关闭")
}
