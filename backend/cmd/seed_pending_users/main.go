package main

import (
	"fmt"
	"log"

	"suxin/internal/appctx"
	"suxin/internal/pkg/config"
	"suxin/internal/pkg/database"
	"suxin/internal/service"
)

func main() {
	// 加载配置，与主服务保持一致
	env := config.AppEnv()
	cfg, err := config.Load(env)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// 连接数据库
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connect db failed: %v", err)
	}

	app := appctx.New(db, cfg)
	authSvc := service.NewAuthService(app)

	// 批量创建 5 个待审核用户
	for i := 1; i <= 5; i++ {
		phone := fmt.Sprintf("139900000%02d", i) // 13990000001 ~ 13990000005
		user, err := authSvc.Register(phone, "123456", "customer")
		if err != nil {
			log.Printf("create user %s failed: %v", phone, err)
			continue
		}
		log.Printf("created pending user: id=%d, phone=%s, status=%s", user.ID, user.Phone, user.Status)
	}
}
