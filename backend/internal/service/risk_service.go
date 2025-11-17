/**
 * 风控引擎服务
 * 
 * 用途：
 * - 监控所有持仓订单的定金率
 * - 触发预警和强平机制
 * - 自动结算强平订单
 * - 记录风控事件
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package service

import (
	"fmt"
	"log"
	"time"

	"suxin/internal/appctx"
	"suxin/internal/model"
	"suxin/internal/repository"
)

/**
 * RiskService 风控引擎服务
 */
type RiskService struct {
	ctx       *appctx.AppContext
	orderRepo *repository.OrderRepository
	userRepo  *repository.UserRepository
}

/**
 * NewRiskService 创建风控引擎服务实例
 * 
 * @param ctx *appctx.AppContext - 应用上下文
 * @return *RiskService
 */
func NewRiskService(ctx *appctx.AppContext) *RiskService {
	return &RiskService{
		ctx:       ctx,
		orderRepo: repository.NewOrderRepository(ctx.DB),
		userRepo:  repository.NewUserRepository(ctx.DB),
	}
}

/**
 * RiskCheckResult 风控检查结果
 */
type RiskCheckResult struct {
	TotalOrders      int                // 总订单数
	NeedForceClose   []*model.Order     // 需要强平的订单
	HighRisk         []*model.Order     // 高风险订单
	Warning          []*model.Order     // 需要预警的订单
	CheckTime        time.Time          // 检查时间
}

/**
 * CheckAllOrders 检查所有持仓订单的风控状态
 * 
 * 业务流程：
 * 1. 获取所有持仓订单
 * 2. 使用当前价格更新每个订单的盈亏和定金率
 * 3. 分类订单：强平/高风险/预警
 * 4. 保存更新后的订单数据
 * 
 * @param currentPrice float64 - 当前市场价格（元/克）
 * @return (*RiskCheckResult, error)
 */
func (s *RiskService) CheckAllOrders(currentPrice float64) (*RiskCheckResult, error) {
	// 1. 获取所有持仓订单
	orders, err := s.orderRepo.FindHoldingOrders()
	if err != nil {
		return nil, fmt.Errorf("获取持仓订单失败: %v", err)
	}

	result := &RiskCheckResult{
		TotalOrders:    len(orders),
		NeedForceClose: make([]*model.Order, 0),
		HighRisk:       make([]*model.Order, 0),
		Warning:        make([]*model.Order, 0),
		CheckTime:      time.Now(),
	}

	// 2. 遍历所有订单，更新价格和定金率
	for _, order := range orders {
		// 更新订单的当前价格、盈亏和定金率
		order.UpdatePnLAndMargin(currentPrice)

		// 3. 保存更新后的订单数据
		if err := s.orderRepo.UpdatePnLAndMargin(order); err != nil {
			log.Printf("[Risk] 更新订单 %s 失败: %v", order.OrderID, err)
			continue
		}

		// 4. 根据定金率分类订单
		if order.IsNeedForceClose() {
			// 定金率 ≤ 20%：需要强制平仓
			result.NeedForceClose = append(result.NeedForceClose, order)
			log.Printf("[Risk] ⚠️ 订单 %s 定金率 %.2f%% ≤ 20%%，需要强制平仓", 
				order.OrderID, order.MarginRate)
		} else if order.IsHighRisk() {
			// 20% < 定金率 < 25%：高风险预警
			result.HighRisk = append(result.HighRisk, order)
			log.Printf("[Risk] ⚠️ 订单 %s 定金率 %.2f%% 在高风险区间", 
				order.OrderID, order.MarginRate)
		} else if order.IsWarning() {
			// 定金率 ≤ 50%：一般预警
			result.Warning = append(result.Warning, order)
			log.Printf("[Risk] ⚠️ 订单 %s 定金率 %.2f%% ≤ 50%%，需要预警", 
				order.OrderID, order.MarginRate)
		}
	}

	log.Printf("[Risk] ✅ 风控检查完成：总计 %d 单，强平 %d 单，高风险 %d 单，预警 %d 单",
		result.TotalOrders, 
		len(result.NeedForceClose), 
		len(result.HighRisk), 
		len(result.Warning))

	return result, nil
}

/**
 * AutoForceClose 自动强制平仓
 * 
 * 业务流程：
 * 1. 获取需要强平的订单列表
 * 2. 对每个订单执行强平操作
 * 3. 更新用户资金（释放定金 + 结算盈亏）
 * 4. 发送强平通知
 * 
 * @param orders []*model.Order - 需要强平的订单列表
 * @param closePrice float64 - 平仓价格
 * @return (int, error) - 成功强平的订单数量
 */
