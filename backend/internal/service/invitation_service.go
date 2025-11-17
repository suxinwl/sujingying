/**
 * 邀请码服务层
 * 
 * 用途：
 * - 邀请码生成和管理
 * - 邀请关系建立
 * - 团队层级管理
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"suxin/internal/appctx"
	"suxin/internal/model"
	"suxin/internal/repository"
)

type InvitationService struct {
	ctx            *appctx.AppContext
	invitationRepo *repository.InvitationRepository
	userRepo       *repository.UserRepository
}

func NewInvitationService(ctx *appctx.AppContext) *InvitationService {
	return &InvitationService{
		ctx:            ctx,
		invitationRepo: repository.NewInvitationRepository(ctx.DB),
		userRepo:       repository.NewUserRepository(ctx.DB),
	}
}

/**
 * GenerateInvitationCode 生成邀请码
 * 
 * @param userID uint - 用户ID
 * @return (string, error)
 */
func (s *InvitationService) GenerateInvitationCode(userID uint) (string, error) {
	// 检查是否已有邀请码
	existingCode, err := s.invitationRepo.FindInvitationCodeByUserID(userID)
	if err == nil && existingCode != nil {
		return existingCode.Code, nil
	}

	// 生成唯一邀请码
	code, err := s.generateUniqueCode()
	if err != nil {
		return "", err
	}

	// 创建邀请码记录
	invCode := &model.InvitationCode{
		UserID:   userID,
		Code:     code,
		IsActive: true,
	}

	if err := s.invitationRepo.CreateInvitationCode(invCode); err != nil {
		return "", fmt.Errorf("创建邀请码失败: %v", err)
	}

	return code, nil
}

/**
 * generateUniqueCode 生成唯一的邀请码
 */
func (s *InvitationService) generateUniqueCode() (string, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		// 生成8位随机码
		bytes := make([]byte, 4)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}
		code := strings.ToUpper(hex.EncodeToString(bytes))

		// 检查是否已存在
		_, err := s.invitationRepo.FindInvitationCodeByCode(code)
		if err != nil {
			// 不存在，可以使用
			return code, nil
		}
	}

	return "", errors.New("生成邀请码失败，请重试")
}

/**
 * GetMyInvitationCode 获取我的邀请码
 */
func (s *InvitationService) GetMyInvitationCode(userID uint) (string, error) {
	code, err := s.invitationRepo.FindInvitationCodeByUserID(userID)
	if err != nil {
		// 如果没有邀请码，自动生成一个
		return s.GenerateInvitationCode(userID)
	}
	return code.Code, nil
}

/**
 * ValidateInvitationCode 验证邀请码
 */
func (s *InvitationService) ValidateInvitationCode(code string) (*model.InvitationCode, error) {
	if code == "" {
		return nil, errors.New("邀请码不能为空")
	}

	invCode, err := s.invitationRepo.FindInvitationCodeByCode(code)
	if err != nil {
		return nil, errors.New("邀请码不存在")
	}

	if !invCode.IsActive {
		return nil, errors.New("邀请码已失效")
	}

	return invCode, nil
}

/**
 * ProcessInvitation 处理邀请关系
 * 
 * 在用户注册时调用，建立邀请关系
 * 
 * @param inviteeID uint - 被邀请人ID（新注册用户）
 * @param inviteCode string - 邀请码
 * @return error
 */
func (s *InvitationService) ProcessInvitation(inviteeID uint, inviteCode string) error {
	if inviteCode == "" {
		// 没有邀请码，跳过
		return nil
	}

	// 1. 验证邀请码
	invCode, err := s.ValidateInvitationCode(inviteCode)
	if err != nil {
		return err
	}

	inviterID := invCode.UserID

	// 2. 不能邀请自己
	if inviterID == inviteeID {
		return errors.New("不能使用自己的邀请码")
	}

	// 3. 开启事务
	tx := s.ctx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 4. 创建邀请记录
	record := &model.InvitationRecord{
		InviterID:    inviterID,
		InviteeID:    inviteeID,
		InviteCode:   inviteCode,
		RewardStatus: "pending",
	}
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建邀请记录失败: %v", err)
	}

	// 5. 更新邀请码统计
	if err := tx.Model(&model.InvitationCode{}).
		Where("user_id = ?", inviterID).
		Updates(map[string]interface{}{
			"invite_count":   invCode.InviteCount + 1,
			"register_count": invCode.RegisterCount + 1,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 6. 建立团队关系
	if err := s.establishTeamRelation(tx, inviteeID, inviterID); err != nil {
		tx.Rollback()
		return err
	}

	// 7. 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

/**
 * establishTeamRelation 建立团队层级关系
 */
func (s *InvitationService) establishTeamRelation(tx *gorm.DB, userID, parentID uint) error {
	// 查询上级的团队关系
	var parentRelation model.TeamRelation
	if err := tx.Where("user_id = ?", parentID).First(&parentRelation).Error; err != nil {
		// 上级没有团队关系，创建一个
		parentRelation = model.TeamRelation{
			UserID:  parentID,
			Level:   1,
		}
		if err := tx.Create(&parentRelation).Error; err != nil {
			return err
		}
	}

	// 创建当前用户的团队关系
	userRelation := &model.TeamRelation{
		UserID:   userID,
		ParentID: parentID,
		Level:    parentRelation.Level + 1,
	}

	// 构建祖先链
	if parentRelation.AncestorIDs != "" {
		userRelation.AncestorIDs = parentRelation.AncestorIDs + "," + fmt.Sprint(parentID)
	} else {
		userRelation.AncestorIDs = fmt.Sprint(parentID)
	}

	if err := tx.Create(userRelation).Error; err != nil {
		return err
	}

	// 更新上级的直接下级计数
	if err := tx.Model(&model.TeamRelation{}).
		Where("user_id = ?", parentID).
		Update("direct_count", gorm.Expr("direct_count + ?", 1)).
		Update("team_count", gorm.Expr("team_count + ?", 1)).
		Error; err != nil {
		return err
	}

	// 更新所有祖先的团队计数
	if userRelation.AncestorIDs != "" {
		ancestorIDs := strings.Split(userRelation.AncestorIDs, ",")
		for _, ancestorIDStr := range ancestorIDs {
			if ancestorIDStr == fmt.Sprint(parentID) {
				continue // 已经更新过了
			}
			if err := tx.Model(&model.TeamRelation{}).
				Where("user_id = ?", ancestorIDStr).
				Update("team_count", gorm.Expr("team_count + ?", 1)).
				Error; err != nil {
				return err
			}
		}
	}

	return nil
}

/**
 * GetMyInvitees 获取我邀请的人
 */
func (s *InvitationService) GetMyInvitees(userID uint, limit, offset int) ([]*model.InvitationRecord, error) {
	return s.invitationRepo.FindInvitationRecordsByInviter(userID, limit, offset)
}

/**
 * GetTeamInfo 获取团队信息
 */
func (s *InvitationService) GetTeamInfo(userID uint) (*model.TeamRelation, error) {
	relation, err := s.invitationRepo.FindTeamRelationByUserID(userID)
	if err != nil {
		// 如果没有团队关系，创建一个
		relation = &model.TeamRelation{
			UserID: userID,
			Level:  1,
		}
		if err := s.invitationRepo.CreateTeamRelation(relation); err != nil {
			return nil, err
		}
	}
	return relation, nil
}

/**
 * GetDirectTeamMembers 获取直接下级
 */
func (s *InvitationService) GetDirectTeamMembers(userID uint) ([]*model.TeamRelation, error) {
	return s.invitationRepo.FindTeamRelationByParentID(userID)
}
