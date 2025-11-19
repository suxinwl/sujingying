# 价格显示调试指南

**问题**: 销售价和回购价依然无显示  
**数据格式**: `"AU":{"C":"AU","Sell":918.99,"Buy":916.99,"Gap":"-0.63","H":925.72,"L":916.29,"is_show":1}`

---

## 🔍 调试步骤

### 步骤1: 刷新浏览器并打开控制台

1. 按 `Ctrl + Shift + R` 强制刷新浏览器
2. 按 `F12` 打开开发者工具
3. 切换到 **Console** 标签

---

### 步骤2: 查看控制台日志

**期望看到的日志顺序**:

```
1. 正在连接行情WebSocket: ws://localhost:8080/ws/quote
2. 行情WebSocket已连接
3. 📨 收到WebSocket原始数据: {"AU":{"C":"AU","Sell":918.99,"Buy":916.99,...}}
4. 📊 解析后的数据对象: {AU: {...}}
5. 💰 AU黄金数据详情: {代码: "AU", 销售价_Sell: 918.99, 回购价_Buy: 916.99, ...}
6. 🔄 开始更新行情数据...
7. 当前 buyPrice: 0 sellPrice: 0
8. ✓ 找到AU数据，Sell: 918.99 Buy: 916.99
9. 解析后 newBuyPrice: 918.99 newSellPrice: 916.99
10. ✅ 价格有效，准备更新 state...
11. ✅ State已更新 - buyPrice: 918.99 sellPrice: 916.99
12. ✅ [AU] 销售价: 918.99 | 回购价: 916.99 | 涨跌: 0.00 (0.00%)
```

---

### 步骤3: 查看页面显示

**页面上应该显示**:

**买入Tab**:
```
┌────────────────────────────┐
│        918.99              │
│   黄金销售价(元/克)         │
│ 调试: buyPrice=918.99 |    │
│ sellPrice=916.99 | 连接=true│
└────────────────────────────┘
```

**卖出Tab**:
```
┌────────────────────────────┐
│        916.99              │
│   黄金回购价(元/克)         │
│ 调试: buyPrice=918.99 |    │
│ sellPrice=916.99 | 连接=true│
└────────────────────────────┘
```

---

## 🐛 常见问题诊断

### 问题A: 控制台没有任何日志

**可能原因**:
- WebSocket未连接
- 后端服务未启动

**检查**:
```javascript
// 在控制台输入
import { useQuoteStore } from './stores/quote'
const store = useQuoteStore()
console.log('WebSocket状态:', store.isConnected)
console.log('WebSocket对象:', store.ws)
```

**解决**:
1. 确认后端 `ws://localhost:8080/ws/quote` 正常运行
2. 检查是否有防火墙阻止WebSocket连接

---

### 问题B: 有日志但价格仍为 `-.--`

**检查日志中是否有**:
```
✅ State已更新 - buyPrice: 918.99 sellPrice: 916.99
```

**如果有这条日志**:
- Store已正确更新
- 问题在Vue响应式系统

**在控制台输入**:
```javascript
const quoteStore = useQuoteStore()
console.log('buyPrice:', quoteStore.buyPrice)
console.log('sellPrice:', quoteStore.sellPrice)
console.log('buyPriceDisplay:', quoteStore.buyPriceDisplay)
console.log('sellPriceDisplay:', quoteStore.sellPriceDisplay)
```

---

### 问题C: 日志显示找不到AU数据

**日志显示**:
```
⚠️ 数据中没有AU字段!
```

**原因**: WebSocket推送的数据格式不正确

**检查**:
在控制台查看 `📨 收到WebSocket原始数据` 的内容

**可能的数据格式问题**:
1. 数据被嵌套在其他对象中
2. 数据是数组而不是对象
3. 数据需要额外的解析

---

### 问题D: parseFloat失败

**日志显示**:
```
✓ 找到AU数据，Sell: 918.99 Buy: 916.99
解析后 newBuyPrice: 0 newSellPrice: 0
❌ newBuyPrice <= 0，价格更新失败！
```

