import dayjs from 'dayjs'

/**
 * 格式化金额
 * @param {number} amount - 金额
 * @param {number} decimals - 小数位数
 * @returns {string}
 */
export function formatMoney(amount, decimals = 2) {
  if (!amount && amount !== 0) return '0.00'
  return Number(amount).toFixed(decimals).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

/**
 * 格式化日期时间
 * @param {string|Date} date - 日期
 * @param {string} format - 格式
 * @returns {string}
 */
export function formatDateTime(date, format = 'YYYY-MM-DD HH:mm:ss') {
  if (!date) return '-'
  return dayjs(date).format(format)
}

/**
 * 格式化日期
 * @param {string|Date} date - 日期
 * @returns {string}
 */
export function formatDate(date) {
  return formatDateTime(date, 'YYYY-MM-DD')
}

/**
 * 格式化时间
 * @param {string|Date} date - 日期
 * @returns {string}
 */
export function formatTime(date) {
  return formatDateTime(date, 'HH:mm:ss')
}

/**
 * 相对时间
 * @param {string|Date} date - 日期
 * @returns {string}
 */
export function timeAgo(date) {
  const now = dayjs()
  const then = dayjs(date)
  const diff = now.diff(then, 'second')
  
  if (diff < 60) return '刚刚'
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`
  if (diff < 2592000) return `${Math.floor(diff / 86400)}天前`
  return formatDate(date)
}

/**
 * 订单状态文本
 */
export const ORDER_STATUS = {
  pending: '待确认',
  confirmed: '已确认',
  filled: '已成交',
  cancelled: '已取消'
}

/**
 * 订单类型文本
 */
export const ORDER_TYPE = {
  buy: '买入',
  sell: '卖出'
}

/**
 * 持仓状态文本
 */
export const POSITION_STATUS = {
  holding: '持仓中',
  closing: '平仓中',
  closed: '已平仓',
  forced_closed: '已强平'
}

/**
 * 审核状态文本
 */
export const REVIEW_STATUS = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已拒绝'
}

/**
 * 用户状态文本
 */
export const USER_STATUS = {
  pending: '待审核',
  active: '正常',
  suspended: '已停用'
}

/**
 * 角色文本
 */
export const ROLE_TEXT = {
  customer: '客户',
  sales: '销售',
  support: '客服',
  super_admin: '超级管理员'
}

/**
 * 计算盈亏
 * @param {number} buyPrice - 买入价
 * @param {number} currentPrice - 当前价
 * @param {number} amount - 数量
 * @returns {number}
 */
export function calculateProfit(buyPrice, currentPrice, amount) {
  return (currentPrice - buyPrice) * amount
}

/**
 * 计算盈亏率
 * @param {number} buyPrice - 买入价
 * @param {number} currentPrice - 当前价
 * @returns {number}
 */
export function calculateProfitRate(buyPrice, currentPrice) {
  if (!buyPrice) return 0
  return ((currentPrice - buyPrice) / buyPrice) * 100
}

/**
 * 复制文本到剪贴板
 * @param {string} text - 文本
 * @returns {Promise}
 */
export async function copyToClipboard(text) {
  if (navigator.clipboard) {
    return navigator.clipboard.writeText(text)
  }
  // 兼容性处理
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  document.execCommand('copy')
  document.body.removeChild(textarea)
}

/**
 * 防抖
 * @param {Function} fn - 函数
 * @param {number} delay - 延迟
 * @returns {Function}
 */
export function debounce(fn, delay = 300) {
  let timer = null
  return function (...args) {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      fn.apply(this, args)
    }, delay)
  }
}

/**
 * 节流
 * @param {Function} fn - 函数
 * @param {number} delay - 延迟
 * @returns {Function}
 */
export function throttle(fn, delay = 300) {
  let last = 0
  return function (...args) {
    const now = Date.now()
    if (now - last > delay) {
      last = now
      fn.apply(this, args)
    }
  }
}
