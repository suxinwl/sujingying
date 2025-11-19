<template>
  <div class="payment-settings-page">
    <van-nav-bar
      title="收款管理"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    />
    
    <!-- 收款方式列表 -->
    <van-cell-group title="收款方式配置" inset style="margin-top: 16px;">
      <!-- 银行卡收款 -->
      <van-cell
        title="银行卡收款"
        is-link
        @click="openBankCardForm"
      >
        <template #label>
          <div v-if="paymentSettings.bank_cards && paymentSettings.bank_cards.length">
            <div>
              {{ paymentSettings.bank_cards[0].bank_name }} {{ paymentSettings.bank_cards[0].account_number }}
            </div>
            <div
              v-if="paymentSettings.bank_cards.length > 1"
              style="color: #999; font-size: 12px;"
            >
              共 {{ paymentSettings.bank_cards.length }} 张银行卡
            </div>
          </div>
          <div v-else style="color: #999;">未设置</div>
        </template>
      </van-cell>
      
      <!-- 微信收款 -->
      <van-cell
        title="微信收款码"
        is-link
        @click="openWechatPopup"
      >
        <template #label>
          <div v-if="paymentSettings.wechat_qrs && paymentSettings.wechat_qrs.length">
            <van-image
              width="60"
              height="60"
              :src="paymentSettings.wechat_qrs[0].qr_url"
              fit="cover"
            />
            <div
              v-if="paymentSettings.wechat_qrs.length > 1"
              style="color: #999; font-size: 12px; margin-top: 4px;"
            >
              共 {{ paymentSettings.wechat_qrs.length }} 个收款码
            </div>
          </div>
          <div v-else style="color: #999;">未设置</div>
        </template>
      </van-cell>
      
      <!-- 支付宝收款 -->
      <van-cell
        title="支付宝收款码"
        is-link
        @click="openAlipayPopup"
      >
        <template #label>
          <div v-if="paymentSettings.alipay_qrs && paymentSettings.alipay_qrs.length">
            <van-image
              width="60"
              height="60"
              :src="paymentSettings.alipay_qrs[0].qr_url"
              fit="cover"
            />
            <div
              v-if="paymentSettings.alipay_qrs.length > 1"
              style="color: #999; font-size: 12px; margin-top: 4px;"
            >
              共 {{ paymentSettings.alipay_qrs.length }} 个收款码
            </div>
          </div>
          <div v-else style="color: #999;">未设置</div>
        </template>
      </van-cell>
    </van-cell-group>
    
    <!-- 银行卡设置弹窗 -->
    <van-popup v-model:show="showBankCardForm" position="bottom" round>
      <div class="popup-content">
        <div class="popup-header">
          <h3>银行卡收款设置</h3>
        </div>
        <van-form @submit="onSaveBankCard">
          <van-field
            v-model="bankCardForm.bank_name"
            label="银行名称"
            placeholder="请选择银行"
            is-link
            readonly
            :rules="[{ required: true, message: '请选择银行名称' }]"
            @click="showBankPicker = true"
          />
          <van-field
            v-model="bankCardForm.account_number"
            type="digit"
            label="银行卡号"
            placeholder="请输入银行卡号"
            :rules="[{ required: true, message: '请输入银行卡号' }]"
          />
          <van-field
            v-model="bankCardForm.account_name"
            label="持卡人"
            placeholder="请输入持卡人姓名"
            :rules="[{ required: true, message: '请输入持卡人姓名' }]"
          />
          <van-field
            v-model="bankCardForm.branch_name"
            label="开户行"
            placeholder="如：北京分行"
          />
          <div style="margin: 16px;">
            <van-button round block type="primary" native-type="submit">
              {{ editingBankIndex === -1 ? '添加银行卡' : '保存修改' }}
            </van-button>
          </div>
        </van-form>
        <div
          v-if="paymentSettings.bank_cards && paymentSettings.bank_cards.length"
          class="bank-card-list"
        >
          <div class="section-title">已配置银行卡</div>
          <div
            v-for="(card, index) in paymentSettings.bank_cards"
            :key="index"
            class="bank-card-item"
          >
            <div class="bank-card-info">
              <div>{{ card.bank_name }} {{ card.account_number }}</div>
              <div class="bank-card-name">{{ card.account_name }}</div>
            </div>
            <div class="bank-card-actions">
              <van-button size="small" type="primary" plain @click="editBankCard(index)">
                编辑
              </van-button>
              <van-button size="small" type="danger" plain @click="removeBankCard(index)">
                删除
              </van-button>
            </div>
          </div>
        </div>
      </div>
    </van-popup>

    <van-action-sheet v-model:show="showBankPicker" title="选择银行">
      <div class="bank-list">
        <div
          v-for="bank in bankOptions"
          :key="bank"
          class="bank-item"
          @click="selectBank(bank)"
        >
          {{ bank }}
        </div>
      </div>
    </van-action-sheet>

    <!-- 微信收款码设置弹窗 -->
    <van-popup v-model:show="showWechatForm" position="bottom" round>
      <div class="popup-content">
        <div class="popup-header">
          <h3>微信收款码设置</h3>
        </div>
        <van-form @submit="onSaveWechat">
          <van-field
            v-model="wechatForm.name"
            label="备注名称"
            placeholder="如：微信收款1"
          />
          <van-field name="uploader" label="收款码">
            <template #input>
              <van-uploader
                v-model="wechatFiles"
                :max-count="1"
                :after-read="afterReadWechat"
              />
            </template>
          </van-field>
          <div v-if="wechatForm.qr_url" style="margin: 16px;">
            <div style="color: #999; font-size: 12px; margin-bottom: 8px;">当前收款码预览：</div>
            <van-image
              width="200"
              height="200"
              :src="wechatForm.qr_url"
              fit="contain"
            />
          </div>
          <div style="margin: 16px;">
            <van-button round block type="primary" native-type="submit">
              {{ editingWechatIndex === -1 ? '添加收款码' : '保存修改' }}
            </van-button>
          </div>
        </van-form>
        <div
          v-if="paymentSettings.wechat_qrs && paymentSettings.wechat_qrs.length"
          class="qr-list"
        >
          <div class="section-title">已配置收款码</div>
          <div
            v-for="(item, index) in paymentSettings.wechat_qrs"
            :key="index"
            class="qr-item"
          >
            <div class="qr-info">
              <div class="qr-name">{{ item.name || '微信收款码' }}</div>
              <van-image width="60" height="60" :src="item.qr_url" fit="cover" />
            </div>
            <div class="qr-actions">
              <van-button size="small" type="primary" plain @click="editWechat(index)">
                编辑
              </van-button>
              <van-button size="small" type="danger" plain @click="removeWechat(index)">
                删除
              </van-button>
            </div>
          </div>
        </div>
      </div>
    </van-popup>
    
    <!-- 支付宝收款码设置弹窗 -->
    <van-popup v-model:show="showAlipayForm" position="bottom" round>
      <div class="popup-content">
        <div class="popup-header">
          <h3>支付宝收款码设置</h3>
        </div>
        <van-form @submit="onSaveAlipay">
          <van-field
            v-model="alipayForm.name"
            label="备注名称"
            placeholder="如：支付宝收款1"
          />
          <van-field name="uploader" label="收款码">
            <template #input>
              <van-uploader
                v-model="alipayFiles"
                :max-count="1"
                :after-read="afterReadAlipay"
              />
            </template>
          </van-field>
          <div v-if="alipayForm.qr_url" style="margin: 16px;">
            <div style="color: #999; font-size: 12px; margin-bottom: 8px;">当前收款码预览：</div>
            <van-image
              width="200"
              height="200"
              :src="alipayForm.qr_url"
              fit="contain"
            />
          </div>
          <div style="margin: 16px;">
            <van-button round block type="primary" native-type="submit">
              {{ editingAlipayIndex === -1 ? '添加收款码' : '保存修改' }}
            </van-button>
          </div>
        </van-form>
        <div
          v-if="paymentSettings.alipay_qrs && paymentSettings.alipay_qrs.length"
          class="qr-list"
        >
          <div class="section-title">已配置收款码</div>
          <div
            v-for="(item, index) in paymentSettings.alipay_qrs"
            :key="index"
            class="qr-item"
          >
            <div class="qr-info">
              <div class="qr-name">{{ item.name || '支付宝收款码' }}</div>
              <van-image width="60" height="60" :src="item.qr_url" fit="cover" />
            </div>
            <div class="qr-actions">
              <van-button size="small" type="primary" plain @click="editAlipay(index)">
                编辑
              </van-button>
              <van-button size="small" type="danger" plain @click="removeAlipay(index)">
                删除
              </van-button>
            </div>
          </div>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import request from '../../utils/request'
