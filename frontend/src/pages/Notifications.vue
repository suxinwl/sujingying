<template>
  <div class="notifications-page">
    <van-nav-bar
      title="系统通知"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    >
      <template #right>
        <van-button size="small" type="primary" @click="markAllRead" v-if="unreadCount > 0">
          全部已读
        </van-button>
      </template>
    </van-nav-bar>
    
    <!-- 通知列表 -->
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        finished-text="没有更多了"
        @load="loadNotifications"
      >
        <div v-if="notifications.length === 0" class="empty">
          <van-empty description="暂无通知" />
        </div>
        
        <div
          v-for="notification in notifications"
          :key="notification.id || notification.ID"
          class="notification-item"
          :class="{ unread: !notification.is_read }"
          @click="markRead(notification)"
        >
          <div class="notification-header">
            <span class="notification-title">{{ notification.title }}</span>
            <span v-if="!notification.is_read" class="unread-badge"></span>
          </div>
          <div class="notification-content">
            {{ notification.content }}
          </div>
          <div class="notification-footer">
            <span class="notification-time">
              {{ formatDateTime(notification.created_at || notification.CreatedAt) }}
            </span>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>
  </div>
</template>

<script setup>
/**
 * @file Notifications.vue
 * @description 系统通知页面
 * @author AI
 * @date 2025-11-18
 */

import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'
import { formatDateTime } from '../utils/helpers'

/**
 * 通知列表
 * @type {import('vue').Ref<Array>}
 */
const notifications = ref([])

/**
 * 未读数量
 * @type {import('vue').Ref<number>}
 */
const unreadCount = ref(0)

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
const pageSize = 10

/**
 * 加载通知列表
 * @async
 * @returns {Promise<void>}
 */
const loadNotifications = async () => {
  try {
    loading.value = true
    const params = {
      limit: pageSize,
      offset: (page.value - 1) * pageSize
    }
    
    const data = await request.get(API_ENDPOINTS.NOTIFICATIONS, { params })
    
    const rawList = data.notifications || data.list || []

    const list = rawList.map(item => {
      const id = (typeof item.id !== 'undefined') ? item.id : item.ID
      const title = item.title || item.Title
      const content = item.content || item.Content
      const createdAt = item.created_at || item.CreatedAt
      const status = item.status || item.Status
      const isRead = (typeof item.is_read !== 'undefined')
        ? item.is_read
        : status === 'read'

      return {
        ...item,
        id,
        title,
        content,
        created_at: createdAt,
        is_read: isRead
      }
    })
    
    if (page.value === 1) {
      notifications.value = list
    } else {
      notifications.value.push(...list)
    }
    
    // 更新未读数
    unreadCount.value = notifications.value.filter(n => !n.is_read).length
    
    if (list.length < pageSize) {
      finished.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载通知失败:', error)
    showToast('加载失败')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

/**
 * 标记通知为已读
 * @async
 * @param {Object} notification - 通知对象
 * @returns {Promise<void>}
 */
const markRead = async (notification) => {
  if (notification.is_read) return
  
  try {
    await request.post(API_ENDPOINTS.NOTIFICATION_READ, {
      ids: [notification.id]
    })
    notification.is_read = true
    unreadCount.value--
  } catch (error) {
    console.error('标记已读失败:', error)
  }
}

/**
 * 标记全部已读
 * @async
 * @returns {Promise<void>}
 */
const markAllRead = async () => {
  try {
    await request.post(API_ENDPOINTS.NOTIFICATION_READ_ALL)
    showToast('已全部标记为已读')
    notifications.value.forEach(n => n.is_read = true)
    unreadCount.value = 0
  } catch (error) {
    console.error('标记全部已读失败:', error)
    showToast('操作失败')
  }
}

/**
 * 下拉刷新
 * @returns {void}
 */
const onRefresh = () => {
  page.value = 1
  finished.value = false
  loadNotifications()
}

onMounted(() => {
  loadNotifications()
})
</script>

<style scoped>
.notifications-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 60px;
}

.notification-item {
  background: #fff;
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
  position: relative;
}

.notification-item.unread {
  background: #f0f9ff;
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.notification-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.unread-badge {
  width: 8px;
  height: 8px;
  background: #f56c6c;
  border-radius: 50%;
}

.notification-content {
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
  margin-bottom: 8px;
}

.notification-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.notification-time {
  font-size: 12px;
  color: #c0c4cc;
}

.empty {
  padding: 100px 0;
}
</style>
