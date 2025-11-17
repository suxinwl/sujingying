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
	"suxin/internal/middleware"
	"suxin/internal/service"
)

type submitSupplementReq struct {
	OrderID    uint    `json:"order_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Method     string  `json:"method" binding:"required"`
	VoucherURL string  `json:"voucher_url"`
}

type reviewSupplementReq struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Note   string `json:"note"`
}

/**
 * RegisterSupplementRoutes 注册补定金路由
 * 
 * 路由列表：
 * - POST /supplements              提交补定金申请(需JWT)
 * - GET  /supplements              查询补定金记录(需JWT)
 * - GET  /supplements/pending      查询待审核列表(需管理员)
 * - POST /supplements/:id/review   审核补定金(需管理员)
 */
func RegisterSupplementRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	supplementSvc := service.NewSupplementService(ctx)
	
	/**
	 * POST /supplements - 提交补定金申请
	 */
	rg.POST("/supplements", func(c *gin.Context) {
		var req submitSupplementReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		userID := c.GetUint("user_id")
		
		supplement, err := supplementSvc.SubmitSupplement(userID, req.OrderID, req.Amount, req.Method, req.VoucherURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"id":         supplement.ID,
			"order_id":   supplement.OrderID,
			"amount":     supplement.Amount,
			"status":     supplement.Status,
			"created_at": supplement.CreatedAt,
		})
	})
	
	/**
	 * GET /supplements - 查询补定金记录
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
	
	/**
	 * GET /supplements/pending - 查询待审核列表(需管理员)
	 */
	rg.GET("/supplements/pending", middleware.RequireAdmin(ctx), func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		
		supplements, err := supplementSvc.GetPendingSupplements(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"supplements": supplements,
			"total":       len(supplements),
		})
	})
	
	/**
	 * POST /supplements/:id/review - 审核补定金(需管理员)
	 */
	rg.POST("/supplements/:id/review", middleware.RequireAdmin(ctx), func(c *gin.Context) {
		supplementID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的补定金ID"})
			return
		}
		
		var req reviewSupplementReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		reviewerID := c.GetUint("user_id")
		
		if req.Action == "approve" {
			err = supplementSvc.ApproveSupplement(uint(supplementID), reviewerID, req.Note)
		} else {
			err = supplementSvc.RejectSupplement(uint(supplementID), reviewerID, req.Note)
		}
		
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "审核成功",
		})
	})
}
