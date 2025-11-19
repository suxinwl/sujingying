<template>
  <div class="invite-codes-page">
    <van-nav-bar
      title="邀请码管理"
      left-arrow
      fixed
      placeholder
      @click-left="$router.back()"
    />

    <div v-if="userStore.isSales" class="content">
      <van-cell-group inset>
        <van-cell
          title="我的邀请码"
          :value="code || '暂无'"
        />
        <van-cell
          title="邀请人数"
          :value="inviteCount >= 0 ? inviteCount : '-'"
        />
        <van-cell
          title="注册人数"
          :value="registerCount >= 0 ? registerCount : '-'"
        />
        <van-cell title="注册链接">
          <template #label>
            <div class="share-url">{{ shareUrl || '暂无' }}</div>
          </template>
        </van-cell>
      </van-cell-group>

      <div v-if="qrCode" class="qr-section">
        <div class="qr-title">邀请二维码</div>
        <van-image
          width="200"
          height="200"
          fit="contain"
          :src="qrCode"
        />
        <div class="qr-tip">客户扫码即可打开注册链接</div>
      </div>

      <div style="margin: 16px;">
        <van-button
          round
          block
          type="primary"
          :loading="loading"
          @click="copyInviteInfo"
        >
          复制邀请码及链接
        </van-button>
      </div>

      <div class="tips">
        <p>将邀请码或注册链接发送给客户，对方使用后会自动成为您的归属客户。</p>
      </div>
    </div>

    <div v-else class="no-permission">
      <van-empty description="仅销售人员可以使用邀请码功能" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import { useUserStore } from '../stores/user'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'

const userStore = useUserStore()

const code = ref('')
const shareUrl = ref('')
const inviteCount = ref(-1)
const registerCount = ref(-1)
const qrCode = ref('')
const loading = ref(false)

const loadInviteCode = async () => {
  if (!userStore.isSales) return
  try {
    loading.value = true
    // 使用 /invitation/my-code 接口获取当前销售的专属邀请码
    const data = await request.get(API_ENDPOINTS.INVITATION_MY_CODE)
    code.value = data.code || ''
    inviteCount.value = typeof data.invite_count === 'number' ? data.invite_count : -1
    registerCount.value = typeof data.register_count === 'number' ? data.register_count : -1
    qrCode.value = data.qr_code || ''
    // 使用当前站点域名拼接本地注册链接，避免生产域名与本地不一致
    if (code.value) {
      shareUrl.value = `${window.location.origin}/register?code=${code.value}`
    } else {
      shareUrl.value = ''
    }
  } catch (error) {
    console.error('获取邀请码失败:', error)
    showToast('获取邀请码失败')
  } finally {
    loading.value = false
  }
}

const copyInviteInfo = async () => {
  if (!code.value) {
    showToast('暂无可用邀请码')
    return
  }
  const text = `邀请码：${code.value}\n注册链接：${shareUrl.value || ''}`
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text)
    } else {
      // 兼容处理
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    showToast('已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    showToast('复制失败，请手动复制')
  }
}

onMounted(() => {
  if (!userStore.isSales) {
    showToast('只有销售人员可以使用邀请码功能')
    return
  }
  loadInviteCode()
})
</script>

<style scoped>
.invite-codes-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}

.content {
  padding-top: 8px;
}

.share-url {
  word-break: break-all;
  color: #1989fa;
  font-size: 13px;
}

.tips {
  padding: 0 24px 24px;
  font-size: 12px;
  color: #999;
  line-height: 1.6;
}

.no-permission {
  padding-top: 80px;
}
</style>
