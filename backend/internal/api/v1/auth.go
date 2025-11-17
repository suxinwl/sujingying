package v1

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"suxin/internal/appctx"
	"suxin/internal/middleware"
	"suxin/internal/repository"
	"suxin/internal/service"
	"suxin/internal/pkg/security"
)

type registerReq struct {
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	InviteCode string `json:"invite_code"` // 邀请码（可选）
}

type loginReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type paypassReq struct {
	PayPassword string `json:"pay_password"`
}

func RegisterAuthRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	authSvc := service.NewAuthService(ctx)
	userRepo := repository.NewUserRepository(ctx.DB)

	rg.POST("/auth/register", func(c *gin.Context) {
		var req registerReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		u, err := authSvc.Register(req.Phone, req.Password, req.Role)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// 处理邀请码（如果提供）
		if req.InviteCode != "" {
			invitationSvc := service.NewInvitationService(ctx)
			if err := invitationSvc.ProcessInvitation(u.ID, req.InviteCode); err != nil {
				// 邀请码处理失败不影响注册，只记录日志
				// TODO: 添加日志记录
			}
		}
		
		c.JSON(http.StatusOK, gin.H{"id": u.ID, "phone": u.Phone, "role": u.Role})
	})

	rg.POST("/auth/login", func(c *gin.Context) {
		var req loginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		ip := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		acc, ref, u, err := authSvc.Login(req.Phone, req.Password, ip, userAgent)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"access_token":  acc,
			"refresh_token": ref,
			"user": gin.H{"id": u.ID, "phone": u.Phone, "role": u.Role},
		})
	})

	rg.POST("/auth/refresh", func(c *gin.Context) {
		var req refreshReq
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		acc, ref, err := authSvc.Refresh(req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"access_token": acc, "refresh_token": ref})
	})

	// 登出：无状态，直接200
	rg.POST("/auth/logout", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 受保护路由：支付密码设置/校验
	pg := rg.Group("/user", middleware.AuthRequired(ctx))
	pg.POST("/paypass/set", func(c *gin.Context) {
		var req paypassReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if err := validatePayPass(req.PayPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h, err := security.HashPassword(req.PayPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
			return
		}
		uid := c.GetUint("user_id")
		if err := userRepo.UpdatePayPassword(uid, h, true); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})

	pg.POST("/paypass/verify", func(c *gin.Context) {
		var req paypassReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		uid := c.GetUint("user_id")
		u, err := userRepo.FindByID(uid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		}
		if !u.HasPayPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pay password not set"})
			return
		}
		if !security.CheckPassword(req.PayPassword, u.PayPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid pay password"})
			return
		}
		c.Status(http.StatusOK)
	})
}

var payPassRe = regexp.MustCompile(`^\d{6}$`)

func validatePayPass(p string) error {
	if !payPassRe.MatchString(p) {
		return errors.New("pay password must be 6 digits")
	}
	return nil
}
