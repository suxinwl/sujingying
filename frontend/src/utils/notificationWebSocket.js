/**
 * 通知 WebSocket 客户端
 *
 * 用途：
 * - 连接後端 /ws/notification
 * - 自动重连
 * - 提供 onMessage 订阅接口
 */

import { WS_BASE_URL, API_ENDPOINTS } from '@/config/api'

class NotificationWebSocket {
  constructor() {
    this.ws = null
    this.reconnectAttempts = 0
    this.messageHandlers = []
    this.isManualClose = false
  }

  get url() {
    const token = localStorage.getItem('access_token') || ''
    const base = WS_BASE_URL.replace(/^http/, 'ws')
    const path = API_ENDPOINTS.WS_NOTIFICATION
    const sep = path.includes('?') ? '&' : '?'
    return `${base}${path}${sep}token=${encodeURIComponent(token)}`
  }

  connect() {
    try {
      const url = this.url
      if (!url.endsWith('token=')) {
        // 只有在有 token 时才连接
        this.ws = new WebSocket(url)
      } else {
        return
      }

      this.ws.onopen = () => {
        this.reconnectAttempts = 0
        this.isManualClose = false
      }

      this.ws.onmessage = (event) => {
        this.handleMessage(event)
      }

      this.ws.onerror = () => {
        // 忽略，重连逻辑在 onclose
      }

      this.ws.onclose = () => {
        if (!this.isManualClose) {
          this.handleReconnect()
        }
      }
    } catch (e) {
      this.handleReconnect()
    }
  }

  handleMessage(event) {
    try {
      const data = JSON.parse(event.data)
      this.messageHandlers.forEach((handler) => {
        try {
          handler(data)
        } catch (err) {
          console.error('[NotifyWS] handler error:', err)
        }
      })
    } catch (err) {
      console.error('[NotifyWS] parse error:', err)
    }
  }

  handleReconnect() {
    if (this.reconnectAttempts >= 10) return
    this.reconnectAttempts++
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000)
    setTimeout(() => {
      this.connect()
    }, delay)
  }

  onMessage(handler) {
    if (typeof handler === 'function') {
      this.messageHandlers.push(handler)
    }
  }

  offMessage(handler) {
    const idx = this.messageHandlers.indexOf(handler)
    if (idx > -1) this.messageHandlers.splice(idx, 1)
  }

  disconnect() {
    if (this.ws) {
      this.isManualClose = true
      this.ws.close()
      this.ws = null
    }
  }
}

export const notifyWS = new NotificationWebSocket()
