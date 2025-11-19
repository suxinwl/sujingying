/**
 * 通知仓储层
 * 
 * 用途：
 * - 封装通知数据访问逻辑
 * - 提供通知CRUD操作
 * - 支持按条件查询通知
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package repository

import (
	"gorm.io/gorm"

	"suxin/internal/model"
)

/**
 * NotificationRepository 通知仓储
 */
type NotificationRepository struct {
	db *gorm.DB
}

/**
 * NewNotificationRepository 创建通知仓储实例
 * 
 * @param db *gorm.DB - 数据库连接
 * @return *NotificationRepository
 */
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

/**
 * Create 创建通知
 * 
 * @param notification *model.Notification - 通知实体
 * @return error
 */
func (r *NotificationRepository) Create(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

/**
 * FindByID 根据ID查找通知
 * 
 * @param id uint - 通知ID
 * @return (*model.Notification, error)
 */
func (r *NotificationRepository) FindByID(id uint) (*model.Notification, error) {
	var notification model.Notification
	if err := r.db.First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

/**
 * FindByUserID 查询用户的通知列表
 * 
 * @param userID uint - 用户ID
 * @param limit int - 查询数量限制
 * @param offset int - 偏移量
 * @return ([]*model.Notification, error)
 */
func (r *NotificationRepository) FindByUserID(userID uint, limit, offset int) ([]*model.Notification, error) {
	var notifications []*model.Notification
	err := r.db.Where("user_id = ? OR (user_id = 0 AND type = ?)", userID, model.NotifyTypeAnnounce).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

/**
 * FindUnreadByUserID 查询用户未读通知
 * 
 * @param userID uint - 用户ID
 * @return ([]*model.Notification, error)
 */
func (r *NotificationRepository) FindUnreadByUserID(userID uint) ([]*model.Notification, error) {
	var notifications []*model.Notification
	err := r.db.Where("user_id = ? AND status != ?", userID, model.NotifyStatusRead).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

/**
 * CountUnreadByUserID 统计用户未读通知数量
 * 
 * @param userID uint - 用户ID
 * @return (int64, error)
 */
func (r *NotificationRepository) CountUnreadByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND status != ?", userID, model.NotifyStatusRead).
		Count(&count).Error
	return count, err
}

/**
 * FindAnnouncements 查询平台公告列表
 * 
 * 公告使用 Notification 表中 type=announce 且 user_id=0 的记录。
 * 
 * @param limit int - 查询数量限制
 * @param offset int - 偏移量
 * @return ([]*model.Notification, error)
 */
func (r *NotificationRepository) FindAnnouncements(limit, offset int) ([]*model.Notification, error) {
	var notifications []*model.Notification
	err := r.db.Where("type = ? AND user_id = 0", model.NotifyTypeAnnounce).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

/**
 * Update 更新通知
 * 
 * @param notification *model.Notification - 通知实体
 * @return error
 */
func (r *NotificationRepository) Update(notification *model.Notification) error {
	return r.db.Save(notification).Error
}

/**
 * MarkAsRead 批量标记通知为已读
 * 
 * @param ids []uint - 通知ID列表
 * @param userID uint - 用户ID（验证归属）
 * @return error
 */
func (r *NotificationRepository) MarkAsRead(ids []uint, userID uint) error {
	return r.db.Model(&model.Notification{}).
		Where("id IN ? AND user_id = ?", ids, userID).
		Updates(map[string]interface{}{
			"status":  model.NotifyStatusRead,
			"read_at": r.db.NowFunc(),
		}).Error
}

/**
 * MarkAllAsRead 标记用户所有通知为已读
 * 
 * @param userID uint - 用户ID
 * @return error
 */
func (r *NotificationRepository) MarkAllAsRead(userID uint) error {
	// 先检查是否有未读通知
	var count int64
	r.db.Model(&model.Notification{}).
		Where("user_id = ? AND status != ?", userID, model.NotifyStatusRead).
		Count(&count)
	
	// 如果没有未读通知，直接返回成功
	if count == 0 {
		return nil
	}
	
	// 有未读通知，执行更新
	return r.db.Model(&model.Notification{}).
		Where("user_id = ? AND status != ?", userID, model.NotifyStatusRead).
		Updates(map[string]interface{}{
			"status":  model.NotifyStatusRead,
			"read_at": r.db.NowFunc(),
		}).Error
}

/**
 * DeleteOldNotifications 删除旧通知（超过30天的已读通知）
 * 
 * @return error
 */
func (r *NotificationRepository) DeleteOldNotifications() error {
	thirtyDaysAgo := r.db.NowFunc().AddDate(0, 0, -30)
	return r.db.Where("status = ? AND created_at < ?", 
		model.NotifyStatusRead, thirtyDaysAgo).
		Delete(&model.Notification{}).Error
}
