# 🧪 快速测试指南 - 价格显示

**问题**: 销售价和回购价不显示  
**数据**: `{"AU":{"Sell":918.99,"Buy":916.99,...}}`

---

## ✅ 最简单的测试方法

### 方法1: 使用页面上的测试按钮（推荐）

1. 访问交易页面: `http://localhost:5173/trade`
2. 在价格卡片下方，点击 **"手动注入测试数据"** 按钮
3. 打开控制台查看日志
4. 观察价格是否变为 918.99 和 916.99

**预期结果**:
- 页面弹出提示: `测试完成: 销售价=918.99, 回购价=916.99`
- 价格卡片显示: **918.99** 或 **916.99**
- 调试信息显示: `buyPrice=918.99 | sellPrice=916.99`

---

### 方法2: 控制台快速测试

在交易页面按 `F12` 打开控制台，复制执行：

```javascript
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

// 获取Vue实例和store
const app = document.querySelector('#app').__vueParentComponent
const quoteStore = app.appContext.config.globalProperties.$pinia._s.get('quote')

// 更新数据
console.log('📝 注入数据:', testData)
quoteStore.updateQuote(testData)

// 查看结果
setTimeout(() => {
  console.log('✅ buyPrice:', quoteStore.buyPrice)
  console.log('✅ sellPrice:', quoteStore.sellPrice)
  console.log('✅ buyPriceDisplay:', quoteStore.buyPriceDisplay)
  console.log('✅ sellPriceDisplay:', quoteStore.sellPriceDisplay)
}, 200)
```

---

## 🔍 调试信息说明

### 页面显示的调试信息

```
调试: buyPrice=918.99 | sellPrice=916.99 | 连接=true
```

**字段含义**:
- `buyPrice`: Store中的销售价（原始数值）
- `sellPrice`: Store中的回购价（原始数值）  
- `连接`: WebSocket连接状态

**正常值**:
- buyPrice > 0（如918.99）
- sellPrice > 0（如916.99）
- 连接=true

**异常值**:
- buyPrice=0 → 未收到数据或解析失败
- sellPrice=0 → 未收到数据或解析失败
- 连接=false → WebSocket未连接

---

## 📊 控制台日志解读

### 成功的日志顺序

```
1. 正在连接行情WebSocket: ws://localhost:8080/ws/quote
2. 行情WebSocket已连接
3. 📨 收到WebSocket原始数据: {"AU":{...}}
4. 📊 解析后的数据对象: {AU: {...}}
5. 💰 AU黄金数据详情: {销售价_Sell: 918.99, 回购价_Buy: 916.99, ...}
6. 🔄 开始更新行情数据...
7. 当前 buyPrice: 0 sellPrice: 0
8. ✓ 找到AU数据，Sell: 918.99 Buy: 916.99
9. 解析后 newBuyPrice: 918.99 newSellPrice: 916.99
10. ✅ 价格有效，准备更新 state...
11. ✅ State已更新 - buyPrice: 918.99 sellPrice: 916.99
12. ✅ [AU] 销售价: 918.99 | 回购价: 916.99 | 涨跌: 0.00
```

### 关键检查点

| 日志 | 说明 | 如果缺失 |
|------|------|----------|
| 行情WebSocket已连接 | WebSocket成功连接 | 检查后端是否启动 |
| 收到WebSocket原始数据 | 接收到推送数据 | 检查后端是否推送 |
| ✓ 找到AU数据 | 成功识别AU字段 | 检查数据格式 |
| State已更新 | Store成功更新 | 检查Pinia是否正常 |

---

## 🐛 常见问题

### Q1: 点击测试按钮后价格仍为 `-.--`

**检查步骤**:
1. 查看控制台是否有"State已更新"日志
2. 查看调试信息中buyPrice是否有值
3. 如果有值但不显示，可能是getter问题

**解决**:
```javascript
// 在控制台检查getter
const quoteStore = app.appContext.config.globalProperties.$pinia._s.get('quote')
console.log('buyPrice:', quoteStore.buyPrice)
console.log('buyPriceDisplay:', quoteStore.buyPriceDisplay)
```

### Q2: 调试信息显示 buyPrice=0

**原因**: updateQuote函数未正确解析数据

**检查**:
- 控制台是否有"❌ newBuyPrice <= 0"
- 数据格式是否正确
- parseFloat是否成功

### Q3: 调试信息显示 连接=false

**原因**: WebSocket未连接

**解决**:
1. 确认后端服务运行: `http://localhost:8080/ws/quote`
2. 检查WebSocket配置: `frontend/src/config/websocket.js`
3. 查看控制台是否有连接错误

### Q4: 手动注入成功，但WebSocket数据不显示

**原因**: WebSocket数据格式与测试数据不同

**检查**:
- 查看控制台"收到WebSocket原始数据"的内容
- 对比数据结构是否一致
- 是否需要额外的数据处理

---

## 🎯 问题定位流程

```
点击测试按钮
    ↓
控制台有"State已更新"？
    ├─ 是 → 调试信息buyPrice有值？
    │         ├─ 是 → 页面显示 -.-- ？
    │         │        ├─ 是 → getter或Vue响应式问题
    │         │        └─ 否 → 正常！
    │         └─ 否 → Vue响应式更新失败
    └─ 否 → updateQuote函数问题
              └─ 查看是否有"✓ 找到AU数据"
                    ├─ 有 → parseFloat失败
                    └─ 无 → 数据格式不匹配
```

---

## 📝 报告问题时提供

如果测试后仍有问题，请提供：

1. **点击测试按钮后的控制台完整日志**
2. **页面调试信息显示的值** (`buyPrice=? | sellPrice=? | 连接=?`)
3. **是否看到弹出提示**
4. **截图**

---

## ✅ 成功标志

测试成功后，您应该看到：

- ✅ 弹出提示: `测试完成: 销售价=918.99, 回购价=916.99`
- ✅ 价格卡片显示 **918.99**（买入）或 **916.99**（卖出）
- ✅ 调试信息: `buyPrice=918.99 | sellPrice=916.99 | 连接=true`
- ✅ 控制台日志完整无报错

---

**现在刷新浏览器，访问 http://localhost:5173/trade 点击测试按钮！**
