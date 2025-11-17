/**
 * 订单服务层
 * 
 * 用途：
 * - 实现订单业务逻辑
 * - 处理订单创建、查询、结算
 * - 集成支付密码验证
 * - 管理资金冻结和释放
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"suxin/internal/appctx"
	"suxin/internal/model"
	"suxin/internal/repository"
)

/**
 * OrderService 订单服务
 */
type OrderService struct {
	ctx       *appctx.AppContext
	orderRepo *repository.OrderRepository
	userRepo  *repository.UserRepository
}

/**
 * NewOrderService 创建订单服务实例
 * 
 * @param ctx *appctx.AppContext - 应用上下文
 * @return *OrderService
 */
func NewOrderService(ctx *appctx.AppContext) *OrderService {
	return &OrderService{
		ctx:       ctx,
		orderRepo: repository.NewOrderRepository(ctx.DB),
		userRepo:  repository.NewUserRepository(ctx.DB),
	}
}

/**
 * CreateOrderRequest 创建订单请求
 */
type CreateOrderRequest struct {
	Type        string  `json:"type" binding:"required"`         // 订单类型
	LockedPrice float64 `json:"locked_price" binding:"required"` // 锁定价格
	WeightG     float64 `json:"weight_g" binding:"required"`     // 克重
	Deposit     float64 `json:"deposit" binding:"required"`      // 定金
}

/**
 * CreateOrder 创建订单
 * 
 * 业务流程：
 * 1. 验证订单类型和参数
 * 2. 检查用户可用定金是否足够
 * 3. 冻结定金（从可用转到已用）
 * 4. 生成订单号并创建订单
 * 5. 初始化订单的盈亏和定金率
 * 
 * @param userID uint - 用户ID
 * @param req CreateOrderRequest - 创建订单请求
 * @return (*model.Order, error)
 */
func (s *OrderService) CreateOrder(userID uint, req CreateOrderRequest) (*model.Order, error) {
	// 1. 验证订单类型
	if req.Type != model.OrderTypeLongBuy && req.Type != model.OrderTypeShortSell {
		return nil, errors.New("无效的订单类型")
	}
	
	// 验证参数合法性
	if req.LockedPrice <= 0 {
		return nil, errors.New("锁定价格必须大于0")
	}
	if req.WeightG <= 0 {
		return nil, errors.New("克重必须大于0")
	}
	if req.Deposit <= 0 {
		return nil, errors.New("定金必须大于0")
	}
	
	// 2. 检查用户可用定金
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	
	if user.AvailableDeposit < req.Deposit {
		return nil, fmt.Errorf("可用定金不足（可用: %.2f, 需要: %.2f）", 
			user.AvailableDeposit, req.Deposit)
	}
	
	// 3. 开启数据库事务
	tx := s.ctx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 4. 冻结定金（可用 -> 已用）
	if err := tx.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"available_deposit": user.AvailableDeposit - req.Deposit,
		"used_deposit":      user.UsedDeposit + req.Deposit,
	}).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("冻结定金失败")
	}
	
	// 5. 生成订单号（格式：年月日时分秒+用户ID）
	orderID := fmt.Sprintf("%s%d", time.Now().Format("20060102150405"), userID)
	
	// 6. 创建订单
	order := &model.Order{
		OrderID:      orderID,
		UserID:       userID,
		Type:         req.Type,
		LockedPrice:  req.LockedPrice,
		CurrentPrice: req.LockedPrice, // 初始当前价格等于锁定价格
		WeightG:      req.WeightG,
		Deposit:      req.Deposit,
		Status:       model.OrderStatusHolding,
	}
	
	// 7. 初始化盈亏和定金率
	order.UpdatePnLAndMargin(req.LockedPrice)
	
	// 8. 保存订单
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("创建订单失败")
	}
	
	// 9. 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("事务提交失败")
	}
	
	return order, nil
}

