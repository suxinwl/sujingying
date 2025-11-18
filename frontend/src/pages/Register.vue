<template>
  <div class="register-page">
    <van-nav-bar
      title="注册"
      left-arrow
      @click-left="$router.back()"
      fixed
      placeholder
    />
    
    <van-form @submit="onSubmit">
      <van-cell-group inset>
        <van-field
          v-model="form.username"
          name="username"
          label="用户名"
          placeholder="请输入用户名"
          :rules="[
            { required: true, message: '请输入用户名' },
            { pattern: /^[a-zA-Z0-9_]{4,20}$/, message: '4-20位字母数字下划线' }
          ]"
        />
        <van-field
          v-model="form.password"
          type="password"
          name="password"
          label="密码"
          placeholder="请输入密码"
          :rules="[
            { required: true, message: '请输入密码' },
            { min: 6, message: '密码至少6位' }
          ]"
        />
        <van-field
          v-model="confirmPassword"
          type="password"
          name="confirmPassword"
          label="确认密码"
          placeholder="请再次输入密码"
          :rules="[
            { required: true, message: '请确认密码' },
            { validator: validatePassword, message: '两次密码不一致' }
          ]"
        />
        <van-field
          v-model="form.real_name"
          name="real_name"
          label="真实姓名"
          placeholder="请输入真实姓名"
          :rules="[{ required: true, message: '请输入真实姓名' }]"
        />
        <van-field
          v-model="form.phone"
          type="tel"
          name="phone"
          label="手机号"
          placeholder="请输入手机号"
          :rules="[
            { required: true, message: '请输入手机号' },
            { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确' }
          ]"
        />
        <van-field
          v-model="form.invite_code"
          name="invite_code"
          label="邀请码"
          placeholder="请输入邀请码"
          :rules="[{ required: true, message: '请输入邀请码' }]"
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
          注册
        </van-button>
      </div>
    </van-form>
    
    <div class="tips">
      <p>· 注册即表示同意《用户协议》和《隐私政策》</p>
      <p>· 注册后需要客服审核通过才能使用</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showDialog } from 'vant'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const form = ref({
  username: '',
  password: '',
  real_name: '',
  phone: '',
  invite_code: ''
})

const confirmPassword = ref('')
const loading = ref(false)

const validatePassword = () => {
  return confirmPassword.value === form.value.password
}

const onSubmit = async () => {
  try {
    loading.value = true
    await userStore.register(form.value)
    
    showDialog({
      title: '注册成功',
      message: '您的账号已提交审核，请等待客服审核通过后即可登录使用',
      confirmButtonText: '返回登录'
    }).then(() => {
      router.replace('/login')
    })
  } catch (error) {
    console.error('注册失败:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 20px;
}

.tips {
  padding: 20px;
  color: #999;
  font-size: 12px;
  text-align: center;
}

.tips p {
  margin: 5px 0;
}
</style>