import { API_ENDPOINTS } from '../../config/api'

const paymentSettings = ref({
  bank_cards: [],
  wechat_qrs: [],
  alipay_qrs: []
})

const showBankCardForm = ref(false)
const showWechatForm = ref(false)
const showAlipayForm = ref(false)

const bankCardForm = ref({
  bank_name: '',
  account_number: '',
  account_name: '',
  branch_name: ''
})

const wechatForm = ref({
  name: '',
  qr_url: ''
})

const alipayForm = ref({
  name: '',
  qr_url: ''
})

const wechatFiles = ref([])
const alipayFiles = ref([])

const editingBankIndex = ref(-1)
const editingWechatIndex = ref(-1)
const editingAlipayIndex = ref(-1)

const showBankPicker = ref(false)
const bankOptions = [
  '中国工商银行',
  '中国农业银行',
  '中国银行',
  '中国建设银行',
  '交通银行',
  '招商银行',
  '中国邮政储蓄银行',
  '中信银行',
  '光大银行',
  '华夏银行',
  '民生银行',
  '兴业银行',
  '广发银行',
  '平安银行'
]

const resetBankCardForm = () => {
  bankCardForm.value = {
    bank_name: '',
    account_number: '',
    account_name: '',
    branch_name: ''
  }
  editingBankIndex.value = -1
}

