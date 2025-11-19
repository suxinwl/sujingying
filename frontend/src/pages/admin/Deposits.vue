<template>
  <div class="admin-deposits-page">
    <van-nav-bar
      title="充值管理"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    />
    
    <!-- Tab切换 -->
    <van-tabs v-model:active="activeTab" @change="onTabChange">
      <van-tab title="待审核" name="pending" />
      <van-tab title="已通过" name="approved" />
      <van-tab title="已拒绝" name="rejected" />
    </van-tabs>
    
    <!-- 充值列表 -->
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        finished-text="没有更多了"
        @load="loadDeposits"
      >
        <div v-if="deposits.length === 0" class="empty">
          <van-empty description="暂无记录" />
        </div>
        
        <div v-for="deposit in deposits" :key="deposit.id || deposit.ID" class="deposit-item" @click="showDepositDetail(deposit)">
          <div class="deposit-header">
            <span class="deposit-amount">¥{{ formatMoney(deposit.amount || deposit.Amount) }}</span>
            <span class="deposit-status" :class="deposit.status || deposit.Status">
              {{ getStatusText(deposit.status || deposit.Status) }}
            </span>
          </div>
          
          <div class="deposit-body">
            <div class="deposit-row">
              <span class="label">订单号:</span>
              <span class="value">#{{ deposit.id || deposit.ID }}</span>
            </div>
            <div class="deposit-row">
              <span class="label">用户:</span>
              <span class="value">{{ getUserDisplay(deposit) }}</span>
            </div>
            <div class="deposit-row" v-if="deposit.voucher_url || deposit.VoucherURL">
              <span class="label">付款凭证:</span>
              <span class="value" style="color: #1989fa; cursor: pointer;" @click.stop="previewVoucher(deposit)">
                查看图片({{ getVoucherCount(deposit) }}张)
              </span>
            </div>
            <div class="deposit-row" v-if="deposit.receipt_voucher || deposit.ReceiptVoucherURL">
              <span class="label">收款凭证:</span>
              <span class="value" style="color: #1989fa; cursor: pointer;" @click.stop="previewReceiptVoucher(deposit)">
                查看图片
              </span>
            </div>
            <div class="deposit-row">
              <span class="label">充值方式:</span>
              <span class="value">{{ getMethodText(deposit.method || deposit.Method) }}</span>
            </div>
            <div class="deposit-row">
              <span class="label">申请时间:</span>
              <span class="value">{{ formatDateTime(deposit.created_at || deposit.CreatedAt) }}</span>
            </div>
            <div class="deposit-row" v-if="deposit.reviewed_at || deposit.ReviewedAt">
              <span class="label">审核时间:</span>
              <span class="value">{{ formatDateTime(deposit.reviewed_at || deposit.ReviewedAt) }}</span>
            </div>
            <div class="deposit-row" v-if="deposit.review_note || deposit.ReviewNote">
              <span class="label">审核备注:</span>
              <span class="value">{{ deposit.review_note || deposit.ReviewNote }}</span>
            </div>
            <div class="deposit-row" v-if="deposit.user_note || deposit.UserNote">
              <span class="label">用户备注:</span>
              <span class="value">{{ deposit.user_note || deposit.UserNote }}</span>
            </div>
          </div>
          
          <div class="deposit-footer" v-if="(deposit.status || deposit.Status) === 'pending'" @click.stop>
            <van-button size="small" type="success" @click="showReviewDialog(deposit.id || deposit.ID, true)">
              通过
            </van-button>
            <van-button size="small" type="danger" @click="showReviewDialog(deposit.id || deposit.ID, false)">
              拒绝
            </van-button>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>
    
    <!-- 审核弹窗 -->
    <van-popup v-model:show="showReviewPopup" position="bottom" round>
      <div class="review-popup">
        <van-nav-bar
          :title="currentReviewApproved ? '通过审核' : '拒绝审核'"
          left-arrow
          @click-left="showReviewPopup = false"
        />
        
        <div class="review-content">
          <van-form>
            <van-field
              v-model="reviewNote"
              type="textarea"
              :label="currentReviewApproved ? '审核备注' : '拒绝原因'"
              :placeholder="currentReviewApproved ? '请输入审核备注（选填）' : '请输入拒绝原因（必填）'"
              rows="3"
              :rules="[{ required: !currentReviewApproved, message: '请输入拒绝原因' }]"
            />
            
            <div v-if="currentReviewApproved" class="receipt-section">
              <div class="section-label">收款凭证（选填）</div>
              <van-uploader
                v-model="receiptVoucherFiles"
                :max-count="1"
                :after-read="afterReadReceipt"
              />
              <div class="section-tip">上传银行收款凭证，方便用户核对</div>
            </div>
            
            <div class="submit-section">
              <van-button 
                round 
                block 
                :type="currentReviewApproved ? 'success' : 'danger'"
                @click="submitReview"
              >
                {{ currentReviewApproved ? '确认通过' : '确认拒绝' }}
              </van-button>
            </div>
          </van-form>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
