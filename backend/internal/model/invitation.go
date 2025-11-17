/**
 * 邀请码系统模型
 * 
 * 用途：
 * - 用户推荐营销
 * - 邀请关系追踪
 * - 推广统计
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
 * InvitationCode 邀请码实体
 * 
 * 字段说明：
 * - UserID: 邀请码所属用户ID
 * - Code: 邀请码（唯一）
 * - InviteCount: 邀请成功数量
 * - RegisterCount: 通过此码注册的用户数
 * - IsActive: 是否激活
 */
type InvitationCode struct {
	ID            uint           `gorm:"primarykey"`
	UserID        uint           `gorm:"uniqueIndex;not null"`        // 用户ID（一个用户一个邀请码）
	Code          string         `gorm:"type:varchar(20);uniqueIndex;not null"` // 邀请码
	InviteCount   int            `gorm:"default:0"`                   // 邀请成功数量
	RegisterCount int            `gorm:"default:0"`                   // 注册数量
	IsActive      bool           `gorm:"default:true"`                // 是否激活
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

/**
 * InvitationRecord 邀请记录
 * 
 * 记录邀请关系和奖励发放
 */
type InvitationRecord struct {
	ID            uint           `gorm:"primarykey"`
	InviterID     uint           `gorm:"index;not null"`              // 邀请人ID
	InviteeID     uint           `gorm:"index;not null"`              // 被邀请人ID
	InviteCode    string         `gorm:"type:varchar(20);index"`      // 使用的邀请码
	RewardAmount  float64        `gorm:"type:decimal(15,2);default:0"` // 奖励金额
	RewardStatus  string         `gorm:"type:varchar(20);default:'pending'"` // 奖励状态: pending/issued
	IssuedAt      *time.Time     // 奖励发放时间
	CreatedAt     time.Time      `gorm:"index"`
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

/**
 * TeamRelation 团队层级关系
 * 
 * 记录上下级关系，支持多级团队
 */
type TeamRelation struct {
	ID           uint           `gorm:"primarykey"`
	UserID       uint           `gorm:"index;not null"`              // 用户ID
	ParentID     uint           `gorm:"index;default:0"`             // 直接上级ID
	AncestorIDs  string         `gorm:"type:varchar(500)"`           // 祖先ID链（逗号分隔）
	Level        int            `gorm:"default:1"`                   // 层级（1=顶级）
	DirectCount  int            `gorm:"default:0"`                   // 直接下级数量
	TeamCount    int            `gorm:"default:0"`                   // 团队总人数
	TeamPoints   float64        `gorm:"type:decimal(15,2);default:0"` // 团队总积分
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

/**
 * GetAncestorList 获取祖先ID列表
 */
func (t *TeamRelation) GetAncestorList() []string {
	if t.AncestorIDs == "" {
		return []string{}
	}
	return []string{} // TODO: 实现逗号分隔字符串解析
}

/**
 * AddAncestor 添加祖先ID
 */
func (t *TeamRelation) AddAncestor(ancestorID uint) {
	if t.AncestorIDs == "" {
		t.AncestorIDs = ""
	}
	// TODO: 实现添加逻辑
}
