import { defineStore } from 'pinia'
import { WS_BASE_URL, API_ENDPOINTS } from '../config/api'

export const useQuoteStore = defineStore('quote', {
  state: () => ({
    ws: null,
    isConnected: false,
    quoteData: null,
    currentPrice: 0,
    priceChange: 0,
    priceChangePercent: 0
  }),
  
  getters: {
    priceDisplay: (state) => {
      return state.currentPrice ? state.currentPrice.toFixed(2) : '-.--'
    },
    isUp: (state) => state.priceChange > 0,
    isDown: (state) => state.priceChange < 0
  },
  
  actions: {
    // 连接WebSocket
    connectWebSocket() {
      if (this.ws && this.isConnected) {
        return
      }
      
      const token = localStorage.getItem('access_token')
      const wsUrl = `${WS_BASE_URL}${API_ENDPOINTS.WS_QUOTE}`
      
      this.ws = new WebSocket(wsUrl)
      
      this.ws.onopen = () => {
        console.log('行情WebSocket已连接')
        this.isConnected = true
      }
      
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this.updateQuote(data)
        } catch (error) {
          console.error('解析行情数据失败:', error)
        }
      }
      
      this.ws.onerror = (error) => {
        console.error('行情WebSocket错误:', error)
        this.isConnected = false
      }
      
      this.ws.onclose = () => {
        console.log('行情WebSocket已断开')
        this.isConnected = false
        
        // 5秒后重连
        setTimeout(() => {
          this.connectWebSocket()
        }, 5000)
      }
    },
    
    // 更新行情数据
    updateQuote(data) {
      const oldPrice = this.currentPrice
      
      // 根据实际数据格式解析
      if (data.data && data.data.au9999) {
        const au9999 = data.data.au9999
        this.currentPrice = parseFloat(au9999.currentPrice) || 0
        this.quoteData = au9999
      }
      
      // 计算涨跌
      if (oldPrice > 0) {
        this.priceChange = this.currentPrice - oldPrice
        this.priceChangePercent = (this.priceChange / oldPrice) * 100
      }
    },
    
    // 断开WebSocket
    disconnectWebSocket() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
        this.isConnected = false
      }
    }
  }
})
