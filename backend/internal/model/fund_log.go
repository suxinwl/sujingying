/**
 * 资金流水模型
 * 
 * 用途：
 * - 记录所有资金变动
 * - 支持台账查询和对账
 * - 追踪资金流向
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
 * 资金流水类型常量
 */
const (
	FundLogTypeDeposit      = "deposit"       // 充值
	FundLogTypeWithdraw     = "withdraw"      // 提现
	FundLogTypeOrderFreeze  = "order_freeze"  // 订单冻结
	FundLogTypeOrderRelease = "order_release" // 订单释放
	FundLogTypeSettle       = "settle"        // 结算
	FundLogTypeForceClose   = "force_close"   // 强平
)

/**
 * FundLog 资金流水实体
 * 
 * 字段说明：
 * - UserID: 用户ID
 * - Type: 流水类型
 * - Amount: 变动金额（正数为增加，负数为减少）
 * - AvailableBefore: 变动前可用定金
 * - AvailableAfter: 变动后可用定金
 * - UsedBefore: 变动前已用定金
 * - UsedAfter: 变动后已用定金
 * - RelatedID: 关联业务ID
 * - RelatedType: 关联业务类型
 * - Note: 备注
 */
type FundLog struct {
	ID              uint           `gorm:"primarykey"`
	UserID          uint           `gorm:"index;not null"`                    // 用户ID
	Type            string         `gorm:"type:varchar(20);index;not null"`   // 流水类型
	Amount          float64        `gorm:"type:decimal(15,2);not null"`       // 变动金额
	AvailableBefore float64        `gorm:"type:decimal(15,2);not null"`       // 变动前可用
	AvailableAfter  float64        `gorm:"type:decimal(15,2);not null"`       // 变动后可用
	UsedBefore      float64        `gorm:"type:decimal(15,2);not null"`       // 变动前已用
	UsedAfter       float64        `gorm:"type:decimal(15,2);not null"`       // 变动后已用
	RelatedID       uint           `gorm:"default:0"`                         // 关联业务ID
	RelatedType     string         `gorm:"type:varchar(50)"`                  // 关联业务类型
	Note            string         `gorm:"type:varchar(500)"`                 // 备注
	CreatedAt       time.Time      `gorm:"index"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

/**
 * IsIncome 判断是否为收入
 * 
 * @return bool
 */
func (f *FundLog) IsIncome() bool {
	return f.Amount > 0
}

/**
 * IsExpense 判断是否为支出
 * 
 * @return bool
 */
func (f *FundLog) IsExpense() bool {
	return f.Amount < 0
}

/**
 * GetTotalBalance 获取变动后总余额
 * 
 * @return float64
 */
func (f *FundLog) GetTotalBalance() float64 {
	return f.AvailableAfter + f.UsedAfter
}
