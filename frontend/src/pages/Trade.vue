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
      <van-tab title="买入" name="buy">
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
      
      <van-tab title="卖出" name="sell">
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
            src="https://j.kingsoftsys.com/h5/#/pages/agreement/index?demp_code=b6b85d1aec49b7db7228ce"
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

// 交易类型：buy=锁价买料，sell=锁价卖料
const tradeType = ref(route.query.type || 'buy')

const form = ref({
  amount: ''
})

const loading = ref(false)
const agreeProtocol = ref(false)
const showAddressPopup = ref(false)
const showAgreementPopup = ref(false)

// 配置：从后台配置中心读取
const config = ref({
  deposit_rate: 0.1,
  min_order_amount: 100,
  service_fee_rate: 0.02, // 兼容旧字段（按金额比例），优先使用交割服务费
  deposit_per_gram: 10,   // 每克定金10元
  delivery_fee_per_gram: 0 // 交割服务费（元/克）
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
// 交割服务费从 delivery_fee_per_gram 读取，若未配置则回退到 service_fee_rate
const serviceFee = computed(() => {
  const amount = parseFloat(form.value.amount) || 0
  const feePerGram =
    (config.value.delivery_fee_per_gram ?? 0) || (config.value.service_fee_rate ?? 0)
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
        if (item.key === 'deposit_rate') {
          config.value.deposit_rate = parseFloat(item.value) || 0.1
        }
        if (item.key === 'min_order_amount') {
          config.value.min_order_amount = parseFloat(item.value) || 100
        }
        if (item.key === 'service_fee_rate') {
          config.value.service_fee_rate = parseFloat(item.value) || 0.02
        }
        if (item.key === 'deposit_per_gram') {
          config.value.deposit_per_gram = parseFloat(item.value) || 10
        }
        if (item.key === 'delivery_fee_per_gram') {
          config.value.delivery_fee_per_gram = parseFloat(item.value) || 0
        }
      })
    }
  } catch (error) {
    console.error('获取配置失败:', error)
    // 使用默认配置兜底
    config.value = {
      deposit_rate: 0.1,
      min_order_amount: 100,
      service_fee_rate: 0.02,
      deposit_per_gram: 10,
      delivery_fee_per_gram: 0
    }
  }
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

// 提交订单
const onSubmit = async () => {
  // 验证是否同意协议
  if (!agreeProtocol.value) {
    showToast('请先阅读并同意服务协议')
    return
  }

  // 验证克重
  const amount = parseFloat(form.value.amount)
  if (!amount || amount < config.value.min_order_amount) {
    showToast(`最低${tradeType.value === 'buy' ? '买入' : '卖出'}克重为${config.value.min_order_amount}克`)
    return
  }

  // 验证价格（来自 WebSocket 实时报价）
  const price = tradeType.value === 'buy' ? quoteStore.buyPrice : quoteStore.sellPrice
  if (!price || price <= 0) {
    showToast('无法获取当前价格，请稍后重试')
    return
  }

  // 计算定金
  const deposit = amount * (config.value.deposit_per_gram || 10)

  // 订单确认弹窗（参考产品原型）
  const typeText = tradeType.value === 'buy' ? '买入' : '卖出'
  const productText = '黄金板料'
  const confirmMessage = `
    <div style="text-align:left;font-size:14px;line-height:1.6;">
      <div style="margin:4px 0;"><span>订单类型：</span><span style="float:right;">${typeText}</span></div>
      <div style="margin:4px 0;"><span>下单品类：</span><span style="float:right;">${productText}</span></div>
      <div style="margin:4px 0;"><span>实时报价：</span><span style="float:right;">${price.toFixed(2)} 元/克</span></div>
      <div style="margin:4px 0;"><span>下单重量：</span><span style="float:right;">${amount} 克</span></div>
      <div style="margin:4px 0;"><span>预估金额：</span><span style="float:right;">${estimatedAmount.value.toFixed(2)} 元</span></div>
      <div style="margin:4px 0;"><span>总服务费(按实收重量收取)：</span><span style="float:right;">${serviceFee.value.toFixed(2)} 元</span></div>
      <div style="margin:4px 0;"><span>定金：</span><span style="float:right;">${deposit.toFixed(2)} 元</span></div>
    </div>
  `

  const confirmed = await new Promise((resolve) => {
    showDialog({
      title: '确认订单',
      message: confirmMessage,
      showCancelButton: true,
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      allowHtml: true,
      beforeClose: (action) => {
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

  // 弹出支付密码输入框
  const payPassword = await new Promise((resolve) => {
    showDialog({
      title: '请输入支付密码',
      message: '请输入6位数字支付密码',
      showCancelButton: true,
      beforeClose: (action) => {
        if (action === 'confirm') {
          const input = document.querySelector('.van-dialog__message input')
          if (input) {
            resolve(input.value)
          } else {
            resolve(null)
          }
        } else {
          resolve(null)
        }
        return true
      }
    })
      .then(() => {
        // 点击确认
      })
      .catch(() => {
        // 点击取消
        resolve(null)
      })

    // 在 message 区域插入输入框
    setTimeout(() => {
      const messageEl = document.querySelector('.van-dialog__message')
      if (messageEl && !messageEl.querySelector('input')) {
        const input = document.createElement('input')
        input.type = 'password'
        input.maxLength = 6
        input.placeholder = '请输入6位数字'
        input.style.cssText = 'width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; margin-top: 8px; font-size: 16px;'
        messageEl.appendChild(input)
        input.focus()
      }
    }, 100)
  })

  if (!payPassword) {
    showToast('请输入支付密码')
    return
  }

  if (!/^\d{6}$/.test(payPassword)) {
    showToast('支付密码必须是6位数字')
    return
  }

  try {
    loading.value = true

    // 转换订单类型：buy -> long_buy, sell -> short_sell
    const orderType = tradeType.value === 'buy' ? 'long_buy' : 'short_sell'

    const orderData = {
      type: orderType,          // long_buy 或 short_sell
      locked_price: price,      // 锁定价格
      weight_g: amount,         // 克重
      deposit: deposit,         // 定金
      pay_password: payPassword // 支付密码
    }

    console.log('📝 提交订单数据:', orderData)

    const data = await request.post(API_ENDPOINTS.ORDER_CREATE, orderData)

    showDialog({
      title: '下单成功',
      message: '订单已提交，等待确认',
      confirmButtonText: '查看订单'
    })
      .then(() => {
        router.push(`/orders/${data.id || data.order_id}`)
      })
      .catch(() => {
        router.push('/orders')
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
</style>
