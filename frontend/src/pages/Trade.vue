<template>
  <div class="trade-page">
    <van-nav-bar
      :title="tradeType === 'buy' ? '锁价买料' : '锁价卖料'"
      left-arrow
      @click-left="$router.back()"
      fixed
      placeholder
    />
    
    <!-- 价格卡片 -->
    <div class="price-card" :class="tradeType">
      <div class="price-value">
        {{ tradeType === 'buy' ? quoteStore.buyPriceDisplay : quoteStore.sellPriceDisplay }}
      </div>
      <div class="price-label">
        {{ tradeType === 'buy' ? '黄金销售价(元/克)' : '黄金回购价(元/克)' }}
      </div>
    </div>
    
    <!-- 买卖切换 -->
    <van-tabs v-model:active="tradeType" color="#f44" title-active-color="#f44">
      <van-tab title="锁价买料" name="buy">
        <div class="trade-form">
          <!-- 买入克重提示 -->
          <div class="tip-row">
            <span class="label">买入克重（需要限制最少于{{ config.min_order_amount }}克）</span>
            <span class="tip">最终以实际报出货重量为准</span>
          </div>
          
          <!-- 克重输入 -->
          <van-field
            v-model="form.amount"
            type="number"
            placeholder="请输入克重"
            class="amount-input"
          />
          <div v-if="form.amount && form.amount < config.min_order_amount" class="error-tip">
            买入克重不能少于{{ config.min_order_amount }}
          </div>
          
          <!-- 快捷克重选择 -->
          <div class="quick-amounts">
            <van-button 
              v-for="amount in quickAmounts" 
              :key="amount"
              size="small"
              plain
              @click="form.amount = amount"
            >
              {{ amount }}g
            </van-button>
          </div>
          
          <!-- 费用明细 -->
          <van-cell-group>
            <van-cell title="预估金额" :value="'¥' + formatMoney(estimatedAmount)" />
            <van-cell 
              title="总服务费(按实收重量收取)" 
              :value="'¥' + formatMoney(serviceFee)" 
              value-class="highlight"
            />
            <van-cell title="定金" :value="'¥' + formatMoney(requiredDepositValue)" />
            <van-cell title="可用定金" :value="'¥' + formatMoney(balance.available_deposit)" />
          </van-cell-group>
          
          <!-- 业务说明 -->
          <div class="business-note">
            <div class="note-title">业务说明</div>
            <div class="note-content">
              当客户买料价格在客户下单价格，延后提货卖料，我司依会约定取款支付取货款及服务费，客户收货当天需要完成支付。
            </div>
          </div>
          
          <!-- 协议 -->
          <van-checkbox v-model="agreeProtocol" class="protocol-check">
            我已阅读并同意
            <span class="link" @click.stop.prevent="openAgreement">《贵金属购销服务协议》</span>
          </van-checkbox>
          
          <!-- 提交按钮 -->
          <van-button
            type="danger"
            size="large"
            round
            block
            :loading="loading"
            :disabled="!agreeProtocol || !form.amount || form.amount < config.min_order_amount"
            @click="onSubmit"
            class="submit-btn"
          >
            立即买入
          </van-button>
        </div>
      </van-tab>
      
      <van-tab title="锁价卖料" name="sell">
        <div class="trade-form">
          <!-- 卖出克重提示 -->
          <div class="tip-row">
            <span class="label">卖出克重</span>
            <span class="tip">长线以实测并足量者结算</span>
          </div>
          
          <!-- 克重输入 -->
          <van-field
            v-model="form.amount"
            type="number"
            placeholder="请输入克重"
            class="amount-input"
          />
          <div v-if="form.amount && form.amount < config.min_order_amount" class="error-tip">
            卖出克重不能少于{{ config.min_order_amount }}
          </div>
          
          <!-- 快捷克重选择 -->
          <div class="quick-amounts">
            <van-button 
              v-for="amount in quickAmounts" 
              :key="amount"
              size="small"
              plain
              @click="form.amount = amount"
            >
              {{ amount }}g
            </van-button>
          </div>
          
          <!-- 费用明细 -->
          <van-cell-group>
            <van-cell title="预估金额" :value="'¥' + formatMoney(estimatedAmount)" />
            <van-cell 
              title="总服务费(按实收重量收取)" 
              :value="'¥' + formatMoney(serviceFee)" 
              value-class="highlight"
            />
            <van-cell title="定金" :value="'¥' + formatMoney(requiredDepositValue)" />
            <van-cell title="可用定金" :value="'¥' + formatMoney(balance.available_deposit)" />
          </van-cell-group>
          
          <!-- 收货地址 -->
          <div class="address-section">
            <div class="section-title">收货地址</div>
            <van-cell
              icon="location-o"
              :title="userAddress.name || '请设置收货地址'"
              :label="userAddress.phone ? `${userAddress.phone}\n${userAddress.address}` : ''"
              is-link
              @click="showAddressPopup = true"
            />
          </div>
          
          <!-- 锁价卖料流程 -->
          <div class="process-section">
            <div class="section-title">锁价卖料流程</div>
            <div class="process-steps">
              <div class="step">
                <div class="step-icon">📱</div>
                <div class="step-text">在线锁价</div>
              </div>
              <div class="step-arrow">···></div>
              <div class="step">
                <div class="step-icon">📦</div>
                <div class="step-text">顺丰保价</div>
              </div>
              <div class="step-arrow">···></div>
              <div class="step">
                <div class="step-icon">🔬</div>
                <div class="step-text">检测报告</div>
              </div>
              <div class="step-arrow">···></div>
              <div class="step">
                <div class="step-icon">💰</div>
                <div class="step-text">结算付款</div>
              </div>
            </div>
          </div>
          
          <!-- 业务说明 -->
          <div class="business-note">
            <div class="note-title">业务说明</div>
            <div class="note-content">
              当客户对外卖价，卖料由客户自身可得利润，可主动联系我司购已卖出支付至期待的定金及服务费，客户发货当天需交结款尾款及承诺一定的结算补足率及服务费。
            </div>
          </div>
          
          <!-- 协议 -->
          <van-checkbox v-model="agreeProtocol" class="protocol-check">
            我已阅读并同意
            <span class="link" @click.stop.prevent="openAgreement">《贵金属购销服务协议》</span>
          </van-checkbox>
          
          <!-- 提交按钮 -->
          <van-button
            type="success"
            size="large"
            round
            block
            :loading="loading"
            :disabled="!agreeProtocol || !form.amount || form.amount < config.min_order_amount"
            @click="onSubmit"
            class="submit-btn"
          >
            立即卖出
          </van-button>
        </div>
      </van-tab>
    </van-tabs>

    <!-- 贵金属购销服务协议弹窗 -->
    <van-popup
      v-model:show="showAgreementPopup"
      position="bottom"
      :style="{ height: '100%', width: '100%' }"
    >
      <div class="agreement-popup">
        <div class="agreement-title">贵金属购销服务协议</div>
        <div class="agreement-body">
          <iframe
            class="agreement-frame"
            :src="config.metal_service_agreement_url || defaultMetalAgreementUrl"
            frameborder="0"
          ></iframe>
        </div>
        <div class="agreement-footer">
          <van-button type="danger" block round @click="onAgreementConfirm">
            确定
          </van-button>
        </div>
      </div>
    </van-popup>

    <!-- 确认订单弹窗（价格与金额实时变化） -->
    <van-dialog
      v-model:show="showConfirmDialog"
      title="确认订单"
      show-cancel-button
      confirm-button-text="确认"
      cancel-button-text="取消"
      :close-on-click-overlay="false"
      :show-confirm-button="true"
      :show-cancel-button="true"
      @confirm="handleConfirmOrder"
    >
      <div class="order-confirm-content">
        <div class="order-confirm-row">
          <span class="label">订单类型：</span>
          <span class="value">{{ tradeType === 'buy' ? '买入' : '卖出' }}</span>
        </div>
        <div class="order-confirm-row">
          <span class="label">下单品类：</span>
          <span class="value">黄金板料</span>
        </div>
        <div class="order-confirm-row">
          <span class="label">锁定单价：</span>
          <span class="value strong">
            {{
              (tradeType === 'buy' ? quoteStore.buyPrice : quoteStore.sellPrice) > 0
                ? (tradeType === 'buy' ? quoteStore.buyPrice : quoteStore.sellPrice).toFixed(2)
                : '--'
            }}
            元/克
          </span>
        </div>
        <div class="order-confirm-row">
          <span class="label">下单重量：</span>
          <span class="value">{{ form.amount }} 克</span>
        </div>
        <div class="order-confirm-row">
          <span class="label">预估金额：</span>
          <span class="value">{{ estimatedAmount > 0 ? estimatedAmount.toFixed(2) : '0.00' }} 元</span>
        </div>
        <div class="order-confirm-row">
          <span class="label">总服务费(按实收重量收取)：</span>
          <span class="value">{{ serviceFee > 0 ? serviceFee.toFixed(2) : '0.00' }} 元</span>
        </div>
        <div class="order-confirm-row">
          <span class="label">定金：</span>
          <span class="value">{{ requiredDepositValue > 0 ? requiredDepositValue.toFixed(2) : '0.00' }} 元</span>
        </div>
        <div class="order-confirm-hint">
          最终锁定价格以点击确认时的实时价格为准
        </div>
      </div>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showDialog, showToast } from 'vant'
