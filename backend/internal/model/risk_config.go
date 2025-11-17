/**
 * 风控配置模型
 * 
 * 用途：
 * - 定义风控阈值配置
 * - 支持动态调整预警和强平阈值
 * - 配置可持久化到数据库
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
 * RiskConfig 风控配置实体
 * 
 * 字段说明：
 * - ForceCloseRate: 强制平仓阈值（%）
 * - HighRiskRateMin: 高风险区间最小值（%）
 * - HighRiskRateMax: 高风险区间最大值（%）
 * - WarningRate: 一般预警阈值（%）
 * - PriceUpdateInterval: 价格更新间隔（秒）
 * - IsActive: 是否启用
 */
type RiskConfig struct {
	ID                  uint           `gorm:"primarykey"`
	ForceCloseRate      float64        `gorm:"type:decimal(5,2);default:20.00"`  // 强平阈值
	HighRiskRateMin     float64        `gorm:"type:decimal(5,2);default:20.00"`  // 高风险区间下限
	HighRiskRateMax     float64        `gorm:"type:decimal(5,2);default:25.00"`  // 高风险区间上限
	WarningRate         float64        `gorm:"type:decimal(5,2);default:50.00"`  // 预警阈值
	PriceUpdateInterval int            `gorm:"default:60"`                       // 价格更新间隔（秒）
	IsActive            bool           `gorm:"default:true"`                     // 是否启用
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}

/**
 * GetDefaultRiskConfig 获取默认风控配置
 * 
 * 默认阈值：
 * - 强制平仓：≤ 20%
 * - 高风险：20% < x < 25%
 * - 一般预警：≤ 50%
 * - 价格更新间隔：60秒
 * 
 * @return *RiskConfig
 */
func GetDefaultRiskConfig() *RiskConfig {
	return &RiskConfig{
		ForceCloseRate:      20.00,
		HighRiskRateMin:     20.00,
		HighRiskRateMax:     25.00,
		WarningRate:         50.00,
		PriceUpdateInterval: 60,
		IsActive:            true,
	}
}