/**
 * @file Deposits.vue
 * @description 充值管理页面
 * @author AI
 * @date 2025-11-18
 */

import { ref, onMounted } from 'vue'
import { showToast, showDialog, showImagePreview } from 'vant'
import request from '../../utils/request'
import { API_ENDPOINTS } from '../../config/api'
import { formatMoney, formatDateTime } from '../../utils/helpers'

const activeTab = ref('pending')
const deposits = ref([])
const refreshing = ref(false)
const loading = ref(false)
const finished = ref(false)

// 将后端存储的图片字段解析为 URL 列表
const parseImageUrls = (raw) => {
  if (!raw) return []
  const str = String(raw).trim()
  if (!str) return []

  // 优先匹配一个字段里包含的多个 data:image...base64,... 段
  // 例如：data:image/jpeg;base64,xxx,data:image/jpeg;base64,yyy,...
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

  // 已经是完整 URL 或 Data URL，直接返回
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

const getStatusText = (status) => {
  const statusMap = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝'
  }
  return statusMap[status] || status
}

const getMethodText = (method) => {
  const methodMap = {
    bank: '银行转账',
    wechat: '微信支付',
    alipay: '支付宝'
  }
  return methodMap[method] || method || '银行转账'
}

// 获取用户显示信息
const getUserDisplay = (deposit) => {
  const user = deposit.user || deposit.User
  if (user) {
    if (user.realname || user.RealName) {
      return user.realname || user.RealName
    }
    if (user.phone || user.Phone) {
      return user.phone || user.Phone
    }
    return `用户${user.id || user.ID || '未知'}`
  }
  return `用户${deposit.user_id || deposit.UserID || '未知'}`
}

// 获取凭证数量
const getVoucherCount = (deposit) => {
  const voucherUrl = deposit.voucher_url || deposit.VoucherURL || ''
  if (!voucherUrl) return 0
  const urls = parseImageUrls(voucherUrl)
  return urls.length
}

// 预览凭证
const previewVoucher = (deposit) => {
  const voucherUrl = deposit.voucher_url || deposit.VoucherURL || ''
  if (!voucherUrl) return
  
  // 解析并规范 URL，兼容 Data URL、多张图片
  const urls = parseImageUrls(voucherUrl)
    .map((url) => normalizeImageUrl(url))
    .filter((url) => url && url.length > 0)
  
  if (urls.length === 0) {
    showToast('暂无凭证图片')
    return
  }
  
  showImagePreview({
    images: urls,
    startPosition: 0
  })
}

// 预览管理员收款凭证
const previewReceiptVoucher = (deposit) => {
  const raw = deposit.receipt_voucher || deposit.ReceiptVoucherURL || ''
  if (!raw) {
    showToast('暂无收款凭证')
    return
  }

  const urls = parseImageUrls(raw)
    .map((url) => normalizeImageUrl(url))
    .filter((url) => url && url.length > 0)

  if (urls.length === 0) {
    showToast('暂无收款凭证')
    return
  }

  showImagePreview({
    images: urls,
    startPosition: 0
  })
}

