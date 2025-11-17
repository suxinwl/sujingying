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
}
