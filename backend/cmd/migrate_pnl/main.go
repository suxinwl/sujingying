package main

import (
	"log"

	"suxin/internal/pkg/config"
	"suxin/internal/pkg/database"
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

	// 执行一次性迁移：为 orders 表增加 pnl_float 列
	// SQLite: REAL DEFAULT 0
	if err := db.Exec("ALTER TABLE orders ADD COLUMN pnl_float REAL DEFAULT 0;").Error; err != nil {
		log.Fatalf("migrate pnl_float failed: %v", err)
	}

	log.Println("[Migrate] ✅ 添加 pnl_float 列成功")
}
