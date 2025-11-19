# 交易页面UI重新设计

**设计时间**: 2025-11-18 13:10  
**参考**: 锁价买料/锁价卖料界面  
**文件**: `frontend/src/pages/Trade.vue`

---

## 🎨 设计改进

### 1. 顶部导航栏

**原设计**:
- 固定标题"交易"

**新设计**:
- ✅ 动态标题：买入显示"锁价买料"，卖出显示"锁价卖料"
- ✅ 添加返回按钮

```vue
<van-nav-bar
  :title="tradeType === 'buy' ? '锁价买料' : '锁价卖料'"
  left-arrow
  @click-left="$router.back()"
/>
```

---

### 2. 价格大卡片

**新增醒目价格展示**:
```vue
<div class="price-card" :class="tradeType">
  <div class="price-value">{{ quoteStore.priceDisplay }}</div>
  <div class="price-label">
    {{ tradeType === 'buy' ? '黄金销售价(元/克)' : '黄金回购价(元/克)' }}
  </div>
</div>
```

**样式**:
- 买入：红色渐变 `#ff6b6b -> #ee5a52`
- 卖出：绿色渐变 `#51cf66 -> #40c057`
- 大字号显示价格（36px）
- 居中布局

---

### 3. 商品Tab

**新增**:
```vue
<van-tabs v-model:active="productType" color="#f44">
  <van-tab title="黄金" name="gold"></van-tab>
  <van-tab title="白银" name="silver"></van-tab>
</van-tabs>
```

---

### 4. 买入界面重新设计

#### 克重输入区域
```vue
<!-- 提示行 -->
<div class="tip-row">
  <span class="label">买入克重（需要限制最少于100克）</span>
  <span class="tip">最终以实际报出货重量为准</span>
</div>

<!-- 输入框 -->
<van-field
  v-model="form.amount"
  type="number"
  placeholder="请输入克重"
  class="amount-input"
/>

<!-- 错误提示 -->
<div v-if="form.amount && form.amount < 100" class="error-tip">
  买入克重不能少于100
</div>
```

#### 快捷克重选择
```vue
<div class="quick-amounts">
  <van-button v-for="amount in [1000, 2000, 3000, 5000, 100, 200, 300, 500]">
    {{ amount }}g
  </van-button>
</div>
```

**布局**: 4列网格，8个快捷选项

#### 费用明细
```vue
<van-cell-group>
  <van-cell title="预估金额" :value="'¥' + formatMoney(estimatedAmount)" />
  <van-cell 
    title="总服务费(按实际出货收取)" 
    :value="'¥' + formatMoney(serviceFee)" 
    value-class="highlight"
  />
  <van-cell title="定金" :value="'¥' + formatMoney(requiredDepositValue)" />
  <van-cell title="可用定金" :value="'¥' + formatMoney(balance.available_deposit)" />
</van-cell-group>
```

**计算逻辑**:
- 预估金额 = 克重 × 当前价格
- 服务费 = 克重 × 服务费率（2%）
- 定金 = 预估金额 × 定金率（10%）

#### 业务说明
```vue
<div class="business-note">
  <div class="note-title">业务说明</div>
  <div class="note-content">
    当客户买料价格在客户下单价格，延后提货卖料，我司依会约定取款支付取货款及服务费，客户收货当天需要完成支付。
  </div>
</div>
```

**样式**: 浅黄色背景 + 红色左边框

#### 协议勾选
```vue
<van-checkbox v-model="agreeProtocol" class="protocol-check">
  我已阅读并同意<span class="link">《黄金居购指服务协议》</span>
</van-checkbox>
```

#### 提交按钮
```vue
<van-button
  type="danger"
  size="large"
  round
  block
  :disabled="!agreeProtocol || !form.amount || form.amount < 100"
  @click="onSubmit"
>
  立即买入
</van-button>
```

---

### 5. 卖出界面重新设计

#### 额外功能

**收货地址**:
```vue
<div class="address-section">
  <div class="section-title">收货地址</div>
  <van-cell
    icon="location-o"
    :title="userAddress.name || '请设置收货地址'"
    :label="userAddress.phone ? `${userAddress.phone}\n${userAddress.address}` : ''"
    is-link
  />
</div>
```

**锁价卖料流程**:
```vue
<div class="process-section">
  <div class="section-title">锁价卖料流程</div>
  <div class="process-steps">
    <div class="step">
      <div class="step-icon">📱</div>
      <div class="step-text">在线锁价</div>
    </div>
    <div class="step-arrow">···></div>
    <div class="step">
      <div class="step-icon">📦</div>
      <div class="step-text">顺丰保价</div>
    </div>
    <div class="step-arrow">···></div>
    <div class="step">
      <div class="step-icon">🔬</div>
      <div class="step-text">检测报告</div>
    </div>
    <div class="step-arrow">···></div>
    <div class="step">
      <div class="step-icon">💰</div>
      <div class="step-text">结算付款</div>
    </div>
  </div>
</div>
```

