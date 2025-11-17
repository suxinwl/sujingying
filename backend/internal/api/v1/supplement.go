/**
 * 补定金API处理器
 * 
 * 用途：
 * - 处理补定金申请和审核
 * - 提供补定金记录查询
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

type submitSupplementReq struct {
	OrderID uint    `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

/**
 * RegisterSupplementRoutes 注册补定金路由（无需审批，自动处理）
 * 
 * 路由列表：
 * - POST /supplements    提交补定金（自动处理，需JWT）
 * - GET  /supplements    查询补定金记录（需JWT）
 */
func RegisterSupplementRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	supplementSvc := service.NewSupplementService(ctx)
	
	/**
	 * POST /supplements - 提交补定金（自动处理）
	 * 
	 * 流程：
	 * 1. 检查可用定金是否充足
	 * 2. 如果充足：直接扣减并增加到订单定金
	 * 3. 如果不足：提示充值
	 * 
	 * 请求示例：
	 * {
	 *   "order_id": 123,
	 *   "amount": 5000.00
	 * }
	 * 
	 * 成功响应：
	 * {
	 *   "message": "补定金成功",
	 *   "supplement": {
	 *     "id": 1,
	 *     "order_id": 123,
	 *     "amount": 5000.00,
	 *     "status": "approved",
	 *     "old_deposit": 10000.00,
	 *     "new_deposit": 15000.00,
	 *     "old_margin_rate": 20.00,
	 *     "new_margin_rate": 30.00
	 *   }
	 * }
	 * 
	 * 失败响应（余额不足）：
	 * {
	 *   "error": "可用定金不足，当前可用: 3000.00 元，需要: 5000.00 元，请先充值"
	 * }
	 */
	rg.POST("/supplements", func(c *gin.Context) {
		var req submitSupplementReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		userID := c.GetUint("user_id")
		
		supplement, err := supplementSvc.SubmitSupplement(userID, req.OrderID, req.Amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "补定金成功",
			"supplement": gin.H{
				"id":         supplement.ID,
				"order_id":   supplement.OrderID,
				"amount":     supplement.Amount,
				"status":     supplement.Status,
				"created_at": supplement.CreatedAt,
			},
		})
	})
	
	/**
	 * GET /supplements - 查询补定金记录
	 * 
	 * 查询参数：
	 * - limit: 每页数量（默认20）
	 * - offset: 偏移量（默认0）
	 * 
	 * 响应：
	 * {
	 *   "supplements": [
	 *     {
	 *       "id": 1,
	 *       "order_id": 123,
	 *       "amount": 5000.00,
	 *       "status": "approved",
	 *       "created_at": "2025-11-18T00:00:00Z"
	 *     }
	 *   ],
	 *   "total": 10
	 * }
	 */
	rg.GET("/supplements", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		supplements, err := supplementSvc.GetUserSupplements(userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"supplements": supplements,
			"total":       len(supplements),
		})
	})
}
