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
        <div class="label">总定金</div>
        <div class="amount">¥{{ formatMoney(totalDeposit) }}</div>
      </div>
      <div class="balance-row">
        <div class="balance-item">
          <div class="label">可用定金</div>
          <div class="amount">¥{{ formatMoney(userInfo.available_deposit) }}</div>
        </div>
        <div class="balance-item">
          <div class="label">持单定金</div>
          <div class="amount">¥{{ formatMoney(holdingMargin) }}</div>
        </div>
      </div>
      
      <div class="actions">
        <van-button type="primary" size="small" @click="showDeposit = true">
          付定金
        </van-button>
        <van-button plain type="primary" size="small" @click="showWithdraw = true">
          退定金
        </van-button>
      </div>
    </div>
    
    <!-- 资金流水 -->
    <van-tabs v-model:active="activeTab" @change="loadRecords">
      <van-tab title="全部" name="all" />
      <van-tab title="付定金" name="deposit" />
      <van-tab title="退定金" name="withdraw" />
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
        
        <div 
          v-for="record in records" 
          :key="record.id" 
          class="record-card"
          @click="showRecordDetail(record)"
        >
          <div class="record-card-header">
            <div
              class="record-type-badge"
              :class="[
                getRecordTypeClass(record.type),
                record.type === 'trade'
                  ? (record.order?.type === 'long_buy'
                      ? 'type-trade-long'
                      : (record.order?.type === 'short_sell' ? 'type-trade-short' : ''))
                  : ''
              ]"
            >
              {{
                record.type === 'trade'
                  ? '交易-' + getOrderTypeText(record.order?.type)
                  : getRecordTypeText(record.type)
              }}
            </div>
            <div class="record-status" :class="record.status">
              {{ getStatusText(record.status) }}
            </div>
          </div>
          
          <div class="record-card-body">
            <div class="record-info-row">
              <span class="label">订单号</span>
              <span class="value">#{{ record.id }}</span>
            </div>
            <div class="record-info-row" v-if="record.type === 'trade'">
              <span class="label">订单类型</span>
              <span class="value">{{ getOrderTypeText(record.order?.type) }}</span>
            </div>
            <div class="record-info-row" v-if="record.type === 'deposit'">
              <span class="label">充值前可用定金</span>
              <span class="value">¥{{ formatMoney(record.before_balance || 0) }}</span>
            </div>
            <div class="record-info-row">
              <span class="label">
                {{ record.type === 'deposit' ? '充值金额' : (record.type === 'trade' ? '定金' : '金额') }}
              </span>
              <span class="value amount-text" :class="{ income: record.amount > 0, expense: record.amount < 0 }">
                {{ record.amount > 0 ? '+' : '' }}¥{{ formatMoney(Math.abs(record.amount)) }}
              </span>
            </div>
            <div class="record-info-row" v-if="record.type === 'trade' && record.order">
              <span class="label">浮动盈亏</span>
              <span
                class="value amount-text"
                :class="{
                  income: calcOrderPnl(record.order) > 0,
                  expense: calcOrderPnl(record.order) < 0
                }"
              >
                {{ calcOrderPnl(record.order) > 0 ? '+' : '' }}¥{{ formatMoney(Math.abs(calcOrderPnl(record.order))) }}
              </span>
            </div>
            <div class="record-info-row" v-if="record.type === 'deposit'">
              <span class="label">充值后可用定金</span>
              <span class="value">¥{{ formatMoney(record.after_balance || 0) }}</span>
            </div>
            <div class="record-info-row">
              <span class="label">{{ record.type === 'deposit' ? '充值日期' : '日期' }}</span>
              <span class="value">{{ formatDateTime(record.created_at) }}</span>
            </div>
            <div class="record-info-row" v-if="record.description">
              <span class="label">备注</span>
              <span class="value desc-text">{{ record.description }}</span>
            </div>
          </div>
          
          <div class="record-card-footer">
            <span class="view-detail">查看详情 ></span>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>
    
    <!-- 付定金弹窗 -->
    <van-popup v-model:show="showDeposit" position="bottom" round :style="{ height: '90%' }">
      <div class="deposit-popup">
        <van-nav-bar
          title="钱包付定金"
          left-arrow
          @click-left="showDeposit = false"
        />
        
        <div class="deposit-content">
          <!-- 金额输入 -->
          <div class="amount-section">
            <div class="amount-label">金额</div>
            <van-field
              v-model="depositForm.amount"
              type="digit"
              placeholder="请输入金额"
              class="amount-input"
            />
          </div>
          
          <!-- 快捷金额 -->
          <div class="quick-amounts">
            <van-button 
              v-for="amount in quickAmounts" 
              :key="amount"
              size="small"
              :type="depositForm.amount == amount ? 'primary' : 'default'"
              @click="depositForm.amount = amount"
            >
              {{ formatQuickAmount(amount) }}
            </van-button>
          </div>
          
          <!-- 付款账户 -->
          <div class="section">
            <div class="section-title">付款账户</div>
            <van-cell
              title="选择付款账户"
              is-link
              :value="selectedPaymentCard ? selectedPaymentCard.bank_name : ''"
              @click="showPaymentCardPicker = true"
            />
            <div v-if="selectedPaymentCard" class="card-info">
              <div class="info-row">
                <span class="label">账户类型</span>
                <span class="value">银行卡</span>
              </div>
              <div class="info-row">
                <span class="label">户名</span>
                <span class="value">{{ selectedPaymentCard.card_holder }}</span>
              </div>
              <div class="info-row">
                <span class="label">产名</span>
                <span class="value">{{ selectedPaymentCard.bank_name }}</span>
              </div>
              <div class="info-row">
                <span class="label">账户</span>
                <span class="value">{{ selectedPaymentCard.card_number }}</span>
              </div>
            </div>
          </div>
          
          <!-- 收款账户 -->
          <div class="section">
            <div class="section-title">收款账户</div>
            <div class="tip-text">点击删除或者点击设置为默认</div>
            <div v-if="paymentInfo.bank_card" class="card-info" style="margin-bottom: 12px;">
              <div class="info-row">
                <span class="label">账户类型</span>
                <span class="value">银行卡</span>
              </div>
              <div class="info-row">
                <span class="label">户名</span>
                <span class="value">{{ paymentInfo.bank_card.account_name }}</span>
              </div>
              <div class="info-row">
                <span class="label">产名</span>
                <span class="value">{{ paymentInfo.bank_card.bank_name }}</span>
              </div>
              <div class="info-row">
                <span class="label">银行</span>
                <span class="value">{{ paymentInfo.bank_card.branch_name || paymentInfo.bank_card.bank_name }}</span>
              </div>
              <div class="info-row">
                <span class="label">账户</span>
                <span class="value">{{ paymentInfo.bank_card.account_number }}</span>
              </div>
            </div>
          </div>
          
          <!-- 支付凭证 -->
          <div class="section">
            <div class="section-title">支付凭证</div>
            <van-uploader
              v-model="voucherFiles"
              :max-count="5"
              :after-read="afterReadVoucher"
              multiple
            />
          </div>
          
          <!-- 温馨提示 -->
          <div class="section">
            <div class="section-title">温馨提示</div>
            <div class="tip-content">
              为保证资金安全请到任何机构或便利店使用现金（不和银卡转的是所有银行支付账单会发给对方支付支口账号）特10001-11本1000,11本-1000）方便快速认证系统自动打款及对账安全请注意选择正确
            </div>
          </div>
          
          <!-- 备注 -->
          <div class="section">
            <div class="section-title">备注</div>
            <van-field
              v-model="depositForm.note"
              type="textarea"
              placeholder="请输入内容"
              rows="2"
            />
          </div>
          
          <!-- 协议 -->
          <div class="agreement">
            <van-checkbox v-model="agreeProtocol">
              请仔细阅读并同意
              <span style="color: #ee0a24;">账户打款者姓名需一致协议</span>
            </van-checkbox>
          </div>
          
          <!-- 提交按钮 -->
          <div class="submit-btn">
            <van-button
              round
              block
              type="danger"
              @click="onDeposit"
              :disabled="!agreeProtocol"
            >
              提交审核
            </van-button>
          </div>
        </div>
      </div>
    </van-popup>
    
    <!-- 付款账户选择器 -->
    <van-popup v-model:show="showPaymentCardPicker" position="bottom" round>
      <van-picker
        :columns="paymentCardColumns"
        @confirm="onSelectPaymentCard"
        @cancel="showPaymentCardPicker = false"
      />
    </van-popup>
    
    <!-- 退定金弹窗 -->
    <van-popup v-model:show="showWithdraw" position="bottom" round>
      <div class="popup-content">
        <div class="popup-header">
          <h3>退定金</h3>
        </div>
        <van-form @submit="onWithdraw">
          <van-field
            v-model="withdrawForm.amount"
            type="number"
            label="退定金金额"
            placeholder="请输入退定金金额"
            :rules="[{ required: true, message: '请输入退定金金额' }]"
          >
            <template #extra>
              <span style="color: #999; font-size: 12px;">
                可用: ¥{{ formatMoney(userInfo.available_deposit) }}
              </span>
            </template>
          </van-field>
          <van-field
            v-model="selectedBankCardText"
            label="银行卡"
            placeholder="请选择银行卡"
            readonly
            is-link
            @click="openBankCardPicker('withdraw')"
            :rules="[{ required: true, message: '请选择银行卡' }]"
          />
	      <van-field
	        v-model="withdrawForm.note"
	        type="textarea"
	        label="备注"
	        placeholder="可填写提现说明（选填）"
	        rows="2"
	      />
          <div style="margin: 16px;">
            <van-button round block type="primary" native-type="submit">
              确认退定金
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
          :key="card.id || card.ID"
          class="bank-card-item"
          @click="selectBankCard(card)"
        >
          <div class="card-info">
            <div class="bank-name">{{ card.bank_name || card.BankName }}</div>
            <div class="card-number">**** **** **** {{ (card.card_number || card.CardNumber || '').slice(-4) }}</div>
          </div>
          <van-icon name="success" v-if="card.is_default || card.IsDefault" color="#07c160" />
        </div>
        <div v-if="bankCards.length === 0" class="empty-tip">
          暂无银行卡，<span style="color: #1989fa; cursor: pointer;" @click="goToAddCard">点击添加</span>
        </div>
      </div>
    </van-action-sheet>
    
    <!-- 详情弹窗 -->
    <van-popup v-model:show="showDepositDetailDialog" position="bottom" round :style="{ height: '80%' }">
      <div class="detail-popup" v-if="currentDetailRecord">
        <van-nav-bar
          :title="getRecordTypeText(currentDetailRecord.type) + '详情'"
          left-arrow
          @click-left="showDepositDetailDialog = false"
        />
        
        <div class="detail-content">
          <van-cell-group inset>
            <!-- 锁价交易详情 -->
            <template v-if="currentDetailRecord.type === 'trade'">
              <van-cell
                title="订单号"
                :value="'#' + (currentDetailRecord.order?.order_id || currentDetailRecord.id)"
              />
              <van-cell
                title="订单类型"
                :value="getOrderTypeText(currentDetailRecord.order?.type)"
              />
              <van-cell
                title="订单状态"
                :value="getStatusText(currentDetailRecord.status || currentDetailRecord.order?.status)"
              />
              <van-cell
                title="锁定单价"
                :value="'¥' + formatMoney(currentDetailRecord.order?.locked_price) + ' /克'"
              />
              <van-cell
                title="锁定货款"
                :value="'¥' + formatMoney((currentDetailRecord.order?.locked_price || 0) * (currentDetailRecord.order?.weight_g || 0))"
              />
              <van-cell
                title="当前价格"
                :value="'¥' + formatMoney(calcOrderCurrentPrice(currentDetailRecord.order)) + ' /克'"
              />
              <van-cell
                title="克重"
                :value="(currentDetailRecord.order?.weight_g || 0) + ' 克'"
              />
              <van-cell
                title="定金"
                :value="'¥' + formatMoney(currentDetailRecord.order?.deposit)"
              />
              <van-cell
                title="浮动盈亏"
                :value="(calcOrderPnl(currentDetailRecord.order) > 0 ? '+' : '') + '¥' + formatMoney(Math.abs(calcOrderPnl(currentDetailRecord.order)))"
              />
              <van-cell
                title="定金率"
                :value="calcOrderMarginRate(currentDetailRecord.order) !== null ? calcOrderMarginRate(currentDetailRecord.order).toFixed(2) + '%' : '-'"
              />
              <van-cell
                title="创建时间"
                :value="formatDateTime(currentDetailRecord.order?.created_at || currentDetailRecord.created_at)"
              />
            </template>

            <!-- 充值/提现明细 -->
            <template v-else>
              <van-cell title="订单号" :value="'#' + currentDetailRecord.id" />
              <van-cell title="类型" :value="getRecordTypeText(currentDetailRecord.type)" />
              <van-cell 
                v-if="currentDetailRecord.status" 
                title="状态" 
                :value="getStatusText(currentDetailRecord.status)" 
                :label-class="currentDetailRecord.status"
              />
              <van-cell 
                title="金额" 
                :value="(currentDetailRecord.amount > 0 ? '+' : '') + '¥' + formatMoney(Math.abs(currentDetailRecord.amount))"
                :class="{ 'income-cell': currentDetailRecord.amount > 0, 'expense-cell': currentDetailRecord.amount < 0 }"
              />
              <van-cell 
                v-if="currentDetailRecord.before_balance !== undefined" 
                title="变动前余额" 
                :value="'¥' + formatMoney(currentDetailRecord.before_balance)" 
              />
              <van-cell 
                v-if="currentDetailRecord.after_balance !== undefined" 
                title="变动后余额" 
                :value="'¥' + formatMoney(currentDetailRecord.after_balance)" 
              />
              <van-cell 
                v-if="currentDetailRecord.method" 
                title="支付方式" 
                :value="currentDetailRecord.method === 'bank' ? '银行转账' : currentDetailRecord.method" 
              />
              <van-cell title="时间" :value="formatDateTime(currentDetailRecord.created_at)" />
              <van-cell 
                v-if="currentDetailRecord.reviewed_at" 
                title="审核时间" 
                :value="formatDateTime(currentDetailRecord.reviewed_at)" 
              />
              <van-cell 
                v-if="currentDetailRecord.paid_at" 
                title="打款时间" 
                :value="formatDateTime(currentDetailRecord.paid_at)" 
              />
              <van-cell 
                v-if="currentDetailRecord.description" 
                title="备注" 
                :value="currentDetailRecord.description" 
              />
            </template>
          </van-cell-group>
          
          <!-- 支付凭证 -->
          <div v-if="currentDetailRecord.voucher_url" class="voucher-section">
            <div class="section-title">支付凭证</div>
            <div class="voucher-images">
              <van-image
                v-for="(url, index) in getVoucherUrls(currentDetailRecord.voucher_url)"
                :key="index"
                :src="url"
                width="100"
                height="100"
                fit="cover"
                @click="previewVoucher(currentDetailRecord.voucher_url, index)"
              />
            </div>
          </div>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { showToast, showDialog, showImagePreview } from 'vant'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'
