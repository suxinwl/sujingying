# WebSocket行情数据说明

## 数据源

**上海黄金交易所实时行情**

- **WebSocket地址**: `wss://push143.jtd9999.vip/ws`
- **数据类型**: Au9999黄金价格（元/克）
- **更新频率**: 实时推送
- **数据权威性**: 上海黄金交易所官方数据

---

## 系统架构

```
上海黄金交易所
    ↓ (WebSocket)
QuoteProxyHub (行情代理中心)
    ├─ 价格提取和缓存
    ├─ 广播给前端客户端
    └─ 提供给风控系统
        ↓
RiskScheduler (风控调度器)
    ↓
订单风险计算
```

---

## 价格提取

### WebSocket消息格式

来自上海黄金交易所的WebSocket消息包含Au9999的实时价格数据。

**示例消息格式**:
```json
{
  "data": {
    "au9999": {
      "currentPrice": "500.12",
      "openPrice": "499.80",
      "highPrice": "501.00",
      "lowPrice": "498.50",
      "updateTime": "2025-11-18 15:30:00"
    }
  }
}
```

### 价格缓存

- **缓存位置**: QuoteProxyHub内存
- **更新方式**: WebSocket实时推送
- **有效期**: 5分钟
- **失效处理**: 自动使用模拟价格fallback

---

## 服务说明

### QuoteProxyHub (行情代理)

**文件**: `internal/websocket/quote_proxy.go`

**功能**:
1. 连接上海黄金交易所WebSocket
2. 接收实时行情数据
3. 提取Au9999价格并缓存
4. 广播给前端客户端
5. 提供价格查询接口

**关键方法**:
```go
// 获取最新价格
func (h *QuoteProxyHub) GetLatestPrice() (float64, time.Time, bool)

// 从消息中提取价格
func (h *QuoteProxyHub) extractPrice(message []byte)
```

### QuoteService (行情服务)

**文件**: `internal/service/quote_service.go`

**功能**:
1. 封装价格查询逻辑
2. 提供统一接口给风控系统
3. 处理价格失效情况

**关键方法**:
```go
// 获取当前价格
func (s *QuoteService) GetCurrentPrice() (float64, error)

// 模拟价格（fallback）
func (s *QuoteService) SimulatePrice() float64

// 价格信息
func (s *QuoteService) GetPriceInfo() map[string]interface{}
```

---

## 容错机制

### 价格失效检测

价格被认为失效的条件：
- WebSocket未连接
- 超过5分钟未更新
- 价格为0或负数

### Fallback策略

```
WebSocket价格
    ↓ (失效)
模拟价格 (500±5元/克)
```

模拟价格特点：
- 基础价格: 500元/克
- 随机波动: ±5元
- 基于时间戳生成，有规律变化

---

## 前端接入

### WebSocket连接

**端点**: `ws://localhost:8080/ws/quote`

**示例**:
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/quote');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('实时行情:', data);
  // 更新UI显示价格
};
```

### 消息格式

前端接收到的消息与上海黄金交易所推送的原始消息一致，包含完整的行情数据。

---

## 风控系统集成

### 价格获取流程

```go
// 创建风控调度器时传入quoteHub
riskScheduler := scheduler.NewRiskScheduler(app, 60, quoteHub)

// 风控检查时获取价格
price := riskScheduler.getCurrentMarketPrice()

// 内部调用
price, err := quoteService.GetCurrentPrice()
```

### 价格更新频率

- **WebSocket推送**: 实时（上游数据变化时）
- **风控检查**: 每60秒
- **价格缓存**: 无需额外更新，WebSocket自动推送

---

## 监控建议

### 关键指标

1. **WebSocket连接状态**
   - 连接是否活跃
   - 重连次数
   - 连接持续时间

2. **价格更新状态**
   - 最后更新时间
   - 价格有效性
   - Fallback触发次数

3. **客户端连接数**
   - 当前连接数
   - 连接/断开频率

### 日志监控

关键日志：
```
[QuoteProxy] ✅ 上游连接成功
[QuoteProxy] 价格更新: Au9999 = 500.12 元/克
[Quote] WebSocket价格无效，使用模拟价格
[Quote] 使用模拟价格: 495.23 元/克
```

---

## 优势

### 相比外部API

| 特性 | WebSocket方案 | 外部API方案 |
|------|--------------|-------------|
| **实时性** | ✅ 实时推送 | ❌ 需要轮询 |
| **成本** | ✅ 免费 | ⚠️ 可能收费 |
| **延迟** | ✅ <1秒 | ⚠️ 数秒 |
| **API限制** | ✅ 无限制 | ❌ 有限额 |
| **数据源** | ✅ 官方直连 | ⚠️ 二手数据 |
| **配置** | ✅ 无需配置 | ❌ 需要密钥 |

---

## 故障排查

### WebSocket连接失败

**症状**: 日志显示连接失败，不断重试

**解决**:
1. 检查网络连接
2. 验证WebSocket URL是否正确
3. 检查防火墙设置

### 价格不更新

**症状**: 价格长时间不变化

**解决**:
1. 检查WebSocket连接状态
2. 查看日志是否有消息接收
3. 验证价格提取逻辑

### 频繁使用模拟价格

**症状**: 日志频繁显示"使用模拟价格"

**解决**:
1. 检查WebSocket连接
2. 验证消息格式是否变化
3. 调整价格有效期阈值

---

## 未来优化

### 可能的改进

1. **多数据源冗余**
   - 同时连接多个WebSocket
   - 主备切换机制

2. **价格历史记录**
   - 存储价格变化历史
   - 支持K线图展示

3. **异常检测**
   - 价格异常波动告警
   - 数据质量监控

4. **性能优化**
   - 消息批处理
   - 选择性广播

---

**文档版本**: 1.0  
**更新时间**: 2025-11-18  
**维护者**: 速金盈技术团队
