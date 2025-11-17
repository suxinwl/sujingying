package appctx

import (
	"suxin/internal/pkg/config"
	"gorm.io/gorm"
)

type AppContext struct {
	DB     *gorm.DB
	Config *config.Config
}

func New(db *gorm.DB, cfg *config.Config) *AppContext {
	return &AppContext{DB: db, Config: cfg}
}
