<template>
  <div class="positions-page">
    <van-nav-bar
      title="料仓"
      fixed
      placeholder
    />

    <!-- 过滤 Tabs：锁价买料 / 锁价卖料 + 待结算 / 已结算 -->
    <div class="trade-filters">
      <div class="trade-type-tabs">
        <div
          class="trade-type-tab"
          :class="{ active: activeType === 'long_buy' }"
          @click="switchType('long_buy')"
        >
          锁价买料
        </div>
        <div
          class="trade-type-tab"
          :class="{ active: activeType === 'short_sell' }"
          @click="switchType('short_sell')"
        >
          锁价卖料
        </div>
      </div>
      <div class="trade-status-tabs">
        <div
          class="trade-status-tab"
          :class="{ active: activeTab === 'unsettled' }"
          @click="switchTab('unsettled')"
        >
          待结算
        </div>
        <div
          class="trade-status-tab"
          :class="{ active: activeTab === 'settled' }"
          @click="switchTab('settled')"
        >
          已结算
        </div>
      </div>
    </div>

    <!-- 顶部汇总区域（采用资金-交易样式） -->
    <div class="trade-summary-card">
      <div class="trade-summary-item">
        <div class="label">待结金料</div>
        <div class="value">{{ totalAmount.toFixed(2) }}克</div>
      </div>
      <div class="trade-summary-item">
        <div class="label">金料均价</div>
        <div class="value">￥{{ avgLockedPrice.toFixed(2) }}</div>
      </div>
      <div class="trade-summary-item">
        <div class="label">{{ activeTab === 'unsettled' ? '浮动金额' : '结算浮动金额' }}</div>
        <div
          class="value"
          :class="{
            profit: totalProfitDisplay > 0,
            loss: totalProfitDisplay < 0
          }"
        >
          {{
            totalProfitDisplay > 0
              ? '+'
              : totalProfitDisplay < 0
                ? '-'
                : ''
          }}¥{{ formatMoney(Math.abs(totalProfitDisplay)) }}
        </div>
      </div>
      <div class="trade-summary-item">
        <div class="label">需补定金</div>
        <div class="value">¥{{ formatMoney(totalNeedSupplement) }}</div>
      </div>
      <div class="trade-summary-item">
        <div class="label">可用定金</div>
        <div class="value">¥{{ formatMoney(userInfo.available_deposit) }}</div>
      </div>
    </div>

    <!-- 订单列表：采用资金-交易的卡片样式与字段布局 -->
    <div class="orders-container">
      <div v-if="sortedPositions.length === 0" class="empty">
        <van-empty :description="activeTab === 'unsettled' ? '暂无未完结订单' : '暂无已完结订单'" />
      </div>

      <div
        v-for="position in sortedPositions"
        :key="position.id"
        class="record-card"
        @click="showPositionDetail(position)"
      >
        <div class="record-card-header">
          <div
            class="record-type-badge"
            :class="position.type === 'long_buy' ? 'type-trade-long' : 'type-trade-short'"
          >
            {{ '交易-' + getOrderTypeText(position.type) }}
          </div>
          <div class="record-status" :class="position.status">
            {{ getStatusText(position.status) }}
          </div>
        </div>

        <div class="record-card-body">
          <div class="record-info-row">
            <span class="label">订单号</span>
            <span class="value">#{{ position.id }}</span>
          </div>
          <div class="record-info-row">
            <span class="label">定金</span>
            <span class="value">¥{{ formatMoney(getBaseDeposit(position)) }}</span>
          </div>
          <div
            class="record-info-row"
            v-if="getSupplementDepositDisplay(position) > 0"
          >
            <span class="label">已补定金</span>
            <span class="value">¥{{ formatMoney(getSupplementDepositDisplay(position)) }}</span>
          </div>
          <div class="record-info-row">
            <span class="label">{{ position.status === 'holding' ? '浮动金额' : '结算浮动金额' }}</span>
            <span
              class="value amount-text"
              :class="{
                income: getPositionPnL(position) > 0,
                expense: getPositionPnL(position) < 0
              }"
            >
              {{
                getPositionPnL(position) > 0
                  ? '+'
                  : getPositionPnL(position) < 0
                    ? '-'
                    : ''
              }}¥{{ formatMoney(Math.abs(getPositionPnL(position))) }}
            </span>
          </div>
          <div
            class="record-info-row"
            v-if="calcOrderNeedSupplement(position) > 0"
          >
            <span class="label">需补定金</span>
            <span class="value need-supplement-value">¥{{ formatMoney(calcOrderNeedSupplement(position)) }}</span>
          </div>
          <div class="record-info-row">
            <span class="label">定金率</span>
            <span class="value risk-line">
              <span
                class="risk-rate"
                :class="getMarginRate(position) >= 100 ? 'risk-rate-high' : 'risk-rate-low'"
              >
                {{ getMarginRate(position).toFixed(2) }}%
              </span>
              <span
                v-if="getRiskLabel(position)"
                class="risk-tag"
                :class="getRiskClass(position)"
              >
                {{ getRiskLabel(position) }}
              </span>
            </span>
          </div>
          <div class="record-info-row">
            <span class="label">{{ position.status === 'holding' ? '下单时间' : '结算时间' }}</span>
            <span class="value">
              {{
                position.status === 'holding'
                  ? formatDateTime(position.created_at)
                  : formatDateTime(position.settled_at || position.created_at)
              }}
            </span>
          </div>
        </div>

        <div class="record-card-footer">
          <div class="record-actions">
            <van-button
              v-if="calcOrderNeedSupplement(position) > 0"
              size="small"
              type="warning"
              plain
              @click.stop="onClickSupplement(position)"
            >
              补定金
            </van-button>
            <van-button
              v-if="position.status === 'holding'"
              size="small"
              type="primary"
              @click.stop="openSettleDialog(position)"
            >
              结算
            </van-button>
          </div>
          <span class="view-detail">查看详情 ></span>
        </div>
      </div>
    </div>
  </div>

  <!-- 详情弹窗：对齐资金-交易锁价详情样式 -->
  <van-popup v-model:show="showDetailDialog" position="bottom" round :style="{ height: '80%' }">
    <div class="detail-popup" v-if="currentDetailPosition">
      <van-nav-bar
        title="订单详情"
        left-arrow
        @click-left="showDetailDialog = false"
      />

      <div class="detail-content">
        <van-cell-group inset>
          <van-cell
            title="订单号"
            :value="'#' + (currentDetailPosition.id || '')"
          />
          <van-cell
            title="订单类型"
            :value="getOrderTypeText(currentDetailPosition.type)"
          />
          <van-cell
            title="订单状态"
            :value="getStatusText(currentDetailPosition.status)"
          />
          <van-cell
            title="锁定单价"
            :value="'¥' + formatMoney(currentDetailPosition.locked_price || 0) + ' /克'"
          />
          <van-cell
            title="锁定货款"
            :value="'¥' + formatMoney((currentDetailPosition.locked_price || 0) * (currentDetailPosition.amount || 0))"
          />
          <van-cell
            title="当前价格"
            :value="'¥' + formatMoney(getRealtimePrice(currentDetailPosition)) + ' /克'"
          />
          <van-cell
            title="克重"
            :value="(currentDetailPosition.amount || 0) + ' 克'"
          />
          <van-cell
            title="定金"
            :value="'¥' + formatMoney(getBaseDeposit(currentDetailPosition))"
          />
          <van-cell
            title="已补定金"
            :value="'¥' + formatMoney(getSupplementDepositDisplay(currentDetailPosition))"
          />
          <van-cell
            :title="currentDetailPosition.status === 'holding' ? '浮动金额' : '结算浮动金额'"
            :value="
              (getPositionPnL(currentDetailPosition) > 0
                ? '+'
                : getPositionPnL(currentDetailPosition) < 0
                  ? '-'
                  : '') +
              '¥' +
              formatMoney(Math.abs(getPositionPnL(currentDetailPosition)))
            "
            :class="{
              'pnl-profit-cell': getPositionPnL(currentDetailPosition) > 0,
              'pnl-loss-cell': getPositionPnL(currentDetailPosition) < 0
            }"
          />
          <van-cell
            v-if="calcOrderNeedSupplement(currentDetailPosition) > 0"
            title="需补定金"
            :value="'¥' + formatMoney(calcOrderNeedSupplement(currentDetailPosition))"
            class="need-supplement-cell"
          />
          <van-cell
            title="定金率"
            :value="getMarginRate(currentDetailPosition).toFixed(2) + '%'"
            :class="
              getMarginRate(currentDetailPosition) >= 100
                ? 'risk-rate-cell-high'
                : 'risk-rate-cell-low'
            "
          />
          <van-cell
            title="创建时间"
            :value="formatDateTime(currentDetailPosition.created_at)"
          />
          <van-cell
            v-if="currentDetailPosition.settled_at"
            title="结算时间"
            :value="formatDateTime(currentDetailPosition.settled_at)"
          />
        </van-cell-group>

        <div class="detail-actions">
          <van-button
            v-if="calcOrderNeedSupplement(currentDetailPosition) > 0"
            type="warning"
            block
            size="small"
            @click="onClickSupplement(currentDetailPosition)"
          >
            补定金
          </van-button>
          <van-button
            v-if="currentDetailPosition.status === 'holding'"
            type="primary"
            block
            size="small"
            style="margin-top: 8px;"
            @click="openSettleDialog(currentDetailPosition)"
          >
            结算
          </van-button>
        </div>
      </div>
    </div>
  </van-popup>

  <van-dialog
    v-model:show="showSettleDialog"
    title="结算订单"
    show-cancel-button
    @confirm="confirmSettle"
  >
    <van-field
      v-model="settlePayPassword"
      label="支付密码"
      type="password"
      placeholder="请输入支付密码"
    />
  </van-dialog>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { showToast, showDialog } from 'vant'
