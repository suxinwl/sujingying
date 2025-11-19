/**
 * 资金流水API处理器
 * 
 * 用途：
 * - 提供资金流水查询功能
 * - 支持按类型和时间筛选
 * - 统计资金变动
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"suxin/internal/appctx"
	"suxin/internal/model"
	"suxin/internal/repository"
)

// FundLogResponse 前端资金流水响应结构，增加了订单信息字段
type FundLogResponse struct {
	ID              uint      `json:"ID"`
	UserID          uint      `json:"UserID"`
	Type            string    `json:"Type"`
	Amount          float64   `json:"Amount"`
	AvailableBefore float64   `json:"AvailableBefore"`
	AvailableAfter  float64   `json:"AvailableAfter"`
	UsedBefore      float64   `json:"UsedBefore"`
	UsedAfter       float64   `json:"UsedAfter"`
	RelatedID       uint      `json:"RelatedID"`
	RelatedType     string    `json:"RelatedType"`
	Note            string    `json:"Note"`
	CreatedAt       time.Time `json:"CreatedAt"`

	// 额外补充的字段：用于前端“补定金”Tab 的料单筛选与详情
	OrderID   string `json:"OrderID,omitempty"`
	OrderType string `json:"OrderType,omitempty"`
}

/**
 * RegisterFundLogRoutes 注册资金流水路由
 * 
 * 路由列表：
 * - GET /fund-logs           查询资金流水（需JWT）
 * - GET /fund-logs/summary   流水统计（需JWT）
 * 
 * @param rg *gin.RouterGroup - 路由组
 * @param ctx *appctx.AppContext - 应用上下文
 * @return void
 */
func RegisterFundLogRoutes(rg *gin.RouterGroup, ctx *appctx.AppContext) {
	fundLogRepo := repository.NewFundLogRepository(ctx.DB)
	orderRepo := repository.NewOrderRepository(ctx.DB)
	
	/**
	 * GET /fund-logs - 查询资金流水
	 * 
	 * 查询参数：
	 * - type: 流水类型（可选）
	 * - start_date: 开始日期（可选，格式：2025-11-01）
	 * - end_date: 结束日期（可选）
	 * - limit: 每页数量（默认20）
	 * - offset: 偏移量（默认0）
	 * 
	 * 响应：
	 * {
	 *   "logs": [...],
	 *   "total": 100
	 * }
	 */
	rg.GET("/fund-logs", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		fundType := c.Query("type")
		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		
		var logs []*model.FundLog
		var err error
		
		// 根据参数选择查询方式
		if startDateStr != "" && endDateStr != "" {
			startDate, _ := time.Parse("2006-01-02", startDateStr)
			endDate, _ := time.Parse("2006-01-02", endDateStr)
			endDate = endDate.Add(24 * time.Hour) // 包含结束日期整天
			
			logs, err = fundLogRepo.FindByDateRange(userID, startDate, endDate)
		} else if fundType != "" {
			logs, err = fundLogRepo.FindByUserIDAndType(userID, fundType, limit)
		} else {
			logs, err = fundLogRepo.FindByUserID(userID, limit, offset)
		}
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// 构造响应结构，针对补定金流水附加订单信息
		responses := make([]FundLogResponse, 0, len(logs))
		orderCache := make(map[uint]*model.Order)
		
		for _, log := range logs {
			if log == nil {
				continue
			}
			resp := FundLogResponse{
				ID:              log.ID,
				UserID:          log.UserID,
				Type:            log.Type,
				Amount:          log.Amount,
				AvailableBefore: log.AvailableBefore,
				AvailableAfter:  log.AvailableAfter,
				UsedBefore:      log.UsedBefore,
				UsedAfter:       log.UsedAfter,
				RelatedID:       log.RelatedID,
				RelatedType:     log.RelatedType,
				Note:            log.Note,
				CreatedAt:       log.CreatedAt,
			}
			
			// 对补定金流水增加 OrderID / OrderType
			if log.Type == "supplement" && log.RelatedType == "order" && log.RelatedID > 0 {
				// 复用缓存，避免重复查库
				order, ok := orderCache[log.RelatedID]
				if !ok {
					order, err = orderRepo.FindByID(log.RelatedID)
					if err != nil {
						// 单条失败不影响整体，记录错误即可
						order = nil
					} else {
						orderCache[log.RelatedID] = order
					}
				}
				if order != nil {
					resp.OrderID = order.OrderID
					resp.OrderType = order.Type
				}
			}
			
			responses = append(responses, resp)
		}
		
		c.JSON(http.StatusOK, gin.H{
			"logs": responses,
		})
	})
	
	/**
	 * GET /fund-logs/summary - 流水统计
	 * 
	 * 查询参数：
	 * - start_date: 开始日期（可选）
	 * - end_date: 结束日期（可选）
	 * 
	 * 响应：
	 * {
	 *   "total_deposit": 50000.00,
	 *   "total_withdraw": 10000.00,
	 *   "total_settle": 5000.00,
	 *   "net_change": 45000.00
	 * }
	 */
	rg.GET("/fund-logs/summary", func(c *gin.Context) {
		userID := c.GetUint("user_id")
		
		// 统计各类型总额
		depositSum, _ := fundLogRepo.GetSumByType(userID, "deposit")
		withdrawSum, _ := fundLogRepo.GetSumByType(userID, "withdraw")
		settleSum, _ := fundLogRepo.GetSumByType(userID, "settle")
		
		netChange := depositSum + withdrawSum + settleSum
		
		c.JSON(http.StatusOK, gin.H{
			"total_deposit":  depositSum,
			"total_withdraw": withdrawSum,
			"total_settle":   settleSum,
			"net_change":     netChange,
		})
	})
}
