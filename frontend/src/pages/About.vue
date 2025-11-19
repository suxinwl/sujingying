<template>
  <div class="about-page">
    <van-nav-bar
      title="关于"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    />
    
    <div class="content">
      <!-- Logo -->
      <div class="logo-section">
        <div class="logo">{{ platformName || '速金盈' }}</div>
        <div class="version">v1.0.0</div>
      </div>
      
      <!-- 信息列表 -->
      <van-cell-group inset>
        <van-cell :title="'公司名称'" :value="companyName || '-'" />
        <van-cell
          title="联系电话"
          :value="servicePhone || '-'"
          is-link
          @click="callPhone"
        />
        <van-cell
          title="客服微信"
          :value="serviceWechat || '-'"
          is-link
          @click="copyWechat"
        />
        <van-cell :title="'营业时间'" :value="businessTimeText || '未配置'" />
      </van-cell-group>
      
      <van-cell-group inset style="margin-top: 20px;">
        <van-cell title="用户协议" is-link @click="openAgreement('user')" />
        <van-cell title="隐私政策" is-link @click="openAgreement('privacy')" />
        <van-cell title="风险提示" is-link @click="openAgreement('risk')" />
        <van-cell title="贵金属购销服务协议" is-link @click="openAgreement('metal')" />
      </van-cell-group>
      
      <!-- 描述 -->
      <div class="description">
        <p v-if="platformIntro">{{ platformIntro }}</p>
        <template v-else>
          <p>速金盈是一家专注于贵金属交易的金融科技公司，为用户提供安全、便捷的黄金交易服务。</p>
          <p>我们致力于打造最专业的贵金属交易平台，让投资更简单。</p>
        </template>
      </div>
      
      <!-- 版权信息 -->
      <div class="copyright">
        <p>Copyright © 2025 {{ companyName || '速金盈科技' }}</p>
        <p>All Rights Reserved</p>
      </div>
    </div>
  </div>
</template>

<script setup>
/**
 * @file About.vue
 * @description 关于页面，展示平台和公司信息，协议入口；数据来源于系统配置
 */

import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'

const router = useRouter()

const platformName = ref('')
const platformIntro = ref('')
const companyName = ref('')
const companyAddress = ref('')
const servicePhone = ref('')
const serviceWechat = ref('')
const tradingStartTime = ref('')
const tradingEndTime = ref('')
const tradingDays = ref('')

const userAgreementUrl = ref('')
const privacyPolicyUrl = ref('')
const riskWarningUrl = ref('')
const metalServiceAgreementUrl = ref('')

const tradingDayMap = {
  1: '周一',
  2: '周二',
  3: '周三',
  4: '周四',
  5: '周五',
  6: '周六',
  7: '周日'
}

const businessTimeText = computed(() => {
  if (!tradingDays.value && !tradingStartTime.value && !tradingEndTime.value) {
    return ''
  }
  let dayText = ''
  if (tradingDays.value) {
    const parts = String(tradingDays.value)
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
      .map((s) => tradingDayMap[s] || s)
    if (parts.length) {
      dayText = parts.join('、')
    }
  }
  const timeText = tradingStartTime.value && tradingEndTime.value
    ? `${tradingStartTime.value}-${tradingEndTime.value}`
    : ''
  if (dayText && timeText) return `${dayText} ${timeText}`
  if (dayText) return dayText
  return timeText
})

const loadConfig = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.CONFIG)
    if (data && data.configs && data.configs.length > 0) {
      data.configs.forEach((item) => {
        const key = item.key || item.Key
        const value = item.value || item.Value
        switch (key) {
          case 'platform_name':
            platformName.value = value
            break
          case 'platform_intro':
            platformIntro.value = value
            break
          case 'company_name':
            companyName.value = value
            break
          case 'company_address':
            companyAddress.value = value
            break
          case 'service_phone':
            servicePhone.value = value
            break
          case 'service_wechat':
            serviceWechat.value = value
            break
          case 'trading_start_time':
            tradingStartTime.value = value
            break
          case 'trading_end_time':
            tradingEndTime.value = value
            break
          case 'trading_days':
            tradingDays.value = value
            break
          case 'user_agreement_url':
            userAgreementUrl.value = value
            break
          case 'privacy_policy_url':
            privacyPolicyUrl.value = value
            break
          case 'risk_warning_url':
            riskWarningUrl.value = value
            break
          case 'metal_service_agreement_url':
            metalServiceAgreementUrl.value = value
            break
          default:
            break
        }
      })
    }
  } catch (error) {
    console.error('加载系统配置失败:', error)
  }
}

const callPhone = () => {
  if (!servicePhone.value) {
    showToast('暂未配置客服电话')
    return
  }
  window.location.href = `tel:${servicePhone.value}`
}

const copyWechat = () => {
  if (!serviceWechat.value) {
    showToast('暂未配置客服微信')
    return
  }
  const wechat = serviceWechat.value
  if (navigator.clipboard) {
    navigator.clipboard.writeText(wechat).then(() => {
      showToast('微信号已复制')
    })
  } else {
    showToast('微信号: ' + wechat)
  }
}

const openAgreement = (type) => {
  const map = {
    user: { title: '用户协议', url: userAgreementUrl.value },
    privacy: { title: '隐私政策', url: privacyPolicyUrl.value },
    risk: { title: '风险提示', url: riskWarningUrl.value },
    metal: { title: '贵金属购销服务协议', url: metalServiceAgreementUrl.value }
  }
  const item = map[type]
  if (!item) return
  if (!item.url) {
    showToast(`${item.title}暂未配置链接`)
    return
  }
  // 统一跳外部 H5，或内部路由
  if (item.url.startsWith('http')) {
    window.location.href = item.url
  } else {
    router.push(item.url)
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.about-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}

.content {
  padding: 20px;
}

.logo-section {
  text-align: center;
  padding: 40px 0;
}

.logo {
  font-size: 36px;
  font-weight: bold;
  color: #667eea;
  margin-bottom: 12px;
}

.version {
  font-size: 14px;
  color: #909399;
}

.description {
  padding: 30px 20px;
  text-align: center;
  color: #606266;
  line-height: 1.8;
}

.description p {
  margin: 8px 0;
}

.copyright {
  text-align: center;
  padding: 40px 0;
  color: #c0c4cc;
  font-size: 12px;
  line-height: 1.6;
}
</style>