import { useQuoteStore } from '../stores/quote'
import request from '../utils/request'
import { API_ENDPOINTS, replaceParams } from '../config/api'
import { formatMoney, formatDateTime } from '../utils/helpers'

const router = useRouter()
const route = useRoute()
const quoteStore = useQuoteStore()

const userInfo = ref({
  available_deposit: 0,
  used_deposit: 0
})

const targetMarginRate = ref(100)

// Tab: 未完结(unsettled) / 已完结(settled)
const activeTab = ref('unsettled')
// 类型：锁价买料(long_buy) / 锁价卖料(short_sell)
const activeType = ref('long_buy')

const positions = ref([])
const loading = ref(false)
const finished = ref(false)

// 排序规则
const sortRule = ref('time_desc')
const sortOptions = [
  { name: '下单时间从晚到早', value: 'time_desc' },
  { name: '下单时间从早到晚', value: 'time_asc' },
  { name: '定金从高到低', value: 'deposit_desc' },
  { name: '定金从低到高', value: 'deposit_asc' },
  { name: '浮动金额从高到低', value: 'pnl_desc' },
  { name: '浮动金额从低到高', value: 'pnl_asc' }
]

const sortLabel = computed(() => {
  const found = sortOptions.find((opt) => opt.value === sortRule.value)
  return found ? found.name : '下单时间从晚到早'
})

