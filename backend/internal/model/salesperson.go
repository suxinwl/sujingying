/**
 * 销售人员领域模型
 * 
 * 用途：
 * - 定义销售人员数据结构
 * - 管理销售业绩和提成
 * - 支持客户归属管理
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package model

import (
	"time"

	"gorm.io/gorm"
)

/**
 * Salesperson 销售人员实体
 * 
 * 字段说明：
 * - UserID: 关联用户ID（销售人员也是系统用户）
 * - SalesCode: 销售编号
 * - Name: 销售姓名
 * - CommissionRate: 提成点数（每克提成点数）
 * - TotalPoints: 总积分（历史累计）
 * - MonthPoints: 本月积分
 * - TotalCustomers: 客户总数
 * - ActiveCustomers: 活跃客户数（有持仓或近30天有交易）
 * - IsActive: 是否在职
 */
type Salesperson struct {
	ID              uint           `gorm:"primarykey"`
	UserID          uint           `gorm:"unique;not null"`                     // 关联用户ID
	SalesCode       string         `gorm:"type:varchar(20);unique;not null"`    // 销售编号
	Name            string         `gorm:"type:varchar(50);not null"`           // 姓名
	CommissionRate  float64        `gorm:"type:decimal(10,4);default:0.0001"`   // 提成点数（默认0.0001）
	TotalPoints     float64        `gorm:"type:decimal(15,2);default:0"`        // 总积分
	MonthPoints     float64        `gorm:"type:decimal(15,2);default:0"`        // 本月积分
	TotalCustomers  int            `gorm:"default:0"`                           // 客户总数
	ActiveCustomers int            `gorm:"default:0"`                           // 活跃客户数
	IsActive        bool           `gorm:"default:true"`                        // 是否在职
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

/**
 * AddPoints 增加积分
 * 
 * @param points float64 - 增加的积分
 * @return void
 */
func (s *Salesperson) AddPoints(points float64) {
	s.TotalPoints += points
	s.MonthPoints += points
}

/**
 * ResetMonthPoints 重置月度积分（每月初调用）
 * 
 * @return void
 */
func (s *Salesperson) ResetMonthPoints() {
	s.MonthPoints = 0
}

/**
 * UpdateCustomerCount 更新客户数量
 * 
 * @param total int - 总客户数
 * @param active int - 活跃客户数
 * @return void
 */
func (s *Salesperson) UpdateCustomerCount(total, active int) {
	s.TotalCustomers = total
	s.ActiveCustomers = active
}

/**
 * Deactivate 离职
 * 
 * @return void
 */
func (s *Salesperson) Deactivate() {
	s.IsActive = false
}

/**
 * Activate 激活
 * 
 * @return void
 */
func (s *Salesperson) Activate() {
	s.IsActive = true
}
