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
	Phone      string `json:"phone" binding:"required"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code" binding:"required"` // 邀请码（必填）
}

type loginReq struct {
	Username string `json:"username"` // 兼容前端字段名（实际为手机号）
	Phone    string `json:"phone"`    // 直接使用phone字段
	Password string `json:"password"`
}

// GetPhone 获取手机号（优先使用phone，兼容username）
func (r *loginReq) GetPhone() string {
	if r.Phone != "" {
		return r.Phone
	}
	return r.Username
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		// 1. 验证邀请码
		invitationSvc := service.NewInvitationService(ctx)
		inviteCode, err := invitationSvc.ValidateInvitationCode(req.InviteCode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "邀请码无效或已失效"})
			return
		}
		
		// 2. 注册用户（role固定为customer，status为pending）
		u, err := authSvc.Register(req.Phone, req.Password, "customer")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// 3. 处理邀请码，建立销售关系
		if err := invitationSvc.ProcessInvitation(u.ID, req.InviteCode); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "邀请关系建立失败"})
			return
		}
		
		// 4. 绑定销售关系（SalesID = 邀请人ID）
		if err := ctx.DB.Model(&u).Update("sales_id", inviteCode.UserID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "销售关系绑定失败"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "注册成功，请等待客服审核后才可登录",
			"user": gin.H{
				"id": u.ID, 
				"phone": u.Phone, 
				"role": u.Role,
				"status": u.Status,
			},
		})
	})

	rg.POST("/auth/login", func(c *gin.Context) {
		var req loginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		
		phone := req.GetPhone()
		if phone == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "手机号和密码不能为空"})
			return
		}
		
		ip := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		acc, ref, u, err := authSvc.Login(phone, req.Password, ip, userAgent)
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
