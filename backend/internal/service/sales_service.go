/**
 * 销售管理服务层
 * 
 * 用途：
 * - 实现销售业务逻辑
 * - 提成计算和积分管理
 * - 客户归属管理
 * - 销售看板数据
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
 * SalesService 销售服务
 */
type SalesService struct {
	ctx            *appctx.AppContext
	salesRepo      *repository.SalespersonRepository
	commissionRepo *repository.CommissionRepository
	userRepo       *repository.UserRepository
	orderRepo      *repository.OrderRepository
}

/**
 * NewSalesService 创建销售服务实例
 * 
 * @param ctx *appctx.AppContext - 应用上下文
 * @return *SalesService
 */
func NewSalesService(ctx *appctx.AppContext) *SalesService {
	return &SalesService{
		ctx:            ctx,
		salesRepo:      repository.NewSalespersonRepository(ctx.DB),
		commissionRepo: repository.NewCommissionRepository(ctx.DB),
		userRepo:       repository.NewUserRepository(ctx.DB),
		orderRepo:      repository.NewOrderRepository(ctx.DB),
	}
}

/**
 * ProcessOrderCommission 处理订单提成（订单结算时调用）
 * 
 * 业务流程：
 * 1. 查找订单和客户信息
 * 2. 获取客户归属的销售人员
 * 3. 检查是否已生成提成
 * 4. 计算提成积分
 * 5. 创建提成记录
 * 6. 更新销售人员积分
 * 
 * @param orderID uint - 订单ID
 * @return error
 */
func (s *SalesService) ProcessOrderCommission(orderID uint) error {
	// 1. 查找订单
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("订单不存在: %v", err)
	}

	// 2. 验证订单状态
	if order.Status != model.OrderStatusSettled {
		return errors.New("只有已结算订单才能计算提成")
	}

	// 3. 查找客户信息
	customer, err := s.userRepo.FindByID(order.UserID)
	if err != nil {
		return fmt.Errorf("客户不存在: %v", err)
	}

	// 4. 检查客户是否有归属销售
	if customer.SalesID == 0 {
		log.Printf("[Sales] 订单 %s 的客户无归属销售，跳过提成", order.OrderID)
		return nil
	}

	// 5. 检查是否已生成提成
	exists, err := s.commissionRepo.ExistsByOrderID(order.ID)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("[Sales] 订单 %s 已生成提成记录，跳过", order.OrderID)
		return nil
	}

	// 6. 查找销售人员
	salesperson, err := s.salesRepo.FindByID(customer.SalesID)
	if err != nil {
		return fmt.Errorf("销售人员不存在: %v", err)
	}

	// 7. 计算提成积分
	points := model.CalculatePoints(order.WeightG, salesperson.CommissionRate)

	// 8. 开启事务
	tx := s.ctx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 9. 创建提成记录
	record := &model.CommissionRecord{
		SalespersonID:  salesperson.ID,
		OrderID:        order.ID,
		CustomerID:     customer.ID,
		WeightG:        order.WeightG,
		CommissionRate: salesperson.CommissionRate,
		Points:         points,
		SettledAt:      *order.SettledAt,
	}
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建提成记录失败: %v", err)
	}

	// 10. 更新销售人员积分
	salesperson.AddPoints(points)
	if err := tx.Save(salesperson).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新销售积分失败: %v", err)
	}

	// 11. 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("事务提交失败: %v", err)
	}

	log.Printf("[Sales] ✅ 订单 %s 提成计算完成：销售=%s, 克重=%.2f, 积分=%.2f",
		order.OrderID, salesperson.Name, order.WeightG, points)

	return nil
}

/**
 * GetSalesRanking 获取销售排行榜
 * 
 * @param limit int - 查询数量限制
 * @param byMonth bool - 是否按本月排名
 * @return ([]*model.Salesperson, error)
 */
func (s *SalesService) GetSalesRanking(limit int, byMonth bool) ([]*model.Salesperson, error) {
	if byMonth {
		return s.salesRepo.FindTopByMonthPoints(limit)
	}
	return s.salesRepo.FindTopByTotalPoints(limit)
}

/**
 * GetSalesDashboard 获取销售看板数据
 * 
 * @param salespersonID uint - 销售人员ID
 * @return (map[string]interface{}, error)
 */
func (s *SalesService) GetSalesDashboard(salespersonID uint) (map[string]interface{}, error) {
	// 1. 查找销售人员
	salesperson, err := s.salesRepo.FindByID(salespersonID)
	if err != nil {
		return nil, errors.New("销售人员不存在")
	}

	// 2. 统计客户持仓订单
	customers, err := s.userRepo.FindBySalesID(salespersonID)
	if err != nil {
		return nil, err
	}

	var totalHoldingOrders int
	var totalHoldingWeight float64
	var totalHoldingDeposit float64

	for _, customer := range customers {
		orders, _ := s.orderRepo.FindHoldingOrdersByUserID(customer.ID)
		for _, order := range orders {
			totalHoldingOrders++
			totalHoldingWeight += order.WeightG
			totalHoldingDeposit += order.Deposit
		}
	}

	// 3. 统计本月提成
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)
	monthPoints, _ := s.commissionRepo.GetTotalPointsByDateRange(salespersonID, monthStart, monthEnd)

	// 4. 返回看板数据
	dashboard := map[string]interface{}{
		"salesperson_name":      salesperson.Name,
		"sales_code":            salesperson.SalesCode,
		"total_points":          salesperson.TotalPoints,
		"month_points":          monthPoints,
		"commission_rate":       salesperson.CommissionRate,
		"total_customers":       len(customers),
		"active_customers":      salesperson.ActiveCustomers,
		"holding_orders_count":  totalHoldingOrders,
		"holding_total_weight":  totalHoldingWeight,
		"holding_total_deposit": totalHoldingDeposit,
	}

	return dashboard, nil
}

/**
 * AssignCustomer 分配客户给销售
 * 
 * @param customerID uint - 客户ID
 * @param salespersonID uint - 销售人员ID
 * @return error
 */
func (s *SalesService) AssignCustomer(customerID, salespersonID uint) error {
	// 1. 验证销售人员
	salesperson, err := s.salesRepo.FindByID(salespersonID)
	if err != nil {
		return errors.New("销售人员不存在")
	}
	if !salesperson.IsActive {
		return errors.New("销售人员已离职")
	}

	// 2. 更新客户归属
	if err := s.userRepo.UpdateSalesID(customerID, salespersonID); err != nil {
		return fmt.Errorf("分配客户失败: %v", err)
	}

	log.Printf("[Sales] 客户 %d 已分配给销售 %s", customerID, salesperson.Name)
	return nil
}

/**
 * ResetMonthlyPoints 重置月度积分（定时任务）
 * 
 * @return error
 */
func (s *SalesService) ResetMonthlyPoints() error {
	if err := s.salesRepo.ResetAllMonthPoints(); err != nil {
		return fmt.Errorf("重置月度积分失败: %v", err)
	}
	log.Println("[Sales] ✅ 所有销售人员月度积分已重置")
	return nil
}
