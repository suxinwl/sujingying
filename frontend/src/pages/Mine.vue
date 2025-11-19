<template>
  <div class="mine-page">
    <van-nav-bar
      title="我的"
      fixed
      placeholder
    />
    
    <!-- 用户信息 -->
    <div class="user-card" @click="$router.push('/profile')">
      <div class="avatar">
        <van-icon name="user-circle-o" size="60" />
      </div>
      <div class="user-info">
        <div class="username">{{ userStore.userInfo?.username }}</div>
        <div class="role">{{ ROLE_TEXT[userStore.userInfo?.role] }}</div>
      </div>
      <van-icon name="arrow" />
    </div>
    
    <!-- 功能菜单：仅客户/销售员显示 -->
    <van-cell-group v-if="userStore.isCustomer || userStore.isSales">
      <van-cell title="我的订单" is-link to="/orders" icon="notes-o" />
      <van-cell title="银行卡管理" is-link to="/bank-cards" icon="credit-pay" />
      <van-cell
        title="消息通知"
        is-link
        to="/notifications"
        icon="bell"
        :badge="unreadCount"
      />
    </van-cell-group>
    
    <!-- 销售功能 -->
    <van-cell-group v-if="userStore.isSales" title="销售中心">
      <van-cell title="邀请码管理" is-link to="/invite-codes" icon="qr" />
      <van-cell title="我的客户" is-link to="/customers" icon="friends-o" />
      <van-cell title="提成记录" is-link to="/commissions" icon="gold-coin-o" />
    </van-cell-group>
    
    <!-- 平台管理：管理员 / 客服 -->
    <van-cell-group v-if="userStore.isAdmin || userStore.isSupport" title="平台管理">
      <van-cell title="平台公告" is-link to="/admin/announcements" icon="volume-o" />
      <van-cell title="消息通知" is-link to="/notifications" icon="bell" :badge="unreadCount" />
      <van-cell title="用户管理" is-link to="/admin/users" icon="manager-o" />
      <van-cell title="销售员管理" is-link to="/admin/sales" icon="friends-o" />
      <van-cell title="充值审核" is-link to="/admin/deposits" icon="completed" />
      <van-cell title="提现审核" is-link to="/admin/withdraws" icon="completed" />
      <van-cell title="平台收货地址" is-link to="/admin/platform-addresses" icon="location-o" />
      <van-cell title="收款管理" is-link to="/admin/payment-settings" icon="balance-pay" />
      <van-cell v-if="userStore.isAdmin" title="系统配置" is-link to="/admin/config" icon="setting-o" />
    </van-cell-group>
    
    <!-- 安全设置 -->
    <van-cell-group title="安全设置">
      <van-cell title="修改密码" is-link @click="showPasswordDialog = true" icon="lock" />
      <van-cell 
        title="设置支付密码" 
        is-link 
        @click="showPayPasswordDialog = true" 
        icon="shield-o"
        :label="userStore.userInfo?.has_pay_password ? '已设置' : '未设置'"
      />
    </van-cell-group>

    <!-- 关于平台 -->
    <van-cell-group>
      <van-cell title="关于平台" is-link to="/about" icon="info-o" />
    </van-cell-group>
    
    <div style="margin: 16px;">
      <van-button round block type="danger" @click="onLogout">
        退出登录
      </van-button>
    </div>
    
    <!-- 修改密码弹窗 -->
    <van-dialog
      v-model:show="showPasswordDialog"
      title="修改密码"
      show-cancel-button
      @confirm="onChangePassword"
    >
      <van-form ref="passwordFormRef">
        <van-field
          v-model="passwordForm.old_password"
          type="password"
          label="旧密码"
          placeholder="请输入旧密码"
          :rules="[{ required: true, message: '请输入旧密码' }]"
        />
        <van-field
          v-model="passwordForm.new_password"
          type="password"
          label="新密码"
          placeholder="请输入新密码"
          :rules="[
            { required: true, message: '请输入新密码' },
            { min: 6, message: '密码至少6位' }
          ]"
        />
        <van-field
          v-model="passwordForm.confirm_password"
          type="password"
          label="确认密码"
          placeholder="请再次输入新密码"
          :rules="[
            { required: true, message: '请确认密码' },
            { validator: validatePassword, message: '两次密码不一致' }
          ]"
        />
      </van-form>
    </van-dialog>
    
    <!-- 设置支付密码弹窗 -->
    <van-dialog
      v-model:show="showPayPasswordDialog"
      :title="userStore.userInfo?.has_pay_password ? '修改支付密码' : '设置支付密码'"
      show-cancel-button
      @confirm="onSetPayPassword"
    >
      <van-form ref="payPasswordFormRef">
        <van-field
          v-if="userStore.userInfo?.has_pay_password"
          v-model="payPasswordForm.old_pay_password"
          type="password"
          label="旧支付密码"
          placeholder="请输入旧支付密码"
          maxlength="6"
          :rules="[
            { required: true, message: '请输入旧支付密码' },
            { pattern: /^\d{6}$/, message: '支付密码为6位数字' }
          ]"
        />
        <van-field
          v-model="payPasswordForm.new_pay_password"
          type="password"
          label="新支付密码"
          placeholder="请输入6位数字"
          maxlength="6"
          :rules="[
            { required: true, message: '请输入新支付密码' },
            { pattern: /^\d{6}$/, message: '支付密码必须是6位数字' }
          ]"
        />
        <van-field
          v-model="payPasswordForm.confirm_pay_password"
          type="password"
          label="确认密码"
          placeholder="请再次输入支付密码"
          maxlength="6"
          :rules="[
            { required: true, message: '请确认支付密码' },
            { validator: validatePayPassword, message: '两次支付密码不一致' }
          ]"
        />
      </van-form>
    </van-dialog>
  </div>
