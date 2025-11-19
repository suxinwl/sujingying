<template>
  <div class="verification-page">
    <van-nav-bar
      title="实名认证"
      left-arrow
      @click-left="$router.back()"
      fixed
      placeholder
    />

    <div class="status-bar" v-if="statusText">
      当前状态：<span class="status-text">{{ statusText }}</span>
    </div>

    <van-form @submit="onSubmit">
      <van-cell-group inset>
        <van-field
          v-model="form.real_name"
          name="real_name"
          label="真实姓名"
          placeholder="请输入真实姓名"
          :readonly="isApproved"
          :rules="[{ required: true, message: '请输入真实姓名' }]"
        />
        <van-field
          v-model="form.id_number"
          name="id_number"
          label="身份证号"
          placeholder="请输入身份证号码"
          :readonly="isApproved"
          :rules="[{ required: true, message: '请输入身份证号' }]"
        />
        <van-field
          v-model="form.receiver_name"
          name="receiver_name"
          label="收货人姓名"
          placeholder="请输入收货人姓名"
          :readonly="isApproved"
        />
        <van-field
          v-model="form.receiver_phone"
          name="receiver_phone"
          label="联系电话"
          type="tel"
          placeholder="请输入联系电话"
          :readonly="isApproved"
        />

        <van-field name="id_front" label="身份证正面">
          <template #input>
            <van-uploader
              v-model="idFrontFiles"
              :max-count="1"
              :after-read="afterReadIdFront"
              :disabled="isApproved"
            />
          </template>
        </van-field>
        <van-field name="id_back" label="身份证反面">
          <template #input>
            <van-uploader
              v-model="idBackFiles"
              :max-count="1"
              :after-read="afterReadIdBack"
              :disabled="isApproved"
            />
          </template>
        </van-field>
      </van-cell-group>

      <div style="margin: 16px;">
        <van-button
          v-if="!isApproved"
          round
          block
          type="primary"
          native-type="submit"
          :loading="submitting"
        >
          提交实名认证
        </van-button>
        <van-button
          v-else
          round
          block
          type="primary"
          native-type="button"
          disabled
        >
          已通过实名认证
        </van-button>
      </div>
    </van-form>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { showToast } from 'vant'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'

const form = ref({
  real_name: '',
  id_number: '',
  receiver_name: '',
  receiver_phone: '',
  id_front_url: '',
  id_back_url: ''
})

const status = ref('')
const submitting = ref(false)

const idFrontFiles = ref([])
const idBackFiles = ref([])

const statusText = computed(() => {
  if (!status.value) return ''
  const map = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已驳回'
  }
  return map[status.value] || status.value
})

const isApproved = computed(() => status.value === 'approved')

onMounted(async () => {
  try {
    const data = await request.get(API_ENDPOINTS.USER_VERIFICATION)
    const v = data.verification || null
    if (v) {
      form.value.real_name = v.real_name || ''
      form.value.id_number = v.id_number || ''
      form.value.receiver_name = v.receiver_name || ''
      form.value.receiver_phone = v.receiver_phone || ''
      form.value.id_front_url = v.id_front_url || ''
      form.value.id_back_url = v.id_back_url || ''
      status.value = v.status || ''

      if (form.value.id_front_url) {
        idFrontFiles.value = [{ url: form.value.id_front_url }]
      }
      if (form.value.id_back_url) {
        idBackFiles.value = [{ url: form.value.id_back_url }]
      }
    }
  } catch (error) {
    console.error('加载实名认证信息失败:', error)
  }
})

const onSubmit = async () => {
  try {
    if (!form.value.id_front_url || !form.value.id_back_url) {
      showToast('请上传身份证正反面照片')
      return
    }

    submitting.value = true
    await request.post(API_ENDPOINTS.USER_VERIFICATION, {
      real_name: form.value.real_name,
      id_number: form.value.id_number,
      receiver_name: form.value.receiver_name,
      receiver_phone: form.value.receiver_phone,
      id_front_url: form.value.id_front_url,
      id_back_url: form.value.id_back_url
    })

    showToast('实名认证信息已提交，等待审核')
    status.value = 'pending'
  } catch (error) {
    console.error('提交实名认证失败:', error)
    showToast('提交失败')
  } finally {
    submitting.value = false
  }
}

const afterReadIdFront = async (file) => {
  try {
    const dataUrl = await readFileAsDataUrl(file.file)
    form.value.id_front_url = dataUrl
    idFrontFiles.value = [{ url: dataUrl }]
  } catch (error) {
    console.error('读取身份证正面失败:', error)
    showToast('身份证正面读取失败')
  }
}

const afterReadIdBack = async (file) => {
  try {
    const dataUrl = await readFileAsDataUrl(file.file)
    form.value.id_back_url = dataUrl
    idBackFiles.value = [{ url: dataUrl }]
  } catch (error) {
    console.error('读取身份证反面失败:', error)
    showToast('身份证反面读取失败')
  }
}

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
.verification-page {
  min-height: 100vh;
  background-color: #f7f8fa;
}

.status-bar {
  padding: 12px 16px;
  font-size: 14px;
  color: #666;
}

.status-text {
  font-weight: 500;
}
</style>
