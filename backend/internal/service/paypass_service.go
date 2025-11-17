package service

import (
	"errors"

	"suxin/internal/appctx"
	"suxin/internal/pkg/security"
	"suxin/internal/repository"
)

type PayPassService struct {
	ctx      *appctx.AppContext
	userRepo *repository.UserRepository
}

func NewPayPassService(ctx *appctx.AppContext) *PayPassService {
	return &PayPassService{
		ctx:      ctx,
		userRepo: repository.NewUserRepository(ctx.DB),
	}
}

// VerifyPayPassword 验证用户的支付密码
// 返回 error 表示验证失败，nil 表示验证通过
func (s *PayPassService) VerifyPayPassword(userID uint, payPassword string) error {
	if payPassword == "" {
		return errors.New("支付密码不能为空")
	}

	u, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if !u.HasPayPassword {
		return errors.New("请先设置支付密码")
	}

	if !security.CheckPassword(payPassword, u.PayPassword) {
		return errors.New("支付密码错误")
	}

	return nil
}

// RequirePayPassword 通用支付密码校验函数，用于关键操作前的验证
// 使用示例：
//   if err := paypassSvc.RequirePayPassword(userID, req.PayPassword); err != nil {
//       return err
//   }
func (s *PayPassService) RequirePayPassword(userID uint, payPassword string) error {
	return s.VerifyPayPassword(userID, payPassword)
}
