# 支付密码集成指南

## 概述

所有涉及资金和关键操作的接口都必须验证支付密码。本文档说明如何在API中集成支付密码验证。

## 需要支付密码验证的操作

- ✅ 充值提交
- ✅ 提现申请  
- ✅ 下单交易（锁价买料/卖料）
- ✅ 现金结算
- ✅ 修改银行卡信息

## 集成步骤

### 1. 在请求结构中添加支付密码字段

```go
type CreateOrderRequest struct {
    // ... 其他字段
    PayPassword string `json:"pay_password" binding:"required"`
}
```

### 2. 在处理器中验证支付密码

```go
func CreateOrder(c *gin.Context, ctx *appctx.AppContext) {
    var req CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }
    
    userID := c.GetUint("user_id")
    
    // 验证支付密码
    paypassSvc := service.NewPayPassService(ctx)
    if err := paypassSvc.RequirePayPassword(userID, req.PayPassword); err != nil {
        c.JSON(401, gin.H{"error": err.Error()})
        return
    }
    
    // 继续业务逻辑...
}
```

### 3. 前端调用示例

```javascript
// 下单时提示用户输入支付密码
const createOrder = async (orderData, payPassword) => {
  const response = await axios.post('/api/v1/orders', {
    ...orderData,
    pay_password: payPassword
  }, {
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  })
  return response.data
}
```

## 错误处理

支付密码验证可能返回以下错误：

- `支付密码不能为空`：前端未传递支付密码
- `请先设置支付密码`：用户尚未设置支付密码
- `支付密码错误`：支付密码不正确
- `用户不存在`：用户ID无效

## 注意事项

1. 支付密码必须是6位数字
2. 支付密码独立加密存储，与登录密码分离
3. 所有关键操作API都应在受保护路由下（需要 JWT 认证）
4. 支付密码验证应在业务逻辑之前进行
5. 验证失败应返回 401 Unauthorized

## 已实现的接口

- `POST /api/v1/user/paypass/set` - 设置支付密码
- `POST /api/v1/user/paypass/verify` - 验证支付密码

## 待集成的模块

以下模块在实现时需要集成支付密码验证：

- [ ] 订单模块 - 创建订单API
- [ ] 订单模块 - 现金结算API
- [ ] 资金模块 - 充值提交API
- [ ] 资金模块 - 提现申请API
- [ ] 银行卡模块 - 新增/编辑/删除银行卡API