const resetWechatForm = () => {
  wechatForm.value = {
    name: '',
    qr_url: ''
  }
  editingWechatIndex.value = -1
  wechatFiles.value = []
}

const resetAlipayForm = () => {
  alipayForm.value = {
    name: '',
    qr_url: ''
  }
  editingAlipayIndex.value = -1
  alipayFiles.value = []
}

const selectBank = (bank) => {
  bankCardForm.value.bank_name = bank || ''
  showBankPicker.value = false
}

const loadPaymentSettings = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.CONFIG)
    const next = {
      bank_cards: [],
      wechat_qrs: [],
      alipay_qrs: []
    }
    if (data && data.configs && data.configs.length > 0) {
      data.configs.forEach(item => {
        const key = item.key || item.Key
        const value = item.value || item.Value
        if (!key || !value) {
          return
        }
        try {
          if (key === 'payment_bank_cards') {
            next.bank_cards = JSON.parse(value) || []
          } else if (key === 'payment_wechat_qrs') {
            next.wechat_qrs = JSON.parse(value) || []
          } else if (key === 'payment_alipay_qrs') {
            next.alipay_qrs = JSON.parse(value) || []
          }
        } catch (_) {
        }
      })
    }
    paymentSettings.value = next
  } catch (error) {
    console.error('加载收款设置失败:', error)
  }
}

const savePaymentSettings = async () => {
  const payload = {
    payment_bank_cards: JSON.stringify(paymentSettings.value.bank_cards || []),
    payment_wechat_qrs: JSON.stringify(paymentSettings.value.wechat_qrs || []),
    payment_alipay_qrs: JSON.stringify(paymentSettings.value.alipay_qrs || [])
  }
  await request.post(API_ENDPOINTS.CONFIG + '/batch', payload)
}

