package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"suxin/internal/pkg/config"
	"suxin/internal/model"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dbType := cfg.Database.Type
	switch dbType {
	case "sqlite":
		// 确保SQLite父目录存在
		p := cfg.Database.Sqlite.Path
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return nil, err
		}
		return gorm.Open(sqlite.Open(p), &gorm.Config{})
	case "mysql":
		mysqlCfg := cfg.Database.MySQL
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
			mysqlCfg.User,
			mysqlCfg.Password,
			mysqlCfg.Host,
			mysqlCfg.Port,
			mysqlCfg.Database,
			mysqlCfg.Charset,
			mysqlCfg.Loc,
		)
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.LoginLog{},
		&model.Order{},
		&model.RiskConfig{},
		&model.Notification{},
		&model.BankCard{},
		&model.DepositRequest{},
		&model.WithdrawRequest{},
		&model.FundLog{},
		&model.Salesperson{},
		&model.CommissionRecord{},
		&model.SystemConfig{},
		&model.SupplementDeposit{},
		&model.InvitationCode{},
		&model.InvitationRecord{},
		&model.TeamRelation{},
	)
}
