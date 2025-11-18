<template>
  <div class="funds-page">
    <van-nav-bar
      title="资金"
      fixed
      placeholder
    />
    
    <!-- 资金概览 -->
    <div class="balance-card">
      <div class="balance-item">
        <div class="label">总资产</div>
        <div class="amount">¥{{ formatMoney(userInfo.total_amount) }}</div>
      </div>
      <div class="balance-row">
        <div class="balance-item">
          <div class="label">可用余额</div>
          <div class="amount">¥{{ formatMoney(userInfo.available_amount) }}</div>
        </div>
        <div class="balance-item">
          <div class="label">冻结资金</div>
          <div class="amount">¥{{ formatMoney(userInfo.frozen_amount) }}</div>
        </div>
      </div>
      
      <div class="actions">
        <van-button type="primary" size="small" @click="showDeposit = true">
          充值
        </van-button>
        <van-button plain type="primary" size="small" @click="showWithdraw = true">
          提现
        </van-button>
      </div>
    </div>
    
    <!-- 资金流水 -->
    <van-tabs v-model:active="activeTab" @change="loadRecords">
      <van-tab title="全部" name="all" />
      <van-tab title="充值" name="deposit" />
      <van-tab title="提现" name="withdraw" />
      <van-tab title="交易" name="trade" />
    </van-tabs>
    
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        finished-text="没有更多了"
        @load="loadRecords"
      >
        <div v-if="records.length === 0" class="empty">
          <van-empty description="暂无记录" />
        </div>
        
        <div v-for="record in records" :key="record.id" class="record-item">
          <div class="record-header">
            <span class="record-type">{{ getRecordTypeText(record.type) }}</span>
            <span class="record-amount" :class="{ income: record.amount > 0, expense: record.amount < 0 }">
              {{ record.amount > 0 ? '+' : '' }}{{ formatMoney(record.amount) }}
            </span>
          </div>
          <div class="record-body">
            <div class="record-desc">{{ record.description }}</div>
            <div class="record-time">{{ formatDateTime(record.created_at) }}</div>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>
    
    <!-- 充值弹窗 -->
    <van-popup v-model:show="showDeposit" position="bottom" round>
      <div class="popup-content">
        <div class="popup-header">
          <h3>充值</h3>
        </div>
        <van-form @submit="onDeposit">
          <van-field
            v-model="depositForm.amount"
            type="number"
            label="充值金额"
            placeholder="请输入充值金额"
            :rules="[{ required: true, message: '请输入充值金额' }]"
          />
          <van-field
            v-model="depositForm.bank_card_id"
            label="银行卡"
            placeholder="请选择银行卡"
            readonly
            is-link
            @click="showBankCardPicker = true"
            :rules="[{ required: true, message: '请选择银行卡' }]"
          />
          <div style="margin: 16px;">
            <van-button round block type="primary" native-type="submit">
              确认充值
            </van-button>
          </div>
        </van-form>
      </div>
    </van-popup>
    
    <!-- 提现弹窗 -->
    <van-popup v-model:show="showWithdraw" position="bottom" round>
      <div class="popup-content">
        <div class="popup-header">
          <h3>提现</h3>
        </div>
        <van-form @submit="onWithdraw">
          <van-field
            v-model="withdrawForm.amount"
            type="number"
            label="提现金额"
            placeholder="请输入提现金额"
            :rules="[{ required: true, message: '请输入提现金额' }]"
          >
            <template #extra>
              <span style="color: #999; font-size: 12px;">
                可用: ¥{{ formatMoney(userInfo.available_amount) }}
              </span>
            </template>
          </van-field>
          <van-field
            v-model="withdrawForm.bank_card_id"
            label="银行卡"
            placeholder="请选择银行卡"
            readonly
            is-link
            @click="showBankCardPicker = true"
            :rules="[{ required: true, message: '请选择银行卡' }]"
          />
          <div style="margin: 16px;">
            <van-button round block type="primary" native-type="submit">
              确认提现
            </van-button>
          </div>
        </van-form>
      </div>
    </van-popup>
    
    <!-- 银行卡选择 -->
    <van-action-sheet v-model:show="showBankCardPicker" title="选择银行卡">
      <div class="bank-card-list">
        <div
          v-for="card in bankCards"
          :key="card.id"
          class="bank-card-item"
          @click="selectBankCard(card)"
        >
          <div class="card-info">
            <div class="bank-name">{{ card.bank_name }}</div>
            <div class="card-number">**** **** **** {{ card.card_number.slice(-4) }}</div>
          </div>
        </div>
        <div v-if="bankCards.length === 0" class="empty-tip">
          暂无银行卡，请先添加
        </div>
      </div>
    </van-action-sheet>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'
import { formatMoney, formatDateTime } from '../utils/helpers'

const activeTab = ref('all')
const records = ref([])
const refreshing = ref(false)
const loading = ref(false)
const finished = ref(false)
const page = ref(1)