/**
 * GetUserOrders 获取用户订单列表
 * 
 * @param userID uint - 用户ID
 * @param status string - 订单状态（可选）
 * @return ([]*model.Order, error)
 */
func (s *OrderService) GetUserOrders(userID uint, status string) ([]*model.Order, error) {
	return s.orderRepo.FindByUserID(userID, status)
}

/**
 * GetOrderDetail 获取订单详情
 * 
 * @param userID uint - 用户ID
 * @param orderID string - 订单号
 * @return (*model.Order, error)
 */
func (s *OrderService) GetOrderDetail(userID uint, orderID string) (*model.Order, error) {
	order, err := s.orderRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	
	// 验证订单归属
	if order.UserID != userID {
		return nil, errors.New("无权访问此订单")
	}
	
	return order, nil
}

/**
 * SettleOrder 现金结算订单
 * 
 * 业务流程：
 * 1. 验证订单状态（只能结算持仓订单）
 * 2. 计算最终盈亏
 * 3. 更新用户资金（释放定金 + 盈亏）
 * 4. 更新订单状态为已结算
 * 5. 记录资金变更日志
 * 
 * @param userID uint - 用户ID
 * @param orderID string - 订单号
 * @param settlePrice float64 - 结算价格
 * @return (*model.Order, error)
 */
func (s *OrderService) SettleOrder(userID uint, orderID string, settlePrice float64) (*model.Order, error) {
	// 1. 查找订单
	order, err := s.orderRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	
	// 验证订单归属
	if order.UserID != userID {
		return nil, errors.New("无权操作此订单")
	}
	
	// 2. 验证订单状态
	if !order.CanSettle() {
		return nil, fmt.Errorf("订单状态不允许结算（当前状态: %s）", order.Status)
	}
	
	// 验证结算价格
	if settlePrice <= 0 {
		return nil, errors.New("结算价格必须大于0")
	}
	
	// 3. 计算结算盈亏
	settledPnL := order.CalculatePnL(settlePrice)
	
	// 4. 查询用户信息
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	
	// 5. 开启事务
	tx := s.ctx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 6. 更新用户资金
	// 释放已用定金，结算金额（定金 + 盈亏）加回可用定金
	newAvailable := user.AvailableDeposit + order.Deposit + settledPnL
	newUsed := user.UsedDeposit - order.Deposit
	
	// 防止资金为负（理论上不应该发生）
	if newAvailable < 0 {
		tx.Rollback()
		return nil, errors.New("结算后资金异常（可用定金为负）")
	}
	
	if err := tx.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"available_deposit": newAvailable,
		"used_deposit":      newUsed,
	}).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("更新用户资金失败")
	}
	
	// 7. 更新订单状态
	order.Settle(settlePrice)
	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("更新订单状态失败")
	}
	
	// TODO: 8. 记录资金变更日志（后续实现）
	
	// 9. 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("事务提交失败")
	}
	
	// 10. 异步计算销售提成
	go func() {
		salesSvc := NewSalesService(s.ctx)
		if err := salesSvc.ProcessOrderCommission(order.ID); err != nil {
			log.Printf("[Order] 计算提成失败: %v", err)
		}
	}()
	
	return order, nil
}

/**
 * UpdateOrderPrices 批量更新订单价格和盈亏
 * 
 * 用途：风控引擎定时调用，更新所有持仓订单的实时数据
 * 
 * @param currentPrice float64 - 当前市场价格
 * @return error
 */
func (s *OrderService) UpdateOrderPrices(currentPrice float64) error {
	// 获取所有持仓订单
	orders, err := s.orderRepo.FindHoldingOrders()
	if err != nil {
		return err
	}
	
	// 批量更新每个订单
	for _, order := range orders {
		order.UpdatePnLAndMargin(currentPrice)
		if err := s.orderRepo.UpdatePnLAndMargin(order); err != nil {
			// 记录错误但继续处理其他订单
			fmt.Printf("更新订单 %s 失败: %v\n", order.OrderID, err)
		}
	}
	
	return nil
}