import { formatMoney, formatDateTime } from '../utils/helpers'
import { useQuoteStore } from '../stores/quote'

const activeTab = ref('all')
const records = ref([])
const refreshing = ref(false)
const loading = ref(false)
const finished = ref(false)
const page = ref(1)

const userInfo = ref({
  available_deposit: 0,
  used_deposit: 0
})

const quoteStore = useQuoteStore()
const holdingOrders = ref([])

const calcOrderCurrentPrice = (order) => {
  if (!order) return 0
  if (order.type === 'long_buy') {
    return quoteStore.sellPrice || quoteStore.buyPrice || order.locked_price || 0
  }
  if (order.type === 'short_sell') {
    return quoteStore.buyPrice || quoteStore.sellPrice || order.locked_price || 0
  }
  return order.locked_price || 0
}

const calcOrderPnl = (order) => {
  if (!order) return 0
  const price = calcOrderCurrentPrice(order)
  const locked = order.locked_price || 0
  const weight = order.weight_g || 0
  if (!weight) return 0
  if (order.type === 'long_buy') {
    return (price - locked) * weight
  }
  if (order.type === 'short_sell') {
    return (locked - price) * weight
  }
  return 0
}

const calcOrderMarginRate = (order) => {
  if (!order) return null
  const deposit = order.deposit || 0
  if (!deposit) return null
  const pnl = calcOrderPnl(order)
  return ((deposit + pnl) / deposit) * 100
}