import { useQuoteStore } from '../stores/quote'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'
import { formatMoney } from '../utils/helpers'

const route = useRoute()
const router = useRouter()
const quoteStore = useQuoteStore()

// 贵金属购销服务协议默认链接（系统未配置时兜底）
const defaultMetalAgreementUrl = 'https://j.kingsoftsys.com/h5/#/pages/agreement/index?demp_code=b6b85d1aec49b7db7228ce'

// 交易类型：buy=锁价买料，sell=锁价卖料
const tradeType = ref(route.query.type || 'buy')

const form = ref({
  amount: ''
})

const loading = ref(false)
const agreeProtocol = ref(false)
const showAddressPopup = ref(false)
const showAgreementPopup = ref(false)
const showConfirmDialog = ref(false)

// 配置：从后台配置中心读取
const config = ref({
  deposit_rate: 0.1,
  min_order_amount: 100,
  service_fee_rate: 0.02, // 兼容旧字段（按金额比例），优先使用交割服务费
  deposit_per_gram: 10,   // 每克定金10元
  delivery_fee_per_gram: 0, // 交割服务费（元/克）
  metal_service_agreement_url: '', // 贵金属购销服务协议链接
  trading_start_time: '',      // 交易开始时间，例如 "09:00"
  trading_end_time: '',        // 交易结束时间，例如 "18:00"
  trading_days: '',            // 交易日列表，例如 "1,2,3,4,5"
  holiday_trading_enabled: '1', // 节假日是否交易：'1' 允许，'0' 休市
  holiday_closed_dates: ''      // 节假日休市日期列表，格式 YYYY-MM-DD,逗号分隔
})

