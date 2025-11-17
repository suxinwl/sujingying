/**
 * 销售管理API处理器
 * 
 * 用途：
 * - 处理销售管理相关HTTP请求
 * - 提供销售看板和排行榜
 * - 客户归属管理
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"suxin/internal/appctx"
	"suxin/internal/middleware"
	"suxin/internal/model"
	"suxin/internal/repository"
	"suxin/internal/service"
)

/**
 * RegisterSalesRoutes 注册销售管理路由
 * 
 * 路由列表：
 * - GET /sales/dashboard      获取销售看板（需JWT）
 * - GET /sales/ranking        获取销售排行榜（需JWT）
 * 
 * @param rg *gin.RouterGroup - 路由组
 * @param ctx *appctx.AppContext - 应用上下文
 * @return void
 */
func RegisterSalesRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	salesSvc := service.NewSalesService(ctx)
	
	/**
	 * GET /sales/dashboard - 获取销售看板数据
	 * 
	 * 响应：
	 * {
	 *   "salesperson_name": "张三",
	 *   "sales_code": "S001",
	 *   "total_points": 1000.50,
	 *   "month_points": 150.25,
	 *   "commission_rate": 0.0001,
	 *   "total_customers": 50,
	 *   "active_customers": 30,
	 *   "holding_orders_count": 15,
	 *   "holding_total_weight": 500.00,
	 *   "holding_total_deposit": 250000.00
	 * }
	 */
	rg.GET("/sales/dashboard", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		
		// TODO: 查询销售人员ID（从user_id）
		// 简化处理：假设salesperson_id = user_id
		dashboard, err := salesSvc.GetSalesDashboard(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, dashboard)
	})
	
	/**
	 * GET /sales/ranking - 获取销售排行榜
	 * 
	 * 查询参数：
	 * - limit: 查询数量（默认10）
	 * - by_month: 是否按本月排名（true/false，默认false）
	 * 
	 * 响应：
	 * {
	 *   "rankings": [
	 *     {
	 *       "id": 1,
	 *       "name": "张三",
	 *       "sales_code": "S001",
	 *       "total_points": 1000.50,
	 *       "month_points": 150.25,
	 *       "total_customers": 50,
	 *       "active_customers": 30
	 *     }
	 *   ]
	 * }
	 */
	rg.GET("/sales/ranking", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		byMonth := c.DefaultQuery("by_month", "false") == "true"
		
		rankings, err := salesSvc.GetSalesRanking(limit, byMonth)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"rankings": rankings,
		})
	})
	
	/**
	 * GET /sales/customers - 获取销售的客户列表
	 * 
	 * 权限说明：
	 * - 销售：只能查看自己的客户
	 * - 客服/超级管理员：可以查看所有销售的客户
	 * 
	 * 查询参数：
	 * - salesperson_id: 销售人员ID（可选，销售默认当前用户）
	 * - limit: 每页数量（默认20）
	 * - offset: 偏移量（默认0）
	 */
	rg.GET("/sales/customers", middleware.CanViewCustomer(ctx), func(c *gin.Context) {
		userID := c.GetUint("user_id")
		userRole := c.GetString("user_role")
		canViewAll := c.GetBool("can_view_all_customers")
		
		// 获取销售人员ID
		salespersonIDStr := c.Query("salesperson_id")
		var salespersonID uint
		
		if canViewAll {
			// 客服/超级管理员可以查看指定销售的客户
			if salespersonIDStr != "" {
				id, _ := strconv.ParseUint(salespersonIDStr, 10, 32)
				salespersonID = uint(id)
			}
		} else if userRole == middleware.RoleSales {
			// 销售只能查看自己的客户
			salespersonID = userID
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			return
		}
		
		// 查询客户列表
		userRepo := repository.NewUserRepository(ctx.DB)
		var customers []*model.User
		var err error
		
		if salespersonID != 0 {
			customers, err = userRepo.FindBySalesID(salespersonID)
		} else {
			// 查询所有客户（仅管理员）
			err = ctx.DB.Where("role = ?", "customer").Find(&customers).Error
		}
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// 脱敏处理（隐藏密码）
		result := make([]map[string]interface{}, 0)
		for _, customer := range customers {
			result = append(result, map[string]interface{}{
				"id":                customer.ID,
				"phone":             customer.Phone,
				"sales_id":          customer.SalesID,
				"created_at":        customer.CreatedAt,
				"available_deposit": customer.AvailableDeposit,
				"used_deposit":      customer.UsedDeposit,
			})
		}
		
		c.JSON(http.StatusOK, gin.H{
			"customers": result,
			"total":     len(result),
		})
	})
	
	/**
	 * GET /sales/commissions - 获取提成明细
	 * 
	 * 权限说明：
	 * - 销售：只能查看自己的提成
	 * - 客服/超级管理员：可以查看所有销售的提成
	 * 
	 * 查询参数：
	 * - salesperson_id: 销售人员ID（可选，销售默认当前用户）
	 * - limit: 每页数量（默认20）
	 * - offset: 偏移量（默认0）
	 */
	rg.GET("/sales/commissions", middleware.CanViewCustomer(ctx), func(c *gin.Context) {
		userID := c.GetUint("user_id")
		userRole := c.GetString("user_role")
		canViewAll := c.GetBool("can_view_all_customers")
		
		// 获取销售人员ID
		salespersonIDStr := c.Query("salesperson_id")
		var salespersonID uint
		
		if canViewAll {
			// 客服/超级管理员可以查看指定销售的提成
			if salespersonIDStr != "" {
				id, _ := strconv.ParseUint(salespersonIDStr, 10, 32)
				salespersonID = uint(id)
			}
		} else if userRole == middleware.RoleSales {
			// 销售只能查看自己的提成
			salespersonID = userID
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足，只有销售人员可以查看提成"})
			return
		}
		
		// 如果没有指定销售ID且是管理员，返回错误
		if salespersonID == 0 && canViewAll {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请指定销售人员ID"})
			return
		}
		
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		// 查询提成记录
		commissionRepo := repository.NewCommissionRepository(ctx.DB)
		commissions, err := commissionRepo.FindBySalespersonID(salespersonID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// 计算总积分
		totalPoints := 0.0
		for _, comm := range commissions {
			totalPoints += comm.Points
		}
		
		c.JSON(http.StatusOK, gin.H{
			"commissions":  commissions,
			"total":        len(commissions),
			"total_points": totalPoints,
		})
	})
}
