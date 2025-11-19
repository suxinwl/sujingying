<template>
  <div class="admin-withdraws-page">
    <van-nav-bar
      title="提现管理"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    />
    
    <van-tabs v-model:active="activeTab" @change="onTabChange">
      <van-tab title="待审核" name="pending" />
      <van-tab title="已通过" name="approved" />
      <van-tab title="已拒绝" name="rejected" />
      <van-tab title="已打款" name="paid" />
    </van-tabs>
    
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        finished-text="没有更多了"
        @load="loadWithdraws"
      >
        <div v-if="withdraws.length === 0" class="empty">
          <van-empty description="暂无记录" />
        </div>
        
        <div
          v-for="withdraw in withdraws"
          :key="withdraw.id || withdraw.ID"
          class="withdraw-item"
        >
          <div class="withdraw-header">
            <span class="withdraw-amount">
              ¥{{ formatMoney(withdraw.amount || withdraw.Amount) }}
            </span>
            <span class="withdraw-status" :class="withdraw.status || withdraw.Status">
              {{ getStatusText(withdraw.status || withdraw.Status) }}
            </span>
          </div>
          
          <div class="withdraw-body">
            <div class="withdraw-row">
              <span class="label">订单号:</span>
              <span class="value">#{{ withdraw.id || withdraw.ID }}</span>
            </div>
            <div class="withdraw-row">
              <span class="label">用户:</span>
              <span class="value">{{ getUserDisplay(withdraw) }}</span>
            </div>
            <div class="withdraw-row">
              <span class="label">银行卡:</span>
              <span class="value">
                ID: {{ withdraw.bank_card_id || withdraw.BankCardID || '未知' }}
              </span>
            </div>
            <div class="withdraw-row">
              <span class="label">申请时间:</span>
              <span class="value">
                {{ formatDateTime(withdraw.created_at || withdraw.CreatedAt) }}
              </span>
            </div>
            <div class="withdraw-row" v-if="withdraw.reviewed_at || withdraw.ReviewedAt">
              <span class="label">审核时间:</span>
              <span class="value">
                {{ formatDateTime(withdraw.reviewed_at || withdraw.ReviewedAt) }}
              </span>
            </div>
            <div class="withdraw-row" v-if="withdraw.review_note || withdraw.ReviewNote">
              <span class="label">审核备注:</span>
              <span class="value">{{ withdraw.review_note || withdraw.ReviewNote }}</span>
            </div>
            <div class="withdraw-row" v-if="withdraw.paid_at || withdraw.PaidAt">
              <span class="label">打款时间:</span>
              <span class="value">
                {{ formatDateTime(withdraw.paid_at || withdraw.PaidAt) }}
              </span>
            </div>
            <div class="withdraw-row" v-if="withdraw.voucher_url || withdraw.VoucherURL">
              <span class="label">打款凭证:</span>
              <span
                class="value"
                style="color: #1989fa; cursor: pointer;"
                @click.stop="previewPayVoucher(withdraw)"
              >
                查看图片
              </span>
            </div>
            <div class="withdraw-row" v-if="withdraw.user_note || withdraw.UserNote">
              <span class="label">用户备注:</span>
              <span class="value">{{ withdraw.user_note || withdraw.UserNote }}</span>
            </div>
          </div>
          
          <div
            class="withdraw-footer"
            v-if="(withdraw.status || withdraw.Status) === 'pending'"
          >
            <van-button
              size="small"
              type="success"
              @click="reviewWithdraw(withdraw.id || withdraw.ID, true)"
            >
              通过
            </van-button>
            <van-button
              size="small"
              type="danger"
              @click="reviewWithdraw(withdraw.id || withdraw.ID, false)"
            >
              拒绝
            </van-button>
          </div>
          <div
            class="withdraw-footer"
            v-else-if="(withdraw.status || withdraw.Status) === 'approved'"
          >
            <van-button
              size="small"
              type="primary"
              @click="showPayDialog(withdraw.id || withdraw.ID)"
            >
              标记已打款
            </van-button>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>
    
    <van-popup v-model:show="showPayPopup" position="bottom" round>
      <div class="pay-popup">
        <van-nav-bar
          title="标记已打款"
          left-arrow
          @click-left="showPayPopup = false"
        />
        <div class="pay-content">
          <div class="receipt-section">
            <div class="section-label">打款凭证（必填）</div>
            <van-uploader
              v-model="payVoucherFiles"
              :max-count="1"
              :after-read="afterReadPayVoucher"
            />
            <div class="section-tip">上传银行打款凭证，方便用户核对</div>
          </div>
          <div class="submit-section">
            <van-button round block type="primary" @click="submitPay">
              确认已打款
            </van-button>
          </div>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
/**
 * @file Withdraws.vue
 * @description 提现管理页面
 * @author AI
 * @date 2025-11-18
 */

import { ref, onMounted } from 'vue'
import { showToast, showDialog, showImagePreview } from 'vant'
import request from '../../utils/request'
import { API_ENDPOINTS } from '../../config/api'
import { formatMoney, formatDateTime } from '../../utils/helpers'

const activeTab = ref('pending')
const withdraws = ref([])
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

const showPayPopup = ref(false)
const currentPayId = ref(null)
const payVoucherFiles = ref([])
const payVoucherUrl = ref('')

