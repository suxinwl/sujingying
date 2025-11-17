/**
 * 邀请码仓储层
 * 
 * 用途：
 * - 邀请码数据访问
 * - 邀请关系查询
 * - 团队数据统计
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package repository

import (
	"gorm.io/gorm"

	"suxin/internal/model"
)

type InvitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

// ===== 邀请码 =====

func (r *InvitationRepository) CreateInvitationCode(code *model.InvitationCode) error {
	return r.db.Create(code).Error
}

func (r *InvitationRepository) FindInvitationCodeByUserID(userID uint) (*model.InvitationCode, error) {
	var code model.InvitationCode
	if err := r.db.Where("user_id = ?", userID).First(&code).Error; err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *InvitationRepository) FindInvitationCodeByCode(code string) (*model.InvitationCode, error) {
	var invCode model.InvitationCode
	if err := r.db.Where("code = ?", code).First(&invCode).Error; err != nil {
		return nil, err
	}
	return &invCode, nil
}

func (r *InvitationRepository) UpdateInvitationCode(code *model.InvitationCode) error {
	return r.db.Save(code).Error
}

func (r *InvitationRepository) IncrementInviteCount(userID uint) error {
	return r.db.Model(&model.InvitationCode{}).
		Where("user_id = ?", userID).
		Update("invite_count", gorm.Expr("invite_count + ?", 1)).
		Update("register_count", gorm.Expr("register_count + ?", 1)).
		Error
}

// ===== 邀请记录 =====

func (r *InvitationRepository) CreateInvitationRecord(record *model.InvitationRecord) error {
	return r.db.Create(record).Error
}

func (r *InvitationRepository) FindInvitationRecordsByInviter(inviterID uint, limit, offset int) ([]*model.InvitationRecord, error) {
	var records []*model.InvitationRecord
	err := r.db.Where("inviter_id = ?", inviterID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error
	return records, err
}

func (r *InvitationRepository) FindInvitationRecordByInvitee(inviteeID uint) (*model.InvitationRecord, error) {
	var record model.InvitationRecord
	if err := r.db.Where("invitee_id = ?", inviteeID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// ===== 团队关系 =====

func (r *InvitationRepository) CreateTeamRelation(relation *model.TeamRelation) error {
	return r.db.Create(relation).Error
}

func (r *InvitationRepository) FindTeamRelationByUserID(userID uint) (*model.TeamRelation, error) {
	var relation model.TeamRelation
	if err := r.db.Where("user_id = ?", userID).First(&relation).Error; err != nil {
		return nil, err
	}
	return &relation, nil
}

func (r *InvitationRepository) FindTeamRelationByParentID(parentID uint) ([]*model.TeamRelation, error) {
	var relations []*model.TeamRelation
	err := r.db.Where("parent_id = ?", parentID).Find(&relations).Error
	return relations, err
}

func (r *InvitationRepository) UpdateTeamRelation(relation *model.TeamRelation) error {
	return r.db.Save(relation).Error
}

func (r *InvitationRepository) IncrementDirectCount(userID uint) error {
	return r.db.Model(&model.TeamRelation{}).
		Where("user_id = ?", userID).
		Update("direct_count", gorm.Expr("direct_count + ?", 1)).
		Error
}

func (r *InvitationRepository) IncrementTeamCount(userID uint) error {
	return r.db.Model(&model.TeamRelation{}).
		Where("user_id = ?", userID).
		Update("team_count", gorm.Expr("team_count + ?", 1)).
		Error
}
