<template>
  <div class="platform-address-page">
    <van-nav-bar
      title="平台收货地址"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    />

    <div class="content">
      <van-cell-group inset>
        <van-cell title="平台收货地址" />
        <div v-if="addresses.length === 0" class="empty">
          <van-empty description="暂未添加平台收货地址" />
        </div>
        <template v-else>
          <van-cell
            v-for="addr in addresses"
            :key="addr.id"
            :title="formatRegion(addr)"
            is-link
            @click="editAddress(addr)"
          >
            <template #label>
              <div class="addr-label">
                <div>收件人：{{ addr.receiver || '-' }} {{ addr.phone || '' }}</div>
                <div>{{ addr.detail }}</div>
              </div>
            </template>
            <template #value>
              <div class="addr-actions">
                <van-tag v-if="addr.is_default" type="danger" round>默认</van-tag>
                <van-button
                  v-else
                  size="mini"
                  type="primary"
                  plain
                  @click.stop="setDefault(addr.id)"
                >
                  设为默认
                </van-button>
                <van-button
                  size="mini"
                  type="danger"
                  plain
                  @click.stop="deleteAddress(addr.id)"
                >
                  删除
                </van-button>
              </div>
            </template>
          </van-cell>
        </template>
      </van-cell-group>

      <div class="btn-wrapper">
        <van-button type="primary" round block icon="plus" @click="openAdd">
          新增平台收货地址
        </van-button>
      </div>
    </div>

    <van-popup v-model:show="showForm" position="bottom" round :style="{ height: '80%' }">
      <div class="popup-content">
        <div class="popup-header">{{ editingId ? '编辑地址' : '新增地址' }}</div>
        <van-form @submit="onSubmit">
          <van-field
            v-model="form.receiver"
            label="收件人"
            placeholder="请输入收件人姓名"
          />
          <van-field
            v-model="form.phone"
            label="联系电话"
            type="tel"
            placeholder="请输入联系电话"
          />
          <van-field
            v-model="regionText"
            label="地区"
            readonly
            is-link
            placeholder="请选择省/市/区"
            @click="showArea = true"
          />
          <van-field
            v-model="form.town"
            label="乡镇街道"
            placeholder="请输入乡镇/街道"
          />
          <van-field
            v-model="form.detail"
            label="详细地址"
            type="textarea"
            rows="2"
            autosize
            placeholder="如：xx路xx号xx楼"
          />
          <div class="form-footer">
            <van-button round block type="primary" native-type="submit">
              保存
            </van-button>
          </div>
        </van-form>
      </div>
    </van-popup>

    <van-popup v-model:show="showArea" position="bottom" :style="{ height: '60%' }">
      <van-area
        :area-list="areaList"
        @confirm="onAreaConfirm"
        @cancel="showArea = false"
      />
    </van-popup>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { showToast, showDialog } from 'vant'
import { areaList } from '@vant/area-data'
import request from '../../utils/request'
import { API_ENDPOINTS } from '../../config/api'

const addresses = ref([])
const showForm = ref(false)
const showArea = ref(false)
const editingId = ref(null)

const form = ref({
  id: null,
  receiver: '',
  phone: '',
  province: '',
  city: '',
  county: '',
  town: '',
  detail: '',
  is_default: false
})

const regionText = ref('')

const loadAddresses = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.CONFIG)
    let list = []
    if (data && data.configs && data.configs.length > 0) {
      data.configs.forEach((item) => {
        const key = item.key || item.Key
        const value = item.value || item.Value
        if (key === 'platform_addresses' && value) {
          try {
            const parsed = JSON.parse(value)
            if (Array.isArray(parsed)) {
              list = parsed
            }
          } catch (err) {
            console.error('解析平台收货地址配置失败:', err)
          }
        }
      })
    }
    addresses.value = list
  } catch (e) {
    console.error('加载平台收货地址失败:', e)
  }
}

const saveAddresses = async () => {
  try {
    const payload = {
      platform_addresses: JSON.stringify(addresses.value || [])
    }
    await request.post(API_ENDPOINTS.CONFIG + '/batch', payload)
  } catch (e) {
    console.error('保存平台收货地址失败:', e)
    showToast('保存失败')
  }
}

const formatRegion = (addr) => {
  return [addr.province, addr.city, addr.county].filter(Boolean).join(' ')
}

const openAdd = () => {
  editingId.value = null
  form.value = {
    id: null,
    receiver: '',
    phone: '',
    province: '',
    city: '',
    county: '',
    town: '',
    detail: '',
    is_default: addresses.value.length === 0
  }
  regionText.value = ''
  showForm.value = true
}

const editAddress = (addr) => {
  editingId.value = addr.id
  form.value = { ...addr }
  regionText.value = formatRegion(addr)
  showForm.value = true
}

const setDefault = async (id) => {
  addresses.value = addresses.value.map((item) => ({
    ...item,
    is_default: item.id === id
  }))
  await saveAddresses()
  showToast('已设为默认地址')
}

const deleteAddress = async (id) => {
  const confirmed = await new Promise((resolve) => {
    showDialog({
      title: '删除地址',
      message: '确定要删除该地址吗？',
      showCancelButton: true,
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      beforeClose: (action) => {
        resolve(action === 'confirm')
        return true
      }
    }).catch(() => {
      resolve(false)
    })
  })

  if (!confirmed) return

  addresses.value = addresses.value.filter((item) => item.id !== id)

  if (!addresses.value.some((item) => item.is_default) && addresses.value.length > 0) {
    addresses.value[0].is_default = true
  }

  await saveAddresses()
  showToast('已删除')
}

const onAreaConfirm = ({ selectedOptions }) => {
  const [p, c, d] = selectedOptions
  form.value.province = p?.text || ''
  form.value.city = c?.text || ''
  form.value.county = d?.text || ''
  regionText.value = formatRegion(form.value)
  showArea.value = false
}

const onSubmit = async () => {
  if (!regionText.value) {
    showToast('请选择地区')
    return
  }
  if (!form.value.receiver) {
    showToast('请输入收件人')
    return
  }
  if (!form.value.phone) {
    showToast('请输入联系电话')
    return
  }
  if (!form.value.detail) {
    showToast('请输入详细地址')
    return
  }

  if (editingId.value) {
    addresses.value = addresses.value.map((item) =>
      item.id === editingId.value ? { ...item, ...form.value } : item
    )
  } else {
    const id = Date.now()
    const isDefault = form.value.is_default || addresses.value.length === 0
    const newAddr = {
      ...form.value,
      id,
      is_default: isDefault
    }
    if (isDefault) {
      addresses.value = addresses.value.map((item) => ({ ...item, is_default: false }))
    }
    addresses.value.push(newAddr)
  }

  await saveAddresses()
  showForm.value = false
  showToast('保存成功')
}

onMounted(() => {
  loadAddresses()
})
</script>

<style scoped>
.platform-address-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 60px;
}

.content {
  padding: 16px 0 80px;
}

.empty {
  padding: 20px 0;
}

.addr-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.addr-label {
  font-size: 12px;
  color: #666;
}

.addr-label div:first-child {
  margin-bottom: 4px;
}

.btn-wrapper {
  margin: 16px;
}

.popup-content {
  padding: 16px;
}

.popup-header {
  text-align: center;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
}

.form-footer {
  margin-top: 16px;
}
</style>
