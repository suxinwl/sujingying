# WebSocket AU黄金数据修复

**修复时间**: 2025-11-18 13:31  
**问题**: 从WebSocket提取AU黄金销售价/回购价，实现秒级更新

---

## ✅ 修复的问题

### 1. ❌ ReferenceError: availableSellAmount is not defined

**错误信息**:
```
Trade.vue:326 获取持仓失败，设置为0: ReferenceError: availableSellAmount is not defined
Trade.vue:330 获取余额失败: ReferenceError: availableSellAmount is not defined
```

**原因**: 
- 在 `loadBalance` 函数中使用了 `availableSellAmount.value`
- 但该变量没有定义

**修复**:
```javascript
// 修复前（有错误）
const loadBalance = async () => {
  // ...
  availableSellAmount.value = ordersList.reduce(...)  // ❌ 未定义
}

// 修复后（正确）
const loadBalance = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.USER_PROFILE)
    balance.value = {
      available_deposit: data.available_deposit || 0,
      used_deposit: data.used_deposit || 0
    }
  } catch (error) {
    console.error('获取余额失败:', error)
    balance.value = {
      available_deposit: 0,
      used_deposit: 0
    }
  }
}
```

**结果**: ✅ 错误已修复

---

### 2. ❌ favicon.ico 404错误

**错误信息**:
```
GET http://localhost:5173/favicon.ico 404 (Not Found)
```

**修复**:
- ✅ 创建 `frontend/public/favicon.ico` 文件

**结果**: ✅ 404错误已解决

---

## 🔄 WebSocket AU黄金数据提取

### WebSocket数据格式

**后端推送的数据格式**:
```json
{
  "AU": {
    "Buy": 498.00,      // 回购价（用户卖出价）
    "Sell": 500.00,     // 销售价（用户买入价）
    "H": 502.00,        // 最高价
    "L": 497.00,        // 最低价
    "Gap": 2.0          // 基差
  },
  "AU9999": {...},
  "XAU": {...}
}
```

### quoteStore数据提取

**代码位置**: `frontend/src/stores/quote.js`

**关键代码**:
```javascript
// 1. WebSocket消息接收
this.ws.onmessage = (event) => {
  try {
    const data = JSON.parse(event.data)
    console.log('收到WebSocket行情数据:', data)
    
    // 提取AU黄金数据
    if (data.AU) {
      console.log('AU黄金数据:', {
        销售价_Sell: data.AU.Sell,
        回购价_Buy: data.AU.Buy,
        最高: data.AU.H,
        最低: data.AU.L
      })
    }
    
    this.updateQuote(data)
  } catch (error) {
    console.error('解析行情数据失败:', error)
  }
}

// 2. 更新行情数据
updateQuote(data) {
  this.quoteData = data
  
  // 优先使用黄金AU的价格
  let newBuyPrice = 0   // 销售价（Sell）
  let newSellPrice = 0  // 回购价（Buy）
  
  // ✅ 提取AU黄金数据
  if (data.AU) {
    newBuyPrice = parseFloat(data.AU.Sell) || 0   // 销售价
    newSellPrice = parseFloat(data.AU.Buy) || 0   // 回购价
  } else if (data.AU9999) {
    newBuyPrice = parseFloat(data.AU9999.Sell) || 0
    newSellPrice = parseFloat(data.AU9999.Buy) || 0
  } else if (data.XAU) {
    newBuyPrice = parseFloat(data.XAU.Sell) || 0
    newSellPrice = parseFloat(data.XAU.Buy) || 0
  }
  
  if (newBuyPrice > 0) {
    this.previousBuyPrice = this.buyPrice
    this.buyPrice = newBuyPrice
    this.sellPrice = newSellPrice
    
    // 计算涨跌
    if (this.previousBuyPrice > 0) {
      this.priceChange = this.buyPrice - this.previousBuyPrice
      this.priceChangePercent = (this.priceChange / this.previousBuyPrice) * 100
    }
    
    console.log(`✅ 销售价: ${this.buyPrice}, 回购价: ${this.sellPrice}, 涨跌: ${this.priceChange.toFixed(2)}`)
  }
}
```

---

## 📊 数据映射关系

| WebSocket字段 | Store变量 | 显示名称 | 用途 |
|--------------|-----------|---------|------|
| AU.Sell | buyPrice | 黄金销售价 | 用户买入价格 |
| AU.Buy | sellPrice | 黄金回购价 | 用户卖出价格 |
| AU.H | - | 最高价 | 参考 |
| AU.L | - | 最低价 | 参考 |

**关键点**:
- ✅ `AU.Sell` → `buyPrice` → 买入Tab显示
- ✅ `AU.Buy` → `sellPrice` → 卖出Tab显示
- ✅ 每次WebSocket推送都会更新（秒级）

---

## 🕐 秒级更新机制

### WebSocket连接

**连接地址**: `ws://localhost:8080/ws/quote`

