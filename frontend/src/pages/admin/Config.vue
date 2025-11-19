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
        <van-field
          v-model="config.auto_supplement_target"
          type="number"
          label="目标定金率(%)"
          placeholder="如 100 表示 100%"
        />
        <van-field
          v-model="config.trading_start_time"
          label="交易开始时间"
          placeholder="请选择开始时间"
          readonly
          is-link
          @click="openStartTimePicker"
        />
        <van-field
          v-model="config.trading_end_time"
          label="交易结束时间"
          placeholder="请选择结束时间"
          readonly
          is-link
          @click="openEndTimePicker"
        />
        <van-cell title="交易日">
          <template #label>
            <van-checkbox-group
              v-model="selectedTradingDays"
              direction="horizontal"
              @change="onTradingDaysChange"
            >
              <van-checkbox
                v-for="item in tradingDayOptions"
                :key="item.value"
                :name="item.value"
              >
                {{ item.label }}
              </van-checkbox>
            </van-checkbox-group>
            <div class="trading-days-hint">
              {{ tradingDaysText || '未选择交易日' }}
            </div>
          </template>
        </van-cell>
        <van-cell title="节假日是否交易" label="关闭后，节假日/临时休市将禁止下单">
          <template #right-icon>
            <van-switch
              v-model="config.holiday_trading_enabled"
              :active-value="'1'"
              :inactive-value="'0'"
            />
          </template>
        </van-cell>
        <van-cell
          title="节假日休市日期"
          :label="holidayDatesText || '未设置节假日休市日期'"
          is-link
          @click="showHolidayCalendar = true"
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
          v-model="config.platform_intro"
          label="平台简介"
          type="textarea"
          rows="2"
          autosize
          placeholder="用于关于平台页面的简介文案"
        />
        <van-field
          v-model="config.company_name"
          label="公司名称"
          placeholder="请输入公司名称"
        />
        <van-field
          v-model="config.company_address"
          label="公司地址"
          placeholder="请输入公司地址"
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
      </van-cell-group>

      <!-- 协议链接设置 -->
      <van-cell-group inset style="margin-top: 20px;">
        <van-cell title="协议链接设置" />
        <van-field
          v-model="config.user_agreement_url"
          label="用户协议链接"
          placeholder="请输入用户协议 H5 地址"
        />
        <van-field
          v-model="config.privacy_policy_url"
          label="隐私政策链接"
          placeholder="请输入隐私政策 H5 地址"
        />
        <van-field
          v-model="config.risk_warning_url"
          label="风险提示链接"
          placeholder="请输入风险提示 H5 地址"
        />
        <van-field
          v-model="config.metal_service_agreement_url"
          label="贵金属购销服务协议"
          placeholder="请输入协议 H5 地址"
        />
      </van-cell-group>
      
      <div class="button-group">
        <van-button type="primary" round block @click="saveConfig">
          保存配置
        </van-button>
      </div>
    </div>

    <van-popup v-model:show="showStartTimePicker" position="bottom" round>
      <div class="time-popup">
        <div class="time-toolbar">
          <span class="time-toolbar-btn" @click="showStartTimePicker = false">取消</span>
          <span class="time-toolbar-title">选择交易开始时间</span>
          <span class="time-toolbar-btn primary" @click="confirmStartTime">确认</span>
        </div>
        <div class="time-columns">
          <div class="time-column">
            <div
              v-for="h in hourOptions"
              :key="h"
              class="time-item"
              :class="{ active: h === startHour }"
              @click="startHour = h"
            >
              {{ h }} 时
            </div>
          </div>
          <div class="time-column">
            <div
              v-for="m in minuteOptions"
              :key="m"
              class="time-item"
              :class="{ active: m === startMinute }"
              @click="startMinute = m"
            >
              {{ m }} 分
            </div>
          </div>
        </div>
      </div>
    </van-popup>

    <van-popup v-model:show="showEndTimePicker" position="bottom" round>
      <div class="time-popup">
        <div class="time-toolbar">
          <span class="time-toolbar-btn" @click="showEndTimePicker = false">取消</span>
          <span class="time-toolbar-title">选择交易结束时间</span>
          <span class="time-toolbar-btn primary" @click="confirmEndTime">确认</span>
        </div>
        <div class="time-columns">
          <div class="time-column">
            <div
              v-for="h in hourOptions"
              :key="h"
              class="time-item"
              :class="{ active: h === endHour }"
              @click="endHour = h"
            >
              {{ h }} 时
            </div>
          </div>
          <div class="time-column">
            <div
              v-for="m in minuteOptions"
              :key="m"
              class="time-item"
              :class="{ active: m === endMinute }"
              @click="endMinute = m"
            >
              {{ m }} 分
            </div>
          </div>
        </div>
      </div>
    </van-popup>

    <!-- 节假日休市日期日历 -->
    <van-calendar
      v-model:show="showHolidayCalendar"
      type="multiple"
      color="#1989fa"
      @confirm="onHolidayDatesConfirm"
    />
  </div>
</template>

<script setup>
/**
 * @file Config.vue
 * @description 系统配置管理页面
 * @author AI
 * @date 2025-11-18
 */