const openBankCardForm = () => {
  resetBankCardForm()
  showBankCardForm.value = true
}

const openWechatPopup = () => {
  resetWechatForm()
  showWechatForm.value = true
}

const openAlipayPopup = () => {
  resetAlipayForm()
  showAlipayForm.value = true
}

const editBankCard = (index) => {
  const cards = paymentSettings.value.bank_cards || []
  const card = cards[index]
  if (!card) {
    return
  }
  bankCardForm.value = { ...card }
  editingBankIndex.value = index
  showBankCardForm.value = true
}

const removeBankCard = async (index) => {
  const cards = paymentSettings.value.bank_cards || []
  if (index < 0 || index >= cards.length) {
    return
  }
  cards.splice(index, 1)
  paymentSettings.value.bank_cards = [...cards]
  try {
    await savePaymentSettings()
    showToast('已删除银行卡')
  } catch (error) {
    console.error('删除银行卡失败:', error)
    showToast('删除失败')
  }
}

const isValidBankCardNumber = (num) => {
  if (!num) {
    return false
  }
  const digits = String(num).replace(/\s+/g, '')
  if (!/^\d{12,19}$/.test(digits)) {
    return false
  }
  let sum = 0
  let shouldDouble = false
  for (let i = digits.length - 1; i >= 0; i -= 1) {
    let d = parseInt(digits[i], 10)
    if (Number.isNaN(d)) {
      return false
    }
    if (shouldDouble) {
      d *= 2
      if (d > 9) {
        d -= 9
      }
    }
    sum += d
    shouldDouble = !shouldDouble
  }
  return sum % 10 === 0
}

const editWechat = (index) => {
  const list = paymentSettings.value.wechat_qrs || []
  const item = list[index]
  if (!item) {
    return
  }
  wechatForm.value = {
    name: item.name || '',
    qr_url: item.qr_url || ''
  }
  editingWechatIndex.value = index
  wechatFiles.value = item.qr_url ? [{ url: item.qr_url, isImage: true }] : []
  showWechatForm.value = true
}

const removeWechat = async (index) => {
  const list = paymentSettings.value.wechat_qrs || []
  if (index < 0 || index >= list.length) {
    return
  }
  list.splice(index, 1)
  paymentSettings.value.wechat_qrs = [...list]
  try {
    await savePaymentSettings()
    showToast('已删除收款码')
  } catch (error) {
    console.error('删除收款码失败:', error)
    showToast('删除失败')
  }
}

const editAlipay = (index) => {
  const list = paymentSettings.value.alipay_qrs || []
  const item = list[index]
  if (!item) {
    return
  }
  alipayForm.value = {
    name: item.name || '',
    qr_url: item.qr_url || ''
  }
  editingAlipayIndex.value = index
  alipayFiles.value = item.qr_url ? [{ url: item.qr_url, isImage: true }] : []
  showAlipayForm.value = true
}

const removeAlipay = async (index) => {
  const list = paymentSettings.value.alipay_qrs || []
  if (index < 0 || index >= list.length) {
    return
  }
  list.splice(index, 1)
  paymentSettings.value.alipay_qrs = [...list]
  try {
    await savePaymentSettings()
    showToast('已删除收款码')
  } catch (error) {
    console.error('删除收款码失败:', error)
    showToast('删除失败')
  }
}

const onSaveBankCard = async () => {
  try {
    const card = { ...bankCardForm.value }
    if (!card.bank_name || !card.account_number || !card.account_name) {
      showToast('请完整填写银行卡信息')
      return
    }
    const normalizedNumber = String(card.account_number).replace(/\s+/g, '')
    if (!isValidBankCardNumber(normalizedNumber)) {
      showToast('请输入有效的银行卡号')
      return
    }
    card.account_number = normalizedNumber
    const cards = paymentSettings.value.bank_cards || []
    if (editingBankIndex.value >= 0 && editingBankIndex.value < cards.length) {
      cards.splice(editingBankIndex.value, 1, card)
    } else {
      cards.push(card)
    }
    paymentSettings.value.bank_cards = [...cards]
    await savePaymentSettings()
    showToast('保存成功')
    showBankCardForm.value = false
  } catch (error) {
    console.error('保存失败:', error)
    showToast('保存失败')
  }
}