const balance = ref({
  available_deposit: 0,
  used_deposit: 0
})

const userAddress = ref({
  name: '',
  phone: '',
  address: ''
})

// 快捷克重选项
const quickAmounts = [1000, 2000, 3000, 5000, 100, 200, 300, 500]

// 计算预估金额：买入用销售价，卖出用回购价
const estimatedAmount = computed(() => {
  const amount = parseFloat(form.value.amount) || 0
  const price = tradeType.value === 'buy' ? quoteStore.buyPrice : quoteStore.sellPrice
  return amount * price
})

// 计算总服务费：交易克重 * 交割服务费（元/克）
// 交割服务费仅从 delivery_fee_per_gram 读取，未配置则视为 0
const serviceFee = computed(() => {
  const amount = parseFloat(form.value.amount) || 0
  const feePerGram = parseFloat(config.value.delivery_fee_per_gram) || 0
  return amount * feePerGram
})

// 计算所需定金：克重 * 每克定金
const requiredDepositValue = computed(() => {
  const amount = parseFloat(form.value.amount) || 0
  return amount * (config.value.deposit_per_gram || 10)
})

// 切换买入/卖出时重置表单与协议勾选
watch(tradeType, () => {
  form.value = { amount: '' }
  agreeProtocol.value = false
})