**原因**: `parseFloat(data.AU.Sell)` 返回 0 或 NaN

**检查数据类型**:
```javascript
console.log('Sell类型:', typeof data.AU.Sell)
console.log('Sell值:', data.AU.Sell)
console.log('parseFloat结果:', parseFloat(data.AU.Sell))
```

---

## 💡 快速诊断命令

**复制以下代码到控制台执行**:

```javascript
// 获取store实例
const { useQuoteStore } = await import('/src/stores/quote.js')
const quoteStore = useQuoteStore()

// 诊断信息
console.log('=== 价格显示诊断 ===')
console.log('1. WebSocket连接状态:', quoteStore.isConnected)
console.log('2. 当前买入价(buyPrice):', quoteStore.buyPrice)
console.log('3. 当前卖出价(sellPrice):', quoteStore.sellPrice)
console.log('4. 买入价显示(buyPriceDisplay):', quoteStore.buyPriceDisplay)
console.log('5. 卖出价显示(sellPriceDisplay):', quoteStore.sellPriceDisplay)
console.log('6. 原始行情数据:', quoteStore.quoteData)
console.log('7. WebSocket对象:', quoteStore.ws)
console.log('8. 涨跌:', quoteStore.priceChange)
console.log('===================')

// 手动触发更新测试
if (quoteStore.quoteData && quoteStore.quoteData.AU) {
  console.log('✅ 已有AU数据，尝试手动更新...')
  quoteStore.updateQuote(quoteStore.quoteData)
} else {
  console.log('❌ 没有AU数据')
}
```

---

## 🔧 手动测试数据更新

**如果WebSocket有问题，可以手动注入测试数据**:

```javascript
const { useQuoteStore } = await import('/src/stores/quote.js')
const quoteStore = useQuoteStore()

// 手动注入测试数据
const testData = {
  "AU": {
    "C": "AU",
    "Sell": 918.99,
    "Buy": 916.99,
    "Gap": "-0.63",
    "H": 925.72,
    "L": 916.29,
    "is_show": 1
  }
}

console.log('手动更新测试数据...')
quoteStore.updateQuote(testData)

console.log('更新后的值:')
console.log('buyPrice:', quoteStore.buyPrice)
console.log('sellPrice:', quoteStore.sellPrice)
console.log('buyPriceDisplay:', quoteStore.buyPriceDisplay)
console.log('sellPriceDisplay:', quoteStore.sellPriceDisplay)
```

---

## 📋 检查清单

请按顺序检查：

- [ ] 后端WebSocket服务是否正常运行
- [ ] 浏览器控制台是否显示"行情WebSocket已连接"
- [ ] 控制台是否收到WebSocket数据
- [ ] 数据中是否包含AU字段
- [ ] AU.Sell 和 AU.Buy 是否为数字
- [ ] 控制台是否显示"State已更新"
- [ ] 页面调试信息显示的值是否正确
- [ ] buyPriceDisplay 和 sellPriceDisplay 是否正确

---

## 🎯 预期结果

**正确的完整流程**:

1. ✅ WebSocket连接成功
2. ✅ 每秒收到数据推送
3. ✅ 正确解析AU数据
4. ✅ Store的buyPrice和sellPrice正确更新
5. ✅ Getter返回正确的格式化字符串
6. ✅ 页面显示正确的价格

**最终显示**:
- 买入Tab显示: **918.99** (黄金销售价)
- 卖出Tab显示: **916.99** (黄金回购价)
- 调试信息显示: buyPrice=918.99 | sellPrice=916.99 | 连接=true

---

## 📞 如果问题仍未解决

请提供以下信息：

1. 控制台完整日志截图
2. 页面显示的调试信息
3. Network标签中WebSocket连接状态
4. 执行快速诊断命令的输出结果

---

**现在请刷新浏览器，打开控制台，查看详细的调试日志！**
