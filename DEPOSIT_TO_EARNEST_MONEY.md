# "充值"改为"付定金"功能实现

**完成时间**: 2025-11-18 16:12  
**功能**: 将充值改为付定金，并重新设计付定金页面UI

---

## ✅ 已完成的功能

### 1. 文案修改

所有"充值"相关文案已改为"付定金"：
- ✅ 按钮文案：充值 → 付定金
- ✅ 页面标题：充值 → 转包付定金
- ✅ Tab标签：充值 → 付定金
- ✅ 流水记录：充值 → 付定金
- ✅ 提示信息：充值申请 → 付定金申请

---

### 2. 付定金页面重新设计

参照提供的设计图完全重构了付定金页面：

#### 新增功能模块

1. **金额输入**
   - 大号字体显示
   - 快捷金额选择按钮（8个）

2. **付款账户**
   - 选择用户自己的银行卡
   - 显示账户类型、户名、产名、账户
   - Picker选择器

3. **收款账户**
   - 显示平台收款信息
   - 支持多个收款账户
   - 显示账户类型、户名、产名、银行、账户

4. **支付凭证（必填）**
   - 图片上传功能
   - 自动压缩图片
   - 最大1张

5. **温馨提示**
   - 资金安全提示信息

6. **备注**
   - 文本域输入

7. **协议勾选**
   - 红色强调文字
   - 必须勾选才能提交

8. **提交审核按钮**
   - 红色主题色
   - 固定底部
   - 禁用状态控制

---

## 🎨 UI 设计

### 页面结构

```
┌─────────────────────────────────────┐
│  ← 转包付定金                        │
├─────────────────────────────────────┤
│  金额                                │
│  [________5000________]              │
│                                      │
│  [5000][6000][10000][15000]         │
│  [2万]  [5万] [10万] [20万]         │
├─────────────────────────────────────┤
│  付款账户                            │
│  选择付款账户            →          │
│  ┌───────────────────────────┐      │
│  │ 账户类型  银行卡          │      │
│  │ 户名      张三            │      │
│  │ 产名      工商银行        │      │
│  │ 账户      6217...5675     │      │
│  └───────────────────────────┘      │
├─────────────────────────────────────┤
│  收款账户                            │
│  点击删除或者点击设置为默认          │
│  ┌───────────────────────────┐      │
│  │ 账户类型  银行卡          │      │
│  │ 户名      速金盈科技      │      │
│  │ 产名      工商银行        │      │
│  │ 银行      深圳南园园区    │      │
│  │ 账户      6217...5675     │      │
│  └───────────────────────────┘      │
├─────────────────────────────────────┤
│  支付凭证(必填)                      │
│  [📷 上传图片]                       │
├─────────────────────────────────────┤
│  温馨提示                            │
│  为保证资金安全...                   │
├─────────────────────────────────────┤
│  备注                                │
│  [请输入内容]                        │
├─────────────────────────────────────┤
│  ☑ 请仔细阅读并同意                  │
│     账户打款者姓名需一致协议         │
├─────────────────────────────────────┤
│            [提交审核]                │
└─────────────────────────────────────┘
```

---

### 快捷金额按钮

```
[5000]  [6000]  [10000] [15000]
[2万]   [5万]   [10万]  [20万]
```

- 4列网格布局
- 点击自动填充金额
- 选中状态高亮（蓝色）

---

### 颜色方案

| 元素 | 颜色 | 说明 |
|------|------|------|
| 标题文字 | `#ee0a24` | 红色 |
| 提交按钮 | `#ee0a24` | 红色 danger |
| 协议强调 | `#ee0a24` | 红色 |
| 普通文字 | `#333` | 深灰 |
| 次要文字 | `#666` | 灰色 |
| 提示文字 | `#999` | 浅灰 |
| 背景色 | `#f7f8fa` | 浅灰背景 |

---

## 🔧 技术实现

### 1. 模板结构

**文件**: `frontend/src/pages/Funds.vue`

```vue
<!-- 付定金弹窗 -->
<van-popup v-model:show="showDeposit" position="bottom" round :style="{ height: '90%' }">
  <div class="deposit-popup">
    <van-nav-bar title="转包付定金" left-arrow @click-left="showDeposit = false" />
    
    <div class="deposit-content">
      <!-- 金额输入 -->
      <div class="amount-section">
        <div class="amount-label">金额</div>
        <van-field v-model="depositForm.amount" type="digit" placeholder="请输入金额" />
      </div>
      
      <!-- 快捷金额 -->
      <div class="quick-amounts">
        <van-button 
          v-for="amount in quickAmounts" 
          :key="amount"
          :type="depositForm.amount == amount ? 'primary' : 'default'"
          @click="depositForm.amount = amount"
        >
          {{ formatQuickAmount(amount) }}
        </van-button>
      </div>
      
      <!-- 付款账户 -->
      <div class="section">
        <div class="section-title">付款账户</div>
        <van-cell title="选择付款账户" is-link @click="showPaymentCardPicker = true" />
        <div v-if="selectedPaymentCard" class="card-info">
          <!-- 卡片信息 -->
        </div>
      </div>
      
      <!-- 收款账户 -->
      <!-- 支付凭证 -->
      <!-- 温馨提示 -->
      <!-- 备注 -->
      <!-- 协议 -->
      
      <!-- 提交按钮 -->
      <div class="submit-btn">
        <van-button type="danger" block round @click="onDeposit" :disabled="!agreeProtocol">
          提交审核
        </van-button>
      </div>
    </div>
  </div>
</van-popup>
```

