# 交易页面最终修复

**修复时间**: 2025-11-18 13:24  
**修复内容**: 按用户需求调整交易逻辑

---

## ✅ 修复清单

### 1. ❌ 移除白银Tab

**修改前**:
```vue
<van-tabs v-model:active="productType">
  <van-tab title="黄金" name="gold"></van-tab>
  <van-tab title="白银" name="silver"></van-tab>  <!-- ❌ 移除 -->
</van-tabs>
```

**修改后**:
```vue
<!-- 只保留黄金交易，移除商品Tab -->
```

**结果**: ✅ 只交易黄金

---

### 2. ✅ 修复价格显示问题

#### quoteStore修改

**问题**: 黄金销售价和回购价显示为 `-.--`

**原因**: 
- 只有一个 `currentPrice`
- 没有区分买入价（销售价）和卖出价（回购价）

**修改后**:
```javascript
state: () => ({
  buyPrice: 0,   // 销售价（用户买入价 = Sell）
  sellPrice: 0,  // 回购价（用户卖出价 = Buy）
  // ...
})

getters: {
  buyPriceDisplay: (state) => {
    return state.buyPrice ? state.buyPrice.toFixed(2) : '-.--'
  },
  sellPriceDisplay: (state) => {
    return state.sellPrice ? state.sellPrice.toFixed(2) : '-.--'
  },
  // ...
}
```

#### 数据解析逻辑

```javascript
updateQuote(data) {
  // 优先使用黄金AU的价格
  let newBuyPrice = 0   // 销售价（Sell）
  let newSellPrice = 0  // 回购价（Buy）
  
  if (data.AU) {
    newBuyPrice = parseFloat(data.AU.Sell) || 0   // ✅ 销售价
    newSellPrice = parseFloat(data.AU.Buy) || 0   // ✅ 回购价
  }
  
  this.buyPrice = newBuyPrice
  this.sellPrice = newSellPrice
  
  console.log(`销售价: ${this.buyPrice}, 回购价: ${this.sellPrice}`)
}
```

#### 价格卡片显示

```vue
<div class="price-card" :class="tradeType">
  <div class="price-value">
    <!-- 买入显示销售价，卖出显示回购价 -->
    {{ tradeType === 'buy' ? quoteStore.buyPriceDisplay : quoteStore.sellPriceDisplay }}
  </div>
  <div class="price-label">
    {{ tradeType === 'buy' ? '黄金销售价(元/克)' : '黄金回购价(元/克)' }}
  </div>
</div>
```

**结果**: ✅ 买入显示销售价，卖出显示回购价

---

### 3. ✅ 修正预估金额计算

**用户要求**: 预估金额 = 黄金销售价 × 克重

**修改前**:
```javascript
const estimatedAmount = computed(() => {
  const amount = parseFloat(form.value.amount) || 0
  const price = quoteStore.currentPrice || 0  // ❌ 不区分买卖
  return amount * price
})
```

**修改后**:
```javascript
const estimatedAmount = computed(() => {
  const amount = parseFloat(form.value.amount) || 0
  // ✅ 买入用销售价，卖出用回购价
  const price = tradeType.value === 'buy' ? quoteStore.buyPrice : quoteStore.sellPrice
  return amount * price
})
```

**计算示例**:
- 买入100克，销售价500元/克
- 预估金额 = 100 × 500 = 50,000元

**结果**: ✅ 买入用销售价计算，卖出用回购价计算

---

### 4. ✅ 服务费从后台配置读取

**用户要求**: 服务费 = 超级管理员后台设置的百分比

**配置读取**:
```javascript
const loadConfig = async () => {
  const data = await request.get(API_ENDPOINTS.CONFIG)
  if (data.configs && Array.isArray(data.configs)) {
    data.configs.forEach(item => {
      if (item.key === 'service_fee_rate') {
        config.value.service_fee_rate = parseFloat(item.value) || 0.02  // ✅
      }
    })
  }
}
```

**服务费计算**:
```javascript
const serviceFee = computed(() => {
  // ✅ 预估金额 × 服务费率
  return estimatedAmount.value * (config.value?.service_fee_rate || 0.02)
})
```

**计算示例**:
- 预估金额：50,000元
- 服务费率：2%（后台配置）
- 服务费 = 50,000 × 0.02 = 1,000元

**后台配置**:
```json
{
  "key": "service_fee_rate",
  "value": "0.02"  // 2%
}
```

**结果**: ✅ 服务费率从后台读取

---

### 5. ✅ 修正定金计算公式

**用户要求**: 定金 = 克重 × 10

**修改前**:
```javascript
const requiredDepositValue = computed(() => {
  const depositRate = config.value?.deposit_rate || 0.1  // ❌ 按比例
  return estimatedAmount.value * depositRate
})
```

**修改后**:
```javascript
const requiredDepositValue = computed(() => {
  const amount = parseFloat(form.value.amount) || 0
  // ✅ 克重 × 10元/克
  return amount * (config.value?.deposit_per_gram || 10)
})
```

**计算示例**:
- 买入100克
- 定金 = 100 × 10 = 1,000元

**配置支持**:
```javascript
config.value = {
  deposit_per_gram: 10  // ✅ 每克定金10元
}

// 也可以从后台配置读取
if (item.key === 'deposit_per_gram') {
  config.value.deposit_per_gram = parseFloat(item.value) || 10
}
```

