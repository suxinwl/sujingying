/**
 * 补定金仓储层
 * 
 * 用途：
 * - 封装补定金数据访问
 * - 提供查询和更新方法
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package repository

import (
	"gorm.io/gorm"

	"suxin/internal/model"
)

type SupplementRepository struct {
	db *gorm.DB
}

func NewSupplementRepository(db *gorm.DB) *SupplementRepository {
	return &SupplementRepository{db: db}
}

func (r *SupplementRepository) Create(supplement *model.SupplementDeposit) error {
	return r.db.Create(supplement).Error
}

func (r *SupplementRepository) FindByID(id uint) (*model.SupplementDeposit, error) {
	var supplement model.SupplementDeposit
	if err := r.db.First(&supplement, id).Error; err != nil {
		return nil, err
	}
	return &supplement, nil
}

func (r *SupplementRepository) FindByUserID(userID uint, limit, offset int) ([]*model.SupplementDeposit, error) {
	var supplements []*model.SupplementDeposit
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&supplements).Error
	return supplements, err
}

func (r *SupplementRepository) FindByOrderID(orderID uint) ([]*model.SupplementDeposit, error) {
	var supplements []*model.SupplementDeposit
	err := r.db.Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&supplements).Error
	return supplements, err
}

func (r *SupplementRepository) FindPending(limit int) ([]*model.SupplementDeposit, error) {
	var supplements []*model.SupplementDeposit
	err := r.db.Where("status = ?", model.SupplementStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&supplements).Error
	return supplements, err
}

func (r *SupplementRepository) Update(supplement *model.SupplementDeposit) error {
	return r.db.Save(supplement).Error
}