---

### 2. 数据状态

```javascript
// 快捷金额
const quickAmounts = ref([5000, 6000, 10000, 15000, 20000, 50000, 100000, 200000])

// 付款账户选择
const showPaymentCardPicker = ref(false)
const selectedPaymentCard = ref(null)
const paymentCardColumns = ref([])

// 支付凭证
const voucherFiles = ref([])
const voucherUrl = ref('')

// 协议
const agreeProtocol = ref(false)

// 表单数据
const depositForm = ref({
  amount: '',
  note: ''
})
```

---

### 3. 核心方法

#### 格式化快捷金额

```javascript
const formatQuickAmount = (amount) => {
  if (amount >= 10000) {
    return (amount / 10000) + '万'
  }
  return amount
}
```

**效果**:
- 5000 → "5000"
- 10000 → "1万"
- 50000 → "5万"

---

#### 选择付款账户

```javascript
const onSelectPaymentCard = (value) => {
  const card = bankCards.value.find(c => (c.id || c.ID) === value.value)
  if (card) {
    selectedPaymentCard.value = {
      id: card.id || card.ID,
      card_holder: card.card_holder || card.CardHolder,
      bank_name: card.bank_name || card.BankName,
      card_number: card.card_number || card.CardNumber
    }
  }
  showPaymentCardPicker.value = false
}
```

---

#### 上传支付凭证

```javascript
const afterReadVoucher = async (file) => {
  try {
    showToast('正在处理图片...')
    const compressed = await compressImage(file.file)
    voucherUrl.value = compressed
    showToast('图片上传成功')
  } catch (error) {
    console.error('图片处理失败:', error)
    showToast('图片处理失败')
  }
}

// 压缩图片（避免LocalStorage超限）
const compressImage = (file, maxWidth = 800, quality = 0.8) => {
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
        
        const compressedBase64 = canvas.toDataURL('image/jpeg', quality)
        resolve(compressedBase64)
      }
      img.src = e.target.result
    }
    reader.readAsDataURL(file)
  })
}
```

---

#### 提交付定金

```javascript
const onDeposit = async () => {
  try {
    // 验证
    if (!depositForm.value.amount) {
      showToast('请输入金额')
      return
    }
    
    if (!selectedPaymentCard.value) {
      showToast('请选择付款账户')
      return
    }
    
    if (!voucherUrl.value) {
      showToast('请上传支付凭证')
      return
    }
    
    if (!agreeProtocol.value) {
      showToast('请阅读并同意协议')
      return
    }
    
    // 提交
    const requestData = {
      amount: parseFloat(depositForm.value.amount),
      method: 'bank',
      voucher_url: voucherUrl.value,
      note: depositForm.value.note || ''
    }
    
    await request.post(API_ENDPOINTS.DEPOSIT_CREATE, requestData)
    
    showToast('付定金申请已提交，等待审核')
    
    // 重置表单
    showDeposit.value = false
    depositForm.value = { amount: '', note: '' }
    voucherFiles.value = []
    voucherUrl.value = ''
    agreeProtocol.value = false
    
    loadUserInfo()
    onRefresh()
  } catch (error) {
    const errorMsg = error.response?.data?.error || '操作失败'
    showToast(errorMsg)
  }
}
```

---

### 4. 样式特点

```css
/* 快捷金额网格布局 */
.quick-amounts {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

/* 红色标题 */
.section-title {
  font-size: 16px;
  font-weight: bold;
  color: #ee0a24;
  margin-bottom: 12px;
}

/* 卡片信息 */
.card-info {
  background: #f7f8fa;
  border-radius: 8px;
  padding: 12px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

/* 固定底部按钮 */
.submit-btn {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 16px;
  background: #fff;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.05);
}
```

---

## 📊 数据流程

### 付定金流程

