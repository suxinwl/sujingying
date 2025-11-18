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
	"fmt"
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
 * 仅使用WebSocket实时数据，不使用任何模拟数据
 * 
 * @return (float64, error)
 */
func (s *QuoteService) GetCurrentPrice() (float64, error) {
	if s.quoteHub == nil {
		return 0, fmt.Errorf("WebSocket行情代理未初始化")
	}
	
	price, lastUpdate, valid := s.quoteHub.GetLatestPrice()
	
	if !valid {
		return 0, fmt.Errorf("WebSocket价格数据无效，请检查行情连接")
	}
	
	// 检查价格时效性（超过5分钟认为过期）
	if time.Since(lastUpdate) > 5*time.Minute {
		return 0, fmt.Errorf("WebSocket价格已过期 (最后更新: %v)", lastUpdate.Format("2006-01-02 15:04:05"))
	}
	
	// 价格合理性检查
	if price <= 0 {
		return 0, fmt.Errorf("WebSocket价格异常: %.2f", price)
	}
	
	return price, nil
}

/**
 * GetPriceInfo 获取价格详细信息
 * 
 * @return map[string]interface{}
 */
func (s *QuoteService) GetPriceInfo() map[string]interface{} {
	if s.quoteHub == nil {
		return map[string]interface{}{
			"error":  "WebSocket行情代理未初始化",
			"valid":  false,
			"source": "websocket",
		}
	}
	
	price, lastUpdate, valid := s.quoteHub.GetLatestPrice()
	
	info := map[string]interface{}{
		"price":       price,
		"last_update": lastUpdate.Format("2006-01-02 15:04:05"),
		"age_seconds": int(time.Since(lastUpdate).Seconds()),
		"source":      "websocket",
		"valid":       valid,
	}
	
	// 如果价格无效，添加错误信息
	if !valid {
		info["error"] = "价格数据无效"
	} else if time.Since(lastUpdate) > 5*time.Minute {
		info["error"] = "价格数据已过期"
		info["valid"] = false
	} else if price <= 0 {
		info["error"] = "价格数据异常"
		info["valid"] = false
	}
	
	return info
}