const loadDeposits = async () => {
  try {
    loading.value = true
    const params = {
      status: activeTab.value,
      limit: 50
    }
    
    const data = await request.get(API_ENDPOINTS.ADMIN_DEPOSITS_PENDING, { params })
    console.log('充值数据:', data)
    const list = data.deposits || data.list || []
    console.log('充值列表:', list)
    if (list.length > 0) {
      console.log('第一条充值数据:', list[0])
    }
    
    deposits.value = list
    finished.value = true
  } catch (error) {
    console.error('加载充值记录失败:', error)
    showToast('加载失败')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const onTabChange = () => {
  finished.value = false
  deposits.value = []
  loadDeposits()
}

const onRefresh = () => {
  finished.value = false
  loadDeposits()
}

// 审核对话框状态
const showReviewPopup = ref(false)
const currentReviewId = ref(null)
const currentReviewApproved = ref(false)
const reviewNote = ref('')
const receiptVoucherFiles = ref([])
const receiptVoucherUrl = ref('')

const showReviewDialog = (depositId, approved) => {
  currentReviewId.value = depositId
  currentReviewApproved.value = approved
  reviewNote.value = ''
  receiptVoucherFiles.value = []
  receiptVoucherUrl.value = ''
  showReviewPopup.value = true
}

// 上传收款凭证
const afterReadReceipt = async (file) => {
  try {
    showToast('正在处理图片...')
    const compressed = await compressImage(file.file)
    receiptVoucherUrl.value = compressed
    showToast('图片上传成功')
  } catch (error) {
    console.error('图片处理失败:', error)
    showToast('图片处理失败')
  }
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

// 提交审核
const submitReview = async () => {
  try {
    if (!currentReviewApproved.value && !reviewNote.value) {
      showToast('请输入拒绝原因')
      return
    }
    
    const requestData = { 
      action: currentReviewApproved.value ? 'approve' : 'reject',
      note: reviewNote.value || '审核通过'
    }
    
    // 如果是通过且上传了收款凭证
    if (currentReviewApproved.value && receiptVoucherUrl.value) {
      requestData.receipt_voucher = receiptVoucherUrl.value
    }
    
    await request.post(
      API_ENDPOINTS.ADMIN_DEPOSIT_REVIEW.replace(':id', currentReviewId.value),
      requestData
    )
    
    showToast(currentReviewApproved.value ? '审核通过' : '已拒绝')
    showReviewPopup.value = false
    onRefresh()
  } catch (error) {
    console.error('审核失败:', error)
    const errorMsg = error.response?.data?.error || error.response?.data?.message || '操作失败'
    showToast(errorMsg)
  }
}

const showDepositDetail = (deposit) => {
  const id = deposit.id || deposit.ID
  const amount = deposit.amount || deposit.Amount
  const user = deposit.user || deposit.User
  const userLabel = user
    ? (user.realname || user.RealName || user.phone || user.Phone || `ID:${user.id || user.ID}`)
    : (deposit.user_id || deposit.UserID)
  const method = deposit.method || deposit.Method
  const createdAt = deposit.created_at || deposit.CreatedAt
  const status = deposit.status || deposit.Status
  const reviewedAt = deposit.reviewed_at || deposit.ReviewedAt
  const reviewNote = deposit.review_note || deposit.ReviewNote
  const hasReceiptVoucher = deposit.receipt_voucher || deposit.ReceiptVoucherURL
  
  const detailInfo = [
    `订单号：#${id}`,
    `金额：¥${formatMoney(amount)}`,
    `用户：${userLabel}`,
    `充值方式：${getMethodText(method)}`,
    `申请时间：${formatDateTime(createdAt)}`,
    `状态：${getStatusText(status)}`,
    reviewedAt ? `审核时间：${formatDateTime(reviewedAt)}` : '',
    reviewNote ? `审核备注：${reviewNote}` : '',
    deposit.user_note || deposit.UserNote ? `用户备注：${deposit.user_note || deposit.UserNote}` : '',
    hasReceiptVoucher ? '管理员收款凭证：已上传' : ''
  ].filter(Boolean).join('\n')
  
  showDialog({
    title: '充值详情',
    message: detailInfo,
    confirmButtonText: '关闭'
  })
}

onMounted(() => {
  loadDeposits()
})
</script>

<style scoped>
.admin-deposits-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 60px;
}

.deposit-item {
  background: #fff;
  margin: 10px;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.deposit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.deposit-amount {
  font-size: 20px;
  font-weight: bold;
  color: #f56c6c;
}

.deposit-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.deposit-status.pending {
  color: #e6a23c;
  background: #fdf6ec;
}

.deposit-status.approved {
  color: #67c23a;
  background: #f0f9ff;
}

.deposit-status.rejected {
  color: #909399;
  background: #f4f4f5;
}

.deposit-body {
  margin: 12px 0;
}

.deposit-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}

.deposit-row .label {
  color: #909399;
}

.deposit-row .value {
  color: #303133;
}

.deposit-footer {
  display: flex;
  gap: 12px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.deposit-footer .van-button {
  flex: 1;
}

.empty {
  padding: 60px 0;
}

/* 审核弹窗样式 */
.review-popup {
  padding-bottom: 20px;
}

.review-content {
  padding: 16px;
}

.receipt-section {
  margin: 16px 0;
  padding: 16px;
  background: #f7f8fa;
  border-radius: 8px;
}

.section-label {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 12px;
  color: #323233;
}

.section-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #969799;
}

.submit-section {
  margin-top: 20px;
}
</style>
