/**
 * 权限管理中间件
 * 
 * 用途：
 * - 实现基于角色的访问控制(RBAC)
 * - 验证用户角色和数据权限
 * - 支持多角色权限检查
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"suxin/internal/appctx"
	"suxin/internal/repository"
)

/**
 * 角色常量定义
 */
const (
	RoleCustomer    = "customer"     // 客户
	RoleSales       = "sales"        // 销售
	RoleSupport     = "support"      // 客服
	RoleSuperAdmin  = "super_admin"  // 超级管理员
)

/**
 * RequireRole 要求特定角色
 * 
 * @param roles ...string - 允许的角色列表
 * @return gin.HandlerFunc
 */
func RequireRole(ctx *appctx.AppContext, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		// 查询用户角色
		userRepo := repository.NewUserRepository(ctx.DB)
		user, err := userRepo.FindByID(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}

		// 检查角色是否匹配
		roleMatched := false
		for _, role := range roles {
			if user.Role == role {
				roleMatched = true
				break
			}
		}

		if !roleMatched {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		// 将用户角色存入上下文
		c.Set("user_role", user.Role)
		c.Next()
	}
}

/**
 * RequireAdmin 要求管理员角色(客服或超级管理员)
 */
func RequireAdmin(ctx *appctx.AppContext) gin.HandlerFunc {
	return RequireRole(ctx, RoleSupport, RoleSuperAdmin)
}

/**
 * RequireSuperAdmin 要求超级管理员角色
 */
func RequireSuperAdmin(ctx *appctx.AppContext) gin.HandlerFunc {
	return RequireRole(ctx, RoleSuperAdmin)
}

/**
 * RequireSales 要求销售角色
 */
func RequireSales(ctx *appctx.AppContext) gin.HandlerFunc {
	return RequireRole(ctx, RoleSales)
}

/**
 * CanViewCustomer 检查是否可以查看指定客户信息
 * 
 * 规则：
 * - 客户本人：可以查看自己
 * - 销售：只能查看自己的客户
 * - 客服/超级管理员：可以查看所有客户
 * 
 * @param ctx *appctx.AppContext
 * @return gin.HandlerFunc
 */
func CanViewCustomer(ctx *appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		userRole := c.GetString("user_role")
		
		// 如果角色未设置，先获取
		if userRole == "" {
			userRepo := repository.NewUserRepository(ctx.DB)
			user, err := userRepo.FindByID(userID)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
				c.Abort()
				return
			}
			userRole = user.Role
			c.Set("user_role", userRole)
		}

		// 客服和超级管理员可以查看所有
		if userRole == RoleSupport || userRole == RoleSuperAdmin {
			c.Set("can_view_all_customers", true)
			c.Next()
			return
		}

		// 销售只能查看自己的客户
		if userRole == RoleSales {
			c.Set("can_view_all_customers", false)
			c.Set("sales_id", userID) // 限制只能查看自己的客户
			c.Next()
			return
		}

		// 客户只能查看自己
		if userRole == RoleCustomer {
			c.Set("can_view_all_customers", false)
			c.Set("customer_id", userID) // 限制只能查看自己
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
		c.Abort()
	}
}

/**
 * CheckCustomerOwnership 检查客户归属权限
 * 
 * 从URL参数或查询参数中获取customer_id，验证权限
 */
func CheckCustomerOwnership(ctx *appctx.AppContext, paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		
		// 获取目标客户ID
		targetCustomerIDStr := c.Param(paramName)
		if targetCustomerIDStr == "" {
			targetCustomerIDStr = c.Query(paramName)
		}
		
		// 如果没有指定客户ID，使用当前用户ID
		if targetCustomerIDStr == "" {
			c.Next()
			return
		}

		// 超级管理员和客服可以访问任何客户
		if userRole == RoleSuperAdmin || userRole == RoleSupport {
			c.Next()
			return
		}

		// 销售需要验证客户归属
		if userRole == RoleSales {
			// 获取目标客户ID
			targetCustomerID, err := strconv.ParseUint(targetCustomerIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
				c.Abort()
				return
			}
			
			// 验证客户是否属于当前销售
			userRepo := repository.NewUserRepository(ctx.DB)
			customer, err := userRepo.FindByID(uint(targetCustomerID))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "客户不存在"})
				c.Abort()
				return
			}
			
			// 检查客户的销售ID是否匹配
			currentUserID := c.GetUint("user_id")
			if customer.SalesID != currentUserID {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该客户信息，该客户不属于您"})
				c.Abort()
				return
			}
			
			c.Next()
			return
		}

		// 客户只能访问自己
		if userRole == RoleCustomer {
			// 获取目标客户ID
			targetCustomerID, err := strconv.ParseUint(targetCustomerIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
				c.Abort()
				return
			}
			
			// 验证是否是本人
			currentUserID := c.GetUint("user_id")
			if uint(targetCustomerID) != currentUserID {
				c.JSON(http.StatusForbidden, gin.H{"error": "您只能访问自己的信息"})
				c.Abort()
				return
			}
			
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "无权限访问该客户信息"})
		c.Abort()
	}
}
