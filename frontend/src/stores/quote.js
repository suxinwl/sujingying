import { defineStore } from 'pinia'
import { WS_CONFIG } from '../config/websocket'

export const useQuoteStore = defineStore('quote', {
  state: () => ({
    ws: null,
    isConnected: false,
    quoteData: {},
    buyPrice: 0,  // 销售价（用户买入价）
    sellPrice: 0,  // 回购价（用户卖出价）
    previousBuyPrice: 0,
    priceChange: 0,
    priceChangePercent: 0
  }),
  
  getters: {
    // 销售价显示
    buyPriceDisplay: (state) => {
      return state.buyPrice ? state.buyPrice.toFixed(2) : '-.--'
    },
    // 回购价显示
    sellPriceDisplay: (state) => {
      return state.sellPrice ? state.sellPrice.toFixed(2) : '-.--'
    },
    // 当前价格（用于兼容）
    currentPrice: (state) => state.buyPrice,
    priceDisplay: (state) => {
      return state.buyPrice ? state.buyPrice.toFixed(2) : '-.--'
    },
    isUp: (state) => state.priceChange > 0,
    isDown: (state) => state.priceChange < 0
  },
  
  actions: {
    // 连接WebSocket
    connectWebSocket() {
      if (this.ws && this.isConnected) {
        console.log('WebSocket已连接，跳过重复连接')
        return
      }
      
      console.log('正在连接行情WebSocket:', WS_CONFIG.QUOTE_WS_URL)
      this.ws = new WebSocket(WS_CONFIG.QUOTE_WS_URL)
      
      this.ws.onopen = () => {
        console.log('行情WebSocket已连接')
        this.isConnected = true
      }
      
      this.ws.onmessage = (event) => {
        try {
          let data = JSON.parse(event.data)
          
          // 检查是否有content字段（需要二次解析）
          if (data.content && typeof data.content === 'string') {
            console.log('检测到嵌套JSON，进行二次解析...')
            data = JSON.parse(data.content)
          }
          
          // 检查是否有items字段
          if (data.items) {
            console.log('从items中提取行情数据')
            data = data.items
          }
          
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
          if (!this.isConnected) {
            console.log('尝试重新连接WebSocket...')
            this.connectWebSocket()
          }
        }, 5000)
      }
    },
    
    // 更新行情数据
    updateQuote(data) {
      // 保存所有行情数据
      this.quoteData = data
      
      // 优先使用黄金AU的价格
      let newBuyPrice = 0   // 销售价（Sell - 用户买入）
      let newSellPrice = 0  // 回购价（Buy - 用户卖出）
      let source = ''       // 数据来源
      
      // 按优先级尝试不同数据源
      if (data.AU && data.AU.Sell) {
        newBuyPrice = parseFloat(data.AU.Sell) || 0
        newSellPrice = parseFloat(data.AU.Buy) || 0
        source = 'AU'
      } else if (data.AU9999 && data.AU9999.Sell) {
        newBuyPrice = parseFloat(data.AU9999.Sell) || 0
        newSellPrice = parseFloat(data.AU9999.Buy) || 0
        source = 'AU9999'
      } else if (data.XAU && data.XAU.Sell) {
        newBuyPrice = parseFloat(data.XAU.Sell) || 0
        newSellPrice = parseFloat(data.XAU.Buy) || 0
        source = 'XAU'
      } else if (typeof data === 'object') {
        // 尝试找到任何包含价格的商品
        for (const key in data) {
          if (data[key] && typeof data[key] === 'object' && data[key].Sell) {
            newBuyPrice = parseFloat(data[key].Sell) || 0
            newSellPrice = parseFloat(data[key].Buy) || 0
            source = key
            break
          }
        }
      }
      
      if (newBuyPrice > 0) {
        this.previousBuyPrice = this.buyPrice
        this.buyPrice = newBuyPrice
        this.sellPrice = newSellPrice
        
        // 计算涨跌
        if (this.previousBuyPrice > 0) {
          this.priceChange = this.buyPrice - this.previousBuyPrice
          this.priceChangePercent = (this.priceChange / this.previousBuyPrice) * 100
        }
        
        // 只在首次更新或价格有变化时输出日志
        if (this.priceChange !== 0 || this.previousBuyPrice === 0) {
          console.log(`[${source}] 销售价: ${this.buyPrice.toFixed(2)} | 回购价: ${this.sellPrice.toFixed(2)}`)
        }
      } else {
        console.warn('未找到有效的价格数据')
      }
    },
    
    // 断开WebSocket
    disconnectWebSocket() {
      if (this.ws) {
        console.log('断开行情WebSocket连接')
        this.ws.close()
        this.ws = null
        this.isConnected = false
      }
    }
  }
})
