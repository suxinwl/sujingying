/**
 * 提现服务层
 * 
 * 用途：
 * - 实现提现业务逻辑
 * - 处理提现申请和审核
 * - 管理资金扣减和流水
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

type WithdrawService struct {
	ctx          *appctx.AppContext
	withdrawRepo *repository.WithdrawRepository
	userRepo     *repository.UserRepository
	fundLogRepo  *repository.FundLogRepository
	cardRepo     *repository.BankCardRepository
	notiSvc      *NotificationService
}

func NewWithdrawService(ctx *appctx.AppContext) *WithdrawService {
	return &WithdrawService{
		ctx:          ctx,
		withdrawRepo: repository.NewWithdrawRepository(ctx.DB),
		userRepo:     repository.NewUserRepository(ctx.DB),
		fundLogRepo:  repository.NewFundLogRepository(ctx.DB),
		cardRepo:     repository.NewBankCardRepository(ctx.DB),
		notiSvc:      NewNotificationService(ctx),
	}
}

/**
 * SubmitWithdraw 提交提现申请
 * 
 * @param userID uint - 用户ID
 * @param bankCardID uint - 银行卡ID
 * @param amount float64 - 提现金额
 * @param note string - 用户备注
 * @return (*model.WithdrawRequest, error)
 */
func (s *WithdrawService) SubmitWithdraw(userID, bankCardID uint, amount float64, note string) (*model.WithdrawRequest, error) {
	// 1. 验证提现金额
	if amount <= 0 {
		return nil, errors.New("提现金额必须大于0")
	}

	// 2. 查询用户信息
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 3. 验证银行卡归属
	card, err := s.cardRepo.FindByID(bankCardID)
	if err != nil {
		return nil, errors.New("银行卡不存在")
	}
	if card.UserID != userID {
		return nil, errors.New("银行卡不属于当前用户")
	}

	// 4. 计算手续费（可配置，这里设为0）
	fee := 0.0
	actualAmount := amount - fee

	// 5. 验证余额是否充足
	if user.AvailableDeposit < amount {
		return nil, fmt.Errorf("可用余额不足，当前可用: %.2f", user.AvailableDeposit)
	}

	// 6. 创建提现申请
	withdraw := &model.WithdrawRequest{
		UserID:       userID,
		BankCardID:   bankCardID,
		Amount:       amount,
		Fee:          fee,
		ActualAmount: actualAmount,
		UserNote:     note,
		Status:       model.WithdrawStatusPending,
	}

	if err := s.withdrawRepo.Create(withdraw); err != nil {
		return nil, fmt.Errorf("创建提现申请失败: %v", err)
	}

	return withdraw, nil
}

/**
 * ApproveWithdraw 审核通过提现
 * 
 * @param withdrawID uint - 提现ID
 * @param reviewerID uint - 审核人ID
 * @param note string - 审核备注
 * @return error
 */
func (s *WithdrawService) ApproveWithdraw(withdrawID, reviewerID uint, note string) error {
	// 1. 查找提现申请
	withdraw, err := s.withdrawRepo.FindByID(withdrawID)
	if err != nil {
		return errors.New("提现申请不存在")
	}

	// 2. 验证状态
	if !withdraw.IsPending() {
		return fmt.Errorf("提现状态不允许审核（当前状态: %s）", withdraw.Status)
	}

	// 3. 查询用户
	user, err := s.userRepo.FindByID(withdraw.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 4. 再次验证余额
	if user.AvailableDeposit < withdraw.Amount {
		return errors.New("用户余额不足")
	}

	// 5. 开启事务
	tx := s.ctx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 6. 扣减用户余额
	newAvailable := user.AvailableDeposit - withdraw.Amount
	if err := tx.Model(&model.User{}).Where("id = ?", user.ID).
		Update("available_deposit", newAvailable).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("扣减余额失败: %v", err)
	}

	// 7. 记录资金流水
	fundLog := &model.FundLog{
		UserID:          user.ID,
		Type:            model.FundLogTypeWithdraw,
		Amount:          -withdraw.Amount,
		AvailableBefore: user.AvailableDeposit,
		AvailableAfter:  newAvailable,
		UsedBefore:      user.UsedDeposit,
		UsedAfter:       user.UsedDeposit,
		RelatedID:       withdraw.ID,
		RelatedType:     "withdraw",
		Note:            fmt.Sprintf("提现: %.2f元", withdraw.Amount),
	}
	if err := tx.Create(fundLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录流水失败: %v", err)
	}

	// 8. 更新提现状态
	withdraw.Approve(reviewerID, note)
	if err := tx.Save(withdraw).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新提现状态失败: %v", err)
	}

	// 9. 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("事务提交失败: %v", err)
	}

	// 10. 发送通知
	notifyMsg := fmt.Sprintf("您的提现申请已通过\n提现金额：%.2f 元\n预计到账：%.2f 元",
		withdraw.Amount, withdraw.ActualAmount)
	s.notiSvc.SendFundNotification(user.ID, "提现通过", notifyMsg)

	return nil
}

