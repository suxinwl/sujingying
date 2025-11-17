/**
 * 风控定时任务调度器
 * 
 * 用途：
 * - 定期执行风控检查
 * - 自动更新订单价格
 * - 触发强平和预警
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package scheduler

import (
	"log"
	"time"

	"suxin/internal/appctx"
	"suxin/internal/service"
)

/**
 * RiskScheduler 风控调度器
 */
type RiskScheduler struct {
	ctx         *appctx.AppContext
	riskService *service.RiskService
	ticker      *time.Ticker
	stopChan    chan bool
	interval    time.Duration
}

/**
 * NewRiskScheduler 创建风控调度器实例
 * 
 * @param ctx *appctx.AppContext - 应用上下文
 * @param intervalSeconds int - 检查间隔（秒）
 * @return *RiskScheduler
 */
func NewRiskScheduler(ctx *appctx.AppContext, intervalSeconds int) *RiskScheduler {
	return &RiskScheduler{
		ctx:         ctx,
		riskService: service.NewRiskService(ctx),
		stopChan:    make(chan bool),
		interval:    time.Duration(intervalSeconds) * time.Second,
	}
}

/**
 * getCurrentMarketPrice 获取当前市场价格
 * 
 * 说明：
 * - 当前版本使用固定价格进行测试
 * - 后续版本将从WebSocket行情数据中获取实时价格
 * - 可以从行情缓存、第三方API或配置文件读取
 * 
 * TODO: 集成实时行情数据源
 * 
 * @return float64 - 当前市场价格（元/克）
 */
func (s *RiskScheduler) getCurrentMarketPrice() float64 {
	// TODO: 从WebSocket行情数据或其他数据源获取实时价格
	// 当前使用测试价格
	return 500.00
}

/**
 * Start 启动风控调度器
 * 
 * 业务流程：
 * 1. 立即执行一次风控检查
 * 2. 启动定时器，按间隔周期性执行检查
 * 3. 监听停止信号
 * 
 * @return void
 */
func (s *RiskScheduler) Start() {
	log.Printf("[RiskScheduler] 🚀 风控调度器启动，检查间隔: %v", s.interval)

	// 立即执行一次检查
	s.runCheck()

	// 创建定时器
	s.ticker = time.NewTicker(s.interval)

	// 启动后台协程
	go func() {
		for {
			select {
			case <-s.ticker.C:
				// 定时执行风控检查
				s.runCheck()

			case <-s.stopChan:
				// 收到停止信号
				log.Println("[RiskScheduler] 收到停止信号，正在关闭...")
				s.ticker.Stop()
				return
			}
		}
	}()

	log.Println("[RiskScheduler] ✅ 风控调度器已启动")
}

/**
 * runCheck 执行风控检查
 * 
 * 说明：
 * - 捕获并记录所有异常，确保调度器不会因单次检查失败而停止
 * - 记录每次检查的开始和结束时间
 * 
 * @return void
 */
func (s *RiskScheduler) runCheck() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[RiskScheduler] ❌ 风控检查发生异常: %v", r)
		}
	}()

	startTime := time.Now()
	log.Println("[RiskScheduler] ⏰ 开始执行风控检查...")

	// 获取当前市场价格
	currentPrice := s.getCurrentMarketPrice()

	// 执行风控检查
	if err := s.riskService.RunRiskCheck(currentPrice); err != nil {
		log.Printf("[RiskScheduler] ❌ 风控检查失败: %v", err)
		return
	}

	elapsed := time.Since(startTime)
	log.Printf("[RiskScheduler] ✅ 风控检查完成，耗时: %v", elapsed)
}

/**
 * Stop 停止风控调度器
 * 
 * @return void
 */
func (s *RiskScheduler) Stop() {
	log.Println("[RiskScheduler] 正在停止风控调度器...")
	s.stopChan <- true
	close(s.stopChan)
	log.Println("[RiskScheduler] ✅ 风控调度器已停止")
}

/**
 * GetStatus 获取调度器状态
 * 
 * @return map[string]interface{}
 */
func (s *RiskScheduler) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"interval":      s.interval.String(),
		"running":       s.ticker != nil,
		"current_price": s.getCurrentMarketPrice(),
	}
}