const onSaveWechat = async () => {
  try {
    if (!wechatForm.value.qr_url) {
      showToast('请上传收款码')
      return
    }
    const list = paymentSettings.value.wechat_qrs || []
    const item = {
      name: wechatForm.value.name || '',
      qr_url: wechatForm.value.qr_url
    }
    if (editingWechatIndex.value >= 0 && editingWechatIndex.value < list.length) {
      list.splice(editingWechatIndex.value, 1, item)
    } else {
      list.push(item)
    }
    paymentSettings.value.wechat_qrs = [...list]
    await savePaymentSettings()
    showToast('保存成功')
    showWechatForm.value = false
  } catch (error) {
    console.error('保存失败:', error)
    showToast('保存失败')
  }
}

const onSaveAlipay = async () => {
  try {
    if (!alipayForm.value.qr_url) {
      showToast('请上传收款码')
      return
    }
    const list = paymentSettings.value.alipay_qrs || []
    const item = {
      name: alipayForm.value.name || '',
      qr_url: alipayForm.value.qr_url
    }
    if (editingAlipayIndex.value >= 0 && editingAlipayIndex.value < list.length) {
      list.splice(editingAlipayIndex.value, 1, item)
    } else {
      list.push(item)
    }
    paymentSettings.value.alipay_qrs = [...list]
    await savePaymentSettings()
    showToast('保存成功')
    showAlipayForm.value = false
  } catch (error) {
    console.error('保存失败:', error)
    showToast('保存失败')
  }
}

// 压缩图片
const compressImage = (file, maxWidth = 600, quality = 0.6) => {
  return new Promise((resolve) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const img = new Image()
      img.onload = () => {
        const canvas = document.createElement('canvas')
        let width = img.width
        let height = img.height
        
        // 更激进的压缩策略
        if (width > maxWidth) {
          height = (height * maxWidth) / width
          width = maxWidth
        }
        
        canvas.width = width
        canvas.height = height
        
        const ctx = canvas.getContext('2d')
        ctx.drawImage(img, 0, 0, width, height)
        
        // 使用更低的质量
        let compressedBase64 = canvas.toDataURL('image/jpeg', quality)
        
        // 如果还是太大，进一步降低质量
        if (compressedBase64.length > 500000) { // 500KB
          compressedBase64 = canvas.toDataURL('image/jpeg', 0.4)
        }
        
        console.log('压缩后图片大小:', Math.round(compressedBase64.length / 1024), 'KB')
        resolve(compressedBase64)
      }
      img.src = e.target.result
    }
    reader.readAsDataURL(file)
  })
}

// 上传微信收款码
const afterReadWechat = async (file) => {
  try {
    showToast('正在处理图片...')
    const compressed = await compressImage(file.file)
    wechatForm.value.qr_url = compressed
    wechatFiles.value = [{ url: compressed, isImage: true }]
    showToast('图片上传成功')
  } catch (error) {
    console.error('图片处理失败:', error)
    showToast('图片处理失败')
  }
}

// 上传支付宝收款码
const afterReadAlipay = async (file) => {
  try {
    showToast('正在处理图片...')
    const compressed = await compressImage(file.file)
    alipayForm.value.qr_url = compressed
    alipayFiles.value = [{ url: compressed, isImage: true }]
    showToast('图片上传成功')
  } catch (error) {
    console.error('图片处理失败:', error)
    showToast('图片处理失败')
  }
}

onMounted(() => {
  loadPaymentSettings()
})
</script>

<style scoped>
.payment-settings-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 60px;
}

.popup-content {
  padding: 20px;
}

.popup-header {
  text-align: center;
  margin-bottom: 20px;
}

.popup-header h3 {
  margin: 0;
  font-size: 18px;
}
</style>
