package repository

import (
	"errors"

	"gorm.io/gorm"

	"suxin/internal/model"
)

// UserVerificationRepository 实名认证仓储
//
// 封装 UserVerification 的常用查询与保存操作。
// 约定：
// - FindByUserID 在记录不存在时返回 (nil, nil)，方便上层做判空处理。
// - 其他数据库错误会原样返回。

type UserVerificationRepository struct {
	db *gorm.DB
}

func NewUserVerificationRepository(db *gorm.DB) *UserVerificationRepository {
	return &UserVerificationRepository{db: db}
}

// FindByUserID 根据用户ID查询实名认证记录
// 不存在时返回 (nil, nil)
func (r *UserVerificationRepository) FindByUserID(userID uint) (*model.UserVerification, error) {
	var v model.UserVerification
	if err := r.db.Where("user_id = ?", userID).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

// Create 创建实名认证记录
func (r *UserVerificationRepository) Create(v *model.UserVerification) error {
	return r.db.Create(v).Error
}

// Update 更新实名认证记录
func (r *UserVerificationRepository) Update(v *model.UserVerification) error {
	return r.db.Save(v).Error
}
