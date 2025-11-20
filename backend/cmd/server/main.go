package main

import (
	"net/http"
	"os"
	"log"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"

	"suxin/internal/pkg/config"
	"suxin/internal/pkg/database"
	"suxin/internal/appctx"
	"suxin/internal/api/v1"
	"suxin/internal/middleware"
	"suxin/internal/scheduler"
	"suxin/internal/service"
	ws "suxin/internal/websocket"
	
	"github.com/gorilla/websocket"
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

	// 启动WebSocket行情代理（上海黄金交易所）
	quoteHub := ws.NewQuoteProxyHub()
	go quoteHub.Run()
	log.Println("[Main] ✅ WebSocket行情代理已启动（数据源: 上海黄金交易所）")
	
	// 启动WebSocket通知推送中心
	notificationHub := ws.NewNotificationHub()
	go notificationHub.Run()
	log.Println("[Main] ✅ WebSocket通知推送中心已启动")
	service.SetDefaultNotificationHub(notificationHub)

	// 启动风控调度器（15秒间隔，使用WebSocket价格）
	riskScheduler := scheduler.NewRiskScheduler(app, 15, quoteHub)
	riskScheduler.Start()
	log.Println("[Main] ✅ 风控调度器已启动（间隔: 15秒，价格来源: WebSocket实时数据）")

	// WebSocket升级器
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源（生产环境需要限制）
		},
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	
	// 配置CORS（跨域资源共享）
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:5175",
			"http://127.0.0.1:5173",
			"http://localhost:8091",
			"http://127.0.0.1:8091",
			// 服务IP前端访问
			"http://192.168.10.8",
			"http://192.168.10.8:8091",
			"http://192.168.2.10",
			"http://192.168.2.10:8091",
			"http://59.36.165.33",
			"http://59.36.165.33:8091",
			"http://59.36.165.33:5173",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12小时
	}))

	// 公开路由（无需认证）
	api := r.Group("/api/v1")
	v1.RegisterAuthRoutes(api, app)

	// 受保护路由（需要JWT认证）
	protected := api.Group("", middleware.AuthRequired(app))
	v1.RegisterOrderRoutes(protected, app)
	v1.RegisterRiskRoutes(protected, app)
	v1.RegisterNotificationRoutes(protected, app)
	v1.RegisterBankCardRoutes(protected, app)
	v1.RegisterSalesRoutes(protected, app)
	v1.RegisterDepositRoutes(protected, app)
	v1.RegisterWithdrawRoutes(protected, app)
	v1.RegisterFundLogRoutes(protected, app)
	v1.RegisterUserManageRoutes(protected, app)
	v1.RegisterConfigRoutes(protected, app)
	v1.RegisterSupplementRoutes(protected, app)
	v1.RegisterInvitationRoutes(protected, app)

	// WebSocket行情代理接口
	r.GET("/ws/quote", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[WebSocket] 升级连接失败: %v", err)
			return
		}
		quoteHub.ServeWs(conn)
	})
	
	// WebSocket通知推送接口（需要JWT认证）
	r.GET("/ws/notification", middleware.AuthRequired(app), func(c *gin.Context) {
		userID := c.GetUint("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}
		
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[WebSocket] 通知连接升级失败: %v", err)
			return
		}
		
		notificationHub.ServeWs(userID, conn)
		log.Printf("[WebSocket] 用户 %d 建立通知连接", userID)
	})

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 启动HTTP服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
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
