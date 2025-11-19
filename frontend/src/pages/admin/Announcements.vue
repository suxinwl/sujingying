<template>
  <div class="admin-announcements-page">
    <van-nav-bar
      title="平台公告"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    />

    <div class="content">
      <!-- 发布公告 -->
      <div class="card publish-card">
        <div class="card-title">发布新公告</div>
        <van-form @submit="onSubmit">
          <van-field
            v-model="form.title"
            label="标题"
            placeholder="请输入公告标题"
            required
          />
          <van-field
            v-model="form.content"
            label="内容"
            type="textarea"
            rows="4"
            autosize
            placeholder="请输入公告内容，将推送给所有用户"
            required
          />
          <div class="submit-row">
            <van-button round block type="primary" native-type="submit">
              发布公告
            </van-button>
          </div>
        </van-form>
      </div>

      <!-- 公告列表 -->
      <div class="card list-card">
        <div class="card-title">公告列表</div>
        <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
          <van-list
            v-model:loading="loading"
            :finished="finished"
            finished-text="没有更多了"
            @load="loadAnnouncements"
          >
            <div v-if="announcements.length === 0" class="empty">
              <van-empty description="暂无公告" />
            </div>

            <div
              v-for="item in announcements"
              :key="item.id || item.ID"
              class="announcement-item"
            >
              <div class="announcement-header">
                <div class="announcement-title">
                  {{ item.title || item.Title }}
                </div>
                <div class="announcement-time">
                  {{ formatDateTime(item.created_at || item.CreatedAt) }}
                </div>
              </div>
              <div class="announcement-content">
                {{ item.content || item.Content }}
              </div>
            </div>
          </van-list>
        </van-pull-refresh>
      </div>
    </div>
  </div>
</template>

<script setup>
// 管理平台公告列表与发布
import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import request from '../../utils/request'
import { API_ENDPOINTS } from '../../config/api'
import { formatDateTime } from '../../utils/helpers'

const form = ref({
  title: '',
  content: ''
})

const announcements = ref([])
const loading = ref(false)
const finished = ref(false)
const refreshing = ref(false)
const page = ref(1)
const pageSize = 20

const loadAnnouncements = async () => {
  try {
    loading.value = true
    const params = {
      limit: pageSize,
      offset: (page.value - 1) * pageSize
    }
    const data = await request.get(API_ENDPOINTS.ADMIN_ANNOUNCEMENTS, { params })
    const list = data.announcements || data.list || []

    if (page.value === 1) {
      announcements.value = list
    } else {
      announcements.value.push(...list)
    }

    if (list.length < pageSize) {
      finished.value = true
    } else {
      page.value++
    }
  } catch (error) {
    console.error('加载公告失败:', error)
    showToast('加载失败')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const onRefresh = () => {
  page.value = 1
  finished.value = false
  loadAnnouncements()
}

const onSubmit = async () => {
  try {
    if (!form.value.title || !form.value.content) {
      showToast('请填写标题和内容')
      return
    }

    await request.post(API_ENDPOINTS.ADMIN_ANNOUNCEMENTS, {
      title: form.value.title,
      content: form.value.content
    })

    showToast('公告已发布')
    form.value = { title: '', content: '' }
    page.value = 1
    finished.value = false
    loadAnnouncements()
  } catch (error) {
    console.error('发布公告失败:', error)
    const msg = error.response?.data?.error || error.message || '发布失败'
    showToast(msg)
  }
}

onMounted(() => {
  loadAnnouncements()
})
</script>

<style scoped>
.admin-announcements-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}

.content {
  padding: 16px;
}

.card {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.card-title {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 12px;
}

.submit-row {
  margin-top: 16px;
}

.announcement-item {
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.announcement-item:last-child {
  border-bottom: none;
}

.announcement-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.announcement-title {
  font-size: 15px;
  font-weight: 500;
  color: #303133;
}

.announcement-time {
  font-size: 12px;
  color: #c0c4cc;
}

.announcement-content {
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
}

.empty {
  padding: 40px 0;
}
</style>