</template>

/**
 * 个人中心页面
 * 
 * @description 用户个人中心，包含用户信息展示、功能菜单、密码修改等功能
 * @author 速金盈技术团队
 * @date 2025-11-18
 */

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showConfirmDialog } from 'vant'
import { useUserStore } from '../stores/user'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'
import { ROLE_TEXT } from '../utils/helpers'

// ==================== 路由和状态管理 ====================
const router = useRouter()
const userStore = useUserStore()

// ==================== 响应式数据 ====================
/** @type {import('vue').Ref<number>} 未读通知数量 */
const unreadCount = ref(0)

/** @type {import('vue').Ref<boolean>} 是否显示修改密码对话框 */
const showPasswordDialog = ref(false)

/** @type {import('vue').Ref<any>} 密码表单引用 */
const passwordFormRef = ref(null)

/** @type {import('vue').Ref<{old_password: string, new_password: string, confirm_password: string}>} 密码表单数据 */
const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

/** @type {import('vue').Ref<boolean>} 是否显示支付密码对话框 */
const showPayPasswordDialog = ref(false)

/** @type {import('vue').Ref<any>} 支付密码表单引用 */
const payPasswordFormRef = ref(null)

/** @type {import('vue').Ref<{old_pay_password: string, new_pay_password: string, confirm_pay_password: string}>} 支付密码表单数据 */
const payPasswordForm = ref({
  old_pay_password: '',
  new_pay_password: '',
  confirm_pay_password: ''
})

// ==================== 表单验证 ====================
/**
 * 验证密码是否一致
 * 
 * @returns {boolean} 两次输入的密码是否一致
 */
const validatePassword = () => {
  return passwordForm.value.confirm_password === passwordForm.value.new_password
}

/**
 * 验证支付密码是否一致
 * 
 * @returns {boolean} 两次输入的支付密码是否一致
 */
const validatePayPassword = () => {
  return payPasswordForm.value.confirm_pay_password === payPasswordForm.value.new_pay_password
}

// ==================== 数据加载 ====================
/**
 * 获取未读通知数量
 * 
 * @async
 * @returns {Promise<void>}
 * @description 从API获取未读通知的数量，用于显示消息角标
 */
const loadUnreadCount = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.NOTIFICATIONS, {
      params: { is_read: false, page_size: 1 }
    })
    unreadCount.value = data?.total || data?.count || 0
  } catch (error) {
    console.error('获取未读通知失败:', error)
    unreadCount.value = 0
  }
}

// ==================== 事件处理 ====================
/**
 * 修改密码
 * 
 * @async
 * @returns {Promise<void>}
 * @description 验证并提交密码修改请求，成功后关闭对话框并清空表单
 */
const onChangePassword = async () => {
  try {
    // 验证表单
    await passwordFormRef.value?.validate()
    
    // 调用API修改密码
    await userStore.changePassword({
      old_password: passwordForm.value.old_password,
      new_password: passwordForm.value.new_password
    })
    
    // 成功提示
    showToast('密码修改成功')
    
    // 重置表单
    passwordForm.value = {
      old_password: '',
      new_password: '',
      confirm_password: ''
    }
    
    // 关闭对话框
    showPasswordDialog.value = false
  } catch (error) {
    console.error('修改密码失败:', error)
  }
}

/**
 * 设置/修改支付密码
 * 
 * @async
 * @returns {Promise<void>}
 * @description 验证并提交支付密码设置请求，成功后关闭对话框并清空表单
 */
const onSetPayPassword = async () => {
  try {
    // 验证表单
    await payPasswordFormRef.value?.validate()
    
    const hasPayPassword = userStore.userInfo?.has_pay_password
    
    // 调用API设置/修改支付密码
    await request.post(API_ENDPOINTS.PAYPASS, {
      old_pay_password: hasPayPassword ? payPasswordForm.value.old_pay_password : undefined,
      new_pay_password: payPasswordForm.value.new_pay_password
    })
    
    // 成功提示
    showToast(hasPayPassword ? '支付密码修改成功' : '支付密码设置成功')
    
    // 更新用户信息
    await userStore.getUserInfo()
    
    // 重置表单
    payPasswordForm.value = {
      old_pay_password: '',
      new_pay_password: '',
      confirm_pay_password: ''
    }
    
    // 关闭对话框
    showPayPasswordDialog.value = false
  } catch (error) {
    console.error('设置支付密码失败:', error)
    showToast(error.response?.data?.error || error.response?.data?.message || '操作失败')
  }
}

/**
 * 退出登录
 * 
 * @async
 * @returns {Promise<void>}
 * @description 确认后退出登录，清除本地数据并跳转到登录页
 */
const onLogout = async () => {
  try {
    // 显示确认对话框
    await showConfirmDialog({
      title: '确认退出',
      message: '确定要退出登录吗？'
    })
    
    // 调用退出API
    await userStore.logout()
    
    // 跳转到登录页
    router.replace('/login')
    
    // 提示
    showToast('已退出登录')
  } catch (error) {
    // 如果是取消操作，不处理
    if (error !== 'cancel') {
      console.error('退出登录失败:', error)
    }
  }
}

// ==================== 生命周期 ====================
/**
 * 组件挂载时执行
 */
onMounted(() => {
  loadUnreadCount()
})
</script>

<style scoped>
.mine-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}

.user-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 30px 20px;
  display: flex;
  align-items: center;
  color: #fff;
  margin-bottom: 10px;
}

.avatar {
  margin-right: 16px;
}

.user-info {
  flex: 1;
}

.username {
  font-size: 20px;
  font-weight: bold;
  margin-bottom: 8px;
}

.role {
  font-size: 14px;
  opacity: 0.9;
}
</style>
