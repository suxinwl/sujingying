<template>
  <div class="admin-config-page">
    <van-nav-bar
      title="系统配置"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    />
    
    <div class="config-container">
      <!-- 交易配置 -->
      <van-cell-group inset>
        <van-cell title="交易配置" />
        <van-field
          v-model="config.min_order_amount"
          type="number"
          label="最小交易克重(克)"
          placeholder="为空则不限制"
        />
        <van-field
          v-model="config.max_order_amount"
          type="number"
          label="最大交易克重(克)"
          placeholder="为空则不限制，上不封顶"
        />
        <van-field
          v-model="config.delivery_fee_per_gram"
          type="number"
          label="交割手续费(元/克)"
          placeholder="请输入交割手续费"
        />
      </van-cell-group>
      
      <!-- 付/退定金配置 -->
      <van-cell-group inset style="margin-top: 20px;">
        <van-cell title="付/退定金配置" />
        <van-field
          v-model="config.min_deposit_amount"
          type="number"
          label="最小付定金额度(元)"
          placeholder="为空则不限制"
        />
        <van-field
          v-model="config.min_withdraw_amount"
          type="number"
          label="最小退定金金额(元)"
          placeholder="为空则不限制"
        />
        <van-cell title="手续费设置" label="笔/元 与 交易金额百分比，二选一填写" />
        <van-field
          v-model="config.withdraw_fee_fixed"
          type="number"
          label="单笔固定(元)"
          placeholder="如 5 表示每笔5元"
        />
        <van-field
          v-model="config.withdraw_fee_rate"
          type="number"
          label="按金额比例(%)"
          placeholder="如 0.5 表示0.5%"
        />
      </van-cell-group>
      
      <!-- 系统设置 -->
      <van-cell-group inset style="margin-top: 20px;">
        <van-cell title="系统设置" />
        <van-field
          v-model="config.platform_name"
          label="平台名称"
          placeholder="请输入平台名称"
        />
        <van-field
          v-model="config.platform_address"
          label="平台地址"
          placeholder="请输入平台地址"
        />
        <van-field name="platform_logo" label="平台Logo">
          <template #input>
            <van-uploader
              :max-count="1"
              :after-read="afterReadLogo"
              v-model="logoFiles"
            />
          </template>
        </van-field>
        <div v-if="config.platform_logo" class="logo-preview">
          <div class="logo-tip">当前Logo预览：</div>
          <van-image :src="config.platform_logo" width="120" height="120" fit="contain" />
        </div>
        <van-field
          v-model="config.service_phone"
          label="客服电话"
          placeholder="示例：18038018206"
        />
        <van-field
          v-model="config.service_wechat"
          label="客服微信"
          placeholder="请输入客服微信号"
        />
        <van-field
          v-model="config.trading_start_time"
          label="交易开始时间"
          placeholder="如 09:00"
        />
        <van-field
          v-model="config.trading_end_time"
          label="交易结束时间"
          placeholder="如 01:00"
        />
        <van-field
          v-model="config.trading_days"
          label="交易日"
          placeholder="如 1,2,3,4,5 表示周一到周五"
        />
      </van-cell-group>
      
      <div class="button-group">
        <van-button type="primary" round block @click="saveConfig">
          保存配置
        </van-button>
      </div>
    </div>
  </div>
</template>

<script setup>
/**
 * @file Config.vue
 * @description 系统配置管理页面
 * @author AI
 * @date 2025-11-18
 */

import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import request from '../../utils/request'
import { API_ENDPOINTS } from '../../config/api'

/**
 * 配置数据
 * @type {import('vue').Ref<Object>}
 */
const config = ref({
  // 交易配置
  min_order_amount: '',
  max_order_amount: '',
  delivery_fee_per_gram: '',

  // 付/退定金配置
  min_deposit_amount: '',
  min_withdraw_amount: '',
  withdraw_fee_fixed: '',
  withdraw_fee_rate: '',

  // 系统设置
  platform_name: '',
  platform_address: '',
  platform_logo: '',
  service_phone: '',
  service_wechat: '',
  trading_start_time: '',
  trading_end_time: '',
  trading_days: ''
})

const logoFiles = ref([])

// 处理平台Logo上传
const compressImage = (file, maxWidth = 600, quality = 0.7) => {
  return new Promise((resolve) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const img = new Image()
      img.onload = () => {
        const canvas = document.createElement('canvas')
        let width = img.width
        let height = img.height

        if (width > maxWidth) {
          height = (height * maxWidth) / width
          width = maxWidth
        }

        canvas.width = width
        canvas.height = height

        const ctx = canvas.getContext('2d')
        ctx.drawImage(img, 0, 0, width, height)

        let compressedBase64 = canvas.toDataURL('image/jpeg', quality)
        if (compressedBase64.length > 800000) {
          compressedBase64 = canvas.toDataURL('image/jpeg', 0.5)
        }
        resolve(compressedBase64)
      }
      img.src = e.target.result
    }
    reader.readAsDataURL(file)
  })
}

const afterReadLogo = async (file) => {
  try {
    const base64 = await compressImage(file.file)
    config.value.platform_logo = base64
    logoFiles.value = [{ url: base64, isImage: true }]
    showToast('Logo上传成功')
  } catch (error) {
    console.error('Logo处理失败:', error)
    showToast('Logo上传失败')
  }
}

/**
 * 加载配置
 * @async
 * @returns {Promise<void>}
 */
const loadConfig = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.CONFIG)
    if (data.configs && data.configs.length > 0) {
      // 将配置数组转换为对象（兼容后端返回 Key/Value 或 key/value）
      data.configs.forEach(item => {
        const key = item.key || item.Key
        const value = item.value || item.Value
        if (key && Object.prototype.hasOwnProperty.call(config.value, key)) {
          config.value[key] = value
        }
      })

      // 平台Logo回显（若存在）
      if (config.value.platform_logo) {
        logoFiles.value = [{ url: config.value.platform_logo, isImage: true }]
      }
    }
  } catch (error) {
    console.error('加载配置失败:', error)
    // 加载失败不阻塞页面显示
  }
}

/**
 * 保存配置
 * @async
 * @returns {Promise<void>}
 */
const saveConfig = async () => {
  try {
    // 过滤掉空值
    const configData = {}
    Object.keys(config.value).forEach(key => {
      if (config.value[key] !== '') {
        configData[key] = config.value[key]
      }
    })
    
    await request.post(API_ENDPOINTS.CONFIG + '/batch', configData)
    showToast('保存成功')
  } catch (error) {
    console.error('保存配置失败:', error)
    showToast('保存失败')
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.admin-config-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 80px;
}

.config-container {
  padding: 20px;
}

.button-group {
  margin-top: 30px;
}
</style>
