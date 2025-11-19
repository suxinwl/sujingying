/**
 * 定金充值服务层
 * 
 * 用途：
 * - 实现充值业务逻辑
 * - 处理充值审核流程
 * - 集成资金流水记录
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package service

import (
	"errors"
	"fmt"
	"log"

	"suxin/internal/appctx"
	"suxin/internal/model"
	"suxin/internal/repository"
)

/**
 * DepositService 充值服务
 */
type DepositService struct {
	ctx          *appctx.AppContext
	depositRepo  *repository.DepositRepository
	userRepo     *repository.UserRepository
	fundLogRepo  *repository.FundLogRepository
	notiSvc      *NotificationService
}

/**
 * NewDepositService 创建充值服务实例
 * 
 * @param ctx *appctx.AppContext - 应用上下文
 * @return *DepositService
 */
func NewDepositService(ctx *appctx.AppContext) *DepositService {
	return &DepositService{
		ctx:         ctx,
		depositRepo: repository.NewDepositRepository(ctx.DB),
		userRepo:    repository.NewUserRepository(ctx.DB),
		fundLogRepo: repository.NewFundLogRepository(ctx.DB),
		notiSvc:     NewNotificationService(ctx),
	}
}

/**
 * SubmitDeposit 提交充值申请
 * 
 * @param userID uint - 用户ID
 * @param amount float64 - 充值金额
 * @param method string - 充值方式
 * @param voucherURL string - 凭证URL
 * @param note string - 用户备注
 * @return (*model.DepositRequest, error)
 */
func (s *DepositService) SubmitDeposit(userID uint, amount float64, method, voucherURL, note string) (*model.DepositRequest, error) {
	// 验证参数
	if amount <= 0 {
		return nil, errors.New("充值金额必须大于0")
	}
	
	// 创建充值申请
	deposit := &model.DepositRequest{
		UserID:     userID,
		Amount:     amount,
		Method:     method,
		VoucherURL: voucherURL,
		UserNote:   note,
		Status:     model.DepositStatusPending,
	}
	
	if err := s.depositRepo.Create(deposit); err != nil {
		return nil, fmt.Errorf("提交充值申请失败: %v", err)
	}
	
	log.Printf("[Deposit] 用户 %d 提交充值申请，金额: %.2f", userID, amount)
	return deposit, nil
}

/**
 * ApproveDeposit 审核通过充值申请
 * 
 * 业务流程：
 * 1. 查找充值申请
 * 2. 验证状态
 * 3. 更新用户可用定金
 * 4. 记录资金流水
 * 5. 更新申请状态
 * 6. 发送通知
 * 
 * @param depositID uint - 充值申请ID
 * @param reviewerID uint - 审核人ID
 * @param note string - 审核备注
 * @param receiptVoucher string - 管理员收款凭证
 * @return error
 */
func (s *DepositService) ApproveDeposit(depositID, reviewerID uint, note, receiptVoucher string) error {
	// 1. 查找充值申请
	deposit, err := s.depositRepo.FindByID(depositID)
	if err != nil {
		return errors.New("充值申请不存在")
	}
	
	// 2. 验证状态
	if !deposit.IsPending() {
		return fmt.Errorf("充值申请状态不允许审核（当前状态: %s）", deposit.Status)
	}
	
	// 3. 查询用户信息
	user, err := s.userRepo.FindByID(deposit.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}
	
	// 4. 开启事务
	tx := s.ctx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 5. 更新用户可用定金
	newAvailable := user.AvailableDeposit + deposit.Amount
	if err := tx.Model(&model.User{}).Where("id = ?", deposit.UserID).
		Update("available_deposit", newAvailable).Error; err != nil {
		tx.Rollback()
		return errors.New("更新用户资金失败")
	}
	
	// 6. 记录资金流水
	fundLog := &model.FundLog{
		UserID:          deposit.UserID,
		Type:            model.FundLogTypeDeposit,
		Amount:          deposit.Amount,
		AvailableBefore: user.AvailableDeposit,
		AvailableAfter:  newAvailable,
		UsedBefore:      user.UsedDeposit,
		UsedAfter:       user.UsedDeposit,
		RelatedID:       deposit.ID,
		RelatedType:     "deposit",
		Note:            fmt.Sprintf("充值审核通过: %s", deposit.Method),
	}
	if err := tx.Create(fundLog).Error; err != nil {
		tx.Rollback()
		return errors.New("记录资金流水失败")
	}
	
	// 7. 更新申请状态和收款凭证
	if receiptVoucher != "" {
		deposit.ReceiptVoucherURL = receiptVoucher
	}
	deposit.Approve(reviewerID, note)
	if err := tx.Save(deposit).Error; err != nil {
		tx.Rollback()
		return errors.New("更新申请状态失败")
	}
	
	// 8. 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.New("事务提交失败")
	}
	
	// 9. 发送通知
	notifyMsg := fmt.Sprintf("您的充值申请已审核通过\n充值金额：%.2f 元\n当前可用定金：%.2f 元", 
		deposit.Amount, newAvailable)
	s.notiSvc.SendFundNotification(deposit.UserID, "充值成功", notifyMsg)
	
	log.Printf("[Deposit] 充值审核通过: ID=%d, 用户=%d, 金额=%.2f", 
		depositID, deposit.UserID, deposit.Amount)
	
	return nil
}

/**
 * RejectDeposit 驳回充值申请
 * 
 * @param depositID uint - 充值申请ID
 * @param reviewerID uint - 审核人ID
 * @param note string - 驳回原因
 * @return error
 */
func (s *DepositService) RejectDeposit(depositID, reviewerID uint, note string) error {
	deposit, err := s.depositRepo.FindByID(depositID)
	if err != nil {
		return errors.New("充值申请不存在")
	}
	
	if !deposit.IsPending() {
		return fmt.Errorf("充值申请状态不允许审核（当前状态: %s）", deposit.Status)
	}
	
	deposit.Reject(reviewerID, note)
	if err := s.depositRepo.Update(deposit); err != nil {
		return errors.New("更新申请状态失败")
	}
	
	// 发送通知
	notifyMsg := fmt.Sprintf("您的充值申请已被驳回\n驳回原因：%s", note)
	s.notiSvc.SendFundNotification(deposit.UserID, "充值驳回", notifyMsg)
	
	log.Printf("[Deposit] 充值驳回: ID=%d, 原因=%s", depositID, note)
	return nil
}

/**
 * GetUserDeposits 获取用户充值记录
 * 
 * @param userID uint - 用户ID
 * @param limit int - 查询数量限制
 * @param offset int - 偏移量
 * @return ([]*model.DepositRequest, error)
 */
func (s *DepositService) GetUserDeposits(userID uint, limit, offset int) ([]*model.DepositRequest, error) {
	return s.depositRepo.FindByUserID(userID, limit, offset)
}

/**
 * GetPendingDeposits 获取待审核的充值申请
 * 
 * @param limit int - 查询数量限制
 * @return ([]*model.DepositRequest, error)
 */
func (s *DepositService) GetPendingDeposits(limit int) ([]*model.DepositRequest, error) {
	return s.depositRepo.FindPending(limit)
}

/**
 * GetDepositsByStatus 根据状态获取充值申请
 * 
 * @param status string - 状态
 * @param limit int - 查询数量限制
 * @return ([]*model.DepositRequest, error)
 */
func (s *DepositService) GetDepositsByStatus(status string, limit int) ([]*model.DepositRequest, error) {
	return s.depositRepo.FindByStatus(status, limit)
}