const userInfo = ref({
  total_amount: 0,
  available_amount: 0,
  frozen_amount: 0
})

const showDeposit = ref(false)
const showWithdraw = ref(false)
const showBankCardPicker = ref(false)
const bankCards = ref([])

const depositForm = ref({
  amount: '',
  bank_card_id: ''
})

const withdrawForm = ref({
  amount: '',
  bank_card_id: ''
})

// 获取记录类型文本
const getRecordTypeText = (type) => {
  const types = {
    deposit: '充值',
    withdraw: '提现',
    buy: '买入',
    sell: '卖出',
    profit: '盈利',
    loss: '亏损',
    commission: '提成',
    supplement_deposit: '补交定金'
  }
  return types[type] || type
}

// 加载用户信息
const loadUserInfo = async () => {
  try {
    const { data } = await request.get(API_ENDPOINTS.USER_PROFILE)
    userInfo.value = data
  } catch (error) {
    console.error('加载用户信息失败:', error)
  }
}

// 加载资金流水
const loadRecords = async () => {
  try {
    const params = {
      page: page.value,
      page_size: 10
    }
    
    if (activeTab.value !== 'all') {
      params.type = activeTab.value
    }
    
    const { data } = await request.get(API_ENDPOINTS.FUND_FLOW, { params })
    
    if (page.value === 1) {
      records.value = data.list || []
    } else {
      records.value.push(...(data.list || []))
    }
    
    if (!data.list || data.list.length < 10) {
      finished.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载资金流水失败:', error)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

// 加载银行卡列表
const loadBankCards = async () => {
  try {
    const { data } = await request.get(API_ENDPOINTS.BANK_CARDS)
    bankCards.value = data.list || []
  } catch (error) {
    console.error('加载银行卡失败:', error)
  }
}

// 选择银行卡
const selectBankCard = (card) => {
  if (showDeposit.value) {
    depositForm.value.bank_card_id = card.id
  } else {
    withdrawForm.value.bank_card_id = card.id
  }
  showBankCardPicker.value = false
}

// 充值
const onDeposit = async () => {
  try {
    await request.post(API_ENDPOINTS.DEPOSIT_CREATE, {
      amount: parseFloat(depositForm.value.amount),
      bank_card_id: depositForm.value.bank_card_id
    })
    
    showToast('充值申请已提交，等待审核')
    showDeposit.value = false
    depositForm.value = { amount: '', bank_card_id: '' }
    
    loadUserInfo()
    onRefresh()
  } catch (error) {
    console.error('充值失败:', error)
  }
}

// 提现
const onWithdraw = async () => {
  try {
    await request.post(API_ENDPOINTS.WITHDRAW_CREATE, {
      amount: parseFloat(withdrawForm.value.amount),
      bank_card_id: withdrawForm.value.bank_card_id
    })
    
    showToast('提现申请已提交，等待审核')
    showWithdraw.value = false
    withdrawForm.value = { amount: '', bank_card_id: '' }
    
    loadUserInfo()
    onRefresh()
  } catch (error) {
    console.error('提现失败:', error)
  }
}

// 下拉刷新
const onRefresh = () => {
  page.value = 1
  finished.value = false
  loadRecords()
}

onMounted(() => {
  loadUserInfo()
  loadRecords()
  loadBankCards()
})
</script>

<style scoped>
.funds-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}

.balance-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 30px 20px;
  color: #fff;
  margin-bottom: 10px;
}

.balance-item {
  margin-bottom: 20px;
}

.balance-item .label {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 8px;
}

.balance-item .amount {
  font-size: 32px;
  font-weight: bold;
}

.balance-row {
  display: flex;
  justify-content: space-between;
}

.balance-row .balance-item .amount {
  font-size: 20px;
}

.actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.actions .van-button {
  flex: 1;
}

.record-item {
  background: #fff;
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.record-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.record-type {
  font-size: 16px;
  font-weight: 500;
}

.record-amount {
  font-size: 18px;
  font-weight: bold;
}

.record-amount.income {
  color: #f56c6c;
}

.record-amount.expense {
  color: #67c23a;
}

.record-body {
  font-size: 14px;
  color: #999;
}

.record-desc {
  margin-bottom: 4px;
}

.popup-content {
  padding: 20px;
}

.popup-header {
  text-align: center;
  margin-bottom: 20px;
}

.popup-header h3 {
  margin: 0;
  font-size: 18px;
}

.bank-card-list {
  padding: 16px;
}

.bank-card-item {
  padding: 16px;
  background: #f7f8fa;
  border-radius: 8px;
  margin-bottom: 12px;
}

.bank-name {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 8px;
}

.card-number {
  font-size: 14px;
  color: #666;
}

.empty-tip {
  text-align: center;
  color: #999;
  padding: 40px 0;
}

.empty {
  padding: 40px 0;
}
</style>