// 当前Tab + 类型下的订单列表
const currentPositions = computed(() => {
  const showHolding = activeTab.value === 'unsettled'

  return positions.value.filter((p) => {
    const isHolding = p.status === 'holding'
    const isFinished = p.status === 'settled' || p.status === 'closed'
    const matchStatus = showHolding ? isHolding : isFinished
    return matchStatus && p.type === activeType.value
  })
})

// 待结金重
const totalAmount = computed(() => {
  return currentPositions.value.reduce((sum, p) => sum + (p.amount || 0), 0)
})

// 订单均价
const avgLockedPrice = computed(() => {
  const list = currentPositions.value
  if (!list.length) return 0
  const totalAmt = list.reduce((sum, p) => sum + (p.amount || 0), 0)
  if (!totalAmt) return 0
  const totalLocked = list.reduce((sum, p) => sum + (p.locked_price || 0) * (p.amount || 0), 0)
  return totalLocked / totalAmt
})

const getBaseDeposit = (position) => {
  if (!position) return 0
  const amount = position.amount || 0
  return amount * 10
}

const getSupplementDepositDisplay = (position) => {
  if (!position) return 0
  const total = position.deposit || 0
  const base = getBaseDeposit(position)
  const extra = total - base
  return extra > 0 ? extra : 0
}

