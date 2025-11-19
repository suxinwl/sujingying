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
        <div class="logo">速金盈</div>
        <div class="version">v1.0.0</div>
      </div>
      
      <!-- 信息列表 -->
      <van-cell-group inset>
        <van-cell title="公司名称" value="速金盈科技有限公司" />
        <van-cell title="联系电话" value="400-888-8888" is-link @click="callPhone" />
        <van-cell title="客服微信" value="suxinying888" is-link @click="copyWechat" />
        <van-cell title="营业时间" value="周一至周五 9:00-18:00" />
      </van-cell-group>
      
      <van-cell-group inset style="margin-top: 20px;">
        <van-cell title="用户协议" is-link @click="showAgreement('user')" />
        <van-cell title="隐私政策" is-link @click="showAgreement('privacy')" />
        <van-cell title="风险提示" is-link @click="showAgreement('risk')" />
      </van-cell-group>
      
      <!-- 描述 -->
      <div class="description">
        <p>速金盈是一家专注于贵金属交易的金融科技公司，为用户提供安全、便捷的黄金交易服务。</p>
        <p>我们致力于打造最专业的贵金属交易平台，让投资更简单。</p>
      </div>
      
      <!-- 版权信息 -->
      <div class="copyright">
        <p>Copyright © 2025 速金盈科技</p>
        <p>All Rights Reserved</p>
      </div>
    </div>
  </div>
</template>

<script setup>
/**
 * @file About.vue
 * @description 关于页面
 * @author AI
 * @date 2025-11-18
 */

import { showToast } from 'vant'

/**
 * 拨打电话
 * @returns {void}
 */
const callPhone = () => {
  window.location.href = 'tel:400-888-8888'
}

/**
 * 复制微信号
 * @returns {void}
 */
const copyWechat = () => {
  // 复制到剪贴板
  const wechat = 'suxinying888'
  if (navigator.clipboard) {
    navigator.clipboard.writeText(wechat).then(() => {
      showToast('微信号已复制')
    })
  } else {
    showToast('微信号: ' + wechat)
  }
}

/**
 * 显示协议内容
 * @param {string} type - 协议类型
 * @returns {void}
 */
const showAgreement = (type) => {
  const titles = {
    user: '用户协议',
    privacy: '隐私政策',
    risk: '风险提示'
  }
  showToast(`${titles[type]}功能开发中`)
}
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
