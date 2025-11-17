package repository

import (
	"time"

	"gorm.io/gorm"

	"suxin/internal/model"
)

type LoginLogRepository struct {
	db *gorm.DB
}

func NewLoginLogRepository(db *gorm.DB) *LoginLogRepository {
	return &LoginLogRepository{db: db}
}

func (r *LoginLogRepository) Create(log *model.LoginLog) error {
	return r.db.Create(log).Error
}

func (r *LoginLogRepository) CountFailedAttempts(phone string, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoginLog{}).
		Where("phone = ? AND status = ? AND created_at > ?", phone, "failed", since).
		Count(&count).Error
	return count, err
}
