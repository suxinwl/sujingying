<template>
  <div class="login-page">
    <div class="logo">
      <h1>速金盈</h1>
      <p>专业的黄金交易平台</p>
    </div>
    
    <van-form @submit="onSubmit">
      <van-cell-group inset>
        <van-field
          v-model="form.username"
          name="username"
          label="用户名"
          placeholder="请输入用户名"
          :rules="[{ required: true, message: '请输入用户名' }]"
        />
        <van-field
          v-model="form.password"
          type="password"
          name="password"
          label="密码"
          placeholder="请输入密码"
          :rules="[{ required: true, message: '请输入密码' }]"
        />
      </van-cell-group>
      
      <div style="margin: 16px;">
        <van-button
          round
          block
          type="primary"
          native-type="submit"
          :loading="loading"
        >
          登录
        </van-button>
      </div>
    </van-form>
    
    <div class="footer">
      <van-button
        type="default"
        size="small"
        @click="$router.push('/register')"
      >
        没有账号？立即注册
      </van-button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const form = ref({
  username: '',
  password: ''
})

const loading = ref(false)

const onSubmit = async () => {
  try {
    loading.value = true
    console.log('登录数据:', form.value)
    await userStore.login(form.value)
    showToast('登录成功')
    router.replace('/')
  } catch (error) {
    console.error('登录失败:', error)
    // 显示具体错误信息
    if (error.response?.data?.error) {
      showToast(error.response.data.error)
    } else {
      showToast(error.message || '登录失败，请重试')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 60px 20px 20px;
}

.logo {
  text-align: center;
  margin-bottom: 60px;
  color: #fff;
}

.logo h1 {
  font-size: 36px;
  margin: 0 0 10px;
}

.logo p {
  font-size: 14px;
  opacity: 0.9;
}

.footer {
  text-align: center;
  margin-top: 20px;
}
</style>