const holdingMargin = computed(() => {
  if (!holdingOrders.value || holdingOrders.value.length === 0) return 0
  return holdingOrders.value.reduce((sum, order) => {
    const deposit = order.deposit || 0
    const pnl = calcOrderPnl(order)
    return sum + deposit + pnl
  }, 0)
})

const totalDeposit = computed(() => {
  const available = userInfo.value.available_deposit || 0
  return available + holdingMargin.value
})

const getOrderTypeText = (type) => {
  if (type === 'long_buy') return '锁价买料'
  if (type === 'short_sell') return '锁价卖料'
  return type || ''
}

const loadHoldingOrders = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.ORDERS, { params: { status: 'holding' } })
    const list = data.orders || data.list || []
    holdingOrders.value = list.map((o) => ({
      order_id: o.order_id || o.OrderID || o.id || '',
      type: o.type || o.Type || '',
      locked_price: o.locked_price ?? o.LockedPrice ?? 0,
      weight_g: o.weight_g ?? o.WeightG ?? 0,
      deposit: o.deposit ?? o.Deposit ?? 0,
      pnl_float: o.pnl_float ?? o.PnLFloat ?? 0,
      status: o.status || o.Status || '',
      created_at: o.created_at || o.CreatedAt || null
    }))
  } catch (error) {
    console.error('加载持仓订单失败:', error)
  }
}

