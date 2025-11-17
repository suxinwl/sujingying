/**
 * 销售人员仓储层
 * 
 * 用途：
 * - 封装销售人员数据访问逻辑
 * - 提供业绩查询和排名
 * - 支持客户归属管理
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
 * SalespersonRepository 销售人员仓储
 */
type SalespersonRepository struct {
	db *gorm.DB
}

/**
 * NewSalespersonRepository 创建销售人员仓储实例
 * 
 * @param db *gorm.DB - 数据库连接
 * @return *SalespersonRepository
 */
func NewSalespersonRepository(db *gorm.DB) *SalespersonRepository {
	return &SalespersonRepository{db: db}
}

/**
 * Create 创建销售人员
 * 
 * @param salesperson *model.Salesperson - 销售人员实体
 * @return error
 */
func (r *SalespersonRepository) Create(salesperson *model.Salesperson) error {
	return r.db.Create(salesperson).Error
}

/**
 * FindByID 根据ID查找销售人员
 * 
 * @param id uint - 销售人员ID
 * @return (*model.Salesperson, error)
 */
func (r *SalespersonRepository) FindByID(id uint) (*model.Salesperson, error) {
	var salesperson model.Salesperson
	if err := r.db.First(&salesperson, id).Error; err != nil {
		return nil, err
	}
	return &salesperson, nil
}

/**
 * FindByUserID 根据用户ID查找销售人员
 * 
 * @param userID uint - 用户ID
 * @return (*model.Salesperson, error)
 */
func (r *SalespersonRepository) FindByUserID(userID uint) (*model.Salesperson, error) {
	var salesperson model.Salesperson
	if err := r.db.Where("user_id = ?", userID).First(&salesperson).Error; err != nil {
		return nil, err
	}
	return &salesperson, nil
}

/**
 * FindAll 查询所有在职销售人员
 * 
 * @return ([]*model.Salesperson, error)
 */
func (r *SalespersonRepository) FindAll() ([]*model.Salesperson, error) {
	var salespeople []*model.Salesperson
	err := r.db.Where("is_active = ?", true).
		Order("total_points DESC").
		Find(&salespeople).Error
	return salespeople, err
}

/**
 * FindTopByTotalPoints 按总积分排名
 * 
 * @param limit int - 查询数量限制
 * @return ([]*model.Salesperson, error)
 */
func (r *SalespersonRepository) FindTopByTotalPoints(limit int) ([]*model.Salesperson, error) {
	var salespeople []*model.Salesperson
	err := r.db.Where("is_active = ?", true).
		Order("total_points DESC").
		Limit(limit).
		Find(&salespeople).Error
	return salespeople, err
}

/**
 * FindTopByMonthPoints 按本月积分排名
 * 
 * @param limit int - 查询数量限制
 * @return ([]*model.Salesperson, error)
 */
func (r *SalespersonRepository) FindTopByMonthPoints(limit int) ([]*model.Salesperson, error) {
	var salespeople []*model.Salesperson
	err := r.db.Where("is_active = ?", true).
		Order("month_points DESC").
		Limit(limit).
		Find(&salespeople).Error
	return salespeople, err
}

/**
 * Update 更新销售人员
 * 
 * @param salesperson *model.Salesperson - 销售人员实体
 * @return error
 */
func (r *SalespersonRepository) Update(salesperson *model.Salesperson) error {
	return r.db.Save(salesperson).Error
}

/**
 * ResetAllMonthPoints 重置所有销售人员的月度积分
 * 
 * @return error
 */
func (r *SalespersonRepository) ResetAllMonthPoints() error {
	return r.db.Model(&model.Salesperson{}).
		Where("is_active = ?", true).
		Update("month_points", 0).Error
}
