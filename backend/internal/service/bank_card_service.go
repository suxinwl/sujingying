/**
 * 银行卡服务层
 * 
 * 用途：
 * - 实现银行卡业务逻辑
 * - 集成支付密码验证
 * - 管理默认卡设置
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

/**
 * BankCardService 银行卡服务
 */
type BankCardService struct {
	ctx      *appctx.AppContext
	cardRepo *repository.BankCardRepository
}

/**
 * NewBankCardService 创建银行卡服务实例
 * 
 * @param ctx *appctx.AppContext - 应用上下文
 * @return *BankCardService
 */
func NewBankCardService(ctx *appctx.AppContext) *BankCardService {
	return &BankCardService{
		ctx:      ctx,
		cardRepo: repository.NewBankCardRepository(ctx.DB),
	}
}

/**
 * AddCard 添加银行卡（需支付密码验证）
 * 
 * 业务流程：
 * 1. 验证支付密码
 * 2. 检查银行卡数量限制（最多5张）
 * 3. 创建银行卡
 * 4. 如果是第一张卡，自动设为默认
 * 
 * @param userID uint - 用户ID
 * @param bankName string - 银行名称
 * @param cardNumber string - 银行卡号
 * @param cardHolder string - 持卡人姓名
 * @return (*model.BankCard, error)
 */
func (s *BankCardService) AddCard(
	userID uint,
	bankName, cardNumber, cardHolder string,
) (*model.BankCard, error) {
	
	// 1. 验证参数
	if bankName == "" || cardNumber == "" || cardHolder == "" {
		return nil, errors.New("银行名称、卡号和持卡人不能为空")
	}
	
	// 2. 检查银行卡数量限制
	count, err := s.cardRepo.CountByUserID(userID)
	if err != nil {
		return nil, err
	}
	if count >= 5 {
		return nil, errors.New("最多只能添加5张银行卡")
	}
	
	// 3. 创建银行卡
	card := &model.BankCard{
		UserID:     userID,
		BankName:   bankName,
		CardNumber: cardNumber,
		CardHolder: cardHolder,
		IsDefault:  count == 0, // 第一张卡自动设为默认
	}
	
	if err := s.cardRepo.Create(card); err != nil {
		return nil, fmt.Errorf("添加银行卡失败: %v", err)
	}
	
	return card, nil
}

/**
 * GetUserCards 获取用户的所有银行卡
 * 
 * @param userID uint - 用户ID
 * @return ([]*model.BankCard, error)
 */
func (s *BankCardService) GetUserCards(userID uint) ([]*model.BankCard, error) {
	return s.cardRepo.FindByUserID(userID)
}

/**
 * SetDefaultCard 设置默认银行卡（需支付密码验证）
 * 
 * 业务流程：
 * 1. 查找银行卡
 * 2. 验证归属
 * 3. 取消其他默认卡
 * 4. 设置新默认卡
 * 
 * @param userID uint - 用户ID
 * @param cardID uint - 银行卡ID
 * @return error
 */
func (s *BankCardService) SetDefaultCard(userID uint, cardID uint) error {
	// 1. 查找银行卡
	card, err := s.cardRepo.FindByID(cardID)
	if err != nil {
		return errors.New("银行卡不存在")
	}
	
	// 2. 验证归属
	if card.UserID != userID {
		return errors.New("无权操作此银行卡")
	}
	
	// 3. 开启事务
	tx := s.ctx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 4. 取消所有默认卡
	if err := s.cardRepo.UnsetAllDefault(userID); err != nil {
		tx.Rollback()
		return err
	}
	
	// 5. 设置新默认卡
	card.SetAsDefault()
	if err := tx.Save(card).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	tx.Commit()
	return nil
}

/**
 * DeleteCard 删除银行卡（需支付密码验证）
 * 
 * 业务流程：
 * 1. 查找银行卡
 * 2. 验证归属
 * 3. 如果是默认卡，提示先设置其他默认卡
 * 4. 删除银行卡
 * 
 * @param userID uint - 用户ID
 * @param cardID uint - 银行卡ID
 * @return error
 */
func (s *BankCardService) DeleteCard(userID uint, cardID uint) error {
	// 1. 查找银行卡
	card, err := s.cardRepo.FindByID(cardID)
	if err != nil {
		return errors.New("银行卡不存在")
	}
	
	// 2. 验证归属
	if card.UserID != userID {
		return errors.New("无权操作此银行卡")
	}
	
	// 3. 如果是默认卡，提示错误
	if card.IsDefault {
		return errors.New("请先将其他银行卡设为默认卡")
	}
	
	// 4. 删除
	return s.cardRepo.Delete(cardID, userID)
}
