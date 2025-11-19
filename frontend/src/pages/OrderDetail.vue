<template>
  <div class="order-detail-page">
    <van-nav-bar
      title="订单详情"
      left-arrow
      fixed
      placeholder
      @click-left="$router.back()"
    />

    <div v-if="loading" class="loading">
      <van-loading size="24px">加载中...</van-loading>
    </div>

    <div v-else-if="!order" class="empty">
      <van-empty description="订单不存在或已删除" />
    </div>

    <div v-else class="content">
      <div class="summary-card" :class="isLongBuy ? 'buy' : 'sell'">
        <div class="summary-header">
          <span class="type-tag">{{ orderTypeText }}</span>
          <span class="status-tag">{{ statusText }}</span>
        </div>
        <div class="summary-body">
          <div class="summary-row">
            <div class="summary-item">
              <div class="label">锁定单价(元/克)</div>
              <div class="value">¥{{ formatMoney(order.locked_price) }}</div>
            </div>
            <div class="summary-item">
              <div class="label">当前价格(元/克)</div>
              <div class="value">
                <span v-if="order.current_price">¥{{ formatMoney(order.current_price) }}</span>
                <span v-else>-</span>
              </div>
            </div>
          </div>
          <div class="summary-row">
            <div class="summary-item">
              <div class="label">浮动盈亏</div>
              <div class="value" :class="{ profit: order.pnl_float > 0, loss: order.pnl_float < 0 }">
                {{ order.pnl_float > 0 ? '+' : '' }}¥{{ formatMoney(Math.abs(order.pnl_float)) }}
              </div>
            </div>
            <div class="summary-item">
              <div class="label">定金率</div>
              <div class="value">
                {{ uiMarginRate !== null ? uiMarginRate.toFixed(2) + '%' : '-' }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="info-section">
        <div class="section-title">订单信息</div>
        <van-cell-group inset>
          <van-cell
            is-link
            title="订单号"
            :value="order.order_id"
            @click="showFieldDetail('订单号', order.order_id)"
          />
          <van-cell
            is-link
            title="订单类型"
            :value="orderTypeText"
            @click="showFieldDetail('订单类型', orderTypeText)"
          />
          <van-cell
            is-link
            title="订单状态"
            :value="statusText"
            @click="showFieldDetail('订单状态', statusText)"
          />
          <van-cell
            is-link
            title="锁定单价"
            :value="`¥${formatMoney(order.locked_price)}`"
            @click="showFieldDetail('锁定单价', `¥${formatMoney(order.locked_price)}`)"
          />
          <van-cell
            is-link
            title="锁定货款"
            :value="`¥${formatMoney(lockedAmount)}`"
            @click="showFieldDetail('锁定货款', `¥${formatMoney(lockedAmount)}`)"
          />
          <van-cell
            is-link
            title="当前价格"
            :value="order.current_price ? `¥${formatMoney(order.current_price)}` : '-'"
            @click="showFieldDetail('当前价格', order.current_price ? `¥${formatMoney(order.current_price)}` : '-')"
          />
          <van-cell
            is-link
            title="克重"
            :value="order.weight_g + ' 克'"
            @click="showFieldDetail('克重', order.weight_g + ' 克')"
          />
          <van-cell
            is-link
            title="定金"
            :value="`¥${formatMoney(order.deposit)}`"
            @click="showFieldDetail('定金', `¥${formatMoney(order.deposit)}`)"
          />
          <van-cell
            is-link
            title="浮动盈亏"
            :value="(order.pnl_float > 0 ? '+' : '') + '¥' + formatMoney(Math.abs(order.pnl_float))"
            @click="showFieldDetail('浮动盈亏', (order.pnl_float > 0 ? '+' : '') + '¥' + formatMoney(Math.abs(order.pnl_float)))"
          />
          <van-cell
            is-link
            title="定金率"
            :value="uiMarginRate !== null ? uiMarginRate.toFixed(2) + '%' : '-'"
            @click="showFieldDetail('定金率', uiMarginRate !== null ? uiMarginRate.toFixed(2) + '%' : '-')"
          />
          <van-cell
            v-if="order.settled_price"
            is-link
            title="结算价格"
            :value="`¥${formatMoney(order.settled_price)}`"
            @click="showFieldDetail('结算价格', `¥${formatMoney(order.settled_price)}`)"
          />
          <van-cell
            v-if="order.settled_pnl"
            is-link
            title="结算盈亏"
            :value="(order.settled_pnl > 0 ? '+' : '') + '¥' + formatMoney(Math.abs(order.settled_pnl))"
            @click="showFieldDetail('结算盈亏', (order.settled_pnl > 0 ? '+' : '') + '¥' + formatMoney(Math.abs(order.settled_pnl)))"
          />
          <van-cell
            is-link
            title="创建时间"
            :value="formatDateTime(order.created_at)"
            @click="showFieldDetail('创建时间', formatDateTime(order.created_at))"
          />
          <van-cell
            v-if="order.settled_at"
            is-link
            title="结算时间"
            :value="formatDateTime(order.settled_at)"
            @click="showFieldDetail('结算时间', formatDateTime(order.settled_at))"
          />
        </van-cell-group>
      </div>

      <div class="bottom-actions" v-if="order.status === 'holding'">
        <van-button type="danger" block round @click="onSettle">
          结算
        </van-button>
      </div>

      <van-dialog v-model:show="showFieldDialog" :title="dialogTitle" :message="dialogContent" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { showToast, showDialog } from 'vant'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'
import { formatMoney, formatDateTime } from '../utils/helpers'

const route = useRoute()
const loading = ref(false)
const order = ref(null)
const showFieldDialog = ref(false)
const dialogTitle = ref('')
const dialogContent = ref('')

const isLongBuy = computed(() => order.value?.type === 'long_buy')

const orderTypeText = computed(() => {
  if (!order.value) return ''
  if (order.value.type === 'long_buy') return '锁价买料'
  if (order.value.type === 'short_sell') return '锁价卖料'
  return '-'
})

const statusText = computed(() => {
  if (!order.value) return ''
  const status = order.value.status
  const map = {
    holding: '待结算',
    settled: '已完结',
    closed: '已完结'
  }
  return map[status] || status || '-'
})

const lockedAmount = computed(() => {
  if (!order.value) return 0
  const price = order.value.locked_price || 0
  const weight = order.value.weight_g || 0
  return price * weight
})

// UI 定金率：(定金 + 浮动盈亏) / 定金 × 100%
const uiMarginRate = computed(() => {
  if (!order.value) return null
  const deposit = order.value.deposit || 0
  if (!deposit) return null
  const pnl = order.value.pnl_float || 0
  return ((deposit + pnl) / deposit) * 100
})

const showFieldDetail = (title, content) => {
  dialogTitle.value = title
  dialogContent.value = content || '-'
  showFieldDialog.value = true
}

const loadOrder = async () => {
  const id = route.params.id
  if (!id) {
    showToast('缺少订单ID')
    return
  }

  try {
    loading.value = true
    const detail = await request.get(API_ENDPOINTS.ORDER_DETAIL.replace(':id', id))
    const raw = detail.order || detail || {}
    order.value = {
      order_id: raw.order_id || raw.OrderID || raw.id || '',
      type: raw.type || raw.Type || '',
      locked_price: raw.locked_price ?? raw.LockedPrice ?? 0,
      current_price: raw.current_price ?? raw.CurrentPrice ?? 0,
      weight_g: raw.weight_g ?? raw.WeightG ?? 0,
      deposit: raw.deposit ?? raw.Deposit ?? 0,
      pnl_float: raw.pnl_float ?? raw.PnLFloat ?? 0,
      margin_rate: raw.margin_rate ?? raw.MarginRate ?? 0,
      status: raw.status || raw.Status || '',
      settled_price: raw.settled_price ?? raw.SettledPrice ?? 0,
      settled_pnl: raw.settled_pnl ?? raw.SettledPnL ?? 0,
      settled_at: raw.settled_at || raw.SettledAt || null,
      created_at: raw.created_at || raw.CreatedAt || null
    }
  } catch (error) {
    console.error('加载订单详情失败:', error)
    showToast('加载订单详情失败')
  } finally {
    loading.value = false
  }
}

const onSettle = async () => {
  if (!order.value) return

  const confirmed = await new Promise(resolve => {
    showDialog({
      title: '确认结算',
      message: '是否立即对该订单进行结算？',
      showCancelButton: true,
      confirmButtonText: '结算',
      cancelButtonText: '取消',
      beforeClose: action => {
        resolve(action === 'confirm')
        return true
      }
    }).catch(() => {
      resolve(false)
    })
  })

  if (!confirmed) {
    return
  }

  try {
    loading.value = true
    await request.post(API_ENDPOINTS.ORDER_SETTLE.replace(':id', order.value.order_id))
    showToast('结算成功')
    await loadOrder()
  } catch (error) {
    console.error('结算失败:', error)
    const msg = error.response?.data?.error || error.response?.data?.message || '结算失败'
    showToast(msg)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadOrder()
})
</script>

<style scoped>
.order-detail-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 60px;
}

.loading {
  padding: 60px 0;
  text-align: center;
}

.empty {
  padding: 60px 0;
}

.content {
  margin-top: 12px;
}

.profit {
  color: #f56c6c;
}

.loss {
  color: #67c23a;
}
</style>
