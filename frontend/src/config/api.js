// API配置
// 自动检测当前访问的域名，设置对应的API地址
function getApiBaseUrl() {
  // 优先使用环境变量
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL
  }
  
  // 自动检测当前访问的域名
  const hostname = window.location.hostname
  const protocol = window.location.protocol
  
  // 如果是公网IP或域名，使用对应的后端地址
  if (hostname === '59.36.165.33') {
    return `${protocol}//${hostname}:8090`
  }
  
  // 如果是localhost或127.0.0.1，使用本地后端
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'http://localhost:8090'
  }
  
  // 默认使用当前域名，端口改为8090
  return `${protocol}//${hostname}:8090`
}

function getWsBaseUrl() {
  // 优先使用环境变量
  if (import.meta.env.VITE_WS_BASE_URL) {
    return import.meta.env.VITE_WS_BASE_URL
  }
  
  // 自动检测当前访问的域名
  const hostname = window.location.hostname
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  
  // 如果是公网IP或域名，使用对应的后端地址
  if (hostname === '59.36.165.33') {
    return `${protocol}//${hostname}:8090`
  }
  
  // 如果是localhost或127.0.0.1，使用本地后端
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'ws://localhost:8090'
  }
  
  // 默认使用当前域名，端口改为8090
  return `${protocol}//${hostname}:8090`
}

export const API_BASE_URL = getApiBaseUrl()
export const WS_BASE_URL = getWsBaseUrl()

// API端点
export const API_ENDPOINTS = {
  // 认证相关
  LOGIN: '/api/v1/auth/login',
  REGISTER: '/api/v1/auth/register',
  REFRESH_TOKEN: '/api/v1/auth/refresh',
  LOGOUT: '/api/v1/auth/logout',
  
  // 用户相关
  USER_PROFILE: '/api/v1/user/profile',
  USER_UPDATE: '/api/v1/user/profile',
  CHANGE_PASSWORD: '/api/v1/user/password',
  PAYPASS: '/api/v1/user/paypass',
  USER_VERIFICATION: '/api/v1/user/verification',
  
  // 订单相关
  ORDERS: '/api/v1/orders',
  ORDER_CREATE: '/api/v1/orders',
  ORDER_DETAIL: '/api/v1/orders/:id',
  ORDER_SETTLE: '/api/v1/orders/:id/settle',
  
  // 持仓相关
  POSITIONS: '/api/v1/positions',
  POSITION_DETAIL: '/api/v1/positions/:id',
  
  // 资金相关
  DEPOSITS: '/api/v1/deposits',
  DEPOSIT_CREATE: '/api/v1/deposits',
  SUPPLEMENTS: '/api/v1/supplements',
  WITHDRAWS: '/api/v1/withdraws',
  WITHDRAW_CREATE: '/api/v1/withdraws',
  FUND_FLOW: '/api/v1/fund-logs',
  
  // 银行卡相关
  BANK_CARDS: '/api/v1/bank-cards',
  BANK_CARD_CREATE: '/api/v1/bank-cards',
  BANK_CARD_DELETE: '/api/v1/bank-cards/:id',
  
  // 通知相关
  NOTIFICATIONS: '/api/v1/notifications',
  NOTIFICATIONS_UNREAD: '/api/v1/notifications/unread',
  NOTIFICATIONS_COUNT: '/api/v1/notifications/count',
  NOTIFICATION_READ: '/api/v1/notifications/read',
  NOTIFICATION_READ_ALL: '/api/v1/notifications/read-all',
  
  // 销售相关
  SALES_INVITE_CODES: '/api/v1/sales/invite-codes',
  SALES_INVITE_CODE_CREATE: '/api/v1/sales/invite-codes',
  SALES_CUSTOMERS: '/api/v1/sales/customers',
  SALES_COMMISSIONS: '/api/v1/sales/commissions',

  // 邀请相关
  INVITATION_MY_CODE: '/api/v1/invitation/my-code',
  
  // 管理员相关
  ADMIN_USERS: '/api/v1/users',
  ADMIN_USER_DETAIL: '/api/v1/users/:id',
  ADMIN_USER_APPROVE: '/api/v1/users/:id/approve',
  ADMIN_USERS_PENDING: '/api/v1/users/pending',
  ADMIN_USER_TOGGLE_AUTO_SUPPLEMENT: '/api/v1/users/:id/toggle-auto-supplement',
  ADMIN_USER_VERIFICATION: '/api/v1/users/:id/verification',
  ADMIN_USER_VERIFICATION_APPROVE: '/api/v1/users/:id/verification/approve',
  ADMIN_USER_VERIFICATION_REJECT: '/api/v1/users/:id/verification/reject',
  ADMIN_DEPOSITS_PENDING: '/api/v1/deposits/pending',
  ADMIN_DEPOSIT_REVIEW: '/api/v1/deposits/:id/review',
  ADMIN_WITHDRAWS_PENDING: '/api/v1/withdraws/pending',
  ADMIN_WITHDRAW_REVIEW: '/api/v1/withdraws/:id/review',
  ADMIN_WITHDRAW_PAY: '/api/v1/withdraws/:id/pay',
  ADMIN_ANNOUNCEMENTS: '/api/v1/admin/announcements',
  ADMIN_SALESPERSONS: '/api/v1/admin/salespersons',
  
  // 配置相关
  CONFIG: '/api/v1/configs',
  CONFIG_UPDATE: '/api/v1/configs',
  
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