import { ref, onMounted, computed } from 'vue'
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
  auto_supplement_target: '',
  holiday_trading_enabled: '1',
  holiday_closed_dates: '',

  // 付/退定金配置
  min_deposit_amount: '',
  min_withdraw_amount: '',
  withdraw_fee_fixed: '',
  withdraw_fee_rate: '',

  // 系统设置
  platform_name: '',
  platform_intro: '',
  company_name: '',
  company_address: '',
  platform_address: '',
  platform_logo: '',
  service_phone: '',
  service_wechat: '',
  trading_start_time: '',
  trading_end_time: '',
  trading_days: '',
  user_agreement_url: '',
  privacy_policy_url: '',
  risk_warning_url: '',
  metal_service_agreement_url: ''
})

const logoFiles = ref([])

const showStartTimePicker = ref(false)
const showEndTimePicker = ref(false)

const hourOptions = [
  '00',
  '01',
  '02',
  '03',
  '04',
  '05',
  '06',
  '07',
  '08',
  '09',
  '10',
  '11',
  '12',
  '13',
  '14',
  '15',
  '16',
  '17',
  '18',
  '19',
  '20',
  '21',
  '22',
  '23'
]

const minuteOptions = ['00', '15', '30', '45']

const startHour = ref('09')
const startMinute = ref('00')
const endHour = ref('18')
const endMinute = ref('00')

const tradingDayOptions = [
  { label: '周一', value: '1' },
  { label: '周二', value: '2' },
  { label: '周三', value: '3' },
  { label: '周四', value: '4' },
  { label: '周五', value: '5' },
  { label: '周六', value: '6' },
  { label: '周日', value: '7' }
]

const selectedTradingDays = ref([])
const tradingDayMap = tradingDayOptions.reduce((acc, item) => {
  acc[item.value] = item.label
  return acc
}, {})

const tradingDaysText = computed(() => {
  if (!selectedTradingDays.value || selectedTradingDays.value.length === 0) {
    return ''
  }
  return selectedTradingDays.value
    .map((v) => tradingDayMap[v] || v)
    .join('、')
})

// 节假日休市日期选择
const showHolidayCalendar = ref(false)
const holidayDates = ref([])

const holidayDatesText = computed(() => {
  if (!holidayDates.value || holidayDates.value.length === 0) return ''
  return holidayDates.value.join('、')
})

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

const openStartTimePicker = () => {
  const raw = config.value.trading_start_time || '09:00'
  const [h, m] = raw.split(':')
  if (h) startHour.value = h.padStart(2, '0')
  if (m) startMinute.value = m.padStart(2, '0')
  showStartTimePicker.value = true
}

const openEndTimePicker = () => {
  const raw = config.value.trading_end_time || '18:00'
  const [h, m] = raw.split(':')
  if (h) endHour.value = h.padStart(2, '0')
  if (m) endMinute.value = m.padStart(2, '0')
  showEndTimePicker.value = true
}

const confirmStartTime = () => {
  config.value.trading_start_time = `${startHour.value}:${startMinute.value}`
  showStartTimePicker.value = false
}

const confirmEndTime = () => {
  config.value.trading_end_time = `${endHour.value}:${endMinute.value}`
  showEndTimePicker.value = false
}

const syncTradingDaysFromConfig = () => {
  const raw = config.value.trading_days
  if (!raw) {
    selectedTradingDays.value = []
    return
  }
  selectedTradingDays.value = String(raw)
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s)
}

const onTradingDaysChange = () => {
  config.value.trading_days = selectedTradingDays.value.join(',')
}

const syncHolidayDatesFromConfig = () => {
  const raw = config.value.holiday_closed_dates
  if (!raw) {
    holidayDates.value = []
    return
  }
  holidayDates.value = String(raw)
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s)
}

const onHolidayDatesConfirm = (values) => {
  const arr = Array.isArray(values) ? values : [values]
  const formatted = arr.map((val) => {
    const d = new Date(val)
    if (Number.isNaN(d.getTime())) return ''
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${y}-${m}-${day}`
  }).filter(Boolean)

  holidayDates.value = formatted
  config.value.holiday_closed_dates = formatted.join(',')
  showHolidayCalendar.value = false
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

      syncTradingDaysFromConfig()

      // 节假日休市日期
      syncHolidayDatesFromConfig()

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

.trading-days-hint {
  margin-top: 8px;
  font-size: 12px;
  color: #999;
}

.time-popup {
  background-color: #fff;
}

.time-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  font-size: 14px;
  border-bottom: 1px solid #f5f5f5;
}

.time-toolbar-title {
  font-size: 15px;
  font-weight: 500;
}

.time-toolbar-btn {
  color: #969799;
}

.time-toolbar-btn.primary {
  color: #1989fa;
}

.time-columns {
  display: flex;
  height: 200px;
}

.time-column {
  flex: 1;
  overflow-y: auto;
}

.time-item {
  text-align: center;
  padding: 10px 0;
  font-size: 16px;
  color: #323233;
}

.time-item.active {
  color: #1989fa;
  font-weight: 500;
}
</style>
