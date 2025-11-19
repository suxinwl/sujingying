<template>
  <div class="profile-page">
    <van-nav-bar
      title="资料设置"
      left-arrow
      @click-left="$router.back()"
    />

    <van-form @submit="onSubmit">
      <van-field
        v-model="form.real_name"
        label="姓名"
        placeholder="实名认证后不可修改"
        :readonly="!!userStore.userInfo?.real_name"
      />
      <van-field
        v-model="form.phone"
        label="手机号"
        placeholder="请输入手机号"
        readonly
      />
      <van-field
        v-model="form.email"
        label="邮箱"
        placeholder="请输入邮箱"
      />
      <!-- 头像暂时占位，不做上传逻辑 -->
      <van-field
        label="头像"
        readonly
        value="暂不支持修改头像"
      />
      <div style="margin: 16px;">
        <van-button round block type="primary" native-type="submit">
          保存
        </van-button>
      </div>
    </van-form>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import { useUserStore } from '../stores/user'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'

const userStore = useUserStore()

const form = ref({
  real_name: '',
  phone: '',
  email: ''
})

onMounted(async () => {
  if (!userStore.userInfo) {
    await userStore.getUserInfo()
  }
  const info = userStore.userInfo || {}
  form.value.real_name = info.real_name || ''
  form.value.phone = info.phone || ''
  form.value.email = info.email || ''
})

const onSubmit = async () => {
  try {
    await request.put(API_ENDPOINTS.USER_UPDATE, {
      email: form.value.email
    })
    await userStore.getUserInfo()
    showToast('资料已更新')
  } catch (error) {
    console.error('更新资料失败:', error)
    showToast('更新失败')
  }
}
</script>

<style scoped>
.profile-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}
</style>
