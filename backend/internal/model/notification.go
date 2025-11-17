/**
 * 通知领域模型
 * 
 * 用途：
 * - 定义系统通知数据结构
 * - 支持多种通知类型和渠道
 * - 记录通知发送状态和历史
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
 * 通知类型常量
 */
const (
	NotifyTypeSystem    = "system"     // 系统通知
	NotifyTypeTrade     = "trade"      // 交易通知
	NotifyTypeRisk      = "risk"       // 风控通知
	NotifyTypeFund      = "fund"       // 资金通知
	NotifyTypeAnnounce  = "announce"   // 系统公告
)

/**
 * 通知级别常量
 */
const (
	NotifyLevelInfo     = "info"       // 普通信息
	NotifyLevelWarning  = "warning"    // 警告
	NotifyLevelCritical = "critical"   // 紧急
)

/**
 * 通知状态常量
 */
const (
	NotifyStatusPending = "pending"    // 待发送
	NotifyStatusSent    = "sent"       // 已发送
	NotifyStatusRead    = "read"       // 已读
	NotifyStatusFailed  = "failed"     // 发送失败
)

/**
 * Notification 通知实体
 * 
 * 字段说明：
 * - UserID: 接收用户ID（0表示系统广播）
 * - Type: 通知类型（system/trade/risk/fund/announce）
 * - Level: 通知级别（info/warning/critical）
 * - Title: 通知标题
 * - Content: 通知内容
 * - RelatedID: 关联业务ID（如订单ID、充值ID）
 * - RelatedType: 关联业务类型（order/deposit/withdraw）
 * - Status: 通知状态
 * - ReadAt: 阅读时间
 */
type Notification struct {
	ID          uint           `gorm:"primarykey"`
	UserID      uint           `gorm:"index;not null"`                        // 接收用户ID
	Type        string         `gorm:"type:varchar(20);index;not null"`       // 通知类型
	Level       string         `gorm:"type:varchar(20);default:'info'"`       // 通知级别
	Title       string         `gorm:"type:varchar(200);not null"`            // 标题
	Content     string         `gorm:"type:text;not null"`                    // 内容
	RelatedID   uint           `gorm:"default:0"`                             // 关联业务ID
	RelatedType string         `gorm:"type:varchar(50)"`                      // 关联业务类型
	Status      string         `gorm:"type:varchar(20);index;default:'pending'"` // 状态
	ReadAt      *time.Time     // 阅读时间
	CreatedAt   time.Time      `gorm:"index"`
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

/**
 * MarkAsRead 标记通知为已读
 * 
 * @return void
 */
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.Status = NotifyStatusRead
	n.ReadAt = &now
}

/**
 * MarkAsSent 标记通知为已发送
 * 
 * @return void
 */
func (n *Notification) MarkAsSent() {
	n.Status = NotifyStatusSent
}

/**
 * MarkAsFailed 标记通知为发送失败
 * 
 * @return void
 */
func (n *Notification) MarkAsFailed() {
	n.Status = NotifyStatusFailed
}

/**
 * IsUnread 判断通知是否未读
 * 
 * @return bool
 */
func (n *Notification) IsUnread() bool {
	return n.Status != NotifyStatusRead
}

/**
 * IsCritical 判断是否为紧急通知
 * 
 * @return bool
 */
func (n *Notification) IsCritical() bool {
	return n.Level == NotifyLevelCritical
}
