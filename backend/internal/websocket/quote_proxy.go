/**
 * 行情WebSocket代理服务
 * 
 * 用途：
 * - 封装外部行情数据源 wss://push143.jtd9999.vip/ws
 * - 提供统一的WebSocket接口给前端
 * - 管理连接池和数据广播
 * - 支持自动重连和心跳保活
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

/**
 * QuoteProxyHub WebSocket代理中心
 * 管理所有客户端连接和上游数据源
 */
type QuoteProxyHub struct {
	clients         map[*Client]bool  // 已连接的客户端
	broadcast       chan []byte       // 广播消息通道
	register        chan *Client      // 注册新客户端
	unregister      chan *Client      // 注销客户端
	upstreamConn    *websocket.Conn   // 上游数据源连接
	mu              sync.RWMutex      // 读写锁
	isUpstreamAlive bool              // 上游连接状态
	
	// 最新价格缓存（用于风控系统）
	latestPrice     float64           // 最新Au9999价格（元/克）
	lastUpdate      time.Time         // 最后更新时间
	priceMutex      sync.RWMutex      // 价格锁
}

/**
 * Client WebSocket客户端
 */
type Client struct {
	hub  *QuoteProxyHub
	conn *websocket.Conn
	send chan []byte
}

/**
 * 上游WebSocket配置
 */