**结果**: ✅ 定金 = 克重 × 10

---

### 6. ✅ 可用定金显示

**用户要求**: 可用定金 = 账户可用定金

**代码**:
```javascript
const balance = ref({
  available_deposit: 0,  // ✅ 账户可用定金
  used_deposit: 0
})

const loadBalance = async () => {
  const data = await request.get(API_ENDPOINTS.USER_PROFILE)
  balance.value = {
    available_deposit: data.available_deposit || 0,  // ✅
    used_deposit: data.used_deposit || 0
  }
}
```

**显示**:
```vue
<van-cell 
  title="可用定金" 
  :value="'¥' + formatMoney(balance.available_deposit)"  <!-- ✅ -->
/>
```

**结果**: ✅ 显示账户可用定金

---

### 7. ✅ 订单提交价格修正

**修改**:
```javascript
const onSubmit = async () => {
  // ✅ 买入用销售价，卖出用回购价
  const price = tradeType.value === 'buy' 
    ? quoteStore.buyPrice 
    : quoteStore.sellPrice
    
  const data = await request.post(API_ENDPOINTS.ORDER_CREATE, {
    direction: tradeType.value,
    price: price,  // ✅ 正确的价格
    quantity: amount
  })
}
```

**结果**: ✅ 提交正确的价格

---

## 📊 完整计算逻辑

### 买入场景

**输入**:
- 克重: 100克
- 销售价: 500元/克
- 服务费率: 2%（后台配置）
- 每克定金: 10元

**计算**:
1. 预估金额 = 100 × 500 = **50,000元**
2. 服务费 = 50,000 × 0.02 = **1,000元**
3. 定金 = 100 × 10 = **1,000元**
4. 可用定金 = **308,072元**（从账户读取）

### 卖出场景

**输入**:
- 克重: 100克
- 回购价: 498元/克
- 服务费率: 2%
- 每克定金: 10元

**计算**:
1. 预估金额 = 100 × 498 = **49,800元**
2. 服务费 = 49,800 × 0.02 = **996元**
3. 定金 = 100 × 10 = **1,000元**
4. 可用定金 = **308,072元**

---

## 🔍 价格逻辑说明

### WebSocket数据格式

```json
{
  "AU": {
    "Buy": 498.00,   // 回购价（用户卖出价）
    "Sell": 500.00,  // 销售价（用户买入价）
    "H": 502.00,
    "L": 497.00
  }
}
```

### 价格映射

| 后端字段 | 前端变量 | 用途 | 显示 |
|---------|---------|------|------|
| AU.Sell | buyPrice | 用户买入 | 黄金销售价 |
| AU.Buy | sellPrice | 用户卖出 | 黄金回购价 |

**关键点**:
- ✅ `Sell` 是销售价（用户买入时支付的价格）
- ✅ `Buy` 是回购价（用户卖出时收到的价格）

---

## 📝 配置项说明

### 前端默认值

```javascript
config.value = {
  deposit_rate: 0.1,        // 定金率（已弃用）
  min_order_amount: 100,     // 最小订单克重
  service_fee_rate: 0.02,    // 服务费率（2%）
  deposit_per_gram: 10       // 每克定金（10元）
}
```

### 后台配置

**需要在超级管理员后台添加**:

| 配置项 | key | value | 说明 |
|-------|-----|-------|------|
| 最小订单克重 | min_order_amount | 100 | 最低交易克重 |
| 服务费率 | service_fee_rate | 0.02 | 2% |
| 每克定金 | deposit_per_gram | 10 | 10元/克 |

---

## 🧪 测试用例

### 测试1: 价格显示

**操作**:
1. 打开交易页面
2. 查看买入Tab价格卡片
3. 切换到卖出Tab

**期望**:
- ✅ 买入Tab显示"黄金销售价" + 具体价格（如500.00）
- ✅ 卖出Tab显示"黄金回购价" + 具体价格（如498.00）
- ✅ 不再显示 `-.--`

### 测试2: 买入计算

**操作**:
1. 买入Tab
2. 输入克重: 100
3. 查看费用明细

**期望**:
- ✅ 预估金额 = 100 × 销售价
- ✅ 服务费 = 预估金额 × 2%
- ✅ 定金 = 100 × 10 = 1,000
- ✅ 可用定金显示账户余额

### 测试3: 卖出计算

**操作**:
1. 卖出Tab
2. 输入克重: 100
3. 查看费用明细

**期望**:
- ✅ 预估金额 = 100 × 回购价
- ✅ 服务费 = 预估金额 × 2%
- ✅ 定金 = 100 × 10 = 1,000

### 测试4: 配置读取

**操作**:
1. 超级管理员后台修改服务费率为3%
2. 刷新交易页面
3. 输入克重100

**期望**:
- ✅ 服务费按3%计算

---

## 📝 修改的文件

1. ✅ `frontend/src/stores/quote.js` - 区分买入价和卖出价
2. ✅ `frontend/src/pages/Trade.vue` - 修正计算逻辑

---

## ✅ 全部修复完成

**修复项**:
1. ✅ 移除白银Tab
2. ✅ 修复价格显示（区分销售价和回购价）
3. ✅ 预估金额 = 价格 × 克重
4. ✅ 服务费从后台配置读取
5. ✅ 定金 = 克重 × 10
6. ✅ 可用定金显示账户余额
7. ✅ 订单提交使用正确价格

**现在刷新浏览器测试交易页面！**
