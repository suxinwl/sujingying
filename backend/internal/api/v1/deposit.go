/**
 * 定金充值API处理器
 * 
 * 用途：
 * - 处理充值申请和审核
 * - 提供充值记录查询
 * - 支持管理员审核功能
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

type submitDepositReq struct {
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Method     string  `json:"method" binding:"required"`
	VoucherURL string  `json:"voucher_url"`
}

type reviewDepositReq struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Note   string `json:"note"`
}

/**
 * RegisterDepositRoutes 注册充值路由
 * 
 * 路由列表：
 * - POST /deposits              提交充值申请（需JWT）
 * - GET  /deposits              查询充值记录（需JWT）
 * - GET  /deposits/pending      查询待审核列表（需JWT+管理员）
 * - POST /deposits/:id/review   审核充值（需JWT+管理员）
 * 
 * @param rg *gin.RouterGroup - 路由组
 * @param ctx *appctx.AppContext - 应用上下文
 * @return void
 */
func RegisterDepositRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	depositSvc := service.NewDepositService(ctx)
	
	/**
	 * POST /deposits - 提交充值申请
	 * 
	 * 请求body：
	 * {
	 *   "amount": 10000.00,
	 *   "method": "bank",
	 *   "voucher_url": "https://example.com/voucher.jpg"
	 * }
	 * 
	 * 响应：
	 * {
	 *   "id": 1,
	 *   "amount": 10000.00,
	 *   "status": "pending",
	 *   "created_at": "2025-11-18T01:00:00Z"
	 * }
	 */
	rg.POST("/deposits", func(c *gin.Context) {
		var req submitDepositReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		userID := c.GetUint("user_id")
		
		deposit, err := depositSvc.SubmitDeposit(userID, req.Amount, req.Method, req.VoucherURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"id":         deposit.ID,
			"amount":     deposit.Amount,
			"method":     deposit.Method,
			"status":     deposit.Status,
			"created_at": deposit.CreatedAt,
		})
	})
	
	/**
	 * GET /deposits - 查询用户充值记录
	 * 
	 * 查询参数：
	 * - limit: 每页数量（默认20）
	 * - offset: 偏移量（默认0）
	 * 
	 * 响应：
	 * {
	 *   "deposits": [...],
	 *   "total": 10
	 * }
	 */
	rg.GET("/deposits", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		deposits, err := depositSvc.GetUserDeposits(userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"deposits": deposits,
			"total":    len(deposits),
		})
	})
	
	/**
	 * GET /deposits/pending - 查询待审核列表（管理员）
	 * 
	 * 响应：
	 * {
	 *   "deposits": [...],
	 *   "total": 10
	 * }
	 */
	admin := rg.Group("", middleware.RequireAdmin(ctx))
	
	admin.GET("/deposits/pending", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		
		deposits, err := depositSvc.GetPendingDeposits(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"deposits": deposits,
			"total":    len(deposits),
		})
	})
	
	/**
	 * POST /deposits/:id/review - 审核充值申请（管理员）
	 * 
	 * 请求body：
	 * {
	 *   "action": "approve",  // approve 或 reject
	 *   "note": "审核备注"
	 * }
	 * 
	 * 响应：
	 * {
	 *   "message": "审核成功",
	 *   "deposit": {...}
	 * }
	 */
	admin.POST("/deposits/:id/review", func(c *gin.Context) {
		depositID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的充值ID"})
			return
		}
		
		var req reviewDepositReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		reviewerID := c.GetUint("user_id")
		
		if req.Action == "approve" {
			err = depositSvc.ApproveDeposit(uint(depositID), reviewerID, req.Note)
		} else {
			err = depositSvc.RejectDeposit(uint(depositID), reviewerID, req.Note)
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

type submitWithdrawReq struct {
	BankCardID uint    `json:"bank_card_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
}

type reviewWithdrawReq struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Note   string `json:"note"`
}

/**
 * RegisterWithdrawRoutes 注册提现路由
 * 
 * 路由列表：
 * - POST /withdraws              提交提现申请（需JWT）
 * - GET  /withdraws              查询提现记录（需JWT）
 * - GET  /withdraws/pending      查询待审核列表（需JWT+管理员）
 * - POST /withdraws/:id/review   审核提现（需JWT+管理员）
 * 
 * @param rg *gin.RouterGroup - 路由组
 * @param ctx *appctx.AppContext - 应用上下文
 * @return void
 */
func RegisterWithdrawRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	withdrawSvc := service.NewWithdrawService(ctx)
	
	// POST /withdraws - 提交提现申请
	rg.POST("/withdraws", func(c *gin.Context) {
		var req submitWithdrawReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		userID := c.GetUint("user_id")
		
		withdraw, err := withdrawSvc.SubmitWithdraw(userID, req.BankCardID, req.Amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"id":            withdraw.ID,
			"amount":        withdraw.Amount,
			"fee":           withdraw.Fee,
			"actual_amount": withdraw.ActualAmount,
			"status":        withdraw.Status,
			"created_at":    withdraw.CreatedAt,
		})
	})
	
	// GET /withdraws - 查询提现记录
	rg.GET("/withdraws", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		withdraws, err := withdrawSvc.GetUserWithdraws(userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"withdraws": withdraws,
			"total":     len(withdraws),
		})
	})
	
	// 管理员权限路由组
	admin := rg.Group("", middleware.RequireAdmin(ctx))
	
	// GET /withdraws/pending - 查询待审核列表（管理员）
	admin.GET("/withdraws/pending", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		
		withdraws, err := withdrawSvc.GetPendingWithdraws(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"withdraws": withdraws,
			"total":     len(withdraws),
		})
	})
	
	// POST /withdraws/:id/review - 审核提现（管理员）
	admin.POST("/withdraws/:id/review", func(c *gin.Context) {
		withdrawID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的提现ID"})
			return
		}
		
		var req reviewWithdrawReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		reviewerID := c.GetUint("user_id")
		
		if req.Action == "approve" {
			err = withdrawSvc.ApproveWithdraw(uint(withdrawID), reviewerID, req.Note)
		} else {
			err = withdrawSvc.RejectWithdraw(uint(withdrawID), reviewerID, req.Note)
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
