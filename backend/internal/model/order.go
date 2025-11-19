/**
 * 订单领域模型
 * 
 * 用途：
 * - 定义订单数据结构
 * - 实现PnL（盈亏）和定金率计算逻辑
 * - 支持锁价买料和卖料两种订单类型
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
 * 订单类型常量
 */
const (
	OrderTypeLongBuy  = "long_buy"  // 锁价买料（看涨）
	OrderTypeShortSell = "short_sell" // 锁价卖料（看跌）
)

/**
 * 订单状态常量
 */
const (
	OrderStatusHolding = "holding" // 持仓中
	OrderStatusSettled = "settled" // 已结算
	OrderStatusClosed  = "closed"  // 已平仓（强制平仓）
)

/**
 * Order 订单实体
 * 
 * 字段说明：
 * - OrderID: 订单号（唯一标识）
 * - Type: 订单类型（long_buy/short_sell）
 * - LockedPrice: 锁定价格（元/克）
 * - CurrentPrice: 当前价格（元/克，实时更新）
 * - WeightG: 克重
 * - Deposit: 定金金额
 * - PnLFloat: 浮动盈亏（实时计算）
 * - MarginRate: 定金率（%，实时计算）
 * - Status: 订单状态
 * - SettledPrice: 结算价格（结算时记录）
 * - SettledPnL: 结算盈亏（结算时记录）
 * - SettledAt: 结算时间
 */
