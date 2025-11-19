<template>
  <div class="home-page">
    <!-- 行情卡片 -->
    <div class="quote-card">
      <div class="price-info">
        <div class="current-price" :class="{ up: quoteStore.isUp, down: quoteStore.isDown }">
          ¥{{ quoteStore.priceDisplay }}
          <span class="unit">/克</span>
        </div>
        <div class="price-change" :class="{ up: quoteStore.isUp, down: quoteStore.isDown }">
          {{ quoteStore.priceChange >= 0 ? '+' : '' }}{{ quoteStore.priceChange.toFixed(2) }}
          ({{ quoteStore.priceChangePercent >= 0 ? '+' : '' }}{{ quoteStore.priceChangePercent.toFixed(2) }}%)
        </div>
      </div>
      <div class="ws-status">
        <van-tag v-if="quoteStore.isConnected" type="success" size="mini">实时</van-tag>
        <van-tag v-else type="warning" size="mini">未连接</van-tag>
      </div>
    </div>
    
    <!-- 快捷操作 -->
    <div class="quick-actions">
      <van-grid :border="false" :column-num="2">
        <van-grid-item
          icon="add-o"
          text="买入"
          @click="$router.push('/trade?type=buy')"
        />
        <van-grid-item
          icon="minus"
          text="卖出"
          @click="$router.push('/trade?type=sell')"
        />
      </van-grid>
    </div>
    
    <!-- 我的订单 -->
    <div class="order-section">
      <div class="section-header">
        <span>我的订单</span>
        <van-button
          plain
          size="small"
          type="primary"
          @click="$router.push('/orders')"
        >
          查看全部
        </van-button>
      </div>
      
      <van-tabs v-model="activeTab" @change="loadOrders">
        <van-tab title="全部" name="all" />
        <van-tab title="待确认" name="pending" />
        <van-tab title="已成交" name="filled" />
      </van-tabs>
      
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
          <div
            v-for="order in orders"
            :key="order.id"
            class="order-item"
            @click="$router.push(`/orders/${order.id}`)"
          >
            <div class="order-header">
              <span class="order-type" :class="order.type">
                {{ order.type === 'buy' ? '买入' : '卖出' }}
              </span>
              <span class="order-status">{{ ORDER_STATUS[order.status] }}</span>
            </div>
            <div class="order-body">
              <div class="order-info">
                <div class="info-row">
                  <span class="label">价格:</span>
                  <span class="value">¥{{ formatMoney(order.price) }}/克</span>
                </div>
                <div class="info-row">
                  <span class="label">数量:</span>
                  <span class="value">{{ order.amount }}克</span>
                </div>
                <div class="info-row">
                  <span class="label">金额:</span>
                  <span class="value">¥{{ formatMoney(order.total_amount) }}</span>
                </div>
              </div>
              <div class="order-time">
                {{ formatDateTime(order.created_at) }}
              </div>
            </div>
          </div>
        </van-list>
      </van-pull-refresh>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useQuoteStore } from '../stores/quote'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'
import { formatMoney, formatDateTime, ORDER_STATUS } from '../utils/helpers'

const quoteStore = useQuoteStore()

const activeTab = ref('all')
const orders = ref([])
const refreshing = ref(false)
const loading = ref(false)
const finished = ref(false)
const page = ref(1)

// 加载订单
const loadOrders = async () => {
  try {
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
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

// 下拉刷新
const onRefresh = () => {
  page.value = 1
  finished.value = false
  loadOrders()
}

onMounted(() => {
  // 连接行情WebSocket
  quoteStore.connectWebSocket()
  // 加载订单
  loadOrders()
})

onUnmounted(() => {
  // 断开WebSocket连接
  quoteStore.disconnectWebSocket()
})
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}

.quote-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 30px 20px;
  color: #fff;
  position: relative;
}

.price-info {
  text-align: center;
}

.current-price {
  font-size: 48px;
  font-weight: bold;
  margin-bottom: 10px;
}

.current-price.up {
  color: #f56c6c;
}

.current-price.down {
  color: #67c23a;
}

.unit {
  font-size: 16px;
  font-weight: normal;
}

.price-change {
  font-size: 16px;
}

.price-change.up {
  color: #f56c6c;
}

.price-change.down {
  color: #67c23a;
}

.ws-status {
  position: absolute;
  top: 20px;
  right: 20px;
}

.quick-actions {
  margin: 16px 0;
}

.order-section {
  background: #fff;
  margin-top: 16px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  font-size: 16px;
  font-weight: bold;
}

.order-item {
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.order-type {
  font-size: 16px;
  font-weight: bold;
}

.order-type.buy {
  color: #f56c6c;
}

.order-type.sell {
  color: #67c23a;
}

.order-status {
  color: #999;
  font-size: 14px;
}

.order-body {
  background: #f7f8fa;
  padding: 12px;
  border-radius: 8px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.label {
  color: #666;
  font-size: 14px;
}

.value {
  font-size: 14px;
  font-weight: 500;
}

.order-time {
  margin-top: 8px;
  color: #999;
  font-size: 12px;
  text-align: right;
}

.empty {
  padding: 40px 0;
}
</style>
