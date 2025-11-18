// API配置
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
export const WS_BASE_URL = import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8080'

// API端点
export const API_ENDPOINTS = {
  // 认证相关
  LOGIN: '/api/v1/auth/login',
  REGISTER: '/api/v1/auth/register',
  REFRESH_TOKEN: '/api/v1/auth/refresh',
  LOGOUT: '/api/v1/auth/logout',
  
  // 用户相关
  USER_PROFILE: '/api/v1/users/profile',
  USER_UPDATE: '/api/v1/users/profile',
  CHANGE_PASSWORD: '/api/v1/users/password',
  
  // 订单相关
  ORDERS: '/api/v1/orders',
  ORDER_BUY: '/api/v1/orders/buy',
  ORDER_SELL: '/api/v1/orders/sell',
  ORDER_DETAIL: '/api/v1/orders/:id',
  ORDER_CANCEL: '/api/v1/orders/:id/cancel',
  
  // 持仓相关
  POSITIONS: '/api/v1/positions',
  POSITION_DETAIL: '/api/v1/positions/:id',
  
  // 资金相关
  DEPOSITS: '/api/v1/deposits',
  DEPOSIT_CREATE: '/api/v1/deposits',
  WITHDRAWS: '/api/v1/withdraws',
  WITHDRAW_CREATE: '/api/v1/withdraws',
  FUND_FLOW: '/api/v1/fund-flow',
  
  // 银行卡相关
  BANK_CARDS: '/api/v1/bank-cards',
  BANK_CARD_CREATE: '/api/v1/bank-cards',
  BANK_CARD_DELETE: '/api/v1/bank-cards/:id',
  
  // 通知相关
  NOTIFICATIONS: '/api/v1/notifications',
  NOTIFICATION_READ: '/api/v1/notifications/:id/read',
  NOTIFICATION_READ_ALL: '/api/v1/notifications/read-all',
  
  // 销售相关
  SALES_INVITE_CODES: '/api/v1/sales/invite-codes',
  SALES_INVITE_CODE_CREATE: '/api/v1/sales/invite-codes',
  SALES_CUSTOMERS: '/api/v1/sales/customers',
  SALES_COMMISSIONS: '/api/v1/sales/commissions',
  
  // 管理员相关
  ADMIN_USERS: '/api/v1/admin/users',
  ADMIN_USER_DETAIL: '/api/v1/admin/users/:id',
  ADMIN_USER_APPROVE: '/api/v1/admin/users/:id/approve',
  ADMIN_DEPOSITS_PENDING: '/api/v1/deposits/pending',
  ADMIN_DEPOSIT_REVIEW: '/api/v1/deposits/:id/review',
  ADMIN_WITHDRAWS_PENDING: '/api/v1/withdraws/pending',
  ADMIN_WITHDRAW_REVIEW: '/api/v1/withdraws/:id/review',
  
  // 配置相关
  CONFIG: '/api/v1/config',
  CONFIG_UPDATE: '/api/v1/config',
  
  // WebSocket相关
  WS_QUOTE: '/ws/quote',
  WS_NOTIFICATION: '/ws/notification'
}

// 替换URL中的参数
export function replaceParams(url, params) {
  let result = url
  Object.keys(params).forEach(key => {
    result = result.replace(`:${key}`, params[key])
  })
  return result
}