const showDeposit = ref(false)
const showWithdraw = ref(false)
const showBankCardPicker = ref(false)
const currentPickerType = ref('deposit')
const selectedBankCardText = ref('')
const bankCards = ref([])

const depositForm = ref({
  amount: '',
  note: ''
})

const withdrawForm = ref({
  amount: '',
  bank_card_id: '',
  note: ''
})

const paymentInfo = ref({
  bank_card: null,
  wechat_qr: '',
  alipay_qr: ''
})

// 快捷金额
const quickAmounts = ref([5000, 6000, 10000, 15000, 20000, 50000, 100000, 200000])

// 付款账户选择
const showPaymentCardPicker = ref(false)
const selectedPaymentCard = ref(null)
const paymentCardColumns = ref([])

// 支付凭证
const voucherFiles = ref([])
const voucherUrl = ref('')

// 协议
const agreeProtocol = ref(false)

// 详情弹窗
const showDepositDetailDialog = ref(false)
const currentDetailRecord = ref(null)

// 获取记录类型文本
const getRecordTypeText = (type) => {
  const types = {
    deposit: '付定金',
    withdraw: '退定金',
    trade: '交易',
    buy: '买入',
    sell: '卖出',
    profit: '盈利',
    loss: '亏损',
    commission: '提成',
    supplement_deposit: '补交定金'
  }
  return types[type] || type
}

