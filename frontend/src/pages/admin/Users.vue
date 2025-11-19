<template>
  <div class="admin-users-page">
    <van-nav-bar
      title="用户管理"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    />
    
    <!-- 搜索和筛选 -->
    <div class="filter-section">
      <van-search
        v-model="searchQuery"
        placeholder="搜索手机号"
        @search="onSearch"
      />
      <van-tabs v-model:active="activeTab" @change="onTabChange">
        <van-tab title="全部" name="all" />
        <van-tab title="待审核" name="pending" />
        <van-tab title="已激活" name="active" />
        <van-tab title="已禁用" name="disabled" />
      </van-tabs>
    </div>
    
    <!-- 用户列表 -->
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        finished-text="没有更多了"
        @load="loadUsers"
      >
        <div v-if="users.length === 0" class="empty">
          <van-empty description="暂无用户" />
        </div>
        
        <div v-for="user in users" :key="user.id" class="user-item">
          <div class="user-header">
            <div class="user-info">
              <span class="user-phone">{{ user.phone }}</span>
              <span class="user-role" :class="user.role">{{ getRoleText(user.role) }}</span>
            </div>
            <span class="user-status" :class="user.status">
              {{ getStatusText(user.status) }}
            </span>
          </div>
          
          <div class="user-body">
            <div class="user-row">
              <span class="label">ID:</span>
              <span class="value">{{ user.id }}</span>
            </div>
            <div class="user-row">
              <span class="label">姓名:</span>
              <span class="value">{{ user.realname || '-' }}</span>
            </div>
            <div class="user-row">
              <span class="label">实名状态:</span>
              <span class="value">{{ getVerifyStatusText(user.verify_status) }}</span>
            </div>
            <div class="user-row">
              <span class="label">可用定金:</span>
              <span class="value">¥{{ formatMoney(user.available_deposit) }}</span>
            </div>
            <div class="user-row">
              <span class="label">已用定金:</span>
              <span class="value">¥{{ formatMoney(user.used_deposit) }}</span>
            </div>
            <div class="user-row">
              <span class="label">待结金料:</span>
              <span class="value">{{ (user.pending_weight_g || 0).toFixed(3) }}g</span>
            </div>
            <div class="user-row">
              <span class="label">已结金料:</span>
              <span class="value">{{ (user.settled_weight_g || 0).toFixed(3) }}g</span>
            </div>
            <div class="user-row">
              <span class="label">上级销售:</span>
              <span class="value">
                {{ user.sales_name || (user.sales_id ? 'ID:' + user.sales_id : '-') }}
              </span>
            </div>
            <div class="user-row">
              <span class="label">自动补定金:</span>
              <span class="value">{{ user.auto_supplement_enabled ? '已启用' : '未启用' }}</span>
            </div>
            <div class="user-row">
              <span class="label">注册时间:</span>
              <span class="value">{{ formatDateTime(user.created_at) }}</span>
            </div>
          </div>
          
          <div class="user-footer" v-if="user.status === 'pending'">
            <van-button size="small" type="success" @click="approveUser(user.id, true)">
              通过
            </van-button>
            <van-button size="small" type="danger" @click="approveUser(user.id, false)">
              拒绝
            </van-button>
          </div>

          <div class="user-footer actions">
            <van-button
              size="small"
              type="primary"
              plain
              @click="openSalesDialog(user)"
            >
              设置上级
            </van-button>
            <van-button
              size="small"
              type="primary"
              plain
              @click="openVerificationDialog(user)"
            >
              实名审核
            </van-button>
            <van-button
              v-if="user.role === 'customer' && user.status === 'active'"
              size="small"
              type="warning"
              plain
              @click="setAsSales(user)"
            >
              设为销售员
            </van-button>
            <van-button
              v-if="user.role === 'sales'"
              size="small"
              type="danger"
              plain
              @click="revertSales(user)"
            >
              撤回为客户
            </van-button>
            <van-button
              v-if="user.role === 'customer'"
              size="small"
              type="success"
              plain
              @click="toggleAutoSupplement(user)"
            >
              {{ user.auto_supplement_enabled ? '禁用自动补定金' : '启用自动补定金' }}
            </van-button>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>

    <!-- 设置上级销售弹窗 -->
    <van-dialog
      v-model:show="showSalesDialog"
      title="设置上级销售"
      show-cancel-button
      @confirm="onConfirmSales"
    >
      <van-field
        v-model="salesIdInput"
        label="上级ID"
        type="number"
        placeholder="请输入上级销售员的用户ID"
        clearable
      />
    </van-dialog>

    <!-- 实名认证详情弹窗 -->
    <van-dialog
      v-model:show="showVerifyDialog"
      title="实名认证信息"
      show-cancel-button
      :confirm-button-text="'关闭'"
      @confirm="showVerifyDialog = false"
    >
      <div class="verify-content">
        <div v-if="verifyLoading" class="verify-row">正在加载...</div>
        <div v-else-if="!verification" class="verify-row">该用户尚未提交实名认证信息</div>
        <div v-else>
          <div class="verify-row"><span class="label">状态：</span><span class="value">{{ getVerifyStatusText(verification.status) }}</span></div>
          <div class="verify-row"><span class="label">姓名：</span><span class="value">{{ verification.real_name }}</span></div>
          <div class="verify-row"><span class="label">身份证号：</span><span class="value">{{ verification.id_number }}</span></div>
          <div class="verify-row" v-if="verification.id_front_url">
            <span class="label">身份证正面：</span>
            <span class="value">
              <van-image
                width="120"
                height="80"
                fit="cover"
                :src="verification.id_front_url"
              />
            </span>
          </div>
          <div class="verify-row" v-if="verification.id_back_url">
            <span class="label">身份证反面：</span>
            <span class="value">
              <van-image
                width="120"
                height="80"
                fit="cover"
                :src="verification.id_back_url"
              />
            </span>
          </div>
          <div class="verify-row"><span class="label">收货人：</span><span class="value">{{ verification.receiver_name }}</span></div>
          <div class="verify-row"><span class="label">联系电话：</span><span class="value">{{ verification.receiver_phone }}</span></div>
          <div class="verify-row"><span class="label">收货地址：</span><span class="value">{{ formatVerifyAddress(verification) }}</span></div>
          <div class="verify-row">
            <span class="label">审核备注：</span>
          </div>
          <van-field
            v-model="verifyRemark"
            type="textarea"
            rows="2"
            placeholder="可填写审核说明（选填）"
          />

          <div
            v-if="verification.status === 'pending'"
            class="verify-actions"
          >
            <van-button
              size="small"
              type="success"
              block
              style="margin-bottom: 8px;"
              @click="approveVerification"
            >
              审核通过
            </van-button>
            <van-button
              size="small"
              type="danger"
              block
              @click="rejectVerification"
            >
              审核驳回
            </van-button>
          </div>
        </div>
      </div>
    </van-dialog>
  </div>
