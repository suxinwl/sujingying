/**
 * 提成记录仓储层
 * 
 * 用途：
 * - 封装提成记录数据访问逻辑
 * - 提供提成统计和查询
 * - 支持对账和报表
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package repository

import (
	"time"

	"gorm.io/gorm"

	"suxin/internal/model"
)

/**
 * CommissionRepository 提成记录仓储
 */
type CommissionRepository struct {
	db *gorm.DB
}

/**
 * NewCommissionRepository 创建提成记录仓储实例
 * 
 * @param db *gorm.DB - 数据库连接
 * @return *CommissionRepository
 */
func NewCommissionRepository(db *gorm.DB) *CommissionRepository {
	return &CommissionRepository{db: db}
}

/**
 * Create 创建提成记录
 * 
 * @param record *model.CommissionRecord - 提成记录实体
 * @return error
 */
func (r *CommissionRepository) Create(record *model.CommissionRecord) error {
	return r.db.Create(record).Error
}

/**
 * FindBySalespersonID 查询销售人员的提成记录
 * 
 * @param salespersonID uint - 销售人员ID
 * @param limit int - 查询数量限制
 * @param offset int - 偏移量
 * @return ([]*model.CommissionRecord, error)
 */
func (r *CommissionRepository) FindBySalespersonID(salespersonID uint, limit, offset int) ([]*model.CommissionRecord, error) {
	var records []*model.CommissionRecord
	err := r.db.Where("salesperson_id = ?", salespersonID).
		Order("settled_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error
	return records, err
}

/**
 * FindByDateRange 查询指定时间范围的提成记录
 * 
 * @param salespersonID uint - 销售人员ID
 * @param startDate time.Time - 开始时间
 * @param endDate time.Time - 结束时间
 * @return ([]*model.CommissionRecord, error)
 */
func (r *CommissionRepository) FindByDateRange(salespersonID uint, startDate, endDate time.Time) ([]*model.CommissionRecord, error) {
	var records []*model.CommissionRecord
	err := r.db.Where("salesperson_id = ? AND settled_at BETWEEN ? AND ?", 
		salespersonID, startDate, endDate).
		Order("settled_at DESC").
		Find(&records).Error
	return records, err
}

/**
 * GetTotalPointsByDateRange 统计指定时间范围的总积分
 * 
 * @param salespersonID uint - 销售人员ID
 * @param startDate time.Time - 开始时间
 * @param endDate time.Time - 结束时间
 * @return (float64, error)
 */
func (r *CommissionRepository) GetTotalPointsByDateRange(salespersonID uint, startDate, endDate time.Time) (float64, error) {
	var sum float64
	err := r.db.Model(&model.CommissionRecord{}).
		Where("salesperson_id = ? AND settled_at BETWEEN ? AND ?", salespersonID, startDate, endDate).
		Select("COALESCE(SUM(points), 0)").
		Scan(&sum).Error
	return sum, err
}

/**
 * ExistsByOrderID 检查订单是否已生成提成记录
 * 
 * @param orderID uint - 订单ID
 * @return (bool, error)
 */
func (r *CommissionRepository) ExistsByOrderID(orderID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.CommissionRecord{}).
		Where("order_id = ?", orderID).
		Count(&count).Error
	return count > 0, err
}