// 获取记录类型样式类
const getRecordTypeClass = (type) => {
  const classMap = {
    deposit: 'type-deposit',
    withdraw: 'type-withdraw',
    buy: 'type-buy',
    sell: 'type-sell',
    profit: 'type-profit',
    loss: 'type-loss'
  }
  return classMap[type] || 'type-default'
}

// 获取状态文本
const getStatusText = (status) => {
  if (!status) return ''
  const normalized = String(status).toLowerCase()
  const statusMap = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝',
    paid: '已打款',
    success: '成功',
    failed: '失败',
    holding: '待结算',
    settled: '已完结',
    closed: '已完结'
  }
  return statusMap[normalized] || status
}

// 格式化快捷金额
const formatQuickAmount = (amount) => {
  if (amount >= 10000) {
    return (amount / 10000) + '万'
  }
  return amount
}

// 选择付款账户
const onSelectPaymentCard = (value) => {
  const card = bankCards.value.find(c => (c.id || c.ID) === value.value)
  if (card) {
    selectedPaymentCard.value = {
      id: card.id || card.ID,
      card_holder: card.card_holder || card.CardHolder,
      bank_name: card.bank_name || card.BankName,
      card_number: card.card_number || card.CardNumber
    }
  }
  showPaymentCardPicker.value = false
}

// 上传凭证图片
const afterReadVoucher = async (file) => {
  try {
    showToast('正在处理图片...')
    
    // 处理单个或多个文件
    const files = Array.isArray(file) ? file : [file]
    
    // 初始化数组
    if (!voucherUrl.value) {
      voucherUrl.value = []
    }
    if (typeof voucherUrl.value === 'string') {
      voucherUrl.value = [voucherUrl.value]
    }
    
    for (const f of files) {
      const compressed = await compressImage(f.file)
      
      // 检查单张图片大小
      const sizeKB = Math.round(compressed.length / 1024)
      if (sizeKB > 800) {
        showToast(`图片过大(${sizeKB}KB)，请重新选择`)
        continue
      }
      
      voucherUrl.value.push(compressed)
    }
    
    // 检查总大小
    const totalSize = voucherUrl.value.join(',').length
    const totalSizeKB = Math.round(totalSize / 1024)
    console.log('凭证总大小:', totalSizeKB, 'KB')
    
    if (totalSizeKB > 2000) {
      showToast('凭证图片总大小超限，请减少数量或降低分辨率')
      return
    }
    
    showToast('图片上传成功')
  } catch (error) {
    console.error('图片处理失败:', error)
    showToast('图片处理失败')
  }
}

