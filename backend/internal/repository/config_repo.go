/**
 * 系统配置仓储层
 * 
 * 用途：
 * - 封装配置数据访问
 * - 提供配置查询和更新
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package repository

import (
	"gorm.io/gorm"

	"suxin/internal/model"
)

type ConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

func (r *ConfigRepository) Create(config *model.SystemConfig) error {
	return r.db.Create(config).Error
}

func (r *ConfigRepository) FindByKey(key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	if err := r.db.Where("key = ?", key).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *ConfigRepository) FindByCategory(category string) ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := r.db.Where("category = ?", category).Find(&configs).Error
	return configs, err
}

func (r *ConfigRepository) FindAll() ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := r.db.Find(&configs).Error
	return configs, err
}

func (r *ConfigRepository) Update(config *model.SystemConfig) error {
	return r.db.Save(config).Error
}

func (r *ConfigRepository) UpdateValue(key, value string) error {
	return r.db.Model(&model.SystemConfig{}).
		Where("key = ?", key).
		Update("value", value).Error
}

func (r *ConfigRepository) Delete(id uint) error {
	return r.db.Delete(&model.SystemConfig{}, id).Error
}
