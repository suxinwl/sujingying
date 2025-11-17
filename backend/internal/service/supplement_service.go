/**
 * 补定金服务层
 * 
 * 用途：
 * - 处理补定金申请
 * - 审核补定金
 * - 增加订单定金
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package service

import (
	"errors"
	"fmt"

	"suxin/internal/appctx"
	"suxin/internal/model"
	"suxin/internal/repository"
)

type SupplementService struct {
	ctx           *appctx.AppContext
	supplementRepo *repository.SupplementRepository
	orderRepo     *repository.OrderRepository
	userRepo      *repository.UserRepository
	fundLogRepo   *repository.FundLogRepository
	notiSvc       *NotificationService
}

func NewSupplementService(ctx *appctx.AppContext) *SupplementService {
	return &SupplementService{
		ctx:           ctx,
		supplementRepo: repository.NewSupplementRepository(ctx.DB),
		orderRepo:     repository.NewOrderRepository(ctx.DB),
		userRepo:      repository.NewUserRepository(ctx.DB),
		fundLogRepo:   repository.NewFundLogRepository(ctx.DB),
		notiSvc:       NewNotificationService(ctx),
	}
}

/**
 * SubmitSupplement 提交补定金申请
 * 
 * @param userID uint - 用户ID
 * @param orderID uint - 订单ID
 * @param amount float64 - 补充金额
 * @param method string - 补充方式
 * @param voucherURL string - 凭证URL
 * @return (*model.SupplementDeposit, error)
 */
func (s *SupplementService) SubmitSupplement(userID, orderID uint, amount float64, method, voucherURL string) (*model.SupplementDeposit, error) {
	// 1. 验证金额
	if amount <= 0 {
		return nil, errors.New("补充金额必须大于0")
	}

	// 2. 验证订单
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}

	// 3. 验证订单归属
	if order.UserID != userID {
		return nil, errors.New("无权操作此订单")
	}

	// 4. 验证订单状态
	if order.Status != model.OrderStatusHolding {
		return nil, errors.New("只能为持仓订单补充定金")
	}

	// 5. 创建补定金申请
	supplement := &model.SupplementDeposit{
		UserID:     userID,
		OrderID:    orderID,
		Amount:     amount,
		Method:     method,
		VoucherURL: voucherURL,
		Status:     model.SupplementStatusPending,
	}

	if err := s.supplementRepo.Create(supplement); err != nil {
		return nil, fmt.Errorf("创建申请失败: %v", err)
	}

	return supplement, nil
}

/**
 * ApproveSupplement 审核通过补定金
 * 
 * @param supplementID uint - 补定金ID
 * @param reviewerID uint - 审核人ID
 * @param note string - 审核备注
 * @return error
 */
func (s *SupplementService) ApproveSupplement(supplementID, reviewerID uint, note string) error {
	// 1. 查找补定金申请
	supplement, err := s.supplementRepo.FindByID(supplementID)
	if err != nil {
		return errors.New("补定金申请不存在")
	}

	// 2. 验证状态
	if !supplement.IsPending() {
		return fmt.Errorf("补定金状态不允许审核（当前状态: %s）", supplement.Status)
	}

	// 3. 查询订单
	order, err := s.orderRepo.FindByID(supplement.OrderID)
	if err != nil {
		return errors.New("订单不存在")
	}

	// 4. 查询用户
	user, err := s.userRepo.FindByID(supplement.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 5. 开启事务
	tx := s.ctx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 6. 增加订单定金
	oldDeposit := order.Deposit
	order.Deposit += supplement.Amount
	
	// 重新计算定金率
	if order.CurrentPrice > 0 {
		order.UpdatePnLAndMargin(order.CurrentPrice)
	}
	
	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单定金失败: %v", err)
	}

	// 7. 扣减用户可用定金，增加已用定金
	newAvailable := user.AvailableDeposit - supplement.Amount
	newUsed := user.UsedDeposit + supplement.Amount
	
	if newAvailable < 0 {
		tx.Rollback()
		return errors.New("用户可用定金不足")
	}
	
	if err := tx.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"available_deposit": newAvailable,
		"used_deposit":      newUsed,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新用户资金失败: %v", err)
	}

	// 8. 记录资金流水
	fundLog := &model.FundLog{
		UserID:          user.ID,
		Type:            "supplement",
		Amount:          -supplement.Amount,
		AvailableBefore: user.AvailableDeposit,
		AvailableAfter:  newAvailable,
		UsedBefore:      user.UsedDeposit,
		UsedAfter:       newUsed,
		RelatedID:       order.ID,
		RelatedType:     "order",
		Note:            fmt.Sprintf("补充定金: %.2f元 (订单%s)", supplement.Amount, order.OrderID),
	}
	if err := tx.Create(fundLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录流水失败: %v", err)
	}

	// 9. 更新补定金状态
	supplement.Approve(reviewerID, note)
	if err := tx.Save(supplement).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新补定金状态失败: %v", err)
	}

	// 10. 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("事务提交失败: %v", err)
	}

	// 11. 发送通知
	notifyMsg := fmt.Sprintf("您的补定金申请已通过\n订单号：%s\n补充金额：%.2f 元\n订单定金：%.2f → %.2f 元\n定金率：%.2f%%",
		order.OrderID, supplement.Amount, oldDeposit, order.Deposit, order.MarginRate)
	s.notiSvc.SendFundNotification(user.ID, "补定金成功", notifyMsg)

	return nil
}

/**
 * RejectSupplement 驳回补定金
 */
func (s *SupplementService) RejectSupplement(supplementID, reviewerID uint, note string) error {
	supplement, err := s.supplementRepo.FindByID(supplementID)
	if err != nil {
		return errors.New("补定金申请不存在")
	}

	if !supplement.IsPending() {
		return errors.New("补定金状态不允许审核")
	}

	supplement.Reject(reviewerID, note)
	if err := s.supplementRepo.Update(supplement); err != nil {
		return err
	}

	// 发送通知
	notifyMsg := fmt.Sprintf("您的补定金申请已被驳回\n驳回原因：%s", note)
	s.notiSvc.SendFundNotification(supplement.UserID, "补定金驳回", notifyMsg)

	return nil
}

/**
 * GetUserSupplements 查询用户补定金记录
 */
func (s *SupplementService) GetUserSupplements(userID uint, limit, offset int) ([]*model.SupplementDeposit, error) {
	return s.supplementRepo.FindByUserID(userID, limit, offset)
}

/**
 * GetPendingSupplements 获取待审核列表
 */
func (s *SupplementService) GetPendingSupplements(limit int) ([]*model.SupplementDeposit, error) {
	return s.supplementRepo.FindPending(limit)
}
