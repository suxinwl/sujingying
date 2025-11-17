<!--
  行情页面组件
  
  用途：
  - 实时显示黄金、白银等贵金属行情数据
  - 连接到真实行情WebSocket数据源
  - 1:1复刻设计模板UI样式（深色主题）
  - 支持数据变化动画效果
  
  作者：速金盈技术团队
  日期：2025-11
-->

<template>
  <div class="quote-page">
    <!-- 顶部标题栏 -->
    <div class="header">
      <h1>速金盈黄金</h1>
    </div>

    <!-- 状态栏：时间 + 连接状态 + 客服按钮 -->
    <div class="status-bar">
      <div class="status-time">{{ currentTime }}</div>
      <div class="status-center">
        <span class="status-badge" :class="{ connected: isConnected }">
          {{ isConnected ? '休息中' : '连接断开' }}
        </span>
      </div>
      <div class="status-actions">
        <a href="tel:18038018206" class="customer-service">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"></path>
          </svg>
          <span>客服</span>
        </a>
      </div>
    </div>

    <!-- 行情数据表格 -->
    <div class="container">
      <!-- 遍历三组行情：现货、国内、国际 -->
      <div 
        v-for="(products, sectionName) in WS_CONFIG.PRODUCT_ORDER" 
        :key="sectionName" 
        class="section"
      >
        <div class="section-title">{{ sectionName }}</div>
        <table class="price-table">
          <thead>
            <tr>
              <th>商品</th>
              <th>回购</th>
              <th>销售</th>
              <th>高/低</th>
              <!-- 只有现货行情显示基差列 -->
              <th v-if="sectionName === '现货行情'">基差</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="product in products" :key="product.code">
              <td class="product-name">{{ product.name }}</td>
              <td class="price-down" :data-field="`${product.code}-buy`">
                {{ formatPrice(quoteData[product.code]?.Buy) }}
              </td>
              <td class="price-up" :data-field="`${product.code}-sell`">
                {{ formatPrice(quoteData[product.code]?.Sell) }}
              </td>
              <td>
                <div class="high-low">
                  <span class="high" :data-field="`${product.code}-high`">
                    {{ formatPrice(quoteData[product.code]?.H) }}
                  </span>
                  <span class="low" :data-field="`${product.code}-low`">
                    {{ formatPrice(quoteData[product.code]?.L) }}
                  </span>
                </div>
              </td>
              <td 
                v-if="sectionName === '现货行情'" 
                :class="getGapClass(quoteData[product.code]?.Gap)"
                class="gap-value"
              >
                {{ formatGap(quoteData[product.code]?.Gap) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 底部说明 -->
    <div class="footer">
      以上数据仅供参考
    </div>
  </div>
</template>

<script setup>
/**
 * 行情页面逻辑
 * 
 * 功能：
 * 1. 组件挂载时连接WebSocket
 * 2. 接收行情数据并更新视图
 * 3. 定时更新当前时间显示
 * 4. 组件卸载时断开WebSocket
 */

import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { quoteWS } from '@/utils/quoteWebSocket'
import { WS_CONFIG } from '@/config/websocket'

// ========== 响应式数据 ==========

/**
 * 当前时间显示
 * @type {Ref<string>}
 */
const currentTime = ref('')

/**
 * WebSocket连接状态
 * @type {Ref<boolean>}
 */
const isConnected = ref(false)

/**
 * 行情数据对象
 * key: 商品代码（如 'AU', 'AG'）
 * value: 行情数据 { Buy, Sell, H, L, Gap }
 * @type {Ref<Object>}
 */
const quoteData = ref({})

/**
 * 上一次的行情数据（用于检测变化）
 * @type {Ref<Object>}
 */
const prevQuoteData = ref({})

// ========== 辅助函数 ==========

/**
 * 格式化价格显示
 * 
 * @param {number|string|null|undefined} value - 价格值
 * @returns {string} 格式化后的价格字符串，无效值返回'--'
 */
const formatPrice = (value) => {
  if (value === null || value === undefined || value === '') {
    return '--'
  }
  return typeof value === 'number' ? value.toFixed(2) : value
}

/**
 * 格式化基差显示
 * 
 * @param {number|string|null|undefined} gap - 基差值
 * @returns {string} 格式化后的基差字符串
 */
const formatGap = (gap) => {
  if (!gap || gap === '--') {
    return '--'
  }
  const gapValue = parseFloat(gap)
  return isNaN(gapValue) ? '--' : gapValue.toFixed(2)
}

/**
 * 获取基差样式类名
 * 
 * @param {number|string|null|undefined} gap - 基差值
 * @returns {string} CSS类名：gap-positive（正）/ gap-negative（负）/ price-neutral（中性）
 */
const getGapClass = (gap) => {
  if (!gap || gap === '--') {
    return 'price-neutral'
  }
  const gapValue = parseFloat(gap)
  if (isNaN(gapValue)) {
    return 'price-neutral'
  }
  if (gapValue > 0) {
    return 'gap-positive'
  }
  if (gapValue < 0) {
    return 'gap-negative'
  }
  return 'price-neutral'
}

/**
 * 更新当前时间显示
 * 格式：YYYY-MM-DD HH:mm:ss
 * 
 * @returns {void}
 */
const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  }).replace(/\//g, '-')
}