// 查看记录详情
const showRecordDetail = (record) => {
  // 显示详情对话框
  showDepositDetailDialog.value = true
  currentDetailRecord.value = record
}

// 将后端存储的图片字段解析为 URL 列表
const parseImageUrls = (raw) => {
  if (!raw) return []
  const str = String(raw).trim()
  if (!str) return []

  // 优先匹配一个字段里包含的多个 data:image...base64,... 段
  const dataUrlMatches = str.match(/data:image[^,]*,[^,]+/g)
  if (dataUrlMatches && dataUrlMatches.length > 0) {
    return dataUrlMatches.map((s) => s.trim())
  }

  // 非 Data URL 情况，再按逗号拆分多张图片
  if (str.includes(',')) {
    return str
      .split(',')
      .map((s) => s.trim())
      .filter((s) => s.length > 0)
  }

  // 单一 URL 或裸 base64
  return [str]
}

// 规范图片 URL：支持 http(s)、/path、本地 base64
const normalizeImageUrl = (raw) => {
  if (!raw) return ''
  const url = String(raw).trim()
  if (!url) return ''

  if (
    url.startsWith('http://') ||
    url.startsWith('https://') ||
    url.startsWith('data:') ||
    url.startsWith('/')
  ) {
    return url
  }

  // 兜底：看起来像裸的 base64 内容，补上 jpeg 前缀
  return `data:image/jpeg;base64,${url}`
}

// 获取凭证URL数组
const getVoucherUrls = (voucherUrl) => {
  const urls = parseImageUrls(voucherUrl)
    .map((url) => normalizeImageUrl(url))
    .filter((url) => url && url.length > 0)
  return urls
}

// 预览凭证
const previewVoucher = (voucherUrl, startPosition = 0) => {
  const urls = getVoucherUrls(voucherUrl)
  showImagePreview({
    images: urls,
    startPosition: startPosition
  })
}

// 压缩图片
const compressImage = (file, maxWidth = 600, quality = 0.6) => {
  return new Promise((resolve) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const img = new Image()
      img.onload = () => {
        const canvas = document.createElement('canvas')
        let width = img.width
        let height = img.height
        
        // 更激进的压缩策略
        if (width > maxWidth) {
          height = (height * maxWidth) / width
          width = maxWidth
        }
        
        canvas.width = width
        canvas.height = height
        
        const ctx = canvas.getContext('2d')
        ctx.drawImage(img, 0, 0, width, height)
        
        // 使用更低的质量
        let compressedBase64 = canvas.toDataURL('image/jpeg', quality)
        
        // 如果还是太大，进一步降低质量
        if (compressedBase64.length > 500000) { // 500KB
          compressedBase64 = canvas.toDataURL('image/jpeg', 0.4)
        }
        
        console.log('压缩后图片大小:', Math.round(compressedBase64.length / 1024), 'KB')
        resolve(compressedBase64)
      }
      img.src = e.target.result
    }
    reader.readAsDataURL(file)
  })
}

// 加载用户信息
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