// 获取后台配置
const loadConfig = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.CONFIG)
    if (data.configs && Array.isArray(data.configs)) {
      data.configs.forEach(item => {
        const key = item.key || item.Key
        const value = item.value || item.Value
        if (!key) return

        if (key === 'deposit_rate') {
          config.value.deposit_rate = parseFloat(value) || 0.1
        }
        if (key === 'min_order_amount') {
          config.value.min_order_amount = parseFloat(value) || 100
        }
        if (key === 'service_fee_rate') {
          config.value.service_fee_rate = parseFloat(value) || 0.02
        }
        if (key === 'deposit_per_gram') {
          config.value.deposit_per_gram = parseFloat(value) || 10
        }
        if (key === 'delivery_fee_per_gram') {
          config.value.delivery_fee_per_gram = parseFloat(value) || 0
        }
        if (key === 'metal_service_agreement_url') {
          config.value.metal_service_agreement_url = value || ''
        }
        if (key === 'trading_start_time') {
          config.value.trading_start_time = value || ''
        }
        if (key === 'trading_end_time') {
          config.value.trading_end_time = value || ''
        }
        if (key === 'trading_days') {
          config.value.trading_days = value || ''
        }
        if (key === 'holiday_trading_enabled') {
          config.value.holiday_trading_enabled = value ?? '1'
        }
        if (key === 'holiday_closed_dates') {
          config.value.holiday_closed_dates = value || ''
        }
      })
    }
  } catch (error) {
    console.error('获取配置失败:', error)
    // 使用默认配置兜底（保留已有默认值）
  }
}

// 校验当前是否处于允许交易的时间与交易日内
const checkTradingStatus = () => {
  const cfg = config.value || {}

  // 1. 节假日开关：'0' 表示休市
  if (cfg.holiday_trading_enabled === '0') {
    return {
      open: false,
      message: '当前为节假日休市，暂不支持交易'
    }
  }

  const now = new Date()

  // 2. 节假日日期列表：holiday_closed_dates，格式 YYYY-MM-DD,逗号分隔
  const closedStr = cfg.holiday_closed_dates || ''
  if (closedStr) {
    const y = now.getFullYear()
    const m = String(now.getMonth() + 1).padStart(2, '0')
    const d = String(now.getDate()).padStart(2, '0')
    const todayStr = `${y}-${m}-${d}`
    const closedList = String(closedStr)
      .split(',')
      .map((s) => s.trim())
      .filter((s) => s)
    if (closedList.includes(todayStr)) {
      return {
        open: false,
        message: '当前为节假日休市，暂不支持交易'
      }
    }
  }

  // 3. 校验交易日（1-7 表示周一到周日）
  const jsDay = now.getDay() // 0=周日,1=周一,...,6=周六
  const weekday = jsDay === 0 ? 7 : jsDay
  const tradingDaysStr = cfg.trading_days || ''
  if (tradingDaysStr) {
    const days = String(tradingDaysStr)
      .split(',')
      .map((s) => parseInt(s.trim(), 10))
      .filter((n) => !Number.isNaN(n))
    if (days.length && !days.includes(weekday)) {
      return {
        open: false,
        message: '当前非交易日，暂不支持交易'
      }
    }
  }

  const parseTimeToMinutes = (str, defaultMinutes) => {
    if (!str) return defaultMinutes
    const parts = String(str).split(':')
    const h = parseInt(parts[0], 10)
    const m = parseInt(parts[1], 10)
    if (Number.isNaN(h) || Number.isNaN(m)) return defaultMinutes
    return h * 60 + m
  }

  const nowMinutes = now.getHours() * 60 + now.getMinutes()
  // 默认全天可交易
  const startMinutes = parseTimeToMinutes(cfg.trading_start_time, 0)
  const endMinutes = parseTimeToMinutes(cfg.trading_end_time, 23 * 60 + 59)

  let inTime = false
  if (endMinutes <= startMinutes) {
    // 跨午夜区间：例如 20:00-06:00
    inTime = nowMinutes >= startMinutes || nowMinutes <= endMinutes
  } else {
    inTime = nowMinutes >= startMinutes && nowMinutes <= endMinutes
  }

  if (!inTime) {
    return {
      open: false,
      message: '当前非交易时间，暂不支持交易'
    }
  }

  return { open: true, message: '' }
}

