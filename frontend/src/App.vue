<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { showDialog } from 'vant'
import { useUserStore } from './stores/user'
import { notifyWS } from '@/utils/notificationWebSocket'

const userStore = useUserStore()

onMounted(async () => {
  // 如果有token，尝试获取用户信息
  if (userStore.isLogin) {
    try {
      await userStore.getUserInfo()
    } catch (error) {
      console.error('获取用户信息失败:', error)
    }
  }

  // 连接通知 WebSocket，用于公告等实时消息
  if (userStore.isLogin) {
    notifyWS.connect()

    // 监听公告、风控、资金等实时通知
    notifyWS.onMessage((n) => {
      // 后端 Notification 结构：Type/Title/Content/Level
      const type = n.type || n.Type
      const title = n.title || n.Title || ''
      const content = n.content || n.Content || ''

      // 平台公告：所有在线用户弹窗提示
      if (type === 'announce') {
        showDialog({
          title: title || '平台公告',
          message: content,
        })
        return
      }

      // 风控通知：阈值预警、自动补定金、自动/强制平仓等
      if (type === 'risk') {
        showDialog({
          title: title || '风险提醒',
          message: content,
        })
        return
      }

      // 资金通知：充值成功、充值驳回、提现通过、提现驳回、已打款等
      if (type === 'fund') {
        showDialog({
          title: title || '资金提醒',
          message: content,
        })
        return
      }

      // 系统通知：新用户注册待审核等（通常发给客服/管理员/销售）
      if (type === 'system') {
        showDialog({
          title: title || '系统通知',
          message: content,
        })
      }
    })
  }
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

#app {
  width: 100%;
  min-height: 100vh;
}

html, body, #app { height: 100%; margin: 0; }
</style>