/**
 * 触发价格变化闪烁动画
 * 
 * @param {string} productCode - 商品代码
 * @param {string} field - 字段名（buy/sell/high/low）
 * @param {string} direction - 变化方向（up/down）
 * @returns {void}
 */
const triggerFlash = (productCode, field, direction) => {
  const selector = `[data-field="${productCode}-${field}"]`
  const element = document.querySelector(selector)
  
  if (element) {
    // 移除旧的动画类
    element.classList.remove('flash-up', 'flash-down')
    
    // 强制重绘以重新触发动画
    void element.offsetWidth
    
    // 添加新的动画类
    element.classList.add(direction === 'up' ? 'flash-up' : 'flash-down')
    
    // 动画结束后移除类
    setTimeout(() => {
      element.classList.remove('flash-up', 'flash-down')
    }, 600)
  }
}

/**
 * 处理行情数据更新
 * WebSocket消息回调函数
 * 
 * 功能：
 * 1. 检测数值变化
 * 2. 触发闪烁动画
 * 3. 更新行情数据
 * 
 * @param {Object} items - 行情数据对象
 * @returns {void}
 */
const handleQuoteUpdate = (items) => {
  // 遍历所有商品，检测价格变化
  Object.keys(items).forEach(productCode => {
    const newData = items[productCode]
    const oldData = prevQuoteData.value[productCode]
    
    if (oldData) {
      // 检测回购价变化
      if (newData.Buy !== undefined && newData.Buy !== oldData.Buy) {
        const direction = newData.Buy > oldData.Buy ? 'up' : 'down'
        nextTick(() => triggerFlash(productCode, 'buy', direction))
      }
      
      // 检测销售价变化
      if (newData.Sell !== undefined && newData.Sell !== oldData.Sell) {
        const direction = newData.Sell > oldData.Sell ? 'up' : 'down'
        nextTick(() => triggerFlash(productCode, 'sell', direction))
      }
      
      // 检测最高价变化
      if (newData.H !== undefined && newData.H !== oldData.H) {
        const direction = newData.H > oldData.H ? 'up' : 'down'
        nextTick(() => triggerFlash(productCode, 'high', direction))
      }
      
      // 检测最低价变化
      if (newData.L !== undefined && newData.L !== oldData.L) {
        const direction = newData.L > oldData.L ? 'up' : 'down'
        nextTick(() => triggerFlash(productCode, 'low', direction))
      }
    }
  })
  
  // 保存旧数据用于下次对比
  prevQuoteData.value = JSON.parse(JSON.stringify(items))
  
  // 更新当前数据
  quoteData.value = items
  isConnected.value = true
}

// ========== 生命周期钩子 ==========

/**
 * 组件挂载时执行
 * 1. 连接WebSocket
 * 2. 注册消息处理器
 * 3. 启动时间更新定时器
 */
onMounted(() => {
  // 连接WebSocket并注册消息处理器
  quoteWS.connect()
  quoteWS.onMessage(handleQuoteUpdate)
  
  // 初始化时间并启动定时器（每秒更新）
  updateTime()
  const timer = setInterval(updateTime, 1000)
  
  // 保存定时器ID用于清理
  onUnmounted(() => {
    clearInterval(timer)
  })
})

/**
 * 组件卸载时执行
 * 1. 移除消息处理器
 * 2. 断开WebSocket连接
 */
onUnmounted(() => {
  quoteWS.offMessage(handleQuoteUpdate)
  quoteWS.disconnect()
})
</script>

<style scoped>
/**
 * 行情页面样式
 * 1:1复刻设计模板UI
 * 深色主题，渐变背景
 */

