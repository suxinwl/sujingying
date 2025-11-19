<template>
  <div class="bank-cards-page">
    <van-nav-bar
      title="银行卡管理"
      fixed
      placeholder
      left-arrow
      @click-left="$router.back()"
    />
    
    <!-- 银行卡列表 -->
    <div class="cards-container">
      <div v-if="cards.length === 0" class="empty">
        <van-empty description="暂无银行卡" />
      </div>
      
      <div v-for="card in cards" :key="card.id || card.ID" class="card-item">
        <div class="card-header">
          <div class="bank-info">
            <span class="bank-name">{{ card.bank_name || card.BankName }}</span>
            <van-tag v-if="card.is_default || card.IsDefault" type="success" size="small">默认</van-tag>
          </div>
          <van-icon name="delete-o" @click="deleteCard(card.id || card.ID)" />
        </div>
        <div class="card-number">{{ formatCardNumber(card.card_number || card.CardNumber) }}</div>
        <div class="card-footer">
          <div class="card-holder">{{ card.card_holder || card.CardHolder }}</div>
          <van-button 
            v-if="!(card.is_default || card.IsDefault)" 
            size="small" 
            type="primary"
            @click="setDefaultCard(card.id || card.ID)"
          >
            设为默认
          </van-button>
        </div>
      </div>
    </div>
    
    <!-- 添加按钮 -->
    <div class="add-button">
      <van-button type="primary" round block @click="showAddDialog = true">
        添加银行卡
      </van-button>
    </div>
    
    <!-- 添加银行卡弹窗 -->
    <van-popup v-model:show="showAddDialog" position="bottom" round>
      <div class="popup-content">
        <div class="popup-header">
          <h3>添加银行卡</h3>
          <p class="tip-text">添加银行卡需要验证支付密码，如未设置请先设置支付密码</p>
        </div>
        <van-form @submit="onSubmit">
          <van-field
            v-model="form.bank_name"
            label="银行名称"
            placeholder="请输入银行名称"
            :rules="[{ required: true, message: '请输入银行名称' }]"
          />
          <van-field
            v-model="form.card_number"
            type="digit"
            label="卡号"
            placeholder="请输入银行卡号"
            :rules="[{ required: true, message: '请输入银行卡号' }]"
          />
          <van-field
            v-model="form.card_holder"
            label="持卡人"
            placeholder="请输入持卡人姓名"
            :rules="[{ required: true, message: '请输入持卡人姓名' }]"
          />
          <van-field
            v-model="form.pay_password"
            type="password"
            label="支付密码"
            placeholder="请输入支付密码"
            :rules="[{ required: true, message: '请输入支付密码' }]"
          />
          <div style="margin: 16px;">
            <van-button round block type="primary" native-type="submit">
              确认添加
            </van-button>
          </div>
        </van-form>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
/**
 * @file BankCards.vue
 * @description 银行卡管理页面
 * @author AI
 * @date 2025-11-18
 */

import { ref, onMounted } from 'vue'
import { showToast, showConfirmDialog } from 'vant'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'

/**
 * 银行卡列表
 * @type {import('vue').Ref<Array>}
 */
const cards = ref([])

/**
 * 显示添加弹窗
 * @type {import('vue').Ref<boolean>}
 */
const showAddDialog = ref(false)

/**
 * 表单数据
 * @type {import('vue').Ref<Object>}
 */
const form = ref({
  bank_name: '',
  card_number: '',
  card_holder: '',
  pay_password: ''
})

/**
 * 格式化卡号
 * @param {string} cardNumber - 卡号
 * @returns {string} 格式化后的卡号
 */
const formatCardNumber = (cardNumber) => {
  if (!cardNumber) return ''
  // 保留前4位和后4位，中间用*代替
  const start = cardNumber.slice(0, 4)
  const end = cardNumber.slice(-4)
  const middle = '*'.repeat(Math.max(0, cardNumber.length - 8))
  return `${start} ${middle} ${end}`
}

/**
 * 加载银行卡列表
 * @async
 * @returns {Promise<void>}
 */
const loadCards = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.BANK_CARDS)
    console.log('银行卡数据:', data)
    cards.value = data.cards || data.list || []
    console.log('解析后的银行卡列表:', cards.value)
  } catch (error) {
    console.error('加载银行卡失败:', error)
    showToast('加载失败')
  }
}

/**
 * 提交添加银行卡
 * @async
 * @returns {Promise<void>}
 */
const onSubmit = async () => {
  try {
    await request.post(API_ENDPOINTS.BANK_CARD_CREATE, form.value)
    showToast('添加成功')
    showAddDialog.value = false
    form.value = {
      bank_name: '',
      card_number: '',
      card_holder: '',
      pay_password: ''
    }
    loadCards()
  } catch (error) {
    console.error('添加银行卡失败:', error)
    // 显示具体错误信息
    if (error.response?.data?.error) {
      showToast(error.response.data.error)
    } else {
      showToast('添加失败')
    }
  }
}

/**
 * 设置默认银行卡
 * @async
 * @param {number} id - 银行卡ID
 * @returns {Promise<void>}
 */
const setDefaultCard = async (id) => {
  try {
    await request.put(`/api/v1/bank-cards/${id}/default`)
    showToast('已设为默认银行卡')
    loadCards()
  } catch (error) {
    console.error('设置默认银行卡失败:', error)
    const errorMsg = error.response?.data?.error || error.response?.data?.message || '设置失败'
    showToast(errorMsg)
  }
}

/**
 * 删除银行卡
 * @async
 * @param {number} id - 银行卡ID
 * @returns {Promise<void>}
 */
const deleteCard = async (id) => {
  try {
    await showConfirmDialog({
      title: '确认删除',
      message: '确定要删除这张银行卡吗？'
    })
    
    await request.delete(API_ENDPOINTS.BANK_CARD_DELETE.replace(':id', id))
    showToast('删除成功')
    loadCards()
  } catch (error) {
    if (error === 'cancel') return
    console.error('删除银行卡失败:', error)
    // 显示具体错误信息
    const errorMsg = error.response?.data?.error || error.response?.data?.message || '删除失败'
    showToast(errorMsg)
  }
}

onMounted(() => {
  loadCards()
})
</script>

<style scoped>
.bank-cards-page {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 80px;
}

.cards-container {
  padding: 16px;
}

.card-item {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  border-radius: 12px;
  margin-bottom: 16px;
  color: #fff;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.bank-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.bank-name {
  font-size: 18px;
  font-weight: 500;
}

.card-number {
  font-size: 20px;
  letter-spacing: 2px;
  margin-bottom: 12px;
  font-family: 'Courier New', monospace;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-holder {
  font-size: 14px;
  opacity: 0.9;
}

.add-button {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 16px;
  background: #fff;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.05);
}

.popup-content {
  padding: 20px;
}

.popup-header {
  text-align: center;
  margin-bottom: 20px;
}

.popup-header h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
}

.tip-text {
  font-size: 12px;
  color: #999;
  margin: 0;
  line-height: 1.5;
}

.empty {
  padding: 100px 0;
}
</style>