const getStatusText = (status) => {
  const statusMap = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝',
    paid: '已打款'
  }
  return statusMap[status] || status
}

const getUserDisplay = (withdraw) => {
	const user = withdraw.user || withdraw.User
	if (user) {
		if (user.realname || user.RealName) {
			return user.realname || user.RealName
		}
		if (user.phone || user.Phone) {
			return user.phone || user.Phone
		}
		return `用户${user.id || user.ID || '未知'}`
	}
	return `用户${withdraw.user_id || withdraw.UserID || '未知'}`
}

const compressImage = (file, maxWidth = 600, quality = 0.6) => {
  return new Promise((resolve) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const img = new Image()
      img.onload = () => {
        const canvas = document.createElement('canvas')
        let width = img.width
        let height = img.height

        if (width > maxWidth) {
          height = (height * maxWidth) / width
          width = maxWidth
        }

        canvas.width = width
        canvas.height = height

        const ctx = canvas.getContext('2d')
        ctx.drawImage(img, 0, 0, width, height)

        let compressedBase64 = canvas.toDataURL('image/jpeg', quality)
        if (compressedBase64.length > 500000) {
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

const loadWithdraws = async () => {
  try {
    loading.value = true
    const params = {
      status: activeTab.value,
      limit: 50
    }
    
    const data = await request.get(API_ENDPOINTS.ADMIN_WITHDRAWS_PENDING, { params })
    const list = data.withdraws || data.list || []
    
    withdraws.value = list
    finished.value = true
  } catch (error) {
    console.error('加载提现记录失败:', error)
    showToast('加载失败')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const onTabChange = () => {
  finished.value = false
  withdraws.value = []
  loadWithdraws()
}

const onRefresh = () => {
  finished.value = false
  loadWithdraws()
}

const reviewWithdraw = async (withdrawId, approved) => {
  try {
    let note = ''
    if (!approved) {
      const result = await showDialog({
        title: '拒绝原因',
        message: '请输入拒绝原因',
        showCancelButton: true
      })
      note = result || '不符合要求'
    }
    
    // 后端要求action字段: "approve" 或 "reject"，note为备注
    await request.post(
      API_ENDPOINTS.ADMIN_WITHDRAW_REVIEW.replace(':id', withdrawId),
      { 
        action: approved ? 'approve' : 'reject',
        note: note
      }
    )
    
    showToast(approved ? '已通过' : '已拒绝')
    onRefresh()
  } catch (error) {
    if (error === 'cancel') return
    console.error('审核失败:', error)
    showToast('操作失败')
  }
}

const showPayDialog = (withdrawId) => {
  currentPayId.value = withdrawId
  payVoucherFiles.value = []
  payVoucherUrl.value = ''
  showPayPopup.value = true
}

const afterReadPayVoucher = async (file) => {
  try {
    showToast('正在处理图片...')
    const compressed = await compressImage(file.file)
    payVoucherUrl.value = compressed
    showToast('图片上传成功')
  } catch (error) {
    console.error('图片处理失败:', error)
    showToast('图片处理失败')
  }
}

// 预览打款凭证
const previewPayVoucher = (withdraw) => {
  const raw = withdraw.voucher_url || withdraw.VoucherURL || ''
  if (!raw) {
    showToast('暂无打款凭证')
    return
  }

  const urls = parseImageUrls(raw)
    .map((url) => normalizeImageUrl(url))
    .filter((url) => url && url.length > 0)

  if (urls.length === 0) {
    showToast('暂无打款凭证')
    return
  }

  showImagePreview({
    images: urls,
    startPosition: 0
  })
}

const submitPay = async () => {
  try {
    if (!payVoucherUrl.value) {
      showToast('请先上传打款凭证')
      return
    }

    await request.post(
      API_ENDPOINTS.ADMIN_WITHDRAW_PAY.replace(':id', currentPayId.value),
      { voucher_url: payVoucherUrl.value }
    )

    showToast('已标记为已打款')
    showPayPopup.value = false
    onRefresh()
  } catch (error) {
    console.error('标记打款失败:', error)
    const msg = error.response?.data?.error || error.response?.data?.message || '操作失败'
    showToast(msg)
  }
}

onMounted(() => {
  loadWithdraws()
})
</script>

<style scoped>
.admin-withdraws-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 60px;
}

.withdraw-item {
  background: #fff;
  margin: 10px;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.withdraw-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.withdraw-amount {
  font-size: 20px;
  font-weight: bold;
  color: #67c23a;
}

.withdraw-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.withdraw-status.pending {
  color: #e6a23c;
  background: #fdf6ec;
}

.withdraw-status.approved {
  color: #67c23a;
  background: #f0f9ff;
}

.withdraw-status.rejected {
  color: #909399;
  background: #f4f4f5;
}

.withdraw-body {
  margin: 12px 0;
}

.withdraw-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}

.withdraw-row .label {
  color: #909399;
}

.withdraw-row .value {
  color: #303133;
}

.withdraw-footer {
  display: flex;
  gap: 12px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.withdraw-footer .van-button {
  flex: 1;
}

.empty {
  padding: 60px 0;
}

.pay-popup {
  padding-bottom: 20px;
}

.pay-content {
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
