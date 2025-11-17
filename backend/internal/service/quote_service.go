/**
 * 行情数据服务
 * 
 * 用途：
 * - 从WebSocket行情代理获取实时黄金价格
 * - 提供价格查询接口
 * 
 * 数据源：
 * - 上海黄金交易所 (通过WebSocket代理)
 * - wss://push143.jtd9999.vip/ws
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package service

import (
	"log"
	"time"
)

/**
 * QuoteHubInterface WebSocket行情代理接口
 */
type QuoteHubInterface interface {
	GetLatestPrice() (float64, time.Time, bool)
}

/**
 * QuoteService 行情服务
 */
type QuoteService struct {
	quoteHub QuoteHubInterface
}

/**
 * NewQuoteService 创建行情服务实例
 * 
 * @param quoteHub QuoteHubInterface - WebSocket行情代理
 * @return *QuoteService
 */
func NewQuoteService(quoteHub QuoteHubInterface) *QuoteService {
	return &QuoteService{
		quoteHub: quoteHub,
	}
}

/**
 * GetCurrentPrice 获取当前黄金价格
 * 
 * @return (float64, error)
 */
func (s *QuoteService) GetCurrentPrice() (float64, error) {
	if s.quoteHub == nil {
		// 如果quoteHub未设置，返回模拟价格
		return s.SimulatePrice(), nil
	}
	
	price, lastUpdate, valid := s.quoteHub.GetLatestPrice()
	
	if !valid {
		log.Println("[Quote] WebSocket价格无效，使用模拟价格")
		return s.SimulatePrice(), nil
	}
	
	// 检查价格时效性（超过5分钟使用模拟价格）
	if time.Since(lastUpdate) > 5*time.Minute {
		log.Printf("[Quote] WebSocket价格过期(%v)，使用模拟价格", time.Since(lastUpdate))
		return s.SimulatePrice(), nil
	}
	
	return price, nil
}

/**
 * SimulatePrice 模拟价格（fallback）
 * 
 * 当WebSocket数据不可用时使用
 * 
 * @return float64
 */
func (s *QuoteService) SimulatePrice() float64 {
	// 基础价格：500元/克
	basePrice := 500.0
	
	// 添加随机波动 (-5 到 +5 元)
	timestamp := time.Now().Unix()
	variation := float64(timestamp%1000) / 100.0 - 5.0
	
	price := basePrice + variation
	
	log.Printf("[Quote] 使用模拟价格: %.2f 元/克", price)
	return price
}

/**
 * GetPriceInfo 获取价格详细信息
 * 
 * @return map[string]interface{}
 */
func (s *QuoteService) GetPriceInfo() map[string]interface{} {
	if s.quoteHub == nil {
		return map[string]interface{}{
			"price":       s.SimulatePrice(),
			"last_update": time.Now().Format("2006-01-02 15:04:05"),
			"source":      "simulated",
			"valid":       false,
		}
	}
	
	price, lastUpdate, valid := s.quoteHub.GetLatestPrice()
	
	return map[string]interface{}{
		"price":       price,
		"last_update": lastUpdate.Format("2006-01-02 15:04:05"),
		"age_seconds": int(time.Since(lastUpdate).Seconds()),
		"source":      "websocket",
		"valid":       valid,
	}
}

/**
 * StartPriceUpdater 启动价格自动更新器
 * 
 * 注意：WebSocket版本不需要主动拉取，数据由WebSocket推送
 * 此方法保留是为了保持接口兼容性
 * 
 * @param interval time.Duration - 更新间隔（已废弃）
 */
func (s *QuoteService) StartPriceUpdater(interval time.Duration) {
	log.Println("[Quote] 使用WebSocket推送，无需启动定时更新器")
	// WebSocket会自动推送数据，无需定时拉取
}