const (
	upstreamURL  = "wss://push143.jtd9999.vip/ws"
	dempCode     = "e2571ebfeb4c217b4f6adac7a1ef3d4d"
	secret       = "ceb1b5791048bb9ca438582b534d005b"
	
	// 客户端配置
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

/**
 * NewQuoteProxyHub 创建行情代理中心实例
 * 
 * @return *QuoteProxyHub
 */
func NewQuoteProxyHub() *QuoteProxyHub {
	return &QuoteProxyHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

/**
 * Run 启动代理中心
 * 
 * 功能：
 * 1. 连接上游数据源
 * 2. 处理客户端注册/注销
 * 3. 广播消息到所有客户端
 * 4. 维护上游连接
 * 
 * @return void
 */
func (h *QuoteProxyHub) Run() {
	// 连接上游数据源
	go h.connectUpstream()
	
	// 主事件循环
	for {
		select {
		case client := <-h.register:
			// 注册新客户端
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[QuoteProxy] 新客户端连接，当前连接数: %d", len(h.clients))
			
		case client := <-h.unregister:
			// 注销客户端
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[QuoteProxy] 客户端断开，当前连接数: %d", len(h.clients))
			
		case message := <-h.broadcast:
			// 广播消息到所有客户端
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// 发送失败，关闭客户端
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

/**
 * connectUpstream 连接上游数据源
 * 
 * 业务流程：
 * 1. 建立WebSocket连接
 * 2. 发送订阅消息
 * 3. 接收上游数据并广播
 * 4. 断线自动重连
 * 
 * @return void
 */
func (h *QuoteProxyHub) connectUpstream() {
	for {
		log.Printf("[QuoteProxy] 正在连接上游数据源: %s", upstreamURL)
		
		// 创建WebSocket连接
		conn, _, err := websocket.DefaultDialer.Dial(upstreamURL, nil)
		if err != nil {
			log.Printf("[QuoteProxy] 连接上游失败: %v，5秒后重试", err)
			time.Sleep(5 * time.Second)
			continue
		}
		
		h.mu.Lock()
		h.upstreamConn = conn
		h.isUpstreamAlive = true
		h.mu.Unlock()
		
		log.Println("[QuoteProxy] ✅ 上游连接成功")
		
		// 发送订阅消息
		h.sendUpstreamSubscribe()
		
		// 接收上游数据
		go h.readUpstream()
		
		// 等待连接断开
		<-time.After(time.Hour * 24)
	}
}

/**
 * sendUpstreamSubscribe 发送订阅消息到上游
 * 
 * @return void
 */
func (h *QuoteProxyHub) sendUpstreamSubscribe() {
	subscribeMsg := map[string]interface{}{
		"userid":           0,
		"dempCode":         dempCode,
		"channel":          "channel",
		"clientIp":         "127.0.0.1",
		"secret":           secret,
		"sessionId":        generateSessionID(),
		"subscriptionType": "all",
		"time":             time.Now().Format("2006-01-02 15:04:05"),
	}
	
	data, err := json.Marshal(subscribeMsg)
	if err != nil {
		log.Printf("[QuoteProxy] 序列化订阅消息失败: %v", err)
		return
	}
	
	h.mu.RLock()
	conn := h.upstreamConn
	h.mu.RUnlock()
	
	if conn != nil {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[QuoteProxy] 发送订阅消息失败: %v", err)
		} else {
			log.Println("[QuoteProxy] ✅ 订阅消息已发送")
		}
	}
}

/**
 * readUpstream 读取上游数据
 * 
 * @return void
 */
func (h *QuoteProxyHub) readUpstream() {
	h.mu.RLock()
	conn := h.upstreamConn
	h.mu.RUnlock()
	
	if conn == nil {
		return
	}
	
	defer func() {
		h.mu.Lock()
		h.isUpstreamAlive = false
		h.mu.Unlock()
		conn.Close()
	}()
	
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[QuoteProxy] 读取上游数据失败: %v", err)
			break
		}
		
		// 解析消息并提取Au9999价格
		h.extractPrice(message)
		
		// 广播到所有客户端
		h.broadcast <- message
	}
}

/**
 * extractPrice 从WebSocket消息中提取Au9999价格
 * 
 * 消息格式示例:
 * {
 *   "data": {
 *     "au9999": {
 *       "currentPrice": "500.12",
 *       ...
 *     }
 *   }
 * }
 * 
 * @param message []byte - WebSocket消息
 * @return void
 */
func (h *QuoteProxyHub) extractPrice(message []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		return
	}

	parsePrice := func(v interface{}) float64 {
		switch t := v.(type) {
		case string:
			var p float64
			if _, err := fmt.Sscanf(t, "%f", &p); err == nil && p > 0 {
				return p
			}
		case float64:
			if t > 0 {
				return t
			}
		}
		return 0
	}

	// 尝试提取Au9999价格
	// 根据实际数据格式调整解析逻辑
	if dataObj, ok := data["data"].(map[string]interface{}); ok {
		if au9999, ok := dataObj["au9999"].(map[string]interface{}); ok {
			if price := parsePrice(au9999["currentPrice"]); price > 0 {
				h.priceMutex.Lock()
				h.latestPrice = price
				h.lastUpdate = time.Now()
				h.priceMutex.Unlock()
				log.Printf("[QuoteProxy] 价格更新: Au9999 = %.2f 元/克", price)
				return
			}
		}
	}

	raw := data
	if content, ok := raw["content"].(string); ok {
		var inner map[string]interface{}
		if err := json.Unmarshal([]byte(content), &inner); err == nil {
			raw = inner
		}
	}

	if items, ok := raw["items"].(map[string]interface{}); ok {
		raw = items
	}

	var price float64
	if au, ok := raw["AU"].(map[string]interface{}); ok {
		if p := parsePrice(au["Sell"]); p > 0 {
			price = p
		}
	}

	if price == 0 {
		if au9999, ok := raw["AU9999"].(map[string]interface{}); ok {
			if p := parsePrice(au9999["Sell"]); p > 0 {
				price = p
			}
		}
	}

	if price == 0 {
		if xau, ok := raw["XAU"].(map[string]interface{}); ok {
			if p := parsePrice(xau["Sell"]); p > 0 {
				price = p
			}
		}
	}

	if price == 0 {
		for _, v := range raw {
			if m, ok := v.(map[string]interface{}); ok {
				if p := parsePrice(m["Sell"]); p > 0 {
					price = p
					break
				}
			}
		}
	}

	if price > 0 {
		h.priceMutex.Lock()
		h.latestPrice = price
		h.lastUpdate = time.Now()
		h.priceMutex.Unlock()
		log.Printf("[QuoteProxy] 价格更新: Sell = %.2f 元/克", price)
	}
}

/**
 * GetLatestPrice 获取最新Au9999价格
 * 
 * @return (float64, time.Time, bool) - 价格、更新时间、是否有效
 */
func (h *QuoteProxyHub) GetLatestPrice() (float64, time.Time, bool) {
	h.priceMutex.RLock()
	defer h.priceMutex.RUnlock()
	
	// 如果超过5分钟没更新，认为数据无效
	if time.Since(h.lastUpdate) > 5*time.Minute {
		return 0, h.lastUpdate, false
	}
	
	return h.latestPrice, h.lastUpdate, h.latestPrice > 0
}

/**
 * ServeWs 处理客户端WebSocket连接
 * 
 * @param conn *websocket.Conn - 客户端连接
 * @return void
 */
func (h *QuoteProxyHub) ServeWs(conn *websocket.Conn) {
	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}
	
	client.hub.register <- client
	
	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

/**
 * readPump 读取客户端消息
 * 
 * @return void
 */
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

/**
 * writePump 向客户端发送消息
 * 
 * @return void
 */
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

/**
 * generateSessionID 生成会话ID
 * 
 * @return string
 */
func generateSessionID() string {
	return time.Now().Format("20060102150405")
}