// 获取实时价格：
// 锁价买料(多)：使用当前回购价 quoteStore.sellPrice（无则用买入价，再无则用锁定价）
// 锁价卖料(空)：使用当前销售价 quoteStore.buyPrice（无则用回购价，再无则用锁定价）
const getRealtimePrice = (position) => {
  if (!position) return 0
  if (position.type === 'long_buy') {
    return quoteStore.sellPrice || quoteStore.buyPrice || position.locked_price || 0
  }
  if (position.type === 'short_sell') {
    // 卖料看销售价
    return quoteStore.buyPrice || quoteStore.sellPrice || position.locked_price || 0
  }
  return position.locked_price || 0
}

// 计算单笔订单浮动盈亏
// 锁价买料（做多）：(当前回购价 - 锁定价) × 克重
// 锁价卖料（做空）：(锁定价 - 当前销售价) × 克重
const computeOrderPnL = (position) => {
  if (!position) return 0
  const price = getRealtimePrice(position)
  const locked = position.locked_price || 0
  const amount = position.amount || 0
  if (!amount) return 0

  if (position.type === 'short_sell') {
    // 做空
    return (locked - price) * amount
  }
  // 做多
  if (position.type === 'long_buy') {
    return (price - locked) * amount
  }
  return 0
}

// 单笔订单当前应展示的浮动金额 / 结算浮动金额
const getPositionPnL = (position) => {
  // 已完结：优先使用结算浮动金额
  if (position.status === 'settled' || position.status === 'closed') {
    if (typeof position.settled_pnl === 'number') return position.settled_pnl
    if (typeof position.pnl_float === 'number') return position.pnl_float
    return 0
  }

  // 未完结：按实时价格计算浮动金额
  return computeOrderPnL(position)
}

// 单笔订单盈亏率
const getPositionPnLRate = (position) => {
  const base = (position.locked_price || 0) * (position.amount || 0)
  if (!base) return 0
  return (getPositionPnL(position) / base) * 100
}

const getMarginRate = (position) => {
  if (!position) return 0
  const base = getBaseDeposit(position)
  if (!base) return 0

  const extra = getSupplementDepositDisplay(position)
  const pnl = getPositionPnL(position)

  return ((base + extra + pnl) / base) * 100
}

// 风险档位：基于定金率
// ≤20%: 强平线
// 20%~25%: 高风险
// ≤50%: 风险预警
const getRiskLabel = (position) => {
  if (position.status !== 'holding') return ''
  const rate = getMarginRate(position)
  if (rate <= 20) return '强平线'
  if (rate > 20 && rate < 25) return '高风险'
  if (rate <= 50) return '风险预警'
  return ''
}

const getRiskClass = (position) => {
  if (position.status !== 'holding') return ''
  const rate = getMarginRate(position)
  if (rate <= 20) return 'danger'
  if (rate > 20 && rate < 25) return 'warning'
  if (rate <= 50) return 'notice'
  return ''
}

// 顶部展示的浮动金额 / 结算盈亏
const totalProfitDisplay = computed(() => {
  const list = currentPositions.value
  if (!list.length) return 0
  return list.reduce((sum, p) => sum + getPositionPnL(p), 0)
})

const calcOrderNeedSupplement = (position) => {
  if (!position) return 0
  const base = getBaseDeposit(position)
  if (!base) return 0

  if (position.status !== 'holding') return 0

  const rate = getMarginRate(position) || 0
  const targetRate = targetMarginRate.value || 100
  if (rate >= targetRate) return 0

  const delta = (base * (targetRate - rate)) / 100
  return delta > 0 ? delta : 0
}

const totalNeedSupplement = computed(() => {
  const list = currentPositions.value
  if (!list.length) return 0
  return list.reduce((sum, p) => sum + calcOrderNeedSupplement(p), 0)
})

const getOrderTypeText = (type) => {
  if (type === 'long_buy') return '锁价买料'
  if (type === 'short_sell') return '锁价卖料'
  return type || ''
}

