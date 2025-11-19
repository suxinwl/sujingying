/**
 * 提现申请领域模型
 * 
 * 用途：
 * - 定义提现申请数据结构
 * - 支持提现审核流程
 * - 记录提现凭证和审核信息
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
 * 提现状态常量
 */
const (
	WithdrawStatusPending  = "pending"  // 待审核
	WithdrawStatusApproved = "approved" // 已通过
	WithdrawStatusRejected = "rejected" // 已驳回
	WithdrawStatusPaid     = "paid"     // 已打款
)

/**
 * WithdrawRequest 提现申请实体
 * 
 * 字段说明：
 * - UserID: 用户ID
 * - BankCardID: 银行卡ID
 * - Amount: 提现金额
 * - Fee: 手续费
 * - ActualAmount: 实际到账金额
 * - Status: 审核状态
 * - ReviewerID: 审核人ID
 * - ReviewNote: 审核备注
 * - VoucherURL: 打款凭证
 */
type WithdrawRequest struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`              // 用户ID
	BankCardID   uint           `gorm:"not null" json:"bank_card_id"`               // 银行卡ID
	Amount       float64        `gorm:"type:decimal(15,2);not null" json:"amount"` // 提现金额
	Fee          float64        `gorm:"type:decimal(15,2);default:0" json:"fee"`    // 手续费
	ActualAmount float64        `gorm:"type:decimal(15,2);not null" json:"actual_amount"` // 实际到账
	Status       string         `gorm:"type:varchar(20);index;default:'pending'" json:"status"` // 状态
	ReviewerID   uint           `gorm:"default:0" json:"reviewer_id"`                       // 审核人ID
	ReviewNote   string         `gorm:"type:varchar(500)" json:"review_note"`               // 审核备注
	UserNote     string         `gorm:"type:varchar(500)" json:"user_note"`                 // 用户备注
	VoucherURL   string         `gorm:"type:longtext" json:"voucher_url"`                   // 打款凭证（Base64或URL）
	User         *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`            // 关联用户
	ReviewedAt   *time.Time     `json:"reviewed_at,omitempty"`                               // 审核时间
	PaidAt       *time.Time     `json:"paid_at,omitempty"`                                   // 打款时间
	CreatedAt    time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

/**
 * Approve 通过审核
 */
func (w *WithdrawRequest) Approve(reviewerID uint, note string) {
	now := time.Now()
	w.Status = WithdrawStatusApproved
	w.ReviewerID = reviewerID
	w.ReviewNote = note
	w.ReviewedAt = &now
}

/**
 * Reject 驳回审核
 */
func (w *WithdrawRequest) Reject(reviewerID uint, note string) {
	now := time.Now()
	w.Status = WithdrawStatusRejected
	w.ReviewerID = reviewerID
	w.ReviewNote = note
	w.ReviewedAt = &now
}

/**
 * MarkAsPaid 标记为已打款
 */
func (w *WithdrawRequest) MarkAsPaid(voucherURL string) {
	now := time.Now()
	w.Status = WithdrawStatusPaid
	w.VoucherURL = voucherURL
	w.PaidAt = &now
}

/**
 * IsPending 判断是否待审核
 */
func (w *WithdrawRequest) IsPending() bool {
	return w.Status == WithdrawStatusPending
}

/**
 * IsApproved 判断是否已通过
 */
func (w *WithdrawRequest) IsApproved() bool {
	return w.Status == WithdrawStatusApproved
}
