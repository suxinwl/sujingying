package repository

import (
	"errors"

	"gorm.io/gorm"

	"suxin/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByPhone(phone string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("phone = ?", phone).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(u *model.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepository) Update(u *model.User) error {
	return r.db.Save(u).Error
}

func (r *UserRepository) UpdatePayPassword(userID uint, hashed string, has bool) error {
	res := r.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
		"pay_password":   hashed,
		"has_pay_password": has,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}