/**
 * MarkWithdrawPaid 标记提现已打款并保存打款凭证
 *
 * @param withdrawID uint - 提现ID
 * @param voucherURL string - 打款凭证（Base64或URL）
 * @return error
 */
func (s *WithdrawService) MarkWithdrawPaid(withdrawID uint, voucherURL string) error {
	withdraw, err := s.withdrawRepo.FindByID(withdrawID)
	if err != nil {
		return errors.New("提现申请不存在")
	}

	if !withdraw.IsApproved() {
		return errors.New("只有已通过的提现才能标记为已打款")
	}

	withdraw.MarkAsPaid(voucherURL)
	if err := s.withdrawRepo.Update(withdraw); err != nil {
		return fmt.Errorf("更新提现状态失败: %v", err)
	}

	notifyMsg := fmt.Sprintf("您的提现已打款\n提现金额：%.2f 元", withdraw.Amount)
	s.notiSvc.SendFundNotification(withdraw.UserID, "提现已打款", notifyMsg)

	return nil
}

/**
 * RejectWithdraw 驳回提现
 * 
 * @param withdrawID uint - 提现ID
 * @param reviewerID uint - 审核人ID
 * @param note string - 驳回原因
 * @return error
 */
func (s *WithdrawService) RejectWithdraw(withdrawID, reviewerID uint, note string) error {
	withdraw, err := s.withdrawRepo.FindByID(withdrawID)
	if err != nil {
		return errors.New("提现申请不存在")
	}

	if !withdraw.IsPending() {
		return errors.New("提现状态不允许审核")
	}

	withdraw.Reject(reviewerID, note)
	if err := s.withdrawRepo.Update(withdraw); err != nil {
		return err
	}

	// 发送通知
	notifyMsg := fmt.Sprintf("您的提现申请已被驳回\n驳回原因：%s", note)
	s.notiSvc.SendFundNotification(withdraw.UserID, "提现驳回", notifyMsg)

	return nil
}

/**
 * GetUserWithdraws 查询用户提现记录
 */
func (s *WithdrawService) GetUserWithdraws(userID uint, limit, offset int) ([]*model.WithdrawRequest, error) {
	return s.withdrawRepo.FindByUserID(userID, limit, offset)
}

/**
 * GetPendingWithdraws 获取待审核列表
 */
func (s *WithdrawService) GetPendingWithdraws(limit int) ([]*model.WithdrawRequest, error) {
	return s.withdrawRepo.FindPending(limit)
}

/**
 * GetWithdrawsByStatus 根据状态获取提现申请
 *
 * @param status string - 状态
 * @param limit int - 查询数量限制
 * @return ([]*model.WithdrawRequest, error)
 */
func (s *WithdrawService) GetWithdrawsByStatus(status string, limit int) ([]*model.WithdrawRequest, error) {
	return s.withdrawRepo.FindByStatus(status, limit)
}