type Order struct {
	ID           uint           `gorm:"primarykey"`
	OrderID      string         `gorm:"type:varchar(50);uniqueIndex;not null"` // 订单号
	UserID       uint           `gorm:"index;not null"`                        // 用户ID
	Type         string         `gorm:"type:varchar(20);not null"`             // 订单类型
	LockedPrice  float64        `gorm:"type:decimal(10,4);not null"`           // 锁定价格（元/克）
	CurrentPrice float64        `gorm:"type:decimal(10,4)"`                    // 当前价格（元/克）
	WeightG      float64        `gorm:"type:decimal(10,3);not null"`           // 克重
	Deposit      float64        `gorm:"type:decimal(15,2);not null"`           // 定金
	PnLFloat     float64        `gorm:"type:decimal(15,2);default:0"`          // 浮动盈亏
	MarginRate   float64        `gorm:"type:decimal(10,2);default:100"`        // 定金率（%）
	Status       string         `gorm:"type:varchar(20);index;default:'holding'"` // 状态
	SettledPrice float64        `gorm:"type:decimal(10,4)"`                    // 结算价格
	SettledPnL   float64        `gorm:"type:decimal(15,2)"`                    // 结算盈亏
	SettledAt    *time.Time     // 结算时间
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

/**
 * CalculatePnL 计算订单盈亏
 * 
 * 计算逻辑：
 * - 买料（long_buy）：盈亏 = (当前价 - 锁定价) × 克重
 * - 卖料（short_sell）：盈亏 = (锁定价 - 当前价) × 克重
 * 
 * @param currentPrice float64 - 当前市场价格（元/克）
 * @return float64 - 浮动盈亏金额（元）
 * 
 * 示例：
 * - 买料：锁定价500，当前价510，克重100 => 盈亏 = (510-500)*100 = 1000元
 * - 卖料：锁定价500，当前价490，克重100 => 盈亏 = (500-490)*100 = 1000元
 */
func (o *Order) CalculatePnL(currentPrice float64) float64 {
	switch o.Type {
	case OrderTypeLongBuy:
		// 买料：价格上涨盈利，下跌亏损
		return (currentPrice - o.LockedPrice) * o.WeightG
	case OrderTypeShortSell:
		// 卖料：价格下跌盈利，上涨亏损
		return (o.LockedPrice - currentPrice) * o.WeightG
	default:
		return 0
	}
}

/**
 * CalculateMarginRate 计算定金率
 * 
 * 计算逻辑：
 * 定金率 = (定金 + 浮动盈亏) / (当前价 × 克重) × 100%
 * 
 * 风控阈值：
 * - 定金率 ≤ 20%：强制平仓
 * - 20% < 定金率 < 25%：高风险预警
 * - 定金率 ≤ 50%：一般预警
 * 
 * @param currentPrice float64 - 当前市场价格（元/克）
 * @return float64 - 定金率（%）
 * 
 * 示例：
 * - 定金10000，盈亏-2000，当前价500，克重100
 *   定金率 = (10000 - 2000) / (500 * 100) * 100 = 16%（触发强平）
 */
func (o *Order) CalculateMarginRate(currentPrice float64) float64 {
	// 计算浮动盈亏
	pnl := o.CalculatePnL(currentPrice)

	// 按前端一致的规则计算基础定金：克重 × 10 元/克
	baseDeposit := o.WeightG * 10
	if baseDeposit <= 0 {
		return 0.0
	}

	// 定金率 = (定金 + 浮动盈亏) / 基础定金 × 100%
	marginRate := (o.Deposit + pnl) / baseDeposit * 100.0
	
	return marginRate
}

/**
 * UpdatePnLAndMargin 更新订单的盈亏和定金率
 * 
 * 用途：
 * - 实时更新订单的浮动盈亏
 * - 实时更新订单的定金率
 * - 更新当前价格
 * 
 * @param currentPrice float64 - 当前市场价格（元/克）
 * @return void
 */
func (o *Order) UpdatePnLAndMargin(currentPrice float64) {
	o.CurrentPrice = currentPrice
	o.PnLFloat = o.CalculatePnL(currentPrice)
	o.MarginRate = o.CalculateMarginRate(currentPrice)
}

/**
 * IsNeedForceClose 判断是否需要强制平仓
 * 
 * 风控规则：定金率 ≤ 20% 时强制平仓
 * 
 * @return bool - true: 需要强平，false: 不需要
 */
func (o *Order) IsNeedForceClose() bool {
	return o.MarginRate <= 20.0 && o.Status == OrderStatusHolding
}

/**
 * IsHighRisk 判断是否为高风险订单
 * 
 * 风控规则：20% < 定金率 < 25% 为高风险区间
 * 
 * @return bool - true: 高风险，false: 正常
 */
func (o *Order) IsHighRisk() bool {
	return o.MarginRate > 20.0 && o.MarginRate < 25.0 && o.Status == OrderStatusHolding
}

/**
 * IsWarning 判断是否需要一般预警
 * 
 * 风控规则：定金率 ≤ 50% 触发一般预警
 * 
 * @return bool - true: 需要预警，false: 安全
 */
func (o *Order) IsWarning() bool {
	return o.MarginRate <= 50.0 && o.Status == OrderStatusHolding
}

/**
 * CanSettle 判断订单是否可以结算
 * 
 * 结算条件：订单状态为持仓中
 * 
 * @return bool - true: 可结算，false: 不可结算
 */
func (o *Order) CanSettle() bool {
	return o.Status == OrderStatusHolding
}

/**
 * Settle 执行订单结算
 * 
 * 结算逻辑：
 * 1. 记录结算价格和结算盈亏
 * 2. 更新订单状态为已结算
 * 3. 记录结算时间
 * 
 * @param settlePrice float64 - 结算价格
 * @return void
 */
func (o *Order) Settle(settlePrice float64) {
	o.SettledPrice = settlePrice
	o.SettledPnL = o.CalculatePnL(settlePrice)
	o.Status = OrderStatusSettled
	now := time.Now()
	o.SettledAt = &now
}

/**
 * ForceClose 执行强制平仓
 * 
 * 平仓逻辑：
 * 1. 记录平仓价格和最终盈亏
 * 2. 更新订单状态为已平仓
 * 3. 记录平仓时间
 * 
 * @param closePrice float64 - 平仓价格
 * @return void
 */
func (o *Order) ForceClose(closePrice float64) {
	o.SettledPrice = closePrice
	o.SettledPnL = o.CalculatePnL(closePrice)
	o.Status = OrderStatusClosed
	now := time.Now()
	o.SettledAt = &now
}
