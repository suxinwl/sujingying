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
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"suxin/internal/appctx"
	"suxin/internal/middleware"
	"suxin/internal/model"
	"suxin/internal/repository"
)

type updateUserReq struct {
	Role    string  `json:"role"`
	SalesID *uint   `json:"sales_id"`
}

type resetPasswordReq struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
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
	 * - limit: 每页数量(默认20)
	 * - offset: 偏移量(默认0)
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
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		var users []*model.User
		var err error
		
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
		
		// 分页查询
		err = query.Limit(limit).Offset(offset).Find(&users).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// 脱敏处理(移除密码)
		result := make([]map[string]interface{}, 0)
		for _, user := range users {
			result = append(result, map[string]interface{}{
				"id":                user.ID,
				"phone":             user.Phone,
				"role":              user.Role,
				"sales_id":          user.SalesID,
				"available_deposit": user.AvailableDeposit,
				"used_deposit":      user.UsedDeposit,
				"has_pay_password":  user.HasPayPassword,
				"created_at":        user.CreatedAt,
			})
		}
		
		c.JSON(http.StatusOK, gin.H{
			"users": result,
			"total": len(result),
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
}
