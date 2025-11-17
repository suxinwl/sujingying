/**
 * 自动补定金服务层
 * 
 * 用途：
 * - 自动检测订单定金率
 * - 触发自动补定金
 * - 只对启用该功能的用户生效
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

type AutoSupplementService struct {
	ctx             *appctx.AppContext
	userRepo        *repository.UserRepository
	orderRepo       *repository.OrderRepository
	configRepo      *repository.ConfigRepository
	supplementSvc   *SupplementService
}

func NewAutoSupplementService(ctx *appctx.AppContext) *AutoSupplementService {
	return &AutoSupplementService{
		ctx:           ctx,
		userRepo:      repository.NewUserRepository(ctx.DB),
		orderRepo:     repository.NewOrderRepository(ctx.DB),
		configRepo:    repository.NewConfigRepository(ctx.DB),
		supplementSvc: NewSupplementService(ctx),
	}
}

/**
 * CheckAndSupplementOrder 检查并自动补定金
 * 
 * 规则：
 * 1. 用户必须启用自动补定金功能
 * 2. 订单定金率低于触发阈值（默认50%）
 * 3. 用户有足够的可用定金
 * 4. 自动补充到目标阈值（默认80%）
 * 
 * @param orderID uint - 订单ID
 * @return bool - 是否执行了自动补定金
 */
func (s *AutoSupplementService) CheckAndSupplementOrder(orderID uint) bool {
	// 1. 查询订单
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		log.Printf("[AutoSupplement] 订单不存在: %d", orderID)
		return false
	}

	// 2. 只处理持仓订单
	if order.Status != model.OrderStatusHolding {
		return false
	}

	// 3. 查询用户
	user, err := s.userRepo.FindByID(order.UserID)
	if err != nil {
		log.Printf("[AutoSupplement] 用户不存在: %d", order.UserID)
		return false
	}

	// 4. 检查用户是否启用自动补定金
	if !user.AutoSupplementEnabled {
		return false
	}

	// 5. 获取配置阈值
	triggerRate := s.getTriggerRate()  // 默认50%
	targetRate := s.getTargetRate()    // 默认80%

	// 6. 检查定金率是否低于触发阈值
	if order.MarginRate >= triggerRate {
		return false // 定金率正常，无需补充
	}

	// 7. 计算需要补充的金额
	// 目标定金率 = 订单定金 / (订单价值 + 订单定金)
	// 设需要补充 X，则：
	// targetRate = (currentDeposit + X) / (orderValue + currentDeposit + X)
	// targetRate * (orderValue + currentDeposit + X) = currentDeposit + X
	// targetRate * orderValue + targetRate * currentDeposit + targetRate * X = currentDeposit + X
	// targetRate * orderValue = currentDeposit - targetRate * currentDeposit + X - targetRate * X
	// targetRate * orderValue = currentDeposit * (1 - targetRate) + X * (1 - targetRate)
	// X = (targetRate * orderValue - currentDeposit * (1 - targetRate)) / (1 - targetRate)
	
	orderValue := order.WeightG / 31.1035 * order.CurrentPrice // 订单价值（克转盎司 * 当前价格）
	currentDeposit := order.Deposit
	
	// 简化计算：补充到目标定金率
	targetDeposit := orderValue * targetRate / (1 - targetRate)
	supplementAmount := targetDeposit - currentDeposit
	
	if supplementAmount <= 0 {
		return false // 无需补充
	}

	// 8. 检查可用定金是否充足
	if user.AvailableDeposit < supplementAmount {
		log.Printf("[AutoSupplement] 可用定金不足，用户=%d, 需要=%.2f, 可用=%.2f", 
			user.ID, supplementAmount, user.AvailableDeposit)
		
		// 发送提醒通知
		s.sendInsufficientBalanceNotification(user.ID, order.OrderID, supplementAmount, user.AvailableDeposit)
		return false
	}

	// 9. 执行自动补定金
	log.Printf("[AutoSupplement] 开始自动补定金: 用户=%d, 订单=%s, 金额=%.2f, 当前定金率=%.2f%%, 目标=%.2f%%",
		user.ID, order.OrderID, supplementAmount, order.MarginRate, targetRate)
	
	_, err = s.supplementSvc.SubmitSupplement(user.ID, order.ID, supplementAmount)
	if err != nil {
		log.Printf("[AutoSupplement] 自动补定金失败: %v", err)
		return false
	}

	log.Printf("[AutoSupplement] 自动补定金成功: 用户=%d, 订单=%s, 金额=%.2f", 
		user.ID, order.OrderID, supplementAmount)
	
	// 发送成功通知
	s.sendSuccessNotification(user.ID, order.OrderID, supplementAmount, order.MarginRate, targetRate)
	
	return true
}

/**
 * getTriggerRate 获取触发阈值（默认50%）
 */
func (s *AutoSupplementService) getTriggerRate() float64 {
	config, err := s.configRepo.FindByKey(model.ConfigKeyAutoSupplementTrigger)
	if err != nil || config == nil {
		return 50.0 // 默认50%
	}
	
	var rate float64
	fmt.Sscanf(config.Value, "%f", &rate)
	if rate <= 0 || rate >= 100 {
		return 50.0
	}
	return rate
}

/**
 * getTargetRate 获取目标阈值（默认80%）
 */
func (s *AutoSupplementService) getTargetRate() float64 {
	config, err := s.configRepo.FindByKey(model.ConfigKeyAutoSupplementTarget)
	if err != nil || config == nil {
		return 80.0 // 默认80%
	}
	
	var rate float64
	fmt.Sscanf(config.Value, "%f", &rate)
	if rate <= 0 || rate >= 100 {
		return 80.0
	}
	return rate
}

/**
 * sendInsufficientBalanceNotification 发送余额不足通知
 */
func (s *AutoSupplementService) sendInsufficientBalanceNotification(userID uint, orderID string, need, available float64) {
	notiSvc := NewNotificationService(s.ctx)
	msg := fmt.Sprintf("自动补定金失败\n订单号：%s\n需要金额：%.2f 元\n可用定金：%.2f 元\n请及时充值以保护订单安全",
		orderID, need, available)
	notiSvc.SendRiskNotification(userID, orderID, msg, false)
}

/**
 * sendSuccessNotification 发送成功通知
 */
func (s *AutoSupplementService) sendSuccessNotification(userID uint, orderID string, amount, oldRate, newRate float64) {
	notiSvc := NewNotificationService(s.ctx)
	msg := fmt.Sprintf("自动补定金成功\n订单号：%s\n补充金额：%.2f 元\n定金率：%.2f%% → %.2f%%\n订单风险已降低",
		orderID, amount, oldRate, newRate)
	notiSvc.SendFundNotification(userID, "自动补定金成功", msg)
}
