/**
 * WebSocket 配置文件
 * 
 * 用途：定义行情WebSocket连接参数和商品显示顺序
 * 作者：速金盈技术团队
 * 日期：2025-11
 */

/**
 * WebSocket 连接配置
 * @type {Object}
 * @property {string} QUOTE_WS_URL - 行情数据WebSocket服务器地址
 * @property {number} MAX_RECONNECT_ATTEMPTS - 最大重连次数
 * @property {number} RECONNECT_BASE_DELAY - 重连基础延迟（毫秒）
 * @property {number} MAX_RECONNECT_DELAY - 最大重连延迟（毫秒）
 */
export const WS_CONFIG = {
  // 后端WebSocket代理接口（已封装外部数据源）
  QUOTE_WS_URL: import.meta.env.DEV 
    ? 'ws://localhost:8080/ws/quote'  // 开发环境
    : 'wss://your-domain.com/ws/quote', // 生产环境
  
  // 重连配置
  MAX_RECONNECT_ATTEMPTS: 10,
  RECONNECT_BASE_DELAY: 1000,
  MAX_RECONNECT_DELAY: 30000,
  
  /**
   * 商品显示顺序配置
   * 定义三组行情数据的商品代码和显示名称
   */
  PRODUCT_ORDER: {
    '现货行情': [
      { code: 'AU', name: '黄金' },
      { code: 'BULLION', name: '金条' },
      { code: 'PT', name: '铂金' },
      { code: 'AG', name: '白银' },
      { code: 'PD', name: '钯金' }
    ],
    '国内行情': [
      { code: 'AU9999', name: '黄金9999' },
      { code: 'AUTD', name: '黄金T+D' },
      { code: 'AGTD', name: '白银T+D' }
    ],
    '国际行情': [
      { code: 'XAU', name: '伦敦金' },
      { code: 'USDCNH', name: '美元' }
    ]
  }
}
