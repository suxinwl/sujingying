/**
 * 风控API处理器
 * 
 * 用途：
 * - 提供风控统计数据查询接口
 * - 支持手动触发风控检查（管理员）
 * - 查看风控历史记录
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"suxin/internal/appctx"
	"suxin/internal/service"
)

/**
 * RegisterRiskRoutes 注册风控路由
 * 
 * 路由列表：
 * - GET /risk/statistics  获取风控统计数据（需JWT）
 * 
 * @param rg *gin.RouterGroup - 路由组
 * @param ctx *appctx.AppContext - 应用上下文
 * @return void
 */
func RegisterRiskRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	riskSvc := service.NewRiskService(ctx)

	/**
	 * GET /risk/statistics - 获取风控统计数据
	 * 
	 * 查询参数：
	 * - current_price: 当前价格（可选，默认500.00）
	 * 
	 * 响应：
	 * {
	 *   "total_orders": 10,
	 *   "force_close_count": 2,
	 *   "high_risk_count": 3,
	 *   "warning_count": 4,
	 *   "safe_count": 1,
	 *   "check_time": "2025-11-18T01:00:00Z",
	 *   "current_price": 500.00
	 * }
	 */
	rg.GET("/risk/statistics", func(c *gin.Context) {
		// 获取当前价格参数（默认500.00）
		currentPrice := 500.00
		if price := c.Query("current_price"); price != "" {
			if p, err := parseFloat(price); err == nil && p > 0 {
				currentPrice = p
			}
		}

		// 获取风控统计数据
		stats, err := riskSvc.GetRiskStatistics(currentPrice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	})
}

/**
 * parseFloat 解析浮点数
 * 
 * @param s string - 字符串
 * @return (float64, error)
 */
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