</template>

<script setup>
/**
 * @file Users.vue
 * @description 用户管理页面 - 管理员查看和审核用户
 * @author AI
 * @date 2025-11-18
 */

import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { showToast, showDialog } from 'vant'
import request from '../../utils/request'
import { API_ENDPOINTS } from '../../config/api'
import { formatMoney, formatDateTime } from '../../utils/helpers'

/**
 * 当前Tab
 * @type {import('vue').Ref<string>}
 */
const activeTab = ref('all')
const route = useRoute()

/**
 * 搜索关键词
 * @type {import('vue').Ref<string>}
 */
const searchQuery = ref('')

/**
 * 用户列表
 * @type {import('vue').Ref<Array>}
 */
const users = ref([])

/**
 * 刷新状态
 * @type {import('vue').Ref<boolean>}
 */
const refreshing = ref(false)

/**
 * 加载状态
 * @type {import('vue').Ref<boolean>}
 */
const loading = ref(false)

/**
 * 是否加载完成
 * @type {import('vue').Ref<boolean>}
 */
const finished = ref(false)

/**
 * 当前页码
 * @type {import('vue').Ref<number>}
 */
const page = ref(1)

const showSalesDialog = ref(false)
const currentUserForSales = ref(null)
const salesIdInput = ref('')

const showVerifyDialog = ref(false)
const verifyLoading = ref(false)
const currentUserForVerify = ref(null)
const verification = ref(null)
const verifyRemark = ref('')

/**
 * 获取角色文本
 * @param {string} role - 角色
 * @returns {string} 角色文本
 */
