/**
 * 用户管理API处理器
 * 
 * 用途：
 * - 管理员管理用户
 * - 用户列表查询
 * - 用户状态管理
 * - 用户信息编辑
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"suxin/internal/appctx"
	"suxin/internal/middleware"
	"suxin/internal/model"
	"suxin/internal/repository"
	"suxin/internal/service"
)

type updateUserReq struct {
	Role    string  `json:"role"`
	SalesID *uint   `json:"sales_id"`
}

type resetPasswordReq struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type toggleAutoSupplementReq struct {
	Enabled bool `json:"enabled"` // 是否启用自动补定金
}

type approveUserReq struct {
	Action string `json:"action" binding:"required,oneof=approve reject"` // approve/reject
	Note   string `json:"note"` // 审核备注
}

type verifyAuditReq struct {
	Remark string `json:"remark"`
}

/**
 * RegisterUserManageRoutes 注册用户管理路由(需管理员权限)
 * 
 * 路由列表：
 * - GET  /users          查询用户列表
 * - GET  /users/:id      查询用户详情
 * - PUT  /users/:id      更新用户信息
 * - POST /users/:id/reset-password  重置密码
 * - POST /users/:id/disable  禁用用户
 * - POST /users/:id/enable   启用用户
 */
func RegisterUserManageRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	userRepo := repository.NewUserRepository(ctx.DB)
	
	// 所有用户管理接口需要管理员权限
	admin := rg.Group("", middleware.RequireAdmin(ctx))
	
	/**
	 * GET /users - 查询用户列表
	 * 
	 * 查询参数：
	 * - role: 角色筛选
	 * - sales_id: 销售ID筛选
	 * - status: 用户状态筛选（pending/active/disabled）
	 * - phone: 手机号模糊查询
	 * - limit, offset: 分页参数（可选）
	 * - page, page_size: 分页参数（可选，优先于 limit/offset）
	 * 
	 * 响应：
	 * {
	 *   "users": [...],
	 *   "total": 100
	 * }
	 */
	admin.GET("/users", func(c *gin.Context) {
		role := c.Query("role")
		salesIDStr := c.Query("sales_id")
		status := c.Query("status")
		phone := c.Query("phone")
		
		// 默认使用 limit/offset
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		// 如果提供了 page/page_size，则优先使用 page 分页
		if pageStr := c.Query("page"); pageStr != "" {
			if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
				if page, err := strconv.Atoi(pageStr); err == nil {
					if pageSize, err2 := strconv.Atoi(pageSizeStr); err2 == nil && page > 0 && pageSize > 0 {
						limit = pageSize
						offset = (page - 1) * pageSize
					}
				}
			}
		}
		
		var users []*model.User
		var total int64
		
		query := ctx.DB.Model(&model.User{})
		
		// 角色筛选
		if role != "" {
			query = query.Where("role = ?", role)
		}
		
		// 销售筛选
		if salesIDStr != "" {
			salesID, _ := strconv.ParseUint(salesIDStr, 10, 32)
			query = query.Where("sales_id = ?", salesID)
		}
		
		// 状态筛选
		if status != "" {
			query = query.Where("status = ?", status)
		}
		
		// 手机号模糊查询
		if phone != "" {
			query = query.Where("phone LIKE ?", "%"+phone+"%")
		}
		
		// 先统计总数
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// 再做分页查询
		if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// 查询实名认证状态映射
		verifyStatusMap := make(map[uint]string)
		if len(users) > 0 {
			var userIDs []uint
			for _, u := range users {
				userIDs = append(userIDs, u.ID)
			}
			var vers []model.UserVerification
			if err := ctx.DB.Where("user_id IN ?", userIDs).Find(&vers).Error; err == nil {
				for _, v := range vers {
					verifyStatusMap[v.UserID] = v.Status
				}
			}
		}
		
		// 查询上级销售员姓名映射
		salesNameMap := make(map[uint]string)
		var salesIDs []uint
		for _, u := range users {
			if u.SalesID != 0 {
				found := false
				for _, id := range salesIDs {
					if id == u.SalesID {
						found = true
						break
					}
				}
				if !found {
					salesIDs = append(salesIDs, u.SalesID)
				}
			}
		}
		if len(salesIDs) > 0 {
			var salesUsers []model.User
			if err := ctx.DB.Where("id IN ?", salesIDs).Find(&salesUsers).Error; err == nil {
				for _, su := range salesUsers {
					name := su.RealName
					if name == "" {
						name = su.Phone
					}
					salesNameMap[su.ID] = name
				}
			}
		}
		
		// 脱敏处理(移除密码)
		result := make([]map[string]interface{}, 0, len(users))
		for _, user := range users {
			result = append(result, map[string]interface{}{
				"id":                      user.ID,
				"phone":                   user.Phone,
				"realname":                user.RealName,
				"role":                    user.Role,
				"status":                  user.Status,
				"verify_status":           verifyStatusMap[user.ID],
				"sales_id":                user.SalesID,
				"sales_name":              salesNameMap[user.SalesID],
				"available_deposit":       user.AvailableDeposit,
				"used_deposit":            user.UsedDeposit,
				"has_pay_password":        user.HasPayPassword,
				"auto_supplement_enabled": user.AutoSupplementEnabled,
				"created_at":              user.CreatedAt,
			})
		}
		
		c.JSON(http.StatusOK, gin.H{
			"users": result,
			"total": total,
		})
	})
	
	/**
	 * GET /users/:id - 查询用户详情
	 */
	admin.GET("/users/:id", func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}
		
		user, err := userRepo.FindByID(uint(userID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"id":                user.ID,
			"phone":             user.Phone,
			"role":              user.Role,
			"sales_id":          user.SalesID,
			"available_deposit": user.AvailableDeposit,
			"used_deposit":      user.UsedDeposit,
			"has_pay_password":  user.HasPayPassword,
			"created_at":        user.CreatedAt,
		})
	})
	
	/**
	 * PUT /users/:id - 更新用户信息
	 * 
	 * 请求body：
	 * {
	 *   "role": "sales",
	 *   "sales_id": 123
	 * }
	 */
	admin.PUT("/users/:id", func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}
		
		var req updateUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		user, err := userRepo.FindByID(uint(userID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		
		// 更新角色
		if req.Role != "" {
			user.Role = req.Role
		}
		
		// 更新销售归属
		if req.SalesID != nil {
			user.SalesID = *req.SalesID
		}
		
		if err := userRepo.Update(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "更新成功",
			"user": map[string]interface{}{
				"id":       user.ID,
				"role":     user.Role,
				"sales_id": user.SalesID,
			},
		})
	})
	
	/**
	 * POST /users/:id/reset-password - 重置用户密码
	 */
	admin.POST("/users/:id/reset-password", func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}
		
		var req resetPasswordReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		// 加密新密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		
		// 更新密码
		if err := ctx.DB.Model(&model.User{}).Where("id = ?", userID).
			Update("password", string(hashedPassword)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码重置失败"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "密码重置成功",
		})
	})
	
	/**
	 * POST /users/:id/toggle-auto-supplement - 启用/禁用自动补定金
	 * 
	 * 权限：只有客服和超级管理员可以操作
	 * 
	 * 功能说明：
	 * - 为用户启用/禁用自动补定金功能
	 * - 启用后，当订单定金率低于50%时，自动从可用定金补充到80%
	 * - 用户自己无法启用该功能，只能由客服操作
	 * 
	 * 请求body：
	 * {
	 *   "enabled": true  // true=启用，false=禁用
	 * }
	 */
	admin.POST("/users/:id/toggle-auto-supplement", func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}
		
		var req toggleAutoSupplementReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		// 查询用户
		user, err := userRepo.FindByID(uint(userID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		
		// 只允许为客户启用
		if user.Role != "customer" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "只能为客户启用自动补定金功能"})
			return
		}
		
		// 更新自动补定金开关
		if err := ctx.DB.Model(&model.User{}).Where("id = ?", userID).
			Update("auto_supplement_enabled", req.Enabled).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
		
		operatorID := c.GetUint("user_id")
		statusText := "禁用"
		if req.Enabled {
			statusText = "启用"
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("已%s用户 %s 的自动补定金功能", statusText, user.Phone),
			"user": map[string]interface{}{
				"id":                      user.ID,
				"phone":                   user.Phone,
				"auto_supplement_enabled": req.Enabled,
			},
			"operator_id": operatorID,
		})
	})
	
	/**
	 * POST /users/:id/approve - 审核用户注册
	 * 
	 * 权限：只有客服和超级管理员可以操作
	 * 
	 * 功能说明：
	 * - 审核新注册用户
	 * - 通过后用户状态变为active，可以正常登录
	 * - 拒绝后用户状态变为disabled，无法登录
	 * 
	 * 请求body：
	 * {
	 *   "action": "approve",  // approve=通过，reject=拒绝
	 *   "note": "审核通过"
	 * }
	 */
	admin.POST("/users/:id/approve", func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}
		
		var req approveUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		// 查询用户
		user, err := userRepo.FindByID(uint(userID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		
		// 检查用户状态
		if user.Status != model.UserStatusPending {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("用户当前状态为%s，无法审核", user.Status),
			})
			return
		}
		
		// 更新用户状态
		var newStatus string
		var message string
		if req.Action == "approve" {
			newStatus = model.UserStatusActive
			message = "审核通过"
		} else {
			newStatus = model.UserStatusDisabled
			message = "审核拒绝"
		}
		
		if err := ctx.DB.Model(&model.User{}).Where("id = ?", userID).
			Update("status", newStatus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
			return
		}
		
		operatorID := c.GetUint("user_id")
		
		// 发送通知
		notiSvc := service.NewNotificationService(ctx)
		if req.Action == "approve" {
			notiSvc.SendNotification(user.ID, "system", "info", "账户审核通过",
				"您的账户已通过审核，现在可以正常登录和交易了！", 0, "")
		} else {
			reason := "审核未通过"
			if req.Note != "" {
				reason = req.Note
			}
			notiSvc.SendNotification(user.ID, "system", "warning", "账户审核未通过",
				fmt.Sprintf("很抱歉，您的账户审核未通过。原因：%s", reason), 0, "")
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("用户 %s %s", user.Phone, message),
			"user": map[string]interface{}{
				"id":     user.ID,
				"phone":  user.Phone,
				"status": newStatus,
			},
			"operator_id": operatorID,
			"note": req.Note,
		})
	})
	
	/**
	 * GET /users/pending - 获取待审核用户列表
	 * 
	 * 权限：客服和超级管理员
	 */
	admin.GET("/users/pending", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		var users []model.User
		var total int64
		
		// 查询待审核用户
		if err := ctx.DB.Model(&model.User{}).
			Where("status = ?", model.UserStatusPending).
			Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
		
		if err := ctx.DB.Where("status = ?", model.UserStatusPending).
			Order("created_at DESC").
			Limit(limit).Offset(offset).
			Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
		
		// 构建响应
		userList := make([]map[string]interface{}, 0)
		for _, u := range users {
			userList = append(userList, map[string]interface{}{
				"id":         u.ID,
				"phone":      u.Phone,
				"role":       u.Role,
				"status":     u.Status,
				"sales_id":   u.SalesID,
				"created_at": u.CreatedAt,
			})
		}
		
		c.JSON(http.StatusOK, gin.H{
			"users": userList,
			"total": total,
		})
	})

	// 获取指定用户的实名认证信息
	admin.GET("/users/:id/verification", func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil || userID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}
		verificationRepo := repository.NewUserVerificationRepository(ctx.DB)
		v, err := verificationRepo.FindByUserID(uint(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询实名认证信息失败"})
			return
		}
		if v == nil {
			c.JSON(http.StatusOK, gin.H{"verification": nil})
			return
		}
		c.JSON(http.StatusOK, gin.H{
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
				"auditor_id":     v.AuditorID,
			},
		})
	})

	// 审核通过实名认证
	admin.POST("/users/:id/verification/approve", func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil || userID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}
		var req verifyAuditReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		verificationRepo := repository.NewUserVerificationRepository(ctx.DB)
		v, err := verificationRepo.FindByUserID(uint(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询实名认证信息失败"})
			return
		}
		if v == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户未提交实名认证信息"})
			return
		}
		if v.Status != model.VerificationStatusPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不可审核"})
			return
		}
		v.Status = model.VerificationStatusApproved
		v.Remark = req.Remark
		v.AuditorID = c.GetUint("user_id")
		if err := verificationRepo.Update(v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新实名认证状态失败"})
			return
		}
		// 同步更新用户表中的真实姓名
		if err := ctx.DB.Model(&model.User{}).Where("id = ?", userID).
			Update("real_name", v.RealName).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "实名认证审核通过"})
	})

	// 驳回实名认证
	admin.POST("/users/:id/verification/reject", func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil || userID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}
		var req verifyAuditReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		verificationRepo := repository.NewUserVerificationRepository(ctx.DB)
		v, err := verificationRepo.FindByUserID(uint(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询实名认证信息失败"})
			return
		}
		if v == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户未提交实名认证信息"})
			return
		}
		if v.Status != model.VerificationStatusPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不可审核"})
			return
		}
		v.Status = model.VerificationStatusRejected
		v.Remark = req.Remark
		v.AuditorID = c.GetUint("user_id")
		if err := verificationRepo.Update(v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新实名认证状态失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "实名认证已驳回"})
	})
}
