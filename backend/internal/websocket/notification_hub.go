/**
 * WebSocket通知推送中心
 * 
 * 用途：
 * - 管理用户WebSocket连接
 * - 推送实时通知
 * - 用户在线状态管理
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"suxin/internal/model"
)

/**
 * NotificationClient 通知客户端连接
 */
type NotificationClient struct {
	UserID uint
	Conn   *websocket.Conn
	Hub    *NotificationHub
	Send   chan *model.Notification
}

/**
 * NotificationHub 通知推送中心
 */
type NotificationHub struct {
	// 已注册的客户端 map[userID]map[*NotificationClient]bool
	clients map[uint]map[*NotificationClient]bool
	
	// 注册请求
	register chan *NotificationClient
	
	// 注销请求
	unregister chan *NotificationClient
	
	// 广播消息
	broadcast chan *NotificationMessage
	
	// 互斥锁
	mutex sync.RWMutex
}

/**
 * NotificationMessage 通知消息
 */
type NotificationMessage struct {
	UserID       uint
	Notification *model.Notification
}

/**
 * NewNotificationHub 创建通知中心
 */
func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		clients:    make(map[uint]map[*NotificationClient]bool),
		register:   make(chan *NotificationClient),
		unregister: make(chan *NotificationClient),
		broadcast:  make(chan *NotificationMessage, 256),
	}
}

/**
 * Run 运行通知中心
 */
func (h *NotificationHub) Run() {
	log.Println("[NotificationHub] ✅ 通知推送中心已启动")
	
	for {
		select {
		case client := <-h.register:
			// 注册新客户端
			h.mutex.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*NotificationClient]bool)
			}
			h.clients[client.UserID][client] = true
			h.mutex.Unlock()
			
			log.Printf("[NotificationHub] 用户 %d 已连接，当前在线设备: %d", 
				client.UserID, len(h.clients[client.UserID]))

		case client := <-h.unregister:
			// 注销客户端
			h.mutex.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					
					// 如果用户没有其他设备在线，删除用户
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
					
					log.Printf("[NotificationHub] 用户 %d 已断开连接，剩余在线设备: %d", 
						client.UserID, len(h.clients[client.UserID]))
				}
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			// 广播消息
			h.mutex.RLock()
			if message.UserID == 0 {
				for _, clients := range h.clients {
					for client := range clients {
						select {
						case client.Send <- message.Notification:
						default:
							close(client.Send)
							delete(clients, client)
						}
					}
				}
			} else {
				if clients, ok := h.clients[message.UserID]; ok {
					for client := range clients {
						select {
						case client.Send <- message.Notification:
						default:
							close(client.Send)
							delete(clients, client)
						}
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

/**
 * SendToUser 发送通知给指定用户
 */
func (h *NotificationHub) SendToUser(userID uint, notification *model.Notification) {
	message := &NotificationMessage{
		UserID:       userID,
		Notification: notification,
	}
	
	select {
	case h.broadcast <- message:
		log.Printf("[NotificationHub] 通知已发送到用户 %d: %s", userID, notification.Title)
	default:
		log.Printf("[NotificationHub] ⚠️ 通知队列已满，丢弃消息: 用户=%d", userID)
	}
}

/**
 * IsUserOnline 检查用户是否在线
 */
func (h *NotificationHub) IsUserOnline(userID uint) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

/**
 * GetOnlineUserCount 获取在线用户数
 */
func (h *NotificationHub) GetOnlineUserCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return len(h.clients)
}

/**
 * GetUserDeviceCount 获取用户在线设备数
 */
func (h *NotificationHub) GetUserDeviceCount(userID uint) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	if clients, ok := h.clients[userID]; ok {
		return len(clients)
	}
	return 0
}

/**
 * ServeWs 处理WebSocket连接
 */
func (h *NotificationHub) ServeWs(userID uint, conn *websocket.Conn) {
	client := &NotificationClient{
		UserID: userID,
		Conn:   conn,
		Hub:    h,
		Send:   make(chan *model.Notification, 256),
	}
	
	// 注册客户端
	h.register <- client
	
	// 启动读写协程
	go client.readPump()
	go client.writePump()
}

/**
 * readPump 读取客户端消息（心跳检测）
 */
func (c *NotificationClient) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	
	// 设置读取超时
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[NotificationHub] WebSocket错误: %v", err)
			}
			break
		}
	}
}

/**
 * writePump 向客户端发送消息
 */
func (c *NotificationClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case notification, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			// 发送JSON消息
			data, err := json.Marshal(notification)
			if err != nil {
				log.Printf("[NotificationHub] JSON序列化失败: %v", err)
				continue
			}
			
			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
			
		case <-ticker.C:
			// 发送心跳
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
