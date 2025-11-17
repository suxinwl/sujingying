/**
 * 提成记录模型
 * 
 * 用途：
 * - 记录每次提成计算详情
 * - 追踪订单与提成的关联
 * - 支持提成对账
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
 * CommissionRecord 提成记录实体
 * 
 * 字段说明：
 * - SalespersonID: 销售人员ID
 * - OrderID: 订单ID
 * - CustomerID: 客户ID
 * - WeightG: 订单克重
 * - CommissionRate: 提成点数（记录当时的点数）
 * - Points: 本次提成积分
 * - SettledAt: 订单结算时间
 */
type CommissionRecord struct {
	ID             uint           `gorm:"primarykey"`
	SalespersonID  uint           `gorm:"index;not null"`                    // 销售人员ID
	OrderID        uint           `gorm:"index;not null"`                    // 订单ID
	CustomerID     uint           `gorm:"index;not null"`                    // 客户ID
	WeightG        float64        `gorm:"type:decimal(10,2);not null"`       // 克重
	CommissionRate float64        `gorm:"type:decimal(10,4);not null"`       // 提成点数
	Points         float64        `gorm:"type:decimal(15,2);not null"`       // 提成积分
	SettledAt      time.Time      `gorm:"index"`                             // 结算时间
	CreatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

/**
 * CalculatePoints 计算提成积分
 * 
 * 计算公式：积分 = 克重 × 提成点数
 * 
 * @param weightG float64 - 克重
 * @param rate float64 - 提成点数
 * @return float64
 */
func CalculatePoints(weightG, rate float64) float64 {
	return weightG * rate
}
