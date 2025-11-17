/**
 * 银行卡仓储层
 * 
 * 用途：
 * - 封装银行卡数据访问逻辑
 * - 提供银行卡CRUD操作
 * - 支持默认卡管理
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package repository

import (
	"gorm.io/gorm"

	"suxin/internal/model"
)

/**
 * BankCardRepository 银行卡仓储
 */
type BankCardRepository struct {
	db *gorm.DB
}

/**
 * NewBankCardRepository 创建银行卡仓储实例
 * 
 * @param db *gorm.DB - 数据库连接
 * @return *BankCardRepository
 */
func NewBankCardRepository(db *gorm.DB) *BankCardRepository {
	return &BankCardRepository{db: db}
}

/**
 * Create 创建银行卡
 * 
 * @param card *model.BankCard - 银行卡实体
 * @return error
 */
func (r *BankCardRepository) Create(card *model.BankCard) error {
	return r.db.Create(card).Error
}

/**
 * FindByID 根据ID查找银行卡
 * 
 * @param id uint - 银行卡ID
 * @return (*model.BankCard, error)
 */
func (r *BankCardRepository) FindByID(id uint) (*model.BankCard, error) {
	var card model.BankCard
	if err := r.db.First(&card, id).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

/**
 * FindByUserID 查询用户的所有银行卡
 * 
 * @param userID uint - 用户ID
 * @return ([]*model.BankCard, error)
 */
func (r *BankCardRepository) FindByUserID(userID uint) ([]*model.BankCard, error) {
	var cards []*model.BankCard
	err := r.db.Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&cards).Error
	return cards, err
}

/**
 * FindDefaultByUserID 查询用户的默认银行卡
 * 
 * @param userID uint - 用户ID
 * @return (*model.BankCard, error)
 */
func (r *BankCardRepository) FindDefaultByUserID(userID uint) (*model.BankCard, error) {
	var card model.BankCard
	if err := r.db.Where("user_id = ? AND is_default = ?", userID, true).
		First(&card).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

/**
 * Update 更新银行卡
 * 
 * @param card *model.BankCard - 银行卡实体
 * @return error
 */
func (r *BankCardRepository) Update(card *model.BankCard) error {
	return r.db.Save(card).Error
}

/**
 * Delete 删除银行卡（软删除）
 * 
 * @param id uint - 银行卡ID
 * @param userID uint - 用户ID（验证归属）
 * @return error
 */
func (r *BankCardRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.BankCard{}).Error
}

/**
 * UnsetAllDefault 取消用户所有银行卡的默认状态
 * 
 * @param userID uint - 用户ID
 * @return error
 */
func (r *BankCardRepository) UnsetAllDefault(userID uint) error {
	return r.db.Model(&model.BankCard{}).
		Where("user_id = ?", userID).
		Update("is_default", false).Error
}

/**
 * CountByUserID 统计用户银行卡数量
 * 
 * @param userID uint - 用户ID
 * @return (int64, error)
 */
func (r *BankCardRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.BankCard{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
