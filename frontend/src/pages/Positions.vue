<template>
  <div class="positions-page">
    <van-nav-bar
      title="持仓"
      fixed
      placeholder
    />
    
    <!-- 统计信息 -->
    <div class="statistics">
      <div class="stat-item">
        <div class="label">总持仓</div>
        <div class="value">{{ totalAmount }}克</div>
      </div>
      <div class="stat-item">
        <div class="label">总盈亏</div>
        <div class="value" :class="{ profit: totalProfit > 0, loss: totalProfit < 0 }">
          ¥{{ formatMoney(Math.abs(totalProfit)) }}
        </div>
      </div>
      <div class="stat-item">
        <div class="label">盈亏率</div>
        <div class="value" :class="{ profit: totalProfitRate > 0, loss: totalProfitRate < 0 }">
          {{ totalProfitRate.toFixed(2) }}%
        </div>
      </div>
    </div>
    
    <!-- 持仓列表 -->
    <van-tabs v-model:active="activeTab" @change="loadPositions">
      <van-tab title="持仓中" name="holding" />
      <van-tab title="已平仓" name="closed" />
      <van-tab title="已强平" name="forced_closed" />
    </van-tabs>
    
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        finished-text="没有更多了"
        @load="loadPositions"
      >
        <div v-if="positions.length === 0" class="empty">
          <van-empty description="暂无持仓" />
        </div>
        
        <div
          v-for="position in positions"
          :key="position.id"
          class="position-item"
          @click="showPositionDetail(position)"
        >
          <div class="position-header">
            <span class="position-id">#{{ position.id }}</span>
            <van-tag :type="getStatusType(position.status)">
              {{ POSITION_STATUS[position.status] }}
            </van-tag>
          </div>
          
          <div class="position-body">
            <div class="info-row">
              <span class="label">买入价:</span>
              <span class="value">¥{{ formatMoney(position.buy_price) }}/克</span>
            </div>
            <div class="info-row">
              <span class="label">数量:</span>
              <span class="value">{{ position.amount }}克</span>
            </div>
            <div class="info-row">
              <span class="label">当前价:</span>
              <span class="value">¥{{ formatMoney(quoteStore.currentPrice) }}/克</span>
            </div>
            <div class="info-row">
              <span class="label">盈亏:</span>
              <span
                class="value"
                :class="{
                  profit: calculateProfit(position) > 0,
                  loss: calculateProfit(position) < 0
                }"
              >
                ¥{{ formatMoney(Math.abs(calculateProfit(position))) }}
                ({{ calculateProfitRate(position).toFixed(2) }}%)
              </span>
            </div>
          </div>
          
          <div class="position-footer" v-if="position.status === 'holding'">
            <van-button
              size="small"
              type="danger"
              @click.stop="sellPosition(position)"
            >
              卖出
            </van-button>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showDialog, showConfirmDialog } from 'vant'
import { useQuoteStore } from '../stores/quote'
import request from '../utils/request'
import { API_ENDPOINTS, replaceParams } from '../config/api'
import { formatMoney, calculateProfit, calculateProfitRate, POSITION_STATUS } from '../utils/helpers'

const router = useRouter()
const quoteStore = useQuoteStore()

const activeTab = ref('holding')
const positions = ref([])
const refreshing = ref(false)
const loading = ref(false)
const finished = ref(false)
const page = ref(1)

// 计算统计数据
const totalAmount = computed(() => {
  return positions.value
    .filter(p => p.status === 'holding')
    .reduce((sum, p) => sum + p.amount, 0)
})

const totalProfit = computed(() => {
  return positions.value
    .filter(p => p.status === 'holding')
    .reduce((sum, p) => sum + calculateProfit(p), 0)
})

const totalProfitRate = computed(() => {
  const totalCost = positions.value
    .filter(p => p.status === 'holding')
    .reduce((sum, p) => sum + p.buy_price * p.amount, 0)
  
  if (totalCost === 0) return 0
  return (totalProfit.value / totalCost) * 100
})

// 获取状态类型
const getStatusType = (status) => {
  const types = {
    holding: 'primary',
    closing: 'warning',
    closed: 'success',
    forced_closed: 'danger'
  }
  return types[status] || 'default'
}

// 加载持仓
const loadPositions = async () => {
  try {
    const params = {
      page: page.value,
      page_size: 10,
      status: activeTab.value
    }
    
    const { data } = await request.get(API_ENDPOINTS.POSITIONS, { params })
    
    if (page.value === 1) {
      positions.value = data.list || []
    } else {
      positions.value.push(...(data.list || []))
    }
    
    if (!data.list || data.list.length < 10) {
      finished.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载持仓失败:', error)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

// 下拉刷新
const onRefresh = () => {
  page.value = 1
  finished.value = false
  loadPositions()
}

// 显示持仓详情
const showPositionDetail = (position) => {
  const profit = calculateProfit(position)
  const profitRate = calculateProfitRate(position)
  
  showDialog({
    title: `持仓详情 #${position.id}`,
    message: `
      买入价: ¥${formatMoney(position.buy_price)}/克
      数量: ${position.amount}克
      当前价: ¥${formatMoney(quoteStore.currentPrice)}/克
      盈亏: ¥${formatMoney(Math.abs(profit))} (${profitRate.toFixed(2)}%)
      买入时间: ${position.created_at}
    `,
    confirmButtonText: '关闭'
  })
}

// 卖出持仓
const sellPosition = async (position) => {
  try {
    await showConfirmDialog({
      title: '确认卖出',
      message: `是否以市价 ¥${quoteStore.priceDisplay}/克 卖出 ${position.amount}克？`
    })
    
    await request.post(API_ENDPOINTS.ORDER_SELL, {
      price: quoteStore.currentPrice,
      amount: position.amount
    })
    
    showDialog({
      title: '卖出成功',
      message: '订单已提交',
      confirmButtonText: '确定'
    }).then(() => {
      onRefresh()
    })
  } catch (error) {
    if (error !== 'cancel') {
      console.error('卖出失败:', error)
    }
  }
}

onMounted(() => {
  quoteStore.connectWebSocket()
  loadPositions()
})
</script>

<style scoped>
.positions-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}

.statistics {
  background: #fff;
  padding: 20px;
  display: flex;
  justify-content: space-around;
  margin-bottom: 10px;
}

.stat-item {
  text-align: center;
}

.stat-item .label {
  font-size: 12px;
  color: #999;
  margin-bottom: 8px;
}

.stat-item .value {
  font-size: 18px;
  font-weight: bold;
  color: #333;
}

.stat-item .value.profit {
  color: #f56c6c;
}

.stat-item .value.loss {
  color: #67c23a;
}

.position-item {
  background: #fff;
  margin: 10px;
  padding: 16px;
  border-radius: 8px;
}

.position-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.position-id {
  font-size: 14px;
  color: #999;
}

.position-body {
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

.value.profit {
  color: #f56c6c;
}

.value.loss {
  color: #67c23a;
}

.position-footer {
  margin-top: 12px;
  text-align: right;
}

.empty {
  padding: 40px 0;
}
</style>
