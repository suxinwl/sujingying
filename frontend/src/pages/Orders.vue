<template>
  <div class="orders-page">
    <van-nav-bar
      title="订单"
      fixed
      placeholder
    />
    
    <!-- Tab切换 -->
    <van-tabs v-model:active="activeTab" @change="onTabChange">
      <van-tab title="全部" name="all" />
      <van-tab title="持仓中" name="open" />
      <van-tab title="已平仓" name="closed" />
    </van-tabs>
    
    <!-- 订单列表 -->
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        finished-text="没有更多了"
        @load="loadOrders"
      >
        <div v-if="orders.length === 0" class="empty">
          <van-empty description="暂无订单" />
        </div>
        
        <div v-for="order in orders" :key="order.id" class="order-item" @click="viewOrder(order)">
          <div class="order-header">
            <span class="order-type" :class="order.direction">
              {{ order.direction === 'buy' ? '买入' : '卖出' }}
            </span>
            <span class="order-status" :class="order.status">
              {{ getStatusText(order.status) }}
            </span>
          </div>
          
          <div class="order-body">
            <div class="order-row">
              <span class="label">品种:</span>
              <span class="value">{{ order.product_name }}</span>
            </div>
            <div class="order-row">
              <span class="label">数量:</span>
              <span class="value">{{ order.quantity }} 克</span>
            </div>
            <div class="order-row">
              <span class="label">开仓价:</span>
              <span class="value">¥{{ formatMoney(order.open_price) }}</span>
            </div>
            <div v-if="order.close_price" class="order-row">
              <span class="label">平仓价:</span>
              <span class="value">¥{{ formatMoney(order.close_price) }}</span>
            </div>
            <div v-if="order.profit_loss" class="order-row">
              <span class="label">盈亏:</span>
              <span class="value" :class="{ profit: order.profit_loss > 0, loss: order.profit_loss < 0 }">
                {{ order.profit_loss > 0 ? '+' : '' }}¥{{ formatMoney(Math.abs(order.profit_loss)) }}
              </span>
            </div>
          </div>
          
          <div class="order-footer">
            <span class="order-time">{{ formatDateTime(order.created_at) }}</span>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>
  </div>
</template>

<script setup>
/**
 * @file Orders.vue
 * @description 订单页面 - 展示用户的所有交易订单
 * @author AI
 * @date 2025-11-18
 */

import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import { useRouter } from 'vue-router'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'
import { formatMoney, formatDateTime } from '../utils/helpers'

/**
 * 当前激活的Tab
 * @type {import('vue').Ref<string>}
 */
const router = useRouter()
const activeTab = ref('all')

/**
 * 订单列表
 * @type {import('vue').Ref<Array>}
 */
const orders = ref([])

/**
 * 刷新状态
 * @type {import('vue').Ref<boolean>}
 */
const refreshing = ref(false)

/**
 * 加载状态
 * @type {import('vue').Ref<boolean>}
 */
const loading = ref(false)

/**
 * 是否加载完成
 * @type {import('vue').Ref<boolean>}
 */
const finished = ref(false)

/**
 * 当前页码
 * @type {import('vue').Ref<number>}
 */
const page = ref(1)

/**
 * 获取订单状态文本
 * @param {string} status - 订单状态
 * @returns {string} 状态文本
 */
const getStatusText = (status) => {
  const statusMap = {
    open: '持仓中',
    closed: '已平仓',
    cancelled: '已取消'
  }
  return statusMap[status] || status
}

/**
 * 加载订单列表
 * @async
 * @returns {Promise<void>}
 */
const loadOrders = async () => {
  try {
    loading.value = true
    const params = {
      page: page.value,
      page_size: 10
    }
    
    if (activeTab.value !== 'all') {
      params.status = activeTab.value
    }
    
    const data = await request.get(API_ENDPOINTS.ORDERS, { params })
    
    const list = data.orders || data.list || []
    
    if (page.value === 1) {
      orders.value = list
    } else {
      orders.value.push(...list)
    }
    
    if (list.length < 10) {
      finished.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载订单失败:', error)
    showToast('加载失败')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

/**
 * Tab切换回调
 * @returns {void}
 */
const onTabChange = () => {
  page.value = 1
  finished.value = false
  orders.value = []
  loadOrders()
}

/**
 * 下拉刷新
 * @returns {void}
 */
const onRefresh = () => {
  page.value = 1
  finished.value = false
  loadOrders()
}

/**
 * 查看订单详情
 * @param {Object} order - 订单对象
 * @returns {void}
 */
const viewOrder = (order) => {
  const id = order.order_id || order.id || order.OrderID
  if (!id) {
    showToast('无法获取订单ID')
    return
  }
  router.push(`/orders/${id}`)
}

onMounted(() => {
  loadOrders()
})
</script>

<style scoped>
.orders-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 60px;
}

.order-item {
  background: #fff;
  margin: 10px;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.order-type {
  font-size: 16px;
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 4px;
}

.order-type.buy {
  color: #f56c6c;
  background: #fef0f0;
}

.order-type.sell {
  color: #67c23a;
  background: #f0f9ff;
}

.order-status {
  font-size: 14px;
  padding: 2px 8px;
  border-radius: 4px;
}

.order-status.open {
  color: #409eff;
  background: #ecf5ff;
}

.order-status.closed {
  color: #909399;
  background: #f4f4f5;
}

.order-body {
  margin: 12px 0;
}

.order-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}

.order-row .label {
  color: #909399;
}

.order-row .value {
  color: #303133;
  font-weight: 500;
}

.order-row .value.profit {
  color: #f56c6c;
}

.order-row .value.loss {
  color: #67c23a;
}

.order-footer {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.order-time {
  font-size: 12px;
  color: #c0c4cc;
}

.empty {
  padding: 60px 0;
}
</style>
