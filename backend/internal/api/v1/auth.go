package v1

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"suxin/internal/appctx"
	"suxin/internal/middleware"
	"suxin/internal/model"
	"suxin/internal/repository"
	"suxin/internal/service"
	"suxin/internal/pkg/security"
	"suxin/internal/pkg/response"
)

type registerReq struct {
	Phone        string                     `json:"phone" binding:"required"`
	Password     string                     `json:"password" binding:"required,min=6"`
	InviteCode   string                     `json:"invite_code" binding:"required"` // 邀请码（必填）
	Verification *userVerificationSubmitReq `json:"verification" binding:"required"`
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

type userVerificationSubmitReq struct {
	RealName       string `json:"real_name" binding:"required"`
	IDNumber       string `json:"id_number" binding:"required"`
	IDFrontURL     string `json:"id_front_url"`
	IDBackURL      string `json:"id_back_url"`
	BankCardID     uint   `json:"bank_card_id"`
	ReceiverName   string `json:"receiver_name"`
	ReceiverPhone  string `json:"receiver_phone"`
	Province       string `json:"province"`
	City           string `json:"city"`
	District       string `json:"district"`
	AddressDetail  string `json:"address_detail"`
}

func RegisterAuthRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	authSvc := service.NewAuthService(ctx)
	userRepo := repository.NewUserRepository(ctx.DB)
	notiSvc := service.NewNotificationService(ctx)
	verificationRepo := repository.NewUserVerificationRepository(ctx.DB)

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
		// 2.1 保存实名认证信息（注册时必须提交）
		if req.Verification == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先完善实名认证信息"})
			return
		}
		v := &model.UserVerification{
			UserID:        u.ID,
			RealName:      req.Verification.RealName,
			IDNumber:      req.Verification.IDNumber,
			IDFrontURL:    req.Verification.IDFrontURL,
			IDBackURL:     req.Verification.IDBackURL,
			BankCardID:    req.Verification.BankCardID,
			ReceiverName:  req.Verification.ReceiverName,
			ReceiverPhone: req.Verification.ReceiverPhone,
			Province:      req.Verification.Province,
			City:          req.Verification.City,
			District:      req.Verification.District,
			AddressDetail: req.Verification.AddressDetail,
			Status:        model.VerificationStatusPending,
		}
		if err := verificationRepo.Create(v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存实名认证信息失败"})
			return
		}
		// 同步用户表中的真实姓名，便于列表展示
		if v.RealName != "" {
			if err := ctx.DB.Model(u).Update("real_name", v.RealName).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
				return
			}
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
		
		// 5. 发送通知给客服/管理员：有新用户待审核
		go notiSvc.SendSystemNotificationToAdmins(
			"新用户注册待审核",
			"有新用户注册并等待审核，手机号："+u.Phone,
			"",
		)

		// 6. 如果有上级销售员，一并提醒该销售员有新客户注册
		if inviteCode.UserID != 0 {
			go notiSvc.SendSystemNotificationToUser(
				inviteCode.UserID,
				"新客户注册",
				"您的客户 " + u.Phone + " 已注册并等待审核",
				"",
			)
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
			response.BadRequest(c, "invalid request")
			return
		}
		
		phone := req.GetPhone()
		if phone == "" || req.Password == "" {
			response.BadRequest(c, "手机号和密码不能为空")
			return
		}
		
		ip := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		acc, ref, u, err := authSvc.Login(phone, req.Password, ip, userAgent)
		if err != nil {
			response.Unauthorized(c, err.Error())
			return
		}
		// 客户角色增加实名认证状态校验：需实名认证通过
		if u.Role == "customer" {
			v, err := verificationRepo.FindByUserID(u.ID)
			if err != nil {
				response.InternalServerError(c, "查询实名认证信息失败")
				return
			}
			if v == nil || v.Status != model.VerificationStatusApproved {
				response.Unauthorized(c, "您的实名认证尚未通过，请完成实名认证并等待审核")
				return
			}
		}
		
		// 使用统一的data格式
		response.Success(c, gin.H{
			"access_token":  acc,
			"refresh_token": ref,
			"user": gin.H{
				"id":    u.ID,
				"phone": u.Phone,
				"role":  u.Role,
			},
		})
	})

	rg.POST("/auth/refresh", func(c *gin.Context) {
		var req refreshReq
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			response.BadRequest(c, "invalid request")
			return
		}
		acc, ref, err := authSvc.Refresh(req.RefreshToken)
		if err != nil {
			response.Unauthorized(c, err.Error())
			return
		}
		response.Success(c, gin.H{
			"access_token":  acc,
			"refresh_token": ref,
		})
	})

	// 登出：无状态，直接200
	rg.POST("/auth/logout", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 受保护路由：用户信息和支付密码
	pg := rg.Group("/user", middleware.AuthRequired(ctx))
	
	// 获取当前用户信息
	pg.GET("/profile", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		user, err := userRepo.FindByID(userID)
		if err != nil {
			response.NotFound(c, "用户不存在")
			return
		}
		
		response.Success(c, gin.H{
			"id":                      user.ID,
			"phone":                   user.Phone,
			"role":                    user.Role,
			"status":                  user.Status,
			"sales_id":                user.SalesID,
			"available_deposit":       user.AvailableDeposit,
			"used_deposit":            user.UsedDeposit,
			"has_pay_password":        user.HasPayPassword,
			"auto_supplement_enabled": user.AutoSupplementEnabled,
			"created_at":              user.CreatedAt,
		})
	})

	// 获取当前用户的实名认证信息
	pg.GET("/verification", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		v, err := verificationRepo.FindByUserID(userID)
		if err != nil {
			response.InternalServerError(c, "查询实名认证信息失败")
			return
		}
		if v == nil {
			response.Success(c, gin.H{"verification": nil})
			return
		}
		response.Success(c, gin.H{
			"verification": gin.H{
				"id":             v.ID,
				"user_id":        v.UserID,
				"real_name":      v.RealName,
				"id_number":      v.IDNumber,
				"id_front_url":   v.IDFrontURL,
				"id_back_url":    v.IDBackURL,
				"bank_card_id":   v.BankCardID,
				"receiver_name":  v.ReceiverName,
				"receiver_phone": v.ReceiverPhone,
				"province":       v.Province,
				"city":           v.City,
				"district":       v.District,
				"address_detail": v.AddressDetail,
				"status":         v.Status,
				"remark":         v.Remark,
			},
		})
	})

	// 提交 / 更新实名认证信息
	pg.POST("/verification", func(c *gin.Context) {
		var req userVerificationSubmitReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "请求参数错误")
			return
		}
		userID := c.GetUint("user_id")
		v, err := verificationRepo.FindByUserID(userID)
		if err != nil {
			response.InternalServerError(c, "查询实名认证信息失败")
			return
		}
		if v == nil {
			v = &model.UserVerification{UserID: userID}
		}
		v.RealName = req.RealName
		v.IDNumber = req.IDNumber
		v.IDFrontURL = req.IDFrontURL
		v.IDBackURL = req.IDBackURL
		v.BankCardID = req.BankCardID
		v.ReceiverName = req.ReceiverName
		v.ReceiverPhone = req.ReceiverPhone
		v.Province = req.Province
		v.City = req.City
		v.District = req.District
		v.AddressDetail = req.AddressDetail
		v.Status = model.VerificationStatusPending
		v.Remark = ""
		v.AuditorID = 0
		if v.ID == 0 {
			if err := verificationRepo.Create(v); err != nil {
				response.InternalServerError(c, "保存实名认证信息失败")
				return
			}
		} else {
			if err := verificationRepo.Update(v); err != nil {
				response.InternalServerError(c, "保存实名认证信息失败")
				return
			}
		}
		response.SuccessWithMessage(c, gin.H{"id": v.ID}, "实名认证信息已提交，等待审核")
	})
	
	// 设置/修改支付密码（统一接口）
	pg.POST("/paypass", func(c *gin.Context) {
		var req struct {
			OldPayPassword string `json:"old_pay_password"` // 修改时需要
			NewPayPassword string `json:"new_pay_password"` // 新密码
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		// 验证新密码格式
		if err := validatePayPass(req.NewPayPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		uid := c.GetUint("user_id")
		user, err := userRepo.FindByID(uid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户不存在"})
			return
		}
		
		// 如果已设置支付密码，需要验证旧密码
		if user.HasPayPassword {
			if req.OldPayPassword == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "请输入旧支付密码"})
				return
			}
			if !security.CheckPassword(req.OldPayPassword, user.PayPassword) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "旧支付密码错误"})
				return
			}
		}
		
		// 加密新密码
		hashed, err := security.HashPassword(req.NewPayPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		
		// 更新支付密码
		if err := userRepo.UpdatePayPassword(uid, hashed, true); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "支付密码设置成功",
		})
	})
	
	// 保留旧接口用于兼容
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
