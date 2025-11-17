/**
 * 通知服务层
 * 
 * 用途：
 * - 实现通知业务逻辑
 * - 处理通知发送、查询、标记已读
 * - 集成多种通知渠道
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package service

import (
	"fmt"
	"log"

	"suxin/internal/appctx"
	"suxin/internal/model"
	"suxin/internal/repository"
)

/**
 * NotificationService 通知服务
 */
type NotificationService struct {
	ctx      *appctx.AppContext
	notiRepo *repository.NotificationRepository
	hub      NotificationHubInterface
}

// NotificationHubInterface 通知Hub接口
type NotificationHubInterface interface {
	SendToUser(userID uint, notification *model.Notification)
	IsUserOnline(userID uint) bool
}

/**
 * NewNotificationService 创建通知服务实例
 * 
 * @param ctx *appctx.AppContext - 应用上下文
 * @return *NotificationService
 */
func NewNotificationService(ctx *appctx.AppContext) *NotificationService {
	return &NotificationService{
		ctx:      ctx,
		notiRepo: repository.NewNotificationRepository(ctx.DB),
		hub:      nil, // Hub会在需要时从AppContext中获取
	}
}

/**
 * SetNotificationHub 设置通知Hub（用于依赖注入）
 */
func (s *NotificationService) SetNotificationHub(hub NotificationHubInterface) {
	s.hub = hub
}

/**
 * SendNotification 发送通知
 * 
 * 业务流程：
 * 1. 创建通知记录
 * 2. 保存到数据库
 * 3. 推送到WebSocket（如果用户在线）
 * 4. 推送到其他渠道（短信/邮件/APP）
 * 
 * @param userID uint - 接收用户ID
 * @param notifyType string - 通知类型
 * @param level string - 通知级别
 * @param title string - 标题
 * @param content string - 内容
 * @param relatedID uint - 关联业务ID
 * @param relatedType string - 关联业务类型
 * @return (*model.Notification, error)
 */
func (s *NotificationService) SendNotification(
	userID uint,
	notifyType, level, title, content string,
	relatedID uint,
	relatedType string,
) (*model.Notification, error) {
	
	// 1. 创建通知实体
	notification := &model.Notification{
		UserID:      userID,
		Type:        notifyType,
		Level:       level,
		Title:       title,
		Content:     content,
		RelatedID:   relatedID,
		RelatedType: relatedType,
		Status:      model.NotifyStatusPending,
	}
	
	// 2. 保存到数据库
	if err := s.notiRepo.Create(notification); err != nil {
		log.Printf("[Notify] 创建通知失败: %v", err)
		return nil, err
	}
	
	// 3. 推送到WebSocket（异步）
	go s.pushToWebSocket(notification)
	
	// 4. 标记为已发送
	notification.MarkAsSent()
	s.notiRepo.Update(notification)
	
	log.Printf("[Notify] ✅ 通知已发送: UserID=%d, Type=%s, Level=%s, Title=%s",
		userID, notifyType, level, title)
	
	return notification, nil
}

/**
 * SendRiskNotification 发送风控通知
 * 
 * 用于：强平通知、预警通知
 * 
 * @param userID uint - 用户ID
 * @param orderID string - 订单号
 * @param message string - 消息内容
 * @param isCritical bool - 是否紧急
 * @return error
 */
func (s *NotificationService) SendRiskNotification(
	userID uint,
	orderID string,
	message string,
	isCritical bool,
) error {
	
	level := model.NotifyLevelWarning
	title := "风控预警"
	
	if isCritical {
		level = model.NotifyLevelCritical
		title = "强制平仓通知"
	}
	
	content := fmt.Sprintf("订单号：%s\n%s", orderID, message)
	
	_, err := s.SendNotification(
		userID,
		model.NotifyTypeRisk,
		level,
		title,
		content,
		0,
		"order",
	)
	
	return err
}

/**
 * SendOrderNotification 发送订单通知
 * 
 * @param userID uint - 用户ID
 * @param orderID uint - 订单ID
 * @param title string - 标题
 * @param content string - 内容
 * @return error
 */
func (s *NotificationService) SendOrderNotification(
	userID uint,
	orderID uint,
	title, content string,
) error {
	
	_, err := s.SendNotification(
		userID,
		model.NotifyTypeTrade,
		model.NotifyLevelInfo,
		title,
		content,
		orderID,
		"order",
	)
	
	return err
}

/**
 * SendFundNotification 发送资金通知
 * 
 * @param userID uint - 用户ID
 * @param title string - 标题
 * @param content string - 内容
 * @return error
 */
func (s *NotificationService) SendFundNotification(
	userID uint,
	title, content string,
) error {
	
	_, err := s.SendNotification(
		userID,
		model.NotifyTypeFund,
		model.NotifyLevelInfo,
		title,
		content,
		0,
		"fund",
	)
	
	return err
}

/**
 * GetUserNotifications 获取用户通知列表
 * 
 * @param userID uint - 用户ID
 * @param limit int - 查询数量限制
 * @param offset int - 偏移量
 * @return ([]*model.Notification, error)
 */
func (s *NotificationService) GetUserNotifications(userID uint, limit, offset int) ([]*model.Notification, error) {
	return s.notiRepo.FindByUserID(userID, limit, offset)
}

/**
 * GetUnreadNotifications 获取用户未读通知
 * 
 * @param userID uint - 用户ID
 * @return ([]*model.Notification, error)
 */
func (s *NotificationService) GetUnreadNotifications(userID uint) ([]*model.Notification, error) {
	return s.notiRepo.FindUnreadByUserID(userID)
}

/**
 * GetUnreadCount 获取用户未读通知数量
 * 
 * @param userID uint - 用户ID
 * @return (int64, error)
 */
func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	return s.notiRepo.CountUnreadByUserID(userID)
}

/**
 * MarkAsRead 标记通知为已读
 * 
 * @param ids []uint - 通知ID列表
 * @param userID uint - 用户ID
 * @return error
 */
func (s *NotificationService) MarkAsRead(ids []uint, userID uint) error {
	return s.notiRepo.MarkAsRead(ids, userID)
}

/**
 * MarkAllAsRead 标记所有通知为已读
 * 
 * @param userID uint - 用户ID
 * @return error
 */
func (s *NotificationService) MarkAllAsRead(userID uint) error {
	return s.notiRepo.MarkAllAsRead(userID)
}

/**
 * pushToWebSocket 推送通知到WebSocket
 * 
 * @param notification *model.Notification - 通知实体
 * @return void
 */
func (s *NotificationService) pushToWebSocket(notification *model.Notification) {
	// 如果Hub未设置，跳过推送
	if s.hub == nil {
		log.Printf("[Notify] ⚠️ NotificationHub未设置，跳过WebSocket推送")
		return
	}
	
	// 检查用户是否在线
	if !s.hub.IsUserOnline(notification.UserID) {
		log.Printf("[Notify] 用户 %d 不在线，跳过WebSocket推送", notification.UserID)
		return
	}
	
	// 推送通知
	s.hub.SendToUser(notification.UserID, notification)
	log.Printf("[Notify] ✅ WebSocket通知已推送到用户 %d: %s", notification.UserID, notification.Title)
}
