/**
 * 行情WebSocket客户端
 * 
 * 用途：
 * - 连接到真实行情数据源 wss://push143.jtd9999.vip/ws
 * - 实现自动重连机制（指数退避策略）
 * - 解析行情数据并触发回调
 * - 处理心跳包和异常情况
 * 
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

import { WS_CONFIG } from '@/config/websocket'

/**
 * 行情WebSocket客户端类
 * 单例模式，全局共享一个连接
 */
class QuoteWebSocket {
  constructor() {
    this.ws = null                    // WebSocket实例
    this.reconnectAttempts = 0        // 当前重连次数
    this.messageHandlers = []         // 消息处理器列表
    this.isManualClose = false        // 是否手动关闭
  }


  /**
   * 连接WebSocket服务器
   * 
   * 流程：
   * 1. 创建WebSocket实例
   * 2. 注册事件监听器（open/message/error/close）
   * 3. 连接成功后发送订阅消息
   * 4. 连接失败自动重连
   * 
   * @returns {void}
   */
  connect() {
    try {
      console.log(`[QuoteWS] 正在连接到 ${WS_CONFIG.QUOTE_WS_URL}`)
      
      // 创建WebSocket连接
      this.ws = new WebSocket(WS_CONFIG.QUOTE_WS_URL)
      
      // 连接成功处理
      this.ws.onopen = () => {
        console.log('[QuoteWS] ✅ 连接成功到后端代理')
        this.reconnectAttempts = 0
        this.isManualClose = false
      }
      
      // 接收消息处理
      this.ws.onmessage = (event) => {
        this.handleMessage(event)
      }
      
      // 错误处理
      this.ws.onerror = (error) => {
        console.error('[QuoteWS] ❌ 连接错误:', error)
      }
      
      // 连接关闭处理
      this.ws.onclose = () => {
        console.log('[QuoteWS] 🔌 连接已关闭')
        
        // 非手动关闭时自动重连
        if (!this.isManualClose) {
          this.handleReconnect()
        }
      }
      
    } catch (error) {
      console.error('[QuoteWS] 连接异常:', error)
      this.handleReconnect()
    }
  }


  /**
   * 处理接收到的消息
   * 
   * 消息类型：
   * - type: 'p' - 心跳包，直接忽略
   * - type: 'messageevent' - 行情数据，解析后触发回调
   * 
   * @param {MessageEvent} event - WebSocket消息事件
   * @returns {void}
   */
  handleMessage(event) {
    try {
      const data = JSON.parse(event.data)
      
      // 处理心跳包
      if (data.type === 'p') {
        return
      }
      
      // 处理行情数据
      if (data.type === 'messageevent' && data.content) {
        const content = JSON.parse(data.content)
        
        if (content.items) {
          // 触发所有注册的消息处理器
          this.messageHandlers.forEach(handler => {
            try {
              handler(content.items)
            } catch (err) {
              console.error('[QuoteWS] 消息处理器执行错误:', err)
            }
          })
        }
      }
    } catch (error) {
      console.error('[QuoteWS] 消息解析错误:', error)
    }
  }

  /**
   * 处理重连逻辑
   * 使用指数退避策略，避免频繁重连
   * 
   * 重连延迟计算：min(1000 * 2^n, 30000) 毫秒
   * 其中 n 为当前重连次数
   * 
   * @returns {void}
   */
  handleReconnect() {
    // 检查是否超过最大重连次数
    if (this.reconnectAttempts >= WS_CONFIG.MAX_RECONNECT_ATTEMPTS) {
      console.error('[QuoteWS] ❌ 已达到最大重连次数，停止重连')
      return
    }
    
    // 计算重连延迟（指数退避）
    this.reconnectAttempts++
    const delay = Math.min(
      WS_CONFIG.RECONNECT_BASE_DELAY * Math.pow(2, this.reconnectAttempts),
      WS_CONFIG.MAX_RECONNECT_DELAY
    )
    
    console.log(`[QuoteWS] ⏰ ${delay / 1000}秒后重连 (尝试 ${this.reconnectAttempts}/${WS_CONFIG.MAX_RECONNECT_ATTEMPTS})`)
    
    // 延迟后重连
    setTimeout(() => {
      this.connect()
    }, delay)
  }

  /**
   * 注册消息处理器
   * 支持多个处理器同时监听消息
   * 
   * @param {Function} handler - 消息处理回调函数
   *   @param {Object} items - 行情数据对象
   * @returns {void}
   */
  onMessage(handler) {
    if (typeof handler === 'function') {
      this.messageHandlers.push(handler)
    }
  }

  /**
   * 移除消息处理器
   * 
   * @param {Function} handler - 要移除的处理器
   * @returns {void}
   */
  offMessage(handler) {
    const index = this.messageHandlers.indexOf(handler)
    if (index > -1) {
      this.messageHandlers.splice(index, 1)
    }
  }

  /**
   * 手动断开连接
   * 设置手动关闭标志，防止自动重连
   * 
   * @returns {void}
   */
  disconnect() {
    if (this.ws) {
      this.isManualClose = true
      this.ws.close()
      this.ws = null
      console.log('[QuoteWS] 手动断开连接')
    }
  }

  /**
   * 获取连接状态
   * 
   * @returns {boolean} true-已连接，false-未连接
   */
  isConnected() {
    return this.ws && this.ws.readyState === WebSocket.OPEN
  }
}

// 导出单例实例
export const quoteWS = new QuoteWebSocket()