const getStatusText = (status) => {
  if (!status) return ''
  const normalized = String(status).toLowerCase()
  if (normalized === 'holding') return '待结算'
  if (normalized === 'settled') return '已结算'
  if (normalized === 'closed') return '已强平'
  return status
}

// 排序后的列表
const sortedPositions = computed(() => {
  const list = [...currentPositions.value]
  const rule = sortRule.value

  list.sort((a, b) => {
    if (rule === 'time_desc' || rule === 'time_asc') {
      const ta = new Date(a.created_at || '').getTime() || 0
      const tb = new Date(b.created_at || '').getTime() || 0
      return rule === 'time_desc' ? tb - ta : ta - tb
    }

    if (rule === 'deposit_desc' || rule === 'deposit_asc') {
      const da = a.deposit || 0
      const db = b.deposit || 0
      return rule === 'deposit_desc' ? db - da : da - db
    }

    if (rule === 'pnl_desc' || rule === 'pnl_asc') {
      const pa = getPositionPnL(a)
      const pb = getPositionPnL(b)
      return rule === 'pnl_desc' ? pb - pa : pa - pb
    }

    return 0
  })

  return list
})

// 状态文案：未完结 / 已完结
const getStatusLabel = (status) => {
  return status === 'holding' ? '未完结' : '已完结'
}

// 切换Tab
const switchTab = (tab) => {
  if (activeTab.value === tab) return
  activeTab.value = tab
  loadPositions()
}

// 切换类型：锁价买料 / 锁价卖料
const switchType = (type) => {
  if (activeType.value === type) return
  activeType.value = type
}

// 加载持仓：复用订单列表接口 /api/v1/orders
const loadPositions = async () => {
  try {
    loading.value = true

    // 将 Tab 状态映射为订单状态
    let statusParam = ''
    if (activeTab.value === 'unsettled') {
      statusParam = 'holding'
    }

    const params = {}
    if (statusParam) {
      params.status = statusParam
    }

    const data = await request.get(API_ENDPOINTS.ORDERS, { params })
    const list = data.orders || data.list || []

    // 将订单结构映射为本页面的持仓结构
    const mapped = list.map((order) => {
      const rawStatus = (order.status || order.Status || '').toLowerCase()
      let status = 'holding'
      if (rawStatus === 'holding') {
        status = 'holding'
      } else if (rawStatus === 'closed') {
        status = 'closed'
      } else {
        status = 'settled'
      }

      return {
        db_id: order.ID || order.id || 0,
        id: order.order_id || order.OrderID || order.id || order.ID,
        type: order.type || order.Type || 'long_buy',
        status,
        locked_price: order.locked_price ?? order.LockedPrice ?? 0,
        current_price: order.current_price ?? order.CurrentPrice ?? 0,
        amount: order.weight_g ?? order.WeightG ?? 0,
        deposit: order.deposit ?? order.Deposit ?? 0,
        margin_rate: order.margin_rate ?? order.MarginRate ?? 0,
        pnl_float: order.pnl_float ?? order.PnLFloat ?? 0,
        settled_pnl: order.settled_pnl ?? order.SettledPnL ?? 0,
        created_at: order.created_at || order.CreatedAt || '',
        settled_at:
          order.settled_at ||
          order.SettledAt ||
          order.closed_at ||
          order.ClosedAt ||
          ''
      }
    })

    positions.value = mapped
    finished.value = true
  } catch (error) {
    console.error('加载持仓失败:', error)
  } finally {
    loading.value = false
  }
}

const showDetailDialog = ref(false)
const currentDetailPosition = ref(null)
const showSettleDialog = ref(false)
const settleOrder = ref(null)
const settlePayPassword = ref('')

const onClickSupplement = async (position) => {
  if (!position) return
  const amount = calcOrderNeedSupplement(position)
  if (!amount) {
    showToast('当前订单无需补定金')
    return
  }

  try {
    await showDialog({
      title: '确认补定金',
      message: `该订单需补定金：¥${formatMoney(amount)}，是否立即从可用定金中补充？`,
      showCancelButton: true
    })
  } catch (error) {
    return
  }

  try {
    loading.value = true
    await request.post(API_ENDPOINTS.SUPPLEMENTS, {
      order_id: position.db_id || position.ID || position.id,
      amount: amount
    })
    showToast('补定金成功')
    loadUserInfo()
    loadPositions()
  } catch (error) {
    const msg = error.response?.data?.error || error.response?.data?.message || '补定金失败'
    showToast(msg)
  } finally {
    loading.value = false
  }
}