func (s *RiskService) AutoForceClose(orders []*model.Order, closePrice float64) (int, error) {
	successCount := 0

	for _, order := range orders {
		// 开启事务
		tx := s.ctx.DB.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 1. 获取用户信息
		user, err := s.userRepo.FindByID(order.UserID)
		if err != nil {
			log.Printf("[Risk] 获取用户 %d 信息失败: %v", order.UserID, err)
			continue
		}

		// 2. 计算最终盈亏
		finalPnL := order.CalculatePnL(closePrice)

		// 3. 执行强平操作
		order.ForceClose(closePrice)

		// 4. 更新用户资金
		// 释放已用定金，结算金额（定金 + 盈亏）加回可用定金
		newAvailable := user.AvailableDeposit + order.Deposit + finalPnL
		newUsed := user.UsedDeposit - order.Deposit

		// 防止资金为负
		if newAvailable < 0 {
			log.Printf("[Risk] ⚠️ 订单 %s 强平后资金异常（可用定金为负: %.2f），跳过", 
				order.OrderID, newAvailable)
			tx.Rollback()
			continue
		}

		if err := tx.Model(&model.User{}).Where("id = ?", order.UserID).Updates(map[string]interface{}{
			"available_deposit": newAvailable,
			"used_deposit":      newUsed,
		}).Error; err != nil {
			log.Printf("[Risk] 更新用户 %d 资金失败: %v", order.UserID, err)
			tx.Rollback()
			continue
		}

		// 5. 保存订单状态
		if err := tx.Save(order).Error; err != nil {
			log.Printf("[Risk] 保存订单 %s 状态失败: %v", order.OrderID, err)
			tx.Rollback()
			continue
		}

		// 6. 提交事务
		if err := tx.Commit().Error; err != nil {
			log.Printf("[Risk] 订单 %s 强平事务提交失败: %v", order.OrderID, err)
			continue
		}

		successCount++
		log.Printf("[Risk] ✅ 订单 %s 强制平仓成功，平仓价 %.2f，最终盈亏 %.2f",
			order.OrderID, closePrice, finalPnL)

		// TODO: 7. 发送强平通知（后续实现）
	}

	log.Printf("[Risk] 🎯 自动强平完成：成功 %d/%d 单", successCount, len(orders))
	return successCount, nil
}

/**
 * RunRiskCheck 执行风控检查（定时任务调用）
 * 
 * 业务流程：
 * 1. 获取当前市场价格
 * 2. 检查所有持仓订单
 * 3. 自动执行强平
 * 4. 发送预警通知
 * 
 * @param currentPrice float64 - 当前市场价格
 * @return error
 */
func (s *RiskService) RunRiskCheck(currentPrice float64) error {
	log.Printf("[Risk] 🔍 开始风控检查，当前价格: %.2f 元/克", currentPrice)

	// 1. 检查所有订单
	result, err := s.CheckAllOrders(currentPrice)
	if err != nil {
		return fmt.Errorf("风控检查失败: %v", err)
	}

	// 2. 自动强平
	if len(result.NeedForceClose) > 0 {
		log.Printf("[Risk] 🚨 发现 %d 单需要强制平仓", len(result.NeedForceClose))
		_, err := s.AutoForceClose(result.NeedForceClose, currentPrice)
		if err != nil {
			log.Printf("[Risk] 自动强平失败: %v", err)
		}
	}

	// 3. 发送高风险预警（后续实现）
	if len(result.HighRisk) > 0 {
		log.Printf("[Risk] ⚠️ 发现 %d 单高风险订单", len(result.HighRisk))
		// TODO: 发送通知
	}

	// 4. 发送一般预警（后续实现）
	if len(result.Warning) > 0 {
		log.Printf("[Risk] ⚠️ 发现 %d 单需要预警", len(result.Warning))
		// TODO: 发送通知
	}

	return nil
}

/**
 * GetRiskStatistics 获取风控统计数据
 * 
 * @param currentPrice float64 - 当前市场价格
 * @return map[string]interface{}
 */
func (s *RiskService) GetRiskStatistics(currentPrice float64) (map[string]interface{}, error) {
	result, err := s.CheckAllOrders(currentPrice)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_orders":       result.TotalOrders,
		"force_close_count":  len(result.NeedForceClose),
		"high_risk_count":    len(result.HighRisk),
		"warning_count":      len(result.Warning),
		"safe_count":         result.TotalOrders - len(result.NeedForceClose) - len(result.HighRisk) - len(result.Warning),
		"check_time":         result.CheckTime,
		"current_price":      currentPrice,
	}

	return stats, nil
}
