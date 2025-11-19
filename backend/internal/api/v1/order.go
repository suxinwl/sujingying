/**
 * 订单API处理器
 * 
 * 用途：
 * - 处理订单相关HTTP请求
 * - 集成支付密码验证
 * - 返回标准化JSON响应
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"suxin/internal/appctx"
	"suxin/internal/service"
)

/**
 * CreateOrderRequest 创建订单请求结构
 */
type createOrderReq struct {
	Type        string  `json:"type" binding:"required"`         // long_buy / short_sell
	LockedPrice float64 `json:"locked_price" binding:"required"` // 锁定价格
	WeightG     float64 `json:"weight_g" binding:"required"`     // 克重
	Deposit     float64 `json:"deposit" binding:"required"`      // 定金
}

/**
 * SettleOrderRequest 结算订单请求结构
 */
type settleOrderReq struct {
	SettlePrice float64 `json:"settle_price" binding:"required"` // 结算价格
	PayPassword string  `json:"pay_password" binding:"required"` // 支付密码
}

/**
 * RegisterOrderRoutes 注册订单路由
 * 
 * 路由列表：
 * - POST   /orders          创建订单（需JWT+支付密码）
 * - GET    /orders          获取订单列表（需JWT）
 * - GET    /orders/:orderID 获取订单详情（需JWT）
 * - POST   /orders/:orderID/settle 现金结算（需JWT+支付密码）
 * 
 * @param rg *gin.RouterGroup - 路由组
 * @param ctx *appctx.AppContext - 应用上下文
 * @return void
 */
func RegisterOrderRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	orderSvc := service.NewOrderService(ctx)
	paypassSvc := service.NewPayPassService(ctx)
	
	/**
	 * POST /orders - 创建订单
	 * 
	 * 请求body：
	 * {
	 *   "type": "long_buy",          // long_buy(买料) 或 short_sell(卖料)
	 *   "locked_price": 500.00,      // 锁定价格（元/克）
	 *   "weight_g": 100.0,           // 克重
	 *   "deposit": 10000.00,         // 定金
	 *   "pay_password": "123456"     // 支付密码
	 * }
	 * 
	 * 响应：
	 * {
	 *   "order_id": "202511180001",
	 *   "type": "long_buy",
	 *   "locked_price": 500.00,
	 *   "weight_g": 100.0,
	 *   "deposit": 10000.00,
	 *   "pnl_float": 0.00,
	 *   "margin_rate": 100.00,
	 *   "status": "holding"
	 * }
	 */
	rg.POST("/orders", func(c *gin.Context) {
		var req createOrderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		userID := c.GetUint("user_id")
		
		// 创建订单
		order, err := orderSvc.CreateOrder(userID, service.CreateOrderRequest{
			Type:        req.Type,
			LockedPrice: req.LockedPrice,
			WeightG:     req.WeightG,
			Deposit:     req.Deposit,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"order_id":     order.OrderID,
			"type":         order.Type,
			"locked_price": order.LockedPrice,
			"weight_g":     order.WeightG,
			"deposit":      order.Deposit,
			"pnl_float":    order.PnLFloat,
			"margin_rate":  order.MarginRate,
			"status":       order.Status,
			"created_at":   order.CreatedAt,
		})
	})
	
	/**
	 * GET /orders - 获取订单列表
	 * 
	 * 查询参数：
	 * - status: 订单状态（可选，holding/settled/closed）
	 * 
	 * 响应：
	 * {
	 *   "orders": [
	 *     {
	 *       "order_id": "202511180001",
	 *       "type": "long_buy",
	 *       ...
	 *     }
	 *   ],
	 *   "total": 10
	 * }
	 */
	rg.GET("/orders", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		status := c.Query("status") // 可选参数
		
		orders, err := orderSvc.GetUserOrders(userID, status)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"orders": orders,
			"total":  len(orders),
		})
	})
	
	/**
	 * GET /orders/:orderID - 获取订单详情
	 * 
	 * 路径参数：
	 * - orderID: 订单号
	 * 
	 * 响应：
	 * {
	 *   "order_id": "202511180001",
	 *   "type": "long_buy",
	 *   "locked_price": 500.00,
	 *   "current_price": 510.00,
	 *   "weight_g": 100.0,
	 *   "deposit": 10000.00,
	 *   "pnl_float": 1000.00,
	 *   "margin_rate": 110.00,
	 *   "status": "holding",
	 *   "created_at": "2025-11-18T00:00:00Z"
	 * }
	 */
	rg.GET("/orders/:orderID", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		orderID := c.Param("orderID")
		
		order, err := orderSvc.GetOrderDetail(userID, orderID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, order)
	})
	
	/**
	 * POST /orders/:orderID/settle - 现金结算订单
	 * 
	 * 路径参数：
	 * - orderID: 订单号
	 * 
	 * 请求body：
	 * {
	 *   "settle_price": 510.00,      // 结算价格
	 *   "pay_password": "123456"     // 支付密码
	 * }
	 * 
	 * 响应：
	 * {
	 *   "order_id": "202511180001",
	 *   "settled_price": 510.00,
	 *   "settled_pnl": 1000.00,
	 *   "status": "settled",
	 *   "settled_at": "2025-11-18T01:00:00Z"
	 * }
	 */
	rg.POST("/orders/:orderID/settle", func(c *gin.Context) {
		var req settleOrderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		userID := c.GetUint("user_id")
		orderID := c.Param("orderID")
		
		// 验证支付密码
		if err := paypassSvc.RequirePayPassword(userID, req.PayPassword); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		
		// 执行结算
		order, err := orderSvc.SettleOrder(userID, orderID, req.SettlePrice)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"order_id":      order.OrderID,
			"settled_price": order.SettledPrice,
			"settled_pnl":   order.SettledPnL,
			"status":        order.Status,
			"settled_at":    order.SettledAt,
		})
	})
}