// 加载资金流水
const loadRecords = async () => {
  try {
    let list = []
    
    if (activeTab.value === 'deposit') {
      // 加载充值记录
      const data = await request.get(API_ENDPOINTS.DEPOSITS)
      const deposits = data.deposits || []
      
      // 转换为统一格式
      list = deposits.map(d => ({
        id: d.ID || d.id,
        type: 'deposit',
        amount: d.Amount || d.amount,
        status: d.Status || d.status,
        created_at: d.CreatedAt || d.created_at,
        voucher_url: d.VoucherURL || d.voucher_url,
        method: d.Method || d.method,
        review_note: d.ReviewNote || d.review_note,
        reviewed_at: d.ReviewedAt || d.reviewed_at,
        description: d.ReviewNote || d.review_note || ''
      }))
    } else if (activeTab.value === 'withdraw') {
      // 加载提现记录
      const data = await request.get(API_ENDPOINTS.WITHDRAWS)
      const withdraws = data.withdraws || []
      
      list = withdraws.map(w => ({
        id: w.ID || w.id,
        type: 'withdraw',
        amount: -(w.Amount || w.amount),
        status: w.Status || w.status,
        created_at: w.CreatedAt || w.created_at,
        bank_card_id: w.BankCardID || w.bank_card_id,
        review_note: w.ReviewNote || w.review_note,
        reviewed_at: w.ReviewedAt || w.reviewed_at,
        paid_at: w.PaidAt || w.paid_at,
        voucher_url: w.VoucherURL || w.voucher_url,
        description: w.UserNote || w.user_note || w.ReviewNote || w.review_note || ''
      }))
    } else if (activeTab.value === 'trade') {
      // 加载持仓订单作为交易列表
      await loadHoldingOrders()
      list = holdingOrders.value.map(order => ({
        id: order.order_id,
        type: 'trade',
        amount: order.deposit || 0,
        status: order.status || 'holding',
        created_at: order.created_at,
        order
      }))
    } else {
      // 加载所有资金流水
      const data = await request.get(API_ENDPOINTS.FUND_FLOW)
      const logs = data.logs || []
      
      list = logs.map(log => ({
        id: log.ID || log.id,
        type: log.Type || log.type,
        amount: log.Amount || log.amount,
        before_balance: log.AvailableBefore || log.available_before,
        after_balance: log.AvailableAfter || log.available_after,
        created_at: log.CreatedAt || log.created_at,
        description: log.Note || log.note || ''
      }))
    }
    
    console.log('加载的记录:', list)
    records.value = list
    finished.value = true
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
    const data = await request.get(API_ENDPOINTS.BANK_CARDS)
    console.log('银行卡数据:', data)
    bankCards.value = data.cards || data.list || []
    console.log('解析后的银行卡列表:', bankCards.value)
    
    // 初始化付款账户选择器
    paymentCardColumns.value = bankCards.value.map(card => ({
      text: `${card.bank_name || card.BankName} (*${(card.card_number || card.CardNumber || '').slice(-4)})`,
      value: card.id || card.ID
    }))
    
    // 默认选择第一张卡
    if (bankCards.value.length > 0) {
      const firstCard = bankCards.value[0]
      selectedPaymentCard.value = {
        id: firstCard.id || firstCard.ID,
        card_holder: firstCard.card_holder || firstCard.CardHolder,
        bank_name: firstCard.bank_name || firstCard.BankName,
        card_number: firstCard.card_number || firstCard.CardNumber
      }
    }
  } catch (error) {
    console.error('加载银行卡失败:', error)
  }
}

// 打开银行卡选择器
const openBankCardPicker = (type) => {
  currentPickerType.value = type
  showBankCardPicker.value = true
}

// 选择银行卡
const selectBankCard = (card) => {
  const cardId = card.id || card.ID
  const bankName = card.bank_name || card.BankName
  const cardNumber = card.card_number || card.CardNumber || ''
  
  // 只用于提现
  withdrawForm.value.bank_card_id = cardId
  selectedBankCardText.value = `${bankName} (*${cardNumber.slice(-4)})`
  showBankCardPicker.value = false
}

// 跳转到添加银行卡
const goToAddCard = () => {
  showBankCardPicker.value = false
  showDeposit.value = false
  showWithdraw.value = false
  window.location.href = '#/bank-cards'
}

// 付定金
const onDeposit = async () => {
  try {
    // 验证
    if (!depositForm.value.amount) {
      showToast('请输入金额')
      return
    }
    
    if (!selectedPaymentCard.value) {
      showToast('请选择付款账户')
      return
    }
    
    if (!agreeProtocol.value) {
      showToast('请阅读并同意协议')
      return
    }
    
    // 处理凭证URL（支持多张图片）
    let voucherUrlString = ''
    if (voucherUrl.value) {
      if (Array.isArray(voucherUrl.value)) {
        // 多张图片用逗号分隔
        voucherUrlString = voucherUrl.value.join(',')
      } else {
        voucherUrlString = voucherUrl.value
      }
    }
    
    const requestData = {
      amount: parseFloat(depositForm.value.amount),
      method: 'bank',
      voucher_url: voucherUrlString,
      note: depositForm.value.note || ''
    }
    
    console.log('付定金请求数据:', requestData)
    
    await request.post(API_ENDPOINTS.DEPOSIT_CREATE, requestData)
    
    showToast('付定金申请已提交，等待审核')
    
    // 重置表单
    showDeposit.value = false
    depositForm.value = { amount: '', note: '' }
    voucherFiles.value = []
    voucherUrl.value = []
    agreeProtocol.value = false
    
    loadUserInfo()
    onRefresh()
  } catch (error) {
    console.error('付定金失败:', error)
    console.error('错误详情:', error.response?.data)
    const errorMsg = error.response?.data?.error || error.response?.data?.message || '付定金失败'
    showToast(errorMsg)
  }
}

