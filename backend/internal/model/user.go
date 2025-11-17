package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID               uint           `gorm:"primarykey"`
	Phone            string         `gorm:"type:varchar(20);uniqueIndex;not null"`
	Password         string         `gorm:"type:varchar(255);not null"`
	Role             string         `gorm:"type:varchar(20);index;not null"` // customer/sales/support/super_admin
	SalesID          uint           `gorm:"index"`                             // 归属销售ID
	AvailableDeposit float64        `gorm:"type:decimal(15,2);default:0"`     // 可用定金
	UsedDeposit      float64        `gorm:"type:decimal(15,2);default:0"`     // 已用定金（冻结）

	PayPassword   string `gorm:"type:varchar(255)"` // 支付密码（单独加密存储）
	HasPayPassword bool   `gorm:"default:false"`
	
	// 自动补定金功能（只能由客服启用）
	AutoSupplementEnabled bool `gorm:"default:false"` // 是否启用自动补定金

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
