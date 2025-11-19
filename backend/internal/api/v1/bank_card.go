/**
 * 银行卡API处理器
 * 
 * 用途：
 * - 处理银行卡相关HTTP请求
 * - 集成支付密码验证
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

type addCardReq struct {
	BankName    string `json:"bank_name" binding:"required"`
	CardNumber  string `json:"card_number" binding:"required"`
	CardHolder  string `json:"card_holder" binding:"required"`
	PayPassword string `json:"pay_password" binding:"required"`
}

/**
 * RegisterBankCardRoutes 注册银行卡路由
 */
func RegisterBankCardRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	cardSvc := service.NewBankCardService(ctx)
	paypassSvc := service.NewPayPassService(ctx)
	
	// POST /bank-cards - 添加银行卡
	rg.POST("/bank-cards", func(c *gin.Context) {
		var req addCardReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
			return
		}
		
		userID := c.GetUint("user_id")
		
		// 验证支付密码
		if err := paypassSvc.RequirePayPassword(userID, req.PayPassword); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		
		card, err := cardSvc.AddCard(userID, req.BankName, req.CardNumber, req.CardHolder)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"id":          card.ID,
			"bank_name":   card.BankName,
			"card_number": card.MaskCardNumber(),
			"card_holder": card.CardHolder,
			"is_default":  card.IsDefault,
		})
	})
	
	// GET /bank-cards - 获取银行卡列表
	rg.GET("/bank-cards", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		cards, err := cardSvc.GetUserCards(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"cards": cards})
	})
	
	// PUT /bank-cards/:id/default - 设置默认银行卡
	rg.PUT("/bank-cards/:id/default", func(c *gin.Context) {
		cardID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
		userID := c.GetUint("user_id")
		
		if err := cardSvc.SetDefaultCard(userID, uint(cardID)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "已设为默认银行卡"})
	})
	
	// DELETE /bank-cards/:id - 删除银行卡
	rg.DELETE("/bank-cards/:id", func(c *gin.Context) {
		cardID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
		userID := c.GetUint("user_id")
		
		if err := cardSvc.DeleteCard(userID, uint(cardID)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "银行卡已删除"})
	})
}
