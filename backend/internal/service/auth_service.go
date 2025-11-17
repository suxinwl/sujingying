package service

import (
	"errors"
	"time"

	"suxin/internal/appctx"
	"suxin/internal/model"
	"suxin/internal/pkg/jwtx"
	"suxin/internal/pkg/security"
	"suxin/internal/repository"
)

type AuthService struct {
	ctx      *appctx.AppContext
	repo     *repository.UserRepository
	logRepo  *repository.LoginLogRepository
}

func NewAuthService(ctx *appctx.AppContext) *AuthService {
	return &AuthService{
		ctx:     ctx,
		repo:    repository.NewUserRepository(ctx.DB),
		logRepo: repository.NewLoginLogRepository(ctx.DB),
	}
}

func (s *AuthService) Register(phone, password, role string) (*model.User, error) {
	if phone == "" || password == "" {
		return nil, errors.New("phone and password required")
	}
	if role == "" {
		role = "customer"
	}
	h, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &model.User{Phone: phone, Password: h, Role: role}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *AuthService) Login(phone, password, ip, userAgent string) (string, string, *model.User, error) {
	// 检查失败次数（1小时内）
	since := time.Now().Add(-1 * time.Hour)
	failCount, err := s.logRepo.CountFailedAttempts(phone, since)
	if err == nil && failCount >= 5 {
		s.logFailedLogin(0, phone, ip, userAgent, "账户已被锁定，请1小时后重试")
		return "", "", nil, errors.New("账户已被锁定，请1小时后重试")
	}

	u, err := s.repo.FindByPhone(phone)
	if err != nil {
		s.logFailedLogin(0, phone, ip, userAgent, "用户不存在")
		return "", "", nil, errors.New("用户名或密码错误")
	}

	if !security.CheckPassword(password, u.Password) {
		s.logFailedLogin(u.ID, phone, ip, userAgent, "密码错误")
		return "", "", nil, errors.New("用户名或密码错误")
	}

	acc, err := jwtx.GenerateAccessToken(u.ID, u.Role, s.ctx.Config.Auth.JWTSecret, s.ctx.Config.Auth.AccessMinutes)
	if err != nil {
		return "", "", nil, err
	}
	ref, err := jwtx.GenerateRefreshToken(u.ID, u.Role, s.ctx.Config.Auth.JWTSecret, s.ctx.Config.Auth.RefreshHours)
	if err != nil {
		return "", "", nil, err
	}

	s.logSuccessLogin(u.ID, phone, ip, userAgent)
	return acc, ref, u, nil
}

func (s *AuthService) logSuccessLogin(userID uint, phone, ip, userAgent string) {
	_ = s.logRepo.Create(&model.LoginLog{
		UserID:    userID,
		Phone:     phone,
		IP:        ip,
		UserAgent: userAgent,
		Status:    "success",
	})
}

func (s *AuthService) logFailedLogin(userID uint, phone, ip, userAgent, reason string) {
	_ = s.logRepo.Create(&model.LoginLog{
		UserID:     userID,
		Phone:      phone,
		IP:         ip,
		UserAgent:  userAgent,
		Status:     "failed",
		FailReason: reason,
	})
}

func (s *AuthService) Refresh(refreshToken string) (string, string, error) {
	claims, err := jwtx.Parse(refreshToken, s.ctx.Config.Auth.JWTSecret)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}
	if claims.Typ != "refresh" {
		return "", "", errors.New("invalid token type")
	}
	acc, err := jwtx.GenerateAccessToken(claims.UserID, claims.Role, s.ctx.Config.Auth.JWTSecret, s.ctx.Config.Auth.AccessMinutes)
	if err != nil {
		return "", "", err
	}
	ref, err := jwtx.GenerateRefreshToken(claims.UserID, claims.Role, s.ctx.Config.Auth.JWTSecret, s.ctx.Config.Auth.RefreshHours)
	if err != nil {
		return "", "", err
	}
	return acc, ref, nil
}
