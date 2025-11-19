<template>
  <div class="register-page">
    <van-nav-bar
      title="注册"
      left-arrow
      @click-left="$router.back()"
      fixed
      placeholder
    />

    <!-- 步骤指示 -->
    <div class="step-indicator">
      <span :class="['step-dot', currentStep === 1 ? 'active' : '']">1. 账号信息</span>
      <span class="step-sep">›</span>
      <span :class="['step-dot', currentStep === 2 ? 'active' : '']">2. 实名认证</span>
    </div>

    <!-- 第一步：账号信息 -->
    <van-form v-if="currentStep === 1" @submit="onFirstStepSubmit">
      <van-cell-group inset>
        <van-field
          v-model="baseForm.phone"
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
          v-model="baseForm.password"
          type="password"
          name="password"
          label="登录密码"
          placeholder="请输入登录密码"
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
          v-model="baseForm.invite_code"
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
        >
          下一步：填写实名认证
        </van-button>
      </div>
    </van-form>

    <!-- 第二步：实名认证信息 -->
    <van-form v-else @submit="onSubmit">
      <van-cell-group inset>
        <van-field
          v-model="verifyForm.real_name"
          name="real_name"
          label="真实姓名"
          placeholder="请输入真实姓名"
          :rules="[{ required: true, message: '请输入真实姓名' }]"
        />
        <van-field
          v-model="verifyForm.id_number"
          name="id_number"
          label="身份证号"
          placeholder="请输入身份证号码"
          :rules="[{ required: true, message: '请输入身份证号' }]"
        />
        <van-field
          v-model="verifyForm.receiver_name"
          name="receiver_name"
          label="收货人姓名"
          placeholder="请输入收货人姓名"
        />
        <van-field
          v-model="verifyForm.receiver_phone"
          name="receiver_phone"
          type="tel"
          label="联系电话"
          placeholder="请输入联系电话"
        />
        <van-field name="id_front" label="身份证正面">
          <template #input>
            <van-uploader
              v-model="idFrontFiles"
              :max-count="1"
              :after-read="afterReadIdFront"
            />
          </template>
        </van-field>
        <van-field name="id_back" label="身份证反面">
          <template #input>
            <van-uploader
              v-model="idBackFiles"
              :max-count="1"
              :after-read="afterReadIdBack"
            />
          </template>
        </van-field>
      </van-cell-group>

      <div style="margin: 16px; display: flex; gap: 12px;">
        <van-button
          round
          block
          type="default"
          native-type="button"
          @click="currentStep = 1"
        >
          上一步
        </van-button>
        <van-button
          round
          block
          type="primary"
          native-type="submit"
          :loading="loading"
        >
          提交实名认证并注册
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
import { showDialog, showToast } from 'vant'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

// 当前步骤：1=账号信息，2=实名认证
const currentStep = ref(1)

// 第一步：账号信息
const baseForm = ref({
  phone: '',
  password: '',
  invite_code: ''
})

// 第二步：实名认证信息
const verifyForm = ref({
  real_name: '',
  id_number: '',
  receiver_name: '',
  receiver_phone: '',
  id_front_url: '',
  id_back_url: ''
})

const confirmPassword = ref('')
const loading = ref(false)

// 身份证图片文件
const idFrontFiles = ref([])
const idBackFiles = ref([])

const validatePassword = () => {
  return confirmPassword.value === baseForm.value.password
}

// 第一步提交：仅校验并切换到第二步
const onFirstStepSubmit = () => {
  currentStep.value = 2
}

// 第二步提交：组装请求体并调用注册接口
const onSubmit = async () => {
  try {
    if (!verifyForm.value.id_front_url || !verifyForm.value.id_back_url) {
      showToast('请上传身份证正反面照片')
      return
    }

    loading.value = true
    const payload = {
      phone: baseForm.value.phone,
      password: baseForm.value.password,
      invite_code: baseForm.value.invite_code,
      verification: {
        real_name: verifyForm.value.real_name,
        id_number: verifyForm.value.id_number,
        receiver_name: verifyForm.value.receiver_name,
        receiver_phone: verifyForm.value.receiver_phone,
        id_front_url: verifyForm.value.id_front_url,
        id_back_url: verifyForm.value.id_back_url
      }
    }

    await userStore.register(payload)

    showDialog({
      title: '注册成功',
      message: '您的账号及实名认证信息已提交，请等待客服审核通过后即可登录使用',
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

// 读取身份证正面
const afterReadIdFront = async (file) => {
  try {
    const dataUrl = await readFileAsDataUrl(file.file)
    verifyForm.value.id_front_url = dataUrl
  } catch (error) {
    console.error('读取身份证正面失败:', error)
    showToast('身份证正面读取失败')
  }
}

// 读取身份证反面
const afterReadIdBack = async (file) => {
  try {
    const dataUrl = await readFileAsDataUrl(file.file)
    verifyForm.value.id_back_url = dataUrl
  } catch (error) {
    console.error('读取身份证反面失败:', error)
    showToast('身份证反面读取失败')
  }
}

// 简单封装：将文件读取为 Data URL
const readFileAsDataUrl = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => resolve(e.target.result)
    reader.onerror = (e) => reject(e)
    reader.readAsDataURL(file)
  })
}
</script>

<style scoped>
.register-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 20px;
}

.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px 16px 0;
  font-size: 13px;
  color: #999;
}

.step-dot {
  padding: 4px 8px;
  border-radius: 12px;
  background-color: #f0f0f0;
}

.step-dot.active {
  background-color: #1989fa;
  color: #fff;
}

.step-sep {
  margin: 0 8px;
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
