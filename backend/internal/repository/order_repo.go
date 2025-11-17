/**
 * 订单仓储层
 * 
 * 用途：
 * - 封装订单数据访问逻辑
 * - 提供订单CRUD操作
 * - 支持按条件查询订单
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package repository

import (
	"gorm.io/gorm"

	"suxin/internal/model"
)

/**
 * OrderRepository 订单仓储
 */
type OrderRepository struct {
	db *gorm.DB
}

/**
 * NewOrderRepository 创建订单仓储实例
 * 
 * @param db *gorm.DB - 数据库连接
 * @return *OrderRepository
 */
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

/**
 * Create 创建订单
 * 
 * @param order *model.Order - 订单实体
 * @return error
 */
func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

/**
 * FindByID 根据ID查找订单
 * 
 * @param id uint - 订单ID
 * @return (*model.Order, error)
 */
func (r *OrderRepository) FindByID(id uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

/**
 * FindByOrderID 根据订单号查找订单
 * 
 * @param orderID string - 订单号
 * @return (*model.Order, error)
 */
func (r *OrderRepository) FindByOrderID(orderID string) (*model.Order, error) {
	var order model.Order
	if err := r.db.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

/**
 * FindByUserID 查询用户的所有订单
 * 
 * @param userID uint - 用户ID
 * @param status string - 订单状态（可选，传空字符串表示查询所有状态）
 * @return ([]*model.Order, error)
 */
func (r *OrderRepository) FindByUserID(userID uint, status string) ([]*model.Order, error) {
	var orders []*model.Order
	query := r.db.Where("user_id = ?", userID)
	
	// 如果指定了状态，则过滤状态
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

/**
 * FindHoldingOrders 查询所有持仓订单
 * 
 * 用途：用于风控引擎批量计算定金率
 * 
 * @return ([]*model.Order, error)
 */
func (r *OrderRepository) FindHoldingOrders() ([]*model.Order, error) {
	var orders []*model.Order
	if err := r.db.Where("status = ?", model.OrderStatusHolding).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

/**
 * FindHoldingOrdersByUserID 查询用户的持仓订单
 * 
 * @param userID uint - 用户ID
 * @return ([]*model.Order, error)
 */
func (r *OrderRepository) FindHoldingOrdersByUserID(userID uint) ([]*model.Order, error) {
	var orders []*model.Order
	if err := r.db.Where("user_id = ? AND status = ?", userID, model.OrderStatusHolding).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

/**
 * Update 更新订单
 * 
 * @param order *model.Order - 订单实体
 * @return error
 */
func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

/**
 * UpdatePnLAndMargin 批量更新订单的盈亏和定金率
 * 
 * 用途：风控引擎定时任务更新所有持仓订单的实时数据
 * 
 * @param order *model.Order - 订单实体
 * @return error
 */
func (r *OrderRepository) UpdatePnLAndMargin(order *model.Order) error {
	// 只更新特定字段，避免覆盖其他字段
	return r.db.Model(order).Updates(map[string]interface{}{
		"current_price": order.CurrentPrice,
		"pnl_float":     order.PnLFloat,
		"margin_rate":   order.MarginRate,
	}).Error
}

/**
 * GetTotalWeightByCustomer 获取客户已结算的总克重
 * 
 * 用途：销售提成计算（积分 = G × 提成点数）
 * 
 * @param customerID uint - 客户ID
 * @return (float64, error) - 总克重
 */
func (r *OrderRepository) GetTotalWeightByCustomer(customerID uint) (float64, error) {
	var total float64
	err := r.db.Model(&model.Order{}).
		Where("user_id = ? AND status = ?", customerID, model.OrderStatusSettled).
		Select("COALESCE(SUM(weight_g), 0)").
		Scan(&total).Error
	return total, err
}