const openSettleDialog = (position) => {
  if (!position) return
  settleOrder.value = position
  settlePayPassword.value = ''
  showSettleDialog.value = true
}

const confirmSettle = async () => {
  if (!settleOrder.value) return
  const order = settleOrder.value
  const price = getRealtimePrice(order)
  if (!price) {
    showToast('当前价格不可用，暂时无法结算')
    return
  }

  try {
    loading.value = true
    await request.post(
      API_ENDPOINTS.ORDER_SETTLE.replace(':id', order.id || ''),
      {
        settle_price: price,
        pay_password: settlePayPassword.value
      }
    )
    showToast('结算成功')
    showSettleDialog.value = false
    settleOrder.value = null
    loadUserInfo()
    loadPositions()
  } catch (error) {
    const msg = error.response?.data?.error || error.response?.data?.message || '结算失败'
    showToast(msg)
  } finally {
    loading.value = false
  }
}

const loadUserInfo = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.USER_PROFILE)
    userInfo.value = {
      available_deposit: data.available_deposit || 0,
      used_deposit: data.used_deposit || 0
    }
  } catch (error) {
    console.error('加载用户信息失败:', error)
  }
}

const loadTargetMarginRate = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.CONFIG)
    const list = data.configs || data.list || []
    const item = list.find((c) => {
      const key = c.key || c.Key || ''
      return key === 'auto_supplement_target'
    })
    if (item) {
      const raw = item.value || item.Value || ''
      const num = parseFloat(raw)
      if (!Number.isNaN(num) && num > 0) {
        targetMarginRate.value = num
      }
    }
  } catch (error) {
    console.error('加载目标定金率失败:', error)
  }
}

// 点击“排序规则”时，在几种排序方式间循环切换
const openSortSelector = () => {
  const order = ['time_desc', 'time_asc', 'deposit_desc', 'deposit_asc', 'pnl_desc', 'pnl_asc']
  const currentIndex = order.indexOf(sortRule.value)
  const nextIndex = (currentIndex + 1) % order.length
  sortRule.value = order[nextIndex] || 'time_desc'
}

// 点击列表卡片，弹出订单详情（对齐资金-交易详情弹窗）
const showPositionDetail = (position) => {
  if (!position) return
  currentDetailPosition.value = position
  showDetailDialog.value = true
}

const openOrderDetailFromRoute = async () => {
  const orderId = route.query.order_id
  if (!orderId) return

  try {
    const url = replaceParams(API_ENDPOINTS.ORDER_DETAIL, { id: orderId })
    const order = await request.get(url)

    const settledAt =
      order.settled_at || order.SettledAt || order.closed_at || order.ClosedAt || ''

    showDialog({
      title: `订单详情 #${order.order_id || orderId}`,
      message: `
锁定价格: ¥${formatMoney(order.locked_price)}/克
当前价格: ¥${formatMoney(order.current_price || quoteStore.currentPrice || order.locked_price)}/克
克重: ${order.weight_g}克
定金: ¥${formatMoney(order.deposit)}
浮动金额: ¥${formatMoney(order.pnl_float || 0)}
定金率: ${order.margin_rate != null ? order.margin_rate.toFixed(2) : '100.00'}%
下单时间: ${order.created_at || ''}
结算时间: ${settledAt || ''}
      `,
      confirmButtonText: '关闭'
    })
  } catch (error) {
    console.error('加载订单详情失败:', error)
  }
}

onMounted(() => {
  quoteStore.connectWebSocket()
  loadUserInfo()
  loadTargetMarginRate()
  loadPositions()
  openOrderDetailFromRoute()
})
</script>

<style scoped>
.positions-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

/* 过滤 Tab 样式：对齐资金-交易 */
.trade-filters {
  padding: 8px 12px 0;
}

