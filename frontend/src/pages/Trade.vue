<template>
  <div class="trade-page">
    <van-nav-bar
      title="交易"
      fixed
      placeholder
    />
    
    <!-- 行情信息 -->
    <div class="quote-info">
      <div class="price">
        ¥{{ quoteStore.priceDisplay }}<span class="unit">/克</span>
      </div>
      <div class="change" :class="{ up: quoteStore.isUp, down: quoteStore.isDown }">
        {{ quoteStore.priceChange >= 0 ? '+' : '' }}{{ quoteStore.priceChange.toFixed(2) }}
        ({{ quoteStore.priceChangePercent >= 0 ? '+' : '' }}{{ quoteStore.priceChangePercent.toFixed(2) }}%)
      </div>
    </div>
    
    <!-- 买卖切换 -->
    <van-tabs v-model:active="tradeType">
      <van-tab title="买入" name="buy">
        <van-form @submit="onSubmit">
          <van-cell-group inset>
            <van-field
              v-model="form.price"
              type="number"
              label="买入价格"
              placeholder="请输入价格"
              :rules="[{ required: true, message: '请输入价格' }]"
            >
              <template #button>
                <van-button size="small" type="primary" @click="form.price = quoteStore.currentPrice">
                  市价
                </van-button>
              </template>
            </van-field>
            
            <van-field
              v-model="form.amount"
              type="number"
              label="买入数量"
              placeholder="请输入数量（克）"
              :rules="[{ required: true, message: '请输入数量' }]"
            />
            
            <van-field
              v-model="totalAmount"
              label="总金额"
              readonly
            />
            
            <van-field
              v-model="requiredDeposit"
              label="所需定金"
              readonly
            >
              <template #extra>
                <span style="color: #999; font-size: 12px;">定金率: {{ (config.deposit_rate * 100).toFixed(0) }}%</span>
              </template>
            </van-field>
          </van-cell-group>
          
          <div class="balance-info">
            <span>可用余额: ¥{{ formatMoney(balance.available_amount) }}</span>
          </div>
          
          <div style="margin: 16px;">
            <van-button
              round
              block
              type="danger"
              native-type="submit"
              :loading="loading"
            >
              买入
            </van-button>
          </div>
        </van-form>
      </van-tab>
      
      <van-tab title="卖出" name="sell">
        <van-form @submit="onSubmit">
          <van-cell-group inset>
            <van-field
              v-model="form.price"
              type="number"
              label="卖出价格"
              placeholder="请输入价格"
              :rules="[{ required: true, message: '请输入价格' }]"
            >
              <template #button>
                <van-button size="small" type="primary" @click="form.price = quoteStore.currentPrice">
                  市价
                </van-button>
              </template>
            </van-field>
            
            <van-field
              v-model="form.amount"
              type="number"
              label="卖出数量"
              placeholder="请输入数量（克）"
              :rules="[{ required: true, message: '请输入数量' }]"
            />
            
            <van-field
              v-model="totalAmount"
              label="总金额"
              readonly
            />
          </van-cell-group>
          
          <div class="balance-info">
            <span>可卖出: {{ formatMoney(availableSellAmount) }}克</span>
          </div>
          
          <div style="margin: 16px;">
            <van-button
              round
              block
              type="success"
              native-type="submit"
              :loading="loading"
            >
              卖出
            </van-button>
          </div>
        </van-form>
      </van-tab>
    </van-tabs>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showDialog } from 'vant'
import { useQuoteStore } from '../stores/quote'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'
import { formatMoney } from '../utils/helpers'

const route = useRoute()
const router = useRouter()
const quoteStore = useQuoteStore()

const tradeType = ref(route.query.type || 'buy')
const form = ref({
  price: '',
  amount: ''
})
const loading = ref(false)
const config = ref({
  deposit_rate: 0.1
})
const balance = ref({
  available_amount: 0
})
const availableSellAmount = ref(0)

// 计算总金额
const totalAmount = computed(() => {
  const price = parseFloat(form.value.price) || 0
  const amount = parseFloat(form.value.amount) || 0
  return formatMoney(price * amount)
})

// 计算所需定金
const requiredDeposit = computed(() => {
  const price = parseFloat(form.value.price) || 0
  const amount = parseFloat(form.value.amount) || 0
  const total = price * amount
  return formatMoney(total * config.value.deposit_rate)
})

// 监听交易类型变化
watch(tradeType, () => {
  form.value = { price: '', amount: '' }
})

// 获取配置
const loadConfig = async () => {
  try {
    const { data } = await request.get(API_ENDPOINTS.CONFIG)
    config.value = data
  } catch (error) {
    console.error('获取配置失败:', error)
  }
}

// 获取余额
const loadBalance = async () => {
  try {
    const { data } = await request.get(API_ENDPOINTS.USER_PROFILE)
    balance.value = data
    
    // 获取可卖出数量（持仓中的数量）
    const { data: positions } = await request.get(API_ENDPOINTS.POSITIONS, {
      params: { status: 'holding' }
    })
    availableSellAmount.value = (positions.list || []).reduce((sum, pos) => sum + pos.amount, 0)
  } catch (error) {
    console.error('获取余额失败:', error)
  }
}

// 提交订单
const onSubmit = async () => {
  try {
    loading.value = true
    
    const endpoint = tradeType.value === 'buy' ? API_ENDPOINTS.ORDER_BUY : API_ENDPOINTS.ORDER_SELL
    const { data } = await request.post(endpoint, {
      price: parseFloat(form.value.price),
      amount: parseFloat(form.value.amount)
    })
    
    showDialog({
      title: '下单成功',
      message: `订单已提交，等待确认`,
      confirmButtonText: '查看订单'
    }).then(() => {
      router.push(`/orders/${data.id}`)
    })
  } catch (error) {
    console.error('下单失败:', error)
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
  background-color: #f7f8fa;
}

.quote-info {
  background: #fff;
  padding: 20px;
  text-align: center;
  margin-bottom: 10px;
}

.price {
  font-size: 32px;
  font-weight: bold;
  color: #333;
  margin-bottom: 8px;
}

.unit {
  font-size: 14px;
  color: #999;
}

.change {
  font-size: 14px;
}

.change.up {
  color: #f56c6c;
}

.change.down {
  color: #67c23a;
}

.balance-info {
  padding: 12px 16px;
  color: #666;
  font-size: 14px;
  text-align: right;
}
</style>