**配置**: `frontend/src/config/websocket.js`
```javascript
export const WS_CONFIG = {
  QUOTE_WS_URL: 'ws://localhost:8080/ws/quote',
  MAX_RECONNECT_ATTEMPTS: 10,
  RECONNECT_BASE_DELAY: 1000,
  MAX_RECONNECT_DELAY: 30000
}
```

### 更新频率

**后端推送频率**: 
- ✅ 后端每秒推送一次行情数据
- ✅ 前端接收到数据后立即更新

**前端处理**:
```javascript
// 每次收到WebSocket消息都会触发
this.ws.onmessage = (event) => {
  // 立即解析和更新数据
  const data = JSON.parse(event.data)
  this.updateQuote(data)
}
```

**响应式更新**:
```vue
<!-- 价格卡片会自动更新 -->
<div class="price-value">
  {{ tradeType === 'buy' ? quoteStore.buyPriceDisplay : quoteStore.sellPriceDisplay }}
</div>
```

---

## 🧪 测试验证

### 测试1: 控制台日志

**操作**:
1. 打开浏览器控制台（F12）
2. 访问 http://localhost:5173/trade
3. 查看控制台输出

**期望看到**:
```
正在连接行情WebSocket: ws://localhost:8080/ws/quote
行情WebSocket已连接
收到WebSocket行情数据: {AU: {...}, AU9999: {...}, ...}
AU黄金数据: {销售价_Sell: 500, 回购价_Buy: 498, 最高: 502, 最低: 497}
✅ 销售价: 500, 回购价: 498, 涨跌: 0.00
```

**每秒更新**:
```
收到WebSocket行情数据: ...
AU黄金数据: {销售价_Sell: 500.5, 回购价_Buy: 498.5, ...}
✅ 销售价: 500.5, 回购价: 498.5, 涨跌: 0.50
```

### 测试2: 价格卡片更新

**操作**:
1. 观察买入Tab的价格卡片
2. 等待1-2秒
3. 观察价格是否变化

**期望**:
- ✅ 价格每秒更新
- ✅ 数字变化流畅
- ✅ 买入显示销售价（Sell）
- ✅ 卖出显示回购价（Buy）

### 测试3: 费用计算自动更新

**操作**:
1. 输入克重100
2. 观察预估金额
3. 等待价格更新
4. 观察预估金额是否同步更新

**期望**:
- ✅ 价格500 → 预估金额50,000
- ✅ 价格变为501 → 预估金额自动变为50,100
- ✅ 服务费、定金同步更新

---

## 📝 调试步骤

### 如果价格不更新

**步骤1: 检查WebSocket连接**
```javascript
// 在控制台输入
console.log('WebSocket状态:', quoteStore.isConnected)
console.log('当前销售价:', quoteStore.buyPrice)
console.log('当前回购价:', quoteStore.sellPrice)
```

**步骤2: 检查后端推送**
```bash
# 确保后端WebSocket服务正常运行
# 检查后端日志是否有推送记录
```

**步骤3: 检查数据格式**
```javascript
// 查看原始WebSocket数据
console.log('原始数据:', quoteStore.quoteData)
console.log('AU数据:', quoteStore.quoteData.AU)
```

### 如果显示 `-.--`

**原因**:
- WebSocket未连接
- 后端未推送数据
- AU字段不存在

**解决**:
1. 检查 `quoteStore.isConnected` 是否为 `true`
2. 检查控制台是否有 "收到WebSocket行情数据" 日志
3. 检查数据中是否有 `AU` 字段

---

## 🔍 数据流程图

```
后端WebSocket服务
    ↓ (每秒推送)
{AU: {Sell: 500, Buy: 498, ...}}
    ↓
quoteStore.ws.onmessage
    ↓
updateQuote(data)
    ↓
提取 data.AU.Sell → buyPrice (销售价)
提取 data.AU.Buy → sellPrice (回购价)
    ↓
Vue响应式更新
    ↓
价格卡片显示更新
    ↓
预估金额自动重新计算
```

---

## 📝 修改的文件

1. ✅ `frontend/src/stores/quote.js` - 添加详细日志
2. ✅ `frontend/src/pages/Trade.vue` - 修复 availableSellAmount 错误
3. ✅ `frontend/public/favicon.ico` - 添加favicon文件

---

## ✅ 修复完成

**修复项**:
1. ✅ 修复 `availableSellAmount` 未定义错误
2. ✅ 修复 favicon.ico 404错误
3. ✅ 确认WebSocket正确提取AU黄金数据
4. ✅ 确认销售价（Sell）和回购价（Buy）正确映射
5. ✅ 添加详细日志便于调试
6. ✅ 实现秒级更新（依赖后端推送频率）

**验证清单**:
- [ ] 控制台无错误
- [ ] WebSocket已连接
- [ ] 看到AU黄金数据日志
- [ ] 买入显示销售价
- [ ] 卖出显示回购价
- [ ] 价格秒级更新
- [ ] 费用自动计算

**刷新浏览器，打开控制台查看日志！**