.trade-type-tabs,
.trade-status-tabs {
  display: flex;
  background: #fff;
  border-radius: 16px;
  overflow: hidden;
  margin-bottom: 8px;
}

.trade-type-tab,
.trade-status-tab {
  flex: 1;
  text-align: center;
  padding: 6px 0;
  font-size: 14px;
  color: #666;
}

.trade-type-tab.active,
.trade-status-tab.active {
  background: #1989fa;
  color: #fff;
  font-weight: 500;
}

/* 汇总卡片样式：对齐资金-交易 */
.trade-summary-card {
  margin: 0 12px 8px;
  padding: 10px 12px;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  display: flex;
  justify-content: space-between;
}

.trade-summary-item {
  flex: 1;
  text-align: center;
}

.trade-summary-item .label {
  font-size: 12px;
  opacity: 0.85;
  margin-bottom: 4px;
}

.trade-summary-item .value {
  font-size: 15px;
  font-weight: 600;
}

.trade-summary-item .value.profit {
  color: #f56c6c;
}

.trade-summary-item .value.loss {
  color: #67c23a;
}

.orders-container {
  padding: 0 12px 16px;
}

/* 卡片式记录列表：对齐资金-交易 */
.record-card {
  background: #fff;
  margin: 12px;
  padding: 16px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  cursor: pointer;
  transition: all 0.3s;
}

.record-card:active {
  transform: scale(0.98);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

.record-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.record-type-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 500;
}

.type-trade-long {
  background: #f56c6c;
  color: #ffffff;
}

.type-trade-short {
  background: #67c23a;
  color: #ffffff;
}

.record-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.record-status.holding {
  background: #fff3e0;
  color: #ff9800;
}

.record-status.settled {
  background: #e8f5e9;
  color: #4caf50;
}

.record-status.closed {
  background: #ffebee;
  color: #f56c6c;
}

.record-card-body {
  margin-bottom: 12px;
}

.record-info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  font-size: 14px;
}

.record-info-row .label {
  color: #999;
}

.record-info-row .value {
  color: #333;
  font-weight: 500;
  text-align: right;
  max-width: 60%;
}

.record-info-row .amount-text {
  font-size: 16px;
  font-weight: bold;
}

.record-info-row .amount-text.income {
  color: #f56c6c;
}

.record-info-row .amount-text.expense {
  color: #67c23a;
}

.record-info-row .desc-text {
  font-size: 13px;
  color: #666;
}

.record-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.record-actions {
  display: inline-flex;
  gap: 8px;
}

.view-detail {
  font-size: 12px;
  color: #999;
}

.risk-tag {
  margin-left: 6px;
  padding: 2px 6px;
  border-radius: 10px;
  font-size: 11px;
  line-height: 1.4;
}

.risk-line {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.risk-rate {
  font-weight: 600;
}

.record-info-row .value.need-supplement-value {
  color: #1989fa;
  font-weight: 600;
}

.risk-rate-high {
  color: #f56c6c;
}

.risk-rate-low {
  color: #67c23a;
}

.risk-rate-cell-high :deep(.van-cell__value) {
  color: #f56c6c;
  font-weight: 600;
}

.risk-rate-cell-low :deep(.van-cell__value) {
  color: #67c23a;
  font-weight: 600;
}

.risk-tag.notice {
  background: #fff7e6;
  color: #faad14;
}

.risk-tag.warning {
  background: #fff1f0;
  color: #f5222d;
}

.risk-tag.danger {
  background: #fff1f0;
  color: #a8071a;
  font-weight: 600;
}

/* 详情弹窗样式：对齐资金-交易 */
.detail-popup {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.detail-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.detail-actions {
  margin: 16px 16px 0;
}

.pnl-profit-cell :deep(.van-cell__value) {
  color: #f56c6c !important;
  font-weight: bold;
}

.pnl-loss-cell :deep(.van-cell__value) {
  color: #67c23a !important;
  font-weight: bold;
}

.need-supplement-cell :deep(.van-cell__value) {
  color: #1989fa;
  font-weight: 600;
}

.empty {
  padding: 40px 0;
}
</style>
