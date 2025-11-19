<template>
  <div class="admin-sales-page">
    <van-nav-bar
      title="销售员管理"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    />

    <div class="content">
      <van-list
        v-model:loading="loading"
        :finished="finished"
        finished-text="没有更多销售员"
        @load="loadSalespersons"
      >
        <div v-if="salespersons.length === 0 && !loading" class="empty">
          <van-empty description="暂无销售员数据" />
        </div>

        <div
          v-for="sp in salespersons"
          :key="sp.id"
          class="sales-item"
        >
          <div class="sales-header">
            <div class="sales-info">
              <div class="sales-name">{{ sp.name }}</div>
              <div class="sales-phone">{{ sp.phone || '-' }}</div>
            </div>
            <div class="sales-status" :class="{ inactive: !sp.is_active }">
              {{ sp.is_active ? '在职' : '离职' }}
            </div>
          </div>

          <div class="sales-body">
            <div class="row">
              <span class="label">销售编号</span>
              <span class="value">{{ sp.sales_code }}</span>
            </div>
            <div class="row">
              <span class="label">总积分</span>
              <span class="value">{{ sp.total_points }}</span>
            </div>
            <div class="row">
              <span class="label">本月积分</span>
              <span class="value">{{ sp.month_points }}</span>
            </div>
            <div class="row">
              <span class="label">客户数</span>
              <span class="value">
                总 {{ sp.total_customers }} / 活跃 {{ sp.active_customers }}
              </span>
            </div>
          </div>

          <div class="sales-footer">
            <van-button
              size="small"
              type="primary"
              block
              @click="goToCustomers(sp)"
            >
              查看客户
            </van-button>
          </div>
        </div>
      </van-list>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import request from '../../utils/request'
import { API_ENDPOINTS } from '../../config/api'

const router = useRouter()

const loading = ref(false)
const finished = ref(false)
const salespersons = ref([])

const loadSalespersons = async () => {
  if (loading.value || finished.value) return
  try {
    loading.value = true
    const data = await request.get(API_ENDPOINTS.ADMIN_SALESPERSONS)
    const list = data.salespersons || data.list || []
    salespersons.value = list
    finished.value = true
  } catch (error) {
    console.error('加载销售员失败:', error)
    showToast('加载失败')
  } finally {
    loading.value = false
  }
}

const goToCustomers = (sp) => {
  if (!sp || !sp.user_id) return
  // 跳转到用户管理页，并按销售员过滤客户
  router.push({
    path: '/admin/users',
    query: { sales_id: sp.user_id }
  })
}

onMounted(() => {
  loadSalespersons()
})
</script>

<style scoped>
.admin-sales-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}

.content {
  padding: 10px;
}

.sales-item {
  background: #fff;
  margin: 8px 8px 12px;
  padding: 12px 12px 10px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.sales-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
}

.sales-name {
  font-size: 16px;
  font-weight: 500;
}

.sales-phone {
  font-size: 13px;
  color: #909399;
}

.sales-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 12px;
  color: #67c23a;
  background: #f0f9eb;
}

.sales-status.inactive {
  color: #909399;
  background: #f4f4f5;
}

.sales-body {
  margin: 8px 0 6px;
}

.row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  margin-bottom: 4px;
}

.label {
  color: #909399;
}

.value {
  color: #303133;
}

.sales-footer {
  margin-top: 8px;
}

.empty {
  padding-top: 60px;
}
</style>
