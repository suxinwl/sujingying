/**
 * 通知API处理器
 * 
 * 用途：
 * - 处理通知相关HTTP请求
 * - 提供通知查询、标记已读等接口
 * - 返回标准化JSON响应
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
 * MarkAsReadRequest 标记已读请求结构
 */
type markAsReadReq struct {
	IDs []uint `json:"ids" binding:"required"` // 通知ID列表
}

/**
 * RegisterNotificationRoutes 注册通知路由
 * 
 * 路由列表：
 * - GET    /notifications          获取通知列表（需JWT）
 * - GET    /notifications/unread   获取未读通知（需JWT）
 * - GET    /notifications/count    获取未读数量（需JWT）
 * - POST   /notifications/read     标记已读（需JWT）
 * - POST   /notifications/read-all 全部标记已读（需JWT）
 * 
 * @param rg *gin.RouterGroup - 路由组
 * @param ctx *appctx.AppContext - 应用上下文
 * @return void
 */
func RegisterNotificationRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	notiSvc := service.NewNotificationService(ctx)
	
	/**
	 * GET /notifications - 获取通知列表
	 * 
	 * 查询参数：
	 * - limit: 每页数量（默认20）
	 * - offset: 偏移量（默认0）
	 * 
	 * 响应：
	 * {
	 *   "notifications": [
	 *     {
	 *       "id": 1,
	 *       "type": "risk",
	 *       "level": "critical",
	 *       "title": "强制平仓通知",
	 *       "content": "订单号：202511180001...",
	 *       "status": "sent",
	 *       "created_at": "2025-11-18T01:00:00Z"
	 *     }
	 *   ],
	 *   "total": 10
	 * }
	 */
	rg.GET("/notifications", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		
		// 解析分页参数
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		// 获取通知列表
		notifications, err := notiSvc.GetUserNotifications(userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"notifications": notifications,
			"total":         len(notifications),
		})
	})
	
	/**
	 * GET /notifications/unread - 获取未读通知
	 * 
	 * 响应：
	 * {
	 *   "notifications": [...],
	 *   "count": 5
	 * }
	 */
	rg.GET("/notifications/unread", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		
		notifications, err := notiSvc.GetUnreadNotifications(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"notifications": notifications,
			"count":         len(notifications),
		})
	})
	
	/**
	 * GET /notifications/count - 获取未读数量
	 * 
	 * 响应：
	 * {
	 *   "count": 5
	 * }
	 */
	rg.GET("/notifications/count", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		
		count, err := notiSvc.GetUnreadCount(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"count": count,
		})
	})
	
	/**
	 * POST /notifications/read - 标记已读
	 * 
	 * 请求body：
	 * {
	 *   "ids": [1, 2, 3]  // 通知ID列表
	 * }
	 * 
	 * 响应：
	 * {
	 *   "message": "已标记为已读"
	 * }
	 */
	rg.POST("/notifications/read", func(c *gin.Context) {
		var req markAsReadReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		userID := c.GetUint("user_id")
		
		if err := notiSvc.MarkAsRead(req.IDs, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "已标记为已读",
		})
	})
	
	/**
	 * POST /notifications/read-all - 全部标记已读
	 * 
	 * 响应：
	 * {
	 *   "message": "所有通知已标记为已读"
	 * }
	 */
	rg.POST("/notifications/read-all", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		
		if err := notiSvc.MarkAllAsRead(userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "所有通知已标记为已读",
		})
	})
}
