/**
 * 提现申请仓储层
 * 
 * 用途：
 * - 封装提现申请数据访问逻辑
 * - 提供CRUD和查询方法
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package repository

import (
	"gorm.io/gorm"

	"suxin/internal/model"
)

type WithdrawRepository struct {
	db *gorm.DB
}

func NewWithdrawRepository(db *gorm.DB) *WithdrawRepository {
	return &WithdrawRepository{db: db}
}

func (r *WithdrawRepository) Create(withdraw *model.WithdrawRequest) error {
	return r.db.Create(withdraw).Error
}

func (r *WithdrawRepository) FindByID(id uint) (*model.WithdrawRequest, error) {
	var withdraw model.WithdrawRequest
	if err := r.db.First(&withdraw, id).Error; err != nil {
		return nil, err
	}
	return &withdraw, nil
}

func (r *WithdrawRepository) FindByUserID(userID uint, limit, offset int) ([]*model.WithdrawRequest, error) {
	var withdraws []*model.WithdrawRequest
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&withdraws).Error
	return withdraws, err
}

func (r *WithdrawRepository) FindPending(limit int) ([]*model.WithdrawRequest, error) {
	var withdraws []*model.WithdrawRequest
	err := r.db.Preload("User").Where("status = ?", model.WithdrawStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&withdraws).Error
	return withdraws, err
}

// FindByStatus 根据状态查询提现申请
func (r *WithdrawRepository) FindByStatus(status string, limit int) ([]*model.WithdrawRequest, error) {
	var withdraws []*model.WithdrawRequest
	query := r.db.Preload("User")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").
		Limit(limit).
		Find(&withdraws).Error
	return withdraws, err
}

func (r *WithdrawRepository) Update(withdraw *model.WithdrawRequest) error {
	return r.db.Save(withdraw).Error
}