// 获取余额
const loadBalance = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.USER_PROFILE)
    balance.value = {
      available_deposit: data.available_deposit || 0,
      used_deposit: data.used_deposit || 0
    }
  } catch (error) {
    console.error('获取余额失败:', error)
    balance.value = {
      available_deposit: 0,
      used_deposit: 0
    }
  }
}

// 打开协议弹窗
const openAgreement = () => {
  showAgreementPopup.value = true
}

// 协议弹窗中点击确定：关闭并勾选协议
const onAgreementConfirm = () => {
  showAgreementPopup.value = false
  agreeProtocol.value = true
}

// 提交订单：先打开确认弹窗，价格在弹窗内实时变化
const onSubmit = () => {
  // 验证是否同意协议
  if (!agreeProtocol.value) {
    showToast('请先阅读并同意服务协议')
    return
  }

  // 验证是否在交易时间与交易日内
  const tradingStatus = checkTradingStatus()
  if (!tradingStatus.open) {
    showToast(tradingStatus.message || '当前为非交易时段，暂不支持交易')
    return
  }

  // 验证克重
  const amount = parseFloat(form.value.amount)
  if (!amount || amount < config.value.min_order_amount) {
    showToast(`最低${tradeType.value === 'buy' ? '买入' : '卖出'}克重为${config.value.min_order_amount}克`)
    return
  }

  // 验证当前价格是否可用（用于确认弹窗展示）
  const price = tradeType.value === 'buy' ? quoteStore.buyPrice : quoteStore.sellPrice
  if (!price || price <= 0) {
    showToast('无法获取当前价格，请稍后重试')
    return
  }

  showConfirmDialog.value = true
}

// 确认下单：点击确认订单弹窗中的“确认”
const handleConfirmOrder = async () => {
  // 再次校验交易时间与交易日
  const tradingStatus = checkTradingStatus()
  if (!tradingStatus.open) {
    showToast(tradingStatus.message || '当前为非交易时段，暂不支持交易')
    showConfirmDialog.value = false
    return
  }

  // 再次校验克重
  const amount = parseFloat(form.value.amount)
  if (!amount || amount < config.value.min_order_amount) {
    showToast(`最低${tradeType.value === 'buy' ? '买入' : '卖出'}克重为${config.value.min_order_amount}克`)
    showConfirmDialog.value = false
    return
  }

  // 用户点击"确认"后，再以当前行情价格作为真正的锁定价格
  const lockedPrice = tradeType.value === 'buy' ? quoteStore.buyPrice : quoteStore.sellPrice
  if (!lockedPrice || lockedPrice <= 0) {
    showToast('当前价格获取失败，请稍后重试')
    return
  }

  // 计算定金
  const deposit = amount * (config.value.deposit_per_gram || 10)

  try {
    loading.value = true

    // 转换订单类型：buy -> long_buy, sell -> short_sell
    const orderType = tradeType.value === 'buy' ? 'long_buy' : 'short_sell'

    const orderData = {
      type: orderType,          // long_buy 或 short_sell
      locked_price: lockedPrice,      // 锁定价格（以点击"确认"时的实时价格为准）
      weight_g: amount,         // 克重
      deposit: deposit          // 定金
    }

    console.log('📝 提交订单数据:', orderData)

    const data = await request.post(API_ENDPOINTS.ORDER_CREATE, orderData)

    showConfirmDialog.value = false

    showDialog({
      title: '下单成功',
      message: '订单已提交，等待确认',
      confirmButtonText: '查看订单'
    })
      .then(() => {
        const orderId = data.order_id || data.id
        router.push({ path: '/positions', query: { order_id: orderId } })
      })
      .catch(() => {
        router.push('/positions')
      })

    // 重新加载余额
    loadBalance()
    // 清空表单
    form.value = { amount: '' }
    agreeProtocol.value = false
  } catch (error) {
    console.error('下单失败:', error)
    console.error('错误详情:', error.response?.data)
    const errorMsg = error.response?.data?.error || error.response?.data?.message || '下单失败'
    showToast(errorMsg)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  quoteStore.connectWebSocket()
  loadConfig()
  loadBalance()
})
</script>

