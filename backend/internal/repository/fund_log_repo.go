/**
 * 资金流水仓储层
 * 
 * 用途：
 * - 封装资金流水数据访问逻辑
 * - 提供流水记录查询
 * - 支持对账和统计
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
 * FundLogRepository 资金流水仓储
 */
type FundLogRepository struct {
	db *gorm.DB
}

/**
 * NewFundLogRepository 创建资金流水仓储实例
 * 
 * @param db *gorm.DB - 数据库连接
 * @return *FundLogRepository
 */
func NewFundLogRepository(db *gorm.DB) *FundLogRepository {
	return &FundLogRepository{db: db}
}

/**
 * Create 创建资金流水记录
 * 
 * @param log *model.FundLog - 流水实体
 * @return error
 */
func (r *FundLogRepository) Create(log *model.FundLog) error {
	return r.db.Create(log).Error
}

/**
 * FindByUserID 查询用户的资金流水
 * 
 * @param userID uint - 用户ID
 * @param limit int - 查询数量限制
 * @param offset int - 偏移量
 * @return ([]*model.FundLog, error)
 */
func (r *FundLogRepository) FindByUserID(userID uint, limit, offset int) ([]*model.FundLog, error) {
	var logs []*model.FundLog
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

/**
 * FindByUserIDAndType 查询用户指定类型的资金流水
 * 
 * @param userID uint - 用户ID
 * @param logType string - 流水类型
 * @param limit int - 查询数量限制
 * @return ([]*model.FundLog, error)
 */
func (r *FundLogRepository) FindByUserIDAndType(userID uint, logType string, limit int) ([]*model.FundLog, error) {
	var logs []*model.FundLog
	err := r.db.Where("user_id = ? AND type = ?", userID, logType).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

/**
 * FindByDateRange 查询指定时间范围的资金流水
 * 
 * @param userID uint - 用户ID
 * @param startDate time.Time - 开始时间
 * @param endDate time.Time - 结束时间
 * @return ([]*model.FundLog, error)
 */
func (r *FundLogRepository) FindByDateRange(userID uint, startDate, endDate time.Time) ([]*model.FundLog, error) {
	var logs []*model.FundLog
	err := r.db.Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

/**
 * GetSumByType 统计指定类型的资金总额
 * 
 * @param userID uint - 用户ID
 * @param logType string - 流水类型
 * @return (float64, error)
 */
func (r *FundLogRepository) GetSumByType(userID uint, logType string) (float64, error) {
	var sum float64
	err := r.db.Model(&model.FundLog{}).
		Where("user_id = ? AND type = ?", userID, logType).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sum).Error
	return sum, err
}
