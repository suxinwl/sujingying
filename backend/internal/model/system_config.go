/**
 * 系统配置模型
 * 
 * 用途：
 * - 系统参数配置
 * - 运营参数管理
 * - 支持热更新
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
 * 配置分类常量
 */
const (
	ConfigCategorySystem   = "system"   // 系统配置
	ConfigCategoryTrading  = "trading"  // 交易配置
	ConfigCategoryFee      = "fee"      // 费用配置
	ConfigCategoryRisk     = "risk"     // 风控配置
	ConfigCategoryTime     = "time"     // 时间配置
)

/**
 * SystemConfig 系统配置实体
 * 
 * 字段说明：
 * - Category: 配置分类
 * - Key: 配置键
 * - Value: 配置值
 * - Description: 配置说明
 * - ValueType: 值类型(string/int/float/bool)
 */
type SystemConfig struct {
	ID          uint           `gorm:"primarykey"`
	Category    string         `gorm:"type:varchar(50);index;not null"` // 配置分类
	Key         string         `gorm:"type:varchar(100);uniqueIndex;not null"` // 配置键
	Value       string         `gorm:"type:varchar(500);not null"` // 配置值
	Description string         `gorm:"type:varchar(500)"` // 说明
	ValueType   string         `gorm:"type:varchar(20);default:'string'"` // 值类型
	IsSystem    bool           `gorm:"default:false"` // 是否系统配置(不可删除)
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

/**
 * 常用配置键定义
 */
const (
	// 交易相关
	ConfigKeyMinDeposit        = "min_deposit"         // 最小充值金额
	ConfigKeyMinWithdraw       = "min_withdraw"        // 最小提现金额
	ConfigKeyWithdrawFeeRate   = "withdraw_fee_rate"   // 提现手续费率
	ConfigKeyMinOrderAmount    = "min_order_amount"    // 最小下单金额
	
	// 时间相关
	ConfigKeyTradingStartTime  = "trading_start_time"  // 交易开始时间
	ConfigKeyTradingEndTime    = "trading_end_time"    // 交易结束时间
	ConfigKeyTradingDays       = "trading_days"        // 交易日(1-7表示周一到周日)
	
	// 点差相关
	ConfigKeyBuySpread         = "buy_spread"          // 买入点差
	ConfigKeySellSpread        = "sell_spread"         // 卖出点差
	
	// 杠杆相关
	ConfigKeyMaxLeverage       = "max_leverage"        // 最大杠杆倍数
	ConfigKeyMinLeverage       = "min_leverage"        // 最小杠杆倍数
	
	// 自动补定金相关
	ConfigKeyAutoSupplementTrigger = "auto_supplement_trigger" // 自动补定金触发阈值（默认50%）
	ConfigKeyAutoSupplementTarget  = "auto_supplement_target"  // 自动补定金目标阈值（默认80%）
)