<style scoped>
.trade-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 20px;
}

/* 价格卡片 */
.price-card {
  margin: 16px;
  padding: 20px;
  border-radius: 8px;
  text-align: center;
  color: #fff;
}

.price-card.buy {
  background: linear-gradient(135deg, #ff6b6b, #ee5a52);
}

.price-card.sell {
  background: linear-gradient(135deg, #51cf66, #40c057);
}

.price-value {
  font-size: 36px;
  font-weight: bold;
  margin-bottom: 8px;
}

.price-label {
  font-size: 14px;
  opacity: 0.9;
}

/* 表单容器 */
.trade-form {
  padding: 16px;
}

/* 提示行 */
.tip-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.tip-row .label {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.tip-row .tip {
  font-size: 12px;
  color: #999;
}

/* 输入框 */
.amount-input {
  margin-bottom: 8px;
}

.amount-input :deep(.van-field__control) {
  font-size: 16px;
  font-weight: bold;
}

/* 错误提示 */
.error-tip {
  color: #ff4444;
  font-size: 12px;
  margin: -4px 0 12px 0;
  padding: 0 16px;
}

/* 快捷克重选择 */
.quick-amounts {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.quick-amounts .van-button {
  height: 36px;
}

/* Cell高亮 */
:deep(.highlight) {
  color: #ff6600 !important;
  font-weight: bold;
}

/* 章节标题 */
.section-title {
  font-size: 14px;
  font-weight: bold;
  color: #333;
  padding: 12px 0 8px;
  border-left: 3px solid #ff4444;
  padding-left: 8px;
  margin-top: 16px;
}

/* 地址区域 */
.address-section {
  margin-top: 16px;
}

/* 流程展示 */
.process-section {
  margin-top: 16px;
}

.process-steps {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.step-icon {
  font-size: 24px;
  margin-bottom: 4px;
}

.step-text {
  font-size: 12px;
  color: #666;
  text-align: center;
}

.step-arrow {
  color: #d4af37;
  font-size: 16px;
  padding: 0 4px;
}

/* 业务说明 */
.business-note {
  margin-top: 16px;
  padding: 12px;
  background: #fff7e6;
  border-left: 3px solid #ff4444;
  border-radius: 4px;
}

.note-title {
  font-size: 14px;
  font-weight: bold;
  color: #ff4444;
  margin-bottom: 8px;
}

.note-content {
  font-size: 12px;
  color: #666;
  line-height: 1.6;
}

/* 协议复选框 */
.protocol-check {
  margin: 16px 0;
  padding: 0 4px;
}

.protocol-check .link {
  color: #ff4444;
  text-decoration: underline;
}

/* 提交按钮 */
.submit-btn {
  margin-top: 16px;
  height: 48px;
  font-size: 16px;
  font-weight: bold;
}

/* 协议弹窗 */
.agreement-popup {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fff;
}

.agreement-title {
  padding: 12px 16px;
  text-align: center;
  font-size: 16px;
  font-weight: 600;
  border-bottom: 1px solid #f0f0f0;
}

.agreement-body {
  flex: 1;
  overflow: hidden;
}

.agreement-frame {
  width: 100%;
  height: 100%;
  border: none;
}

.agreement-footer {
  padding: 12px 16px 20px;
  border-top: 1px solid #f0f0f0;
  background: #fff;
}

/* 确认订单弹窗样式 */
.order-confirm-content {
  padding: 8px 4px 4px;
}

.order-confirm-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 2px 0;
  font-size: 14px;
}

.order-confirm-row .label {
  color: #333;
}

.order-confirm-row .value {
  color: #333;
}

.order-confirm-row .value.strong {
  font-weight: 600;
}

.order-confirm-hint {
  margin-top: 6px;
  text-align: center;
  font-size: 12px;
  color: #999;
}
</style>
