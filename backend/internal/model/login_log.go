package model

import (
	"time"
)

type LoginLog struct {
	ID         uint      `gorm:"primarykey"`
	UserID     uint      `gorm:"index"`
	Phone      string    `gorm:"type:varchar(20);index"`
	IP         string    `gorm:"type:varchar(45)"`
	UserAgent  string    `gorm:"type:varchar(200)"`
	Status     string    `gorm:"type:varchar(20);index"` // success/failed/blocked
	FailReason string    `gorm:"type:varchar(200)"`
	CreatedAt  time.Time `gorm:"index"`
}
