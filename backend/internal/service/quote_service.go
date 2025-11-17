/**
 * 行情数据服务
 * 
 * 用途：
 * - 从外部API获取实时黄金价格
 * - 缓存行情数据
 * - 提供价格查询接口
 * 
 * 数据源：
 * - 主要：上海黄金交易所API
 * - 备用：第三方黄金价格API（如：金色数据、东方财富等）
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

/**
 * QuoteService 行情服务
 */
type QuoteService struct {
	// 当前价格缓存
	currentPrice float64
	
	// 最后更新时间
	lastUpdate time.Time
	
	// 互斥锁
	mutex sync.RWMutex
	
	// HTTP客户端
	httpClient *http.Client
	
	// API配置
	apiURL string
	apiKey string
}

/**
 * GoldPriceResponse 黄金价格API响应（示例结构）
 */
type GoldPriceResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Price     float64 `json:"price"`      // 当前价格（元/克）
		UpdatedAt string  `json:"updated_at"` // 更新时间
	} `json:"data"`
}

/**
 * NewQuoteService 创建行情服务实例
 */
func NewQuoteService() *QuoteService {
	return &QuoteService{
		currentPrice: 500.00, // 初始价格
		lastUpdate:   time.Now(),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		// TODO: 从配置文件读取
		apiURL: "https://api.example.com/gold/price", // 示例API
		apiKey: "",
	}
}

/**
 * GetCurrentPrice 获取当前黄金价格
 * 
 * @return (float64, error)
 */
func (s *QuoteService) GetCurrentPrice() (float64, error) {
	s.mutex.RLock()
	
	// 如果缓存有效（1分钟内），直接返回
	if time.Since(s.lastUpdate) < time.Minute {
		price := s.currentPrice
		s.mutex.RUnlock()
		return price, nil
	}
	s.mutex.RUnlock()
	
	// 缓存过期，重新获取
	return s.FetchAndUpdatePrice()
}

/**
 * FetchAndUpdatePrice 从API获取并更新价格
 * 
 * @return (float64, error)
 */
func (s *QuoteService) FetchAndUpdatePrice() (float64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// 双重检查，避免重复请求
	if time.Since(s.lastUpdate) < time.Minute {
		return s.currentPrice, nil
	}
	
	// 尝试多个数据源
	price, err := s.fetchFromPrimarySource()
	if err != nil {
		log.Printf("[Quote] ⚠️ 主数据源获取失败: %v，尝试备用源", err)
		price, err = s.fetchFromBackupSource()
		if err != nil {
			log.Printf("[Quote] ❌ 备用数据源也失败: %v，使用缓存价格", err)
			// 返回上次的价格
			return s.currentPrice, nil
		}
	}
	
	// 更新缓存
	s.currentPrice = price
	s.lastUpdate = time.Now()
	
	log.Printf("[Quote] ✅ 价格更新成功: %.2f 元/克", price)
	return price, nil
}

/**
 * fetchFromPrimarySource 从主数据源获取价格
 * 
 * 可接入：
 * 1. 上海黄金交易所API
 * 2. 中国黄金协会
 * 3. 新浪财经黄金接口
 * 
 * @return (float64, error)
 */
func (s *QuoteService) fetchFromPrimarySource() (float64, error) {
	// TODO: 接入真实API
	// 示例：调用上海黄金交易所API
	
	/*
	// 真实API调用示例：
	req, err := http.NewRequest("GET", s.apiURL, nil)
	if err != nil {
		return 0, err
	}
	
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API返回错误: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	
	var priceResp GoldPriceResponse
	if err := json.Unmarshal(body, &priceResp); err != nil {
		return 0, err
	}
	
	if priceResp.Code != 0 {
		return 0, fmt.Errorf("API错误: %s", priceResp.Msg)
	}
	
	return priceResp.Data.Price, nil
	*/
	
	// 当前返回模拟数据
	return 0, errors.New("主数据源未配置")
}

/**
 * fetchFromBackupSource 从备用数据源获取价格
 * 
 * 可接入：
 * 1. 东方财富网API
 * 2. 金色数据
 * 3. CoinGecko（国际金价）
 * 
 * @return (float64, error)
 */
func (s *QuoteService) fetchFromBackupSource() (float64, error) {
	// 示例：调用第三方API
	// TODO: 接入备用API
	
	// 方案1：使用新浪财经接口（免费）
	url := "https://hq.sinajs.cn/list=hf_GC" // 纽约黄金期货
	
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("新浪接口返回错误: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	
	// 解析新浪返回数据（格式：var hq_str_hf_GC="...";）
	// TODO: 实现真实解析
	
	log.Printf("[Quote] 新浪接口响应: %s", string(body))
	
	// 当前返回模拟价格（避免使用固定值）
	// 在真实场景中，这里应该解析API返回的数据
	return 0, errors.New("备用数据源解析待实现")
}

/**
 * SimulatePrice 模拟价格（用于开发测试）
 * 
 * 在真实API未接入前，生成模拟的价格波动
 * 
 * @return float64
 */
func (s *QuoteService) SimulatePrice() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// 基础价格：500元/克
	basePrice := 500.0
	
	// 添加随机波动 (-5 到 +5 元)
	// 使用时间戳作为随机种子
	timestamp := time.Now().Unix()
	variation := float64(timestamp%1000) / 100.0 - 5.0
	
	price := basePrice + variation
	
	// 更新缓存
	s.currentPrice = price
	s.lastUpdate = time.Now()
	
	return price
}

/**
 * StartPriceUpdater 启动价格自动更新器
 * 
 * 每隔一定时间自动获取最新价格
 * 
 * @param interval time.Duration - 更新间隔
 */
func (s *QuoteService) StartPriceUpdater(interval time.Duration) {
	ticker := time.NewTicker(interval)
	
	go func() {
		log.Printf("[Quote] 🚀 价格自动更新器已启动，间隔: %v", interval)
		
		for range ticker.C {
			// 尝试更新价格
			price, err := s.FetchAndUpdatePrice()
			if err != nil {
				log.Printf("[Quote] ⚠️ 自动更新价格失败: %v", err)
				// 使用模拟价格作为fallback
				price = s.SimulatePrice()
				log.Printf("[Quote] 使用模拟价格: %.2f 元/克", price)
			}
			
			log.Printf("[Quote] 📊 当前黄金价格: %.2f 元/克", price)
		}
	}()
}

/**
 * GetPriceInfo 获取价格详细信息
 * 
 * @return map[string]interface{}
 */
func (s *QuoteService) GetPriceInfo() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return map[string]interface{}{
		"price":       s.currentPrice,
		"last_update": s.lastUpdate.Format("2006-01-02 15:04:05"),
		"age_seconds": int(time.Since(s.lastUpdate).Seconds()),
	}
}
