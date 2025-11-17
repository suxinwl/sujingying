/**
 * 银行卡领域模型
 * 
 * 用途：
 * - 定义用户银行卡数据结构
 * - 支持银行卡增删改查
 * - 需要支付密码验证的敏感操作
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
 * BankCard 银行卡实体
 * 
 * 字段说明：
 * - UserID: 用户ID
 * - BankName: 银行名称
 * - CardNumber: 银行卡号（加密存储）
 * - CardHolder: 持卡人姓名
 * - IsDefault: 是否默认卡
 */
type BankCard struct {
	ID         uint           `gorm:"primarykey"`
	UserID     uint           `gorm:"index;not null"`                          // 用户ID
	BankName   string         `gorm:"type:varchar(100);not null"`              // 银行名称
	CardNumber string         `gorm:"type:varchar(50);not null"`               // 银行卡号
	CardHolder string         `gorm:"type:varchar(50);not null"`               // 持卡人姓名
	IsDefault  bool           `gorm:"default:false"`                           // 是否默认卡
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

/**
 * MaskCardNumber 脱敏显示银行卡号
 * 显示格式：622202******1234
 * 
 * @return string
 */
func (c *BankCard) MaskCardNumber() string {
	if len(c.CardNumber) <= 8 {
		return c.CardNumber
	}
	return c.CardNumber[:6] + "******" + c.CardNumber[len(c.CardNumber)-4:]
}

/**
 * SetAsDefault 设置为默认卡
 * 
 * @return void
 */
func (c *BankCard) SetAsDefault() {
	c.IsDefault = true
}

/**
 * UnsetDefault 取消默认卡
 * 
 * @return void
 */
func (c *BankCard) UnsetDefault() {
	c.IsDefault = false
}