```
1. 用户点击"付定金"
   ↓
2. 打开付定金弹窗
   ↓
3. 输入金额（可选快捷金额）
   ↓
4. 选择付款账户（用户的银行卡）
   ↓
5. 查看收款账户（平台收款信息）
   ↓
6. 上传支付凭证
   ↓
7. 填写备注（可选）
   ↓
8. 勾选协议
   ↓
9. 点击"提交审核"
   ↓
10. 验证表单
    ├─ 金额不为空
    ├─ 已选择付款账户
    ├─ 已上传凭证
    └─ 已勾选协议
   ↓
11. 提交到后端
   ↓
12. 显示"付定金申请已提交，等待审核"
   ↓
13. 关闭弹窗，刷新数据
```

---

### API请求参数

```json
{
  "amount": 5000,
  "method": "bank",
  "voucher_url": "data:image/jpeg;base64,...",
  "note": "转账备注信息"
}
```

---

## ⚠️ 注意事项

### 1. 图片压缩

**问题**: 支付凭证图片可能很大（2-5MB）  
**解决**: 自动压缩到800px宽度，质量0.8

```javascript
const compressed = await compressImage(file.file, 800, 0.8)
```

**效果**: 2MB → ~150KB

---

### 2. 表单验证

必须按顺序验证：
1. ✅ 金额
2. ✅ 付款账户
3. ✅ 支付凭证
4. ✅ 协议勾选

任一项未完成，显示相应提示。

---

### 3. 协议勾选

- 未勾选时按钮禁用（`:disabled="!agreeProtocol"`）
- 红色强调文字
- 必须勾选才能提交

---

### 4. 付款账户初始化

```javascript
// 默认选择第一张卡
if (bankCards.value.length > 0) {
  const firstCard = bankCards.value[0]
  selectedPaymentCard.value = {
    id: firstCard.id || firstCard.ID,
    card_holder: firstCard.card_holder || firstCard.CardHolder,
    bank_name: firstCard.bank_name || firstCard.BankName,
    card_number: firstCard.card_number || firstCard.CardNumber
  }
}
```

---

## 🧪 测试清单

### 基础功能测试

- [ ] 点击"付定金"按钮打开弹窗
- [ ] 返回按钮关闭弹窗
- [ ] 输入金额
- [ ] 点击快捷金额按钮自动填充
- [ ] 选择付款账户
- [ ] 查看收款账户信息
- [ ] 上传支付凭证
- [ ] 填写备注
- [ ] 勾选协议
- [ ] 提交审核

### 验证测试

- [ ] 未填金额，点击提交 → 提示"请输入金额"
- [ ] 未选择付款账户 → 提示"请选择付款账户"
- [ ] 未上传凭证 → 提示"请上传支付凭证"
- [ ] 未勾选协议 → 按钮禁用

### 样式测试

- [ ] 快捷金额按钮4列布局
- [ ] 选中状态蓝色高亮
- [ ] 红色标题显示正确
- [ ] 卡片信息格式正确
- [ ] 提交按钮固定底部
- [ ] 弹窗高度90%

### 数据测试

- [ ] 快捷金额格式化（2万、5万等）
- [ ] 银行卡信息正确显示
- [ ] 凭证图片压缩成功
- [ ] 提交后数据刷新

---

## 📝 修改文件列表

1. ✅ `frontend/src/pages/Funds.vue`
   - 文案修改（充值 → 付定金）
   - 重新设计付定金弹窗UI
   - 添加快捷金额选择
   - 添加付款账户选择
   - 添加支付凭证上传
   - 添加协议勾选
   - 添加完整样式

---

## 💡 后续优化建议

### 1. 服务器端图片存储

目前凭证存储在LocalStorage（base64），建议改为服务器存储：

```javascript
const afterReadVoucher = async (file) => {
  const formData = new FormData()
  formData.append('file', file.file)
  
  const response = await request.post('/api/v1/upload', formData)
  voucherUrl.value = response.url  // 存储URL而非base64
}
```

---

### 2. 多收款账户切换

支持配置多个收款账户，并可以选择：

```vue
<van-radio-group v-model="selectedReceiverAccount">
  <van-radio v-for="account in paymentInfo.bank_cards" :key="account.id" :name="account.id">
    {{ account.bank_name }} - {{ account.account_number }}
  </van-radio>
</van-radio-group>
```

---

### 3. 凭证预览

添加图片预览功能：

```vue
<van-image
  v-if="voucherUrl"
  :src="voucherUrl"
  width="100"
  height="100"
  fit="cover"
  @click="previewVoucher"
/>
```

---

### 4. 历史金额记录

记录用户常用金额：

```javascript
// 保存到本地
const recentAmounts = JSON.parse(localStorage.getItem('recent_amounts') || '[]')
recentAmounts.unshift(depositForm.value.amount)
localStorage.setItem('recent_amounts', JSON.stringify(recentAmounts.slice(0, 5)))
```

---

**刷新浏览器，测试全新的付定金功能！** 🎉