// 提现
const onWithdraw = async () => {
  try {
    await request.post(API_ENDPOINTS.WITHDRAW_CREATE, {
      amount: parseFloat(withdrawForm.value.amount),
      bank_card_id: withdrawForm.value.bank_card_id,
      note: withdrawForm.value.note || ''
    })
    
    showToast('提现申请已提交，等待审核')
    showWithdraw.value = false
    withdrawForm.value = { amount: '', bank_card_id: '', note: '' }
    selectedBankCardText.value = ''
    
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

// 加载收款信息
const loadPaymentInfo = () => {
  try {
    const settings = localStorage.getItem('payment_settings')
    if (settings) {
      paymentInfo.value = JSON.parse(settings)
    }
  } catch (error) {
    console.error('加载收款信息失败:', error)
  }
}

onMounted(() => {
  quoteStore.connectWebSocket()
  loadUserInfo()
  loadHoldingOrders()
  loadRecords()
  loadBankCards()
  loadPaymentInfo()
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

/* 卡片式记录列表 */
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

.type-deposit {
  background: #fff3e0;
  color: #ff9800;
}

.type-withdraw {
  background: #e3f2fd;
  color: #2196f3;
}

.type-buy {
  background: #f3e5f5;
  color: #9c27b0;
}

.type-sell {
  background: #e8f5e9;
  color: #4caf50;
}

.type-profit {
  background: #ffebee;
  color: #f44336;
}

.type-loss {
  background: #fce4ec;
  color: #e91e63;
}

.type-default {
  background: #f5f5f5;
  color: #666;
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

.record-status.pending {
  background: #fff3e0;
  color: #ff9800;
}

.record-status.approved,
.record-status.success,
.record-status.paid {
  background: #e8f5e9;
  color: #4caf50;
}

.record-status.rejected,
.record-status.failed {
  background: #ffebee;
  color: #f44336;
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
  text-align: right;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.view-detail {
  color: #1989fa;
  font-size: 13px;
}

.popup-content {
  padding: 20px;
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

/* 付定金弹窗样式 */
.deposit-popup {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.deposit-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  padding-bottom: 80px;
}

.amount-section {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.amount-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.amount-input {
  font-size: 24px;
  font-weight: bold;
  padding: 0;
}

.amount-input :deep(.van-field__control) {
  font-size: 24px;
  font-weight: bold;
}

.quick-amounts {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.quick-amounts .van-button {
  border-radius: 4px;
}

.section {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: bold;
  color: #ee0a24;
  margin-bottom: 12px;
}

.section-title.required::after {
  content: '';
}

.tip-text {
  font-size: 12px;
  color: #ee0a24;
  margin-bottom: 12px;
}

.card-info {
  background: #f7f8fa;
  border-radius: 8px;
  padding: 12px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

.info-row:last-child {
  border-bottom: none;
}

.info-row .label {
  color: #666;
  font-size: 14px;
}

.info-row .value {
  color: #333;
  font-size: 14px;
  font-weight: 500;
}

.tip-content {
  font-size: 12px;
  color: #999;
  line-height: 1.6;
}

.agreement {
  margin: 16px 0;
}

.submit-btn {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 16px;
  background: #fff;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.05);
}

/* 详情弹窗样式 */
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

.income-cell :deep(.van-cell__value) {
  color: #f56c6c !important;
  font-weight: bold;
}

.expense-cell :deep(.van-cell__value) {
  color: #67c23a !important;
  font-weight: bold;
}

.voucher-section {
  margin-top: 16px;
  background: #fff;
  border-radius: 8px;
  padding: 16px;
}

.voucher-section .section-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 12px;
  color: #333;
}

.voucher-images {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.voucher-images .van-image {
  border-radius: 4px;
  cursor: pointer;
}
</style>
