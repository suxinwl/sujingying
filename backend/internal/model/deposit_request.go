/**
 * 定金充值领域模型
 * 
 * 用途：
 * - 定义定金充值申请数据结构
 * - 支持充值审核流程
 * - 记录充值凭证和审核信息
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
 * 充值状态常量
 */
const (
	DepositStatusPending  = "pending"  // 待审核
	DepositStatusApproved = "approved" // 已通过
	DepositStatusRejected = "rejected" // 已驳回
)

/**
 * 充值方式常量
 */
const (
	DepositMethodBank   = "bank"   // 银行转账
	DepositMethodWechat = "wechat" // 微信支付
	DepositMethodAlipay = "alipay" // 支付宝
)

/**
 * DepositRequest 定金充值申请实体
 * 
 * 字段说明：
 * - UserID: 用户ID
 * - Amount: 充值金额
 * - Method: 充值方式
 * - VoucherURL: 凭证图片URL
 * - Status: 审核状态
 * - ReviewerID: 审核人ID
 * - ReviewNote: 审核备注
 * - ReviewedAt: 审核时间
 */
type DepositRequest struct {
	ID          uint           `gorm:"primarykey"`
	UserID      uint           `gorm:"index;not null"`                           // 用户ID
	Amount      float64        `gorm:"type:decimal(15,2);not null"`              // 充值金额
	Method      string         `gorm:"type:varchar(20);not null"`                // 充值方式
	VoucherURL  string         `gorm:"type:varchar(500)"`                        // 凭证URL
	Status      string         `gorm:"type:varchar(20);index;default:'pending'"` // 状态
	ReviewerID  uint           `gorm:"default:0"`                                // 审核人ID
	ReviewNote  string         `gorm:"type:varchar(500)"`                        // 审核备注
	ReviewedAt  *time.Time     // 审核时间
	CreatedAt   time.Time      `gorm:"index"`
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

/**
 * Approve 通过审核
 * 
 * @param reviewerID uint - 审核人ID
 * @param note string - 审核备注
 * @return void
 */
func (d *DepositRequest) Approve(reviewerID uint, note string) {
	now := time.Now()
	d.Status = DepositStatusApproved
	d.ReviewerID = reviewerID
	d.ReviewNote = note
	d.ReviewedAt = &now
}

/**
 * Reject 驳回审核
 * 
 * @param reviewerID uint - 审核人ID
 * @param note string - 驳回原因
 * @return void
 */
func (d *DepositRequest) Reject(reviewerID uint, note string) {
	now := time.Now()
	d.Status = DepositStatusRejected
	d.ReviewerID = reviewerID
	d.ReviewNote = note
	d.ReviewedAt = &now
}

/**
 * IsPending 判断是否待审核
 * 
 * @return bool
 */
func (d *DepositRequest) IsPending() bool {
	return d.Status == DepositStatusPending
}

/**
 * IsApproved 判断是否已通过
 * 
 * @return bool
 */
func (d *DepositRequest) IsApproved() bool {
	return d.Status == DepositStatusApproved
}
