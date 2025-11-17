/**
 * 补定金申请模型
 * 
 * 用途：
 * - 客户补充定金
 * - 提高订单定金率
 * - 降低风控风险
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
 * 补定金状态常量
 */
const (
	SupplementStatusPending  = "pending"  // 待审核
	SupplementStatusApproved = "approved" // 已通过
	SupplementStatusRejected = "rejected" // 已驳回
)

/**
 * SupplementDeposit 补定金申请实体
 * 
 * 字段说明：
 * - UserID: 用户ID
 * - OrderID: 订单ID
 * - Amount: 补充金额
 * - Method: 补充方式(bank/wechat/alipay)
 * - VoucherURL: 凭证URL
 * - Status: 审核状态
 * - ReviewerID: 审核人ID
 * - ReviewNote: 审核备注
 */
type SupplementDeposit struct {
	ID         uint           `gorm:"primarykey"`
	UserID     uint           `gorm:"index;not null"`                           // 用户ID
	OrderID    uint           `gorm:"index;not null"`                           // 订单ID
	Amount     float64        `gorm:"type:decimal(15,2);not null"`              // 补充金额
	Method     string         `gorm:"type:varchar(20);not null"`                // 补充方式
	VoucherURL string         `gorm:"type:varchar(500)"`                        // 凭证URL
	Status     string         `gorm:"type:varchar(20);index;default:'pending'"` // 状态
	ReviewerID uint           `gorm:"default:0"`                                // 审核人ID
	ReviewNote string         `gorm:"type:varchar(500)"`                        // 审核备注
	ReviewedAt *time.Time     // 审核时间
	CreatedAt  time.Time      `gorm:"index"`
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

/**
 * Approve 通过审核
 */
func (s *SupplementDeposit) Approve(reviewerID uint, note string) {
	now := time.Now()
	s.Status = SupplementStatusApproved
	s.ReviewerID = reviewerID
	s.ReviewNote = note
	s.ReviewedAt = &now
}

/**
 * Reject 驳回审核
 */
func (s *SupplementDeposit) Reject(reviewerID uint, note string) {
	now := time.Now()
	s.Status = SupplementStatusRejected
	s.ReviewerID = reviewerID
	s.ReviewNote = note
	s.ReviewedAt = &now
}

/**
 * IsPending 是否待审核
 */
func (s *SupplementDeposit) IsPending() bool {
	return s.Status == SupplementStatusPending
}
