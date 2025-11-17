/**
 * 定金充值仓储层
 * 
 * 用途：
 * - 封装充值申请数据访问逻辑
 * - 提供充值记录查询
 * - 支持审核状态管理
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
 * DepositRepository 充值仓储
 */
type DepositRepository struct {
	db *gorm.DB
}

/**
 * NewDepositRepository 创建充值仓储实例
 * 
 * @param db *gorm.DB - 数据库连接
 * @return *DepositRepository
 */
func NewDepositRepository(db *gorm.DB) *DepositRepository {
	return &DepositRepository{db: db}
}

/**
 * Create 创建充值申请
 * 
 * @param deposit *model.DepositRequest - 充值申请实体
 * @return error
 */
func (r *DepositRepository) Create(deposit *model.DepositRequest) error {
	return r.db.Create(deposit).Error
}

/**
 * FindByID 根据ID查找充值申请
 * 
 * @param id uint - 充值申请ID
 * @return (*model.DepositRequest, error)
 */
func (r *DepositRepository) FindByID(id uint) (*model.DepositRequest, error) {
	var deposit model.DepositRequest
	if err := r.db.First(&deposit, id).Error; err != nil {
		return nil, err
	}
	return &deposit, nil
}

/**
 * FindByUserID 查询用户的充值记录
 * 
 * @param userID uint - 用户ID
 * @param limit int - 查询数量限制
 * @param offset int - 偏移量
 * @return ([]*model.DepositRequest, error)
 */
func (r *DepositRepository) FindByUserID(userID uint, limit, offset int) ([]*model.DepositRequest, error) {
	var deposits []*model.DepositRequest
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&deposits).Error
	return deposits, err
}

/**
 * FindPending 查询待审核的充值申请
 * 
 * @param limit int - 查询数量限制
 * @return ([]*model.DepositRequest, error)
 */
func (r *DepositRepository) FindPending(limit int) ([]*model.DepositRequest, error) {
	var deposits []*model.DepositRequest
	err := r.db.Where("status = ?", model.DepositStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&deposits).Error
	return deposits, err
}

/**
 * Update 更新充值申请
 * 
 * @param deposit *model.DepositRequest - 充值申请实体
 * @return error
 */
func (r *DepositRepository) Update(deposit *model.DepositRequest) error {
	return r.db.Save(deposit).Error
}

/**
 * CountByStatus 统计指定状态的充值申请数量
 * 
 * @param status string - 状态
 * @return (int64, error)
 */
func (r *DepositRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.DepositRequest{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}