const getRoleText = (role) => {
  const roleMap = {
    customer: '客户',
    sales: '销售',
    support: '客服',
    super_admin: '管理员'
  }
  return roleMap[role] || role
}

/**
 * 获取状态文本
 * @param {string} status - 状态
 * @returns {string} 状态文本
 */
const getStatusText = (status) => {
  const statusMap = {
    pending: '待审核',
    active: '已激活',
    disabled: '已禁用'
  }
  return statusMap[status] || status
}

/**
 * 加载用户列表
 * @async
 * @returns {Promise<void>}
 */
const loadUsers = async () => {
  try {
    loading.value = true
    const params = {
      page: page.value,
      page_size: 10
    }
    
    if (activeTab.value !== 'all') {
      params.status = activeTab.value
    }
    
    if (searchQuery.value) {
      params.phone = searchQuery.value
    }

    // 如果通过路由 query 传入 sales_id，则按销售员过滤客户列表
    const salesIdFromRoute = route.query.sales_id
    if (salesIdFromRoute) {
      params.sales_id = salesIdFromRoute
    }
    
    const data = await request.get(API_ENDPOINTS.ADMIN_USERS, { params })
    
    const list = data.users || data.list || []
    
    if (page.value === 1) {
      users.value = list
    } else {
      users.value.push(...list)
    }
    
    if (list.length < 10) {
      finished.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载用户失败:', error)
    showToast('加载失败')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const getVerifyStatusText = (status) => {
  const map = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已驳回'
  }
  return map[status] || status || '未提交'
}

const formatVerifyAddress = (v) => {
  if (!v) return ''
  const parts = [v.province, v.city, v.district, v.address_detail].filter(Boolean)
  return parts.join(' ')
}

const revertSales = async (user) => {
  try {
    await showDialog({
      title: '撤回为客户',
      message: `确定将 ${user.phone} 撤回为普通客户吗？`,
      showCancelButton: true,
      closeOnClickOverlay: true
    })

    await request.put(
      API_ENDPOINTS.ADMIN_USER_DETAIL.replace(':id', user.id),
      {
        role: 'customer'
      }
    )

    user.role = 'customer'
    showToast('已撤回为客户')
  } catch (error) {
    if (error === 'cancel') return
    console.error('撤回销售员失败:', error)
    showToast('操作失败')
  }
}

const openSalesDialog = (user) => {
  currentUserForSales.value = user
  salesIdInput.value = user.sales_id || ''
  showSalesDialog.value = true
}

const onConfirmSales = async () => {
  if (!currentUserForSales.value) return
  try {
    const userId = currentUserForSales.value.id
    const salesId = Number(salesIdInput.value) || 0
    await request.put(
      API_ENDPOINTS.ADMIN_USER_DETAIL.replace(':id', userId),
      {
        sales_id: salesId,
      }
    )
    currentUserForSales.value.sales_id = salesId
    showToast('上级销售已更新')
    showSalesDialog.value = false
  } catch (error) {
    console.error('设置上级销售失败:', error)
    showToast('操作失败')
  }
}

const setAsSales = async (user) => {
  try {
    await request.put(
      API_ENDPOINTS.ADMIN_USER_DETAIL.replace(':id', user.id),
      {
        role: 'sales',
      }
    )
    user.role = 'sales'
    showToast('已设置为销售员')
  } catch (error) {
    console.error('设置为销售员失败:', error)
    showToast('操作失败')
  }
}

const toggleAutoSupplement = async (user) => {
  try {
    const enabled = !user.auto_supplement_enabled
    await request.post(
      API_ENDPOINTS.ADMIN_USER_TOGGLE_AUTO_SUPPLEMENT.replace(':id', user.id),
      {
        enabled,
      }
    )
    user.auto_supplement_enabled = enabled
    showToast(enabled ? '已启用自动补定金' : '已禁用自动补定金')
  } catch (error) {
    console.error('自动补定金开关失败:', error)
    showToast('操作失败')
  }
}

const openVerificationDialog = async (user) => {
  currentUserForVerify.value = user
  showVerifyDialog.value = true
  verifyLoading.value = true
  verification.value = null
  verifyRemark.value = ''
  try {
    const data = await request.get(
      API_ENDPOINTS.ADMIN_USER_VERIFICATION.replace(':id', user.id)
    )
    verification.value = data.verification || null
    if (verification.value && verification.value.remark) {
      verifyRemark.value = verification.value.remark
    }
  } catch (error) {
    console.error('加载实名认证信息失败:', error)
    showToast('加载失败')
  } finally {
    verifyLoading.value = false
  }
}

const approveVerification = async () => {
  if (!currentUserForVerify.value || !verification.value) return
  try {
    await request.post(
      API_ENDPOINTS.ADMIN_USER_VERIFICATION_APPROVE.replace(
        ':id',
        currentUserForVerify.value.id
      ),
      { remark: verifyRemark.value || '' }
    )
    verification.value.status = 'approved'
    verification.value.remark = verifyRemark.value || ''
    showToast('实名认证已通过')
  } catch (error) {
    console.error('实名认证审核通过失败:', error)
    showToast('操作失败')
  }
}

const rejectVerification = async () => {
  if (!currentUserForVerify.value || !verification.value) return
  try {
    await showDialog({
      title: '驳回实名认证',
      message: '确定要驳回该用户的实名认证吗？',
      showCancelButton: true,
      closeOnClickOverlay: true
    })

    await request.post(
      API_ENDPOINTS.ADMIN_USER_VERIFICATION_REJECT.replace(
        ':id',
        currentUserForVerify.value.id
      ),
      { remark: verifyRemark.value || '' }
    )
    verification.value.status = 'rejected'
    verification.value.remark = verifyRemark.value || ''
    showToast('实名认证已驳回')
  } catch (error) {
    if (error === 'cancel') return
    console.error('实名认证驳回失败:', error)
    showToast('操作失败')
  }
}

/**
 * Tab切换
 * @returns {void}
 */
const onTabChange = () => {
  page.value = 1
  finished.value = false
  users.value = []
  loadUsers()
}

/**
 * 搜索
 * @returns {void}
 */
const onSearch = () => {
  page.value = 1
  finished.value = false
  users.value = []
  loadUsers()
}

/**
 * 下拉刷新
 * @returns {void}
 */
const onRefresh = () => {
  page.value = 1
  finished.value = false
  loadUsers()
}

/**
 * 审核用户
 * @async
 * @param {number} userId - 用户ID
 * @param {boolean} approved - 是否通过
 * @returns {Promise<void>}
 */
const approveUser = async (userId, approved) => {
  try {
    let note = ''
    if (!approved) {
      const result = await showDialog({
        title: '拒绝原因',
        message: '请输入拒绝原因',
        showCancelButton: true,
        closeOnClickOverlay: true
      })
      note = result || '不符合审核标准'
    }
    
    // 后端要求action字段: "approve" 或 "reject"，note为备注
    await request.post(
      API_ENDPOINTS.ADMIN_USER_APPROVE.replace(':id', userId),
      { 
        action: approved ? 'approve' : 'reject',
        note: note
      }
    )
    
    showToast(approved ? '审核通过' : '已拒绝')
    onRefresh()
  } catch (error) {
    if (error === 'cancel') return
    console.error('审核失败:', error)
    showToast('操作失败')
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.admin-users-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 60px;
}

.filter-section {
  background: #fff;
  margin-bottom: 10px;
}

.user-item {
  background: #fff;
  margin: 10px;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.user-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-phone {
  font-size: 16px;
  font-weight: 500;
}

.user-role {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.user-role.customer {
  color: #409eff;
  background: #ecf5ff;
}

.user-role.sales {
  color: #67c23a;
  background: #f0f9ff;
}

.user-role.super_admin {
  color: #f56c6c;
  background: #fef0f0;
}

.user-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.user-status.pending {
  color: #e6a23c;
  background: #fdf6ec;
}

.user-status.active {
  color: #67c23a;
  background: #f0f9ff;
}

.user-status.disabled {
  color: #909399;
  background: #f4f4f5;
}

.user-body {
  margin: 12px 0;
}

.user-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}

.user-row .label {
  color: #909399;
}

.user-row .value {
  color: #303133;
}

.user-footer {
  display: flex;
  gap: 12px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.user-footer .van-button {
  flex: 1;
}

.empty {
  padding: 60px 0;
}
</style>