/* ========== 全局容器 ========== */
.quote-page {
  background: linear-gradient(to bottom, #1a2332, #283446);
  min-height: 100vh;
  color: #fff;
  font-family: 'Microsoft YaHei', Arial, sans-serif;
}

/* ========== 顶部标题栏 ========== */
.header {
  background: #000;
  padding: 15px;
  text-align: center;
}

.header h1 {
  font-size: 32px;
  color: #d4a45a;
  letter-spacing: 8px;
  margin: 0;
}

/* ========== 状态栏 ========== */
.status-bar {
  background: #1a2332;
  padding: 12px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #2a3a4f;
}

.status-time {
  flex: 0 0 auto;
  font-size: 16px;
  font-weight: 500;
}

.status-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.status-badge {
  background: #dc3545;
  color: white;
  padding: 2px 12px;
  border-radius: 3px;
  font-size: 12px;
}

.status-badge.connected {
  background: #28a745;
}

/* 客服按钮 */
.status-actions {
  flex: 0 0 auto;
}

.customer-service {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #d4a45a, #f4c06f);
  color: #000;
  text-decoration: none;
  border-radius: 20px;
  font-size: 14px;
  font-weight: bold;
  transition: all 0.3s;
  box-shadow: 0 2px 8px rgba(212, 164, 90, 0.3);
}

.customer-service:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(212, 164, 90, 0.5);
}

/* ========== 行情表格 ========== */
.container {
  padding: 0;
}

.section {
  margin: 0;
  border-bottom: 3px solid #2a3a4f;
}

.section-title {
  background: #2a3a4f;
  padding: 12px 20px;
  font-size: 16px;
  font-weight: bold;
  text-align: center;
}

.price-table {
  width: 100%;
  border-collapse: collapse;
  background: #1a2332;
}

.price-table thead {
  background: #0d1419;
}

.price-table th {
  padding: 12px;
  text-align: center;
  font-weight: normal;
  color: #999;
  font-size: 15px;
  border-bottom: 1px solid #2a3a4f;
}

.price-table td {
  padding: 12px;
  text-align: center;
  font-size: 16px;
  border-bottom: 1px solid #222d3d;
  transition: background 0.2s;
}

.price-table tbody tr:hover {
  background: rgba(212, 164, 90, 0.1);
}

/* 商品名称列左对齐 */
.product-name {
  text-align: left !important;
  padding-left: 20px !important;
  font-weight: 500;
}

/* ========== 价格颜色 ========== */
.price-up {
  color: #ff4444;
  font-weight: bold;
}

.price-down {
  color: #00ff00;
  font-weight: bold;
}

.price-neutral {
  color: #ffffff;
}

/* ========== 高低价显示 ========== */
.high-low {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.high-low .high {
  color: #ff4444;
  font-size: 14px;
}

.high-low .low {
  color: #00ff00;
  font-size: 14px;
}

/* ========== 基差样式 ========== */
.gap-value {
  font-weight: bold;
}

.gap-positive {
  color: #ff4444;
}

.gap-negative {
  color: #00ff00;
}

/* ========== 数据变化动画 ========== */
[data-field] {
  position: relative;
  transition: all 0.3s ease;
}

/* 价格上涨闪烁动画 */
@keyframes flashUp {
  0% {
    background-color: rgba(255, 68, 68, 0.4);
    transform: scale(1.05);
  }
  50% {
    background-color: rgba(255, 68, 68, 0.6);
  }
  100% {
    background-color: transparent;
    transform: scale(1);
  }
}

/* 价格下跌闪烁动画 */
@keyframes flashDown {
  0% {
    background-color: rgba(0, 255, 0, 0.4);
    transform: scale(1.05);
  }
  50% {
    background-color: rgba(0, 255, 0, 0.6);
  }
  100% {
    background-color: transparent;
    transform: scale(1);
  }
}

/* 应用上涨动画 */
.flash-up {
  animation: flashUp 0.6s ease-out;
}

/* 应用下跌动画 */
.flash-down {
  animation: flashDown 0.6s ease-out;
}

/* ========== 底部说明 ========== */
.footer {
  text-align: center;
  padding: 20px;
  color: #666;
  font-size: 12px;
}

/* ========== 响应式设计 ========== */
@media (max-width: 768px) {
  .header h1 {
    font-size: 24px;
    letter-spacing: 4px;
  }
  
  .status-bar {
    padding: 8px 12px;
  }
  
  .status-time {
    font-size: 14px;
  }
  
  .customer-service {
    padding: 6px 12px;
    font-size: 12px;
  }
  
  .price-table th,
  .price-table td {
    padding: 8px;
    font-size: 14px;
  }
}
</style>