---

## 📝 代码结构变化

### 响应式变量

**移除**:
```javascript
form.price  // ❌ 不再手动输入价格
```

**新增**:
```javascript
const productType = ref('gold')  // 商品类型
const agreeProtocol = ref(false)  // 协议同意状态
const showAddressPopup = ref(false)  // 地址选择弹窗
const userAddress = ref({})  // 用户地址
const quickAmounts = [1000, 2000, 3000, 5000, 100, 200, 300, 500]  // 快捷克重
```

### 计算属性

```javascript
// 预估金额
const estimatedAmount = computed(() => {
  const amount = parseFloat(form.value.amount) || 0
  const price = quoteStore.currentPrice || 0
  return amount * price
})

// 服务费
const serviceFee = computed(() => {
  const amount = parseFloat(form.value.amount) || 0
  return amount * (config.value?.service_fee_rate || 0.02)
})

// 所需定金
const requiredDepositValue = computed(() => {
  const depositRate = config.value?.deposit_rate || 0.1
  return estimatedAmount.value * depositRate
})
```

### 提交逻辑

**关键改变**:
- ✅ 使用市价 `quoteStore.currentPrice`
- ✅ 验证协议同意状态
- ✅ 验证最小克重（100克）
- ✅ 自动计算所有费用

```javascript
const onSubmit = async () => {
  if (!agreeProtocol.value) {
    showToast('请先阅读并同意服务协议')
    return
  }
  
  const amount = parseFloat(form.value.amount)
  if (!amount || amount < config.value.min_order_amount) {
    showToast(`最低克重为${config.value.min_order_amount}克`)
    return
  }
  
  const data = await request.post(API_ENDPOINTS.ORDER_CREATE, {
    direction: tradeType.value,
    price: quoteStore.currentPrice,  // ✅ 使用市价
    quantity: amount
  })
  // ...
}
```

---

## 🎨 样式特点

### 1. 颜色方案
- **买入**: 红色系 `#ff6b6b`, `#ee5a52`, `#ff4444`
- **卖出**: 绿色系 `#51cf66`, `#40c057`
- **强调**: 橙色 `#ff6600`
- **背景**: `#f5f5f5`

### 2. 布局
- 16px 边距统一
- 8px 圆角
- 网格布局（快捷克重）
- Flexbox（流程展示）

### 3. 字体
- 标题: 14px, bold
- 正文: 12-14px
- 价格: 36px, bold
- 提示: 12px, #999

### 4. 交互反馈
- 按钮禁用状态
- 错误提示颜色
- 表单验证
- 加载状态

---

## 📊 对比总结

### 原设计
- ❌ 手动输入价格
- ❌ 简单表单
- ❌ 无快捷选项
- ❌ 无业务说明
- ❌ 无流程展示

### 新设计
- ✅ 自动使用市价
- ✅ 醒目价格卡片
- ✅ 快捷克重选择（8个选项）
- ✅ 详细费用明细
- ✅ 业务说明区域
- ✅ 协议勾选
- ✅ 流程可视化（卖出）
- ✅ 收货地址（卖出）
- ✅ 最小克重验证（100克）

---

## 🧪 测试清单

### 买入功能
- [ ] 价格卡片显示正确（红色渐变）
- [ ] 商品Tab切换正常
- [ ] 输入克重，费用自动计算
- [ ] 点击快捷克重按钮，自动填充
- [ ] 克重小于100，显示错误提示
- [ ] 未勾选协议，按钮禁用
- [ ] 勾选协议后，可提交
- [ ] 提交成功，清空表单

### 卖出功能
- [ ] 价格卡片显示正确（绿色渐变）
- [ ] 显示收货地址
- [ ] 显示流程步骤
- [ ] 费用计算正确
- [ ] 提交验证正常

### 响应式
- [ ] 移动端显示正常
- [ ] 按钮大小合适
- [ ] 字体清晰可读

---

## 📝 修改的文件

- ✅ `frontend/src/pages/Trade.vue` - 完全重构

---

## ✅ 重新设计完成

**改进点**:
1. ✅ 醒目的价格展示（红/绿渐变卡片）
2. ✅ 商品Tab（黄金/白银）
3. ✅ 快捷克重选择（8个选项）
4. ✅ 详细费用明细
5. ✅ 业务说明
6. ✅ 协议勾选
7. ✅ 流程可视化
8. ✅ 收货地址管理

**现在刷新浏览器查看全新的锁价买料/卖料界面！**
