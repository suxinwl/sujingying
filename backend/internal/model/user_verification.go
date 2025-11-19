package model

import (
	"time"

	"gorm.io/gorm"
)

// UserVerification 实名认证信息
//
// 用途：
// - 存储用户提交的身份证、银行卡、收货地址等实名资料
// - 支持管理员在后台进行审核
//
// 状态：
// - pending  待审核
// - approved 已通过
// - rejected 已驳回
//
// 注意：图片字段为外部存储的 URL，由上传接口单独负责生成
//
// 仅在实名认证审核相关接口中使用，不在普通业务接口中直接暴露
//
// 为兼容 AutoMigrate，请在 internal/pkg/database/database.go 中注册本模型。
//
// 字段说明：
// - UserID: 关联的用户ID
// - RealName: 真实姓名
// - IDNumber: 身份证号码
// - IDFrontURL: 身份证正面照片 URL
// - IDBackURL: 身份证反面照片 URL
// - BankCardID: 绑定的银行卡ID（来自 bank_cards 表）
// - ReceiverName/Phone/Province/City/District/AddressDetail: 收货/联系地址
// - Status: 审核状态（pending/approved/rejected）
// - Remark: 审核备注
// - AuditorID: 审核管理员用户ID
//
// 软删除用于保留历史记录。
//
//go:generate stringer -type=UserVerificationStatus

// UserVerification 实名认证记录
type UserVerification struct {
	ID             uint           `gorm:"primarykey"`
	UserID         uint           `gorm:"uniqueIndex;not null"`                  // 用户ID（一个用户一条最新记录）
	RealName       string         `gorm:"type:varchar(50);not null"`             // 真实姓名
	IDNumber       string         `gorm:"type:varchar(32);not null"`             // 身份证号码
	IDFrontURL     string         `gorm:"type:varchar(255)"`                     // 身份证正面照片
	IDBackURL      string         `gorm:"type:varchar(255)"`                     // 身份证反面照片
	BankCardID     uint           `gorm:"index"`                                 // 绑定银行卡ID
	ReceiverName   string         `gorm:"type:varchar(50)"`                      // 收货人姓名
	ReceiverPhone  string         `gorm:"type:varchar(20)"`                      // 收货人电话
	Province       string         `gorm:"type:varchar(50)"`                      // 省
	City           string         `gorm:"type:varchar(50)"`                      // 市
	District       string         `gorm:"type:varchar(50)"`                      // 区/县
	AddressDetail  string         `gorm:"type:varchar(255)"`                     // 详细地址
	Status         string         `gorm:"type:varchar(20);default:'pending'"`    // 审核状态
	Remark         string         `gorm:"type:varchar(255)"`                     // 审核备注
	AuditorID      uint           `gorm:"index"`                                 // 审核管理员用户ID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// 审核状态常量
const (
	VerificationStatusPending  = "pending"  // 待审核
	VerificationStatusApproved = "approved" // 已通过
	VerificationStatusRejected = "rejected" // 已驳回
)
