# 功能测试清单

**测试日期**: 2025-11-18  
**测试账号**: 见下方

---

## 🧪 测试账号

### 超级管理员（测试所有功能）
```
用户名: 13900000000
密码: admin123
```

### 普通用户（测试客户功能）
```
用户名: 13800000001
密码: 123456
```

---

## ✅ 测试步骤

### 1. 资金页面测试

**URL**: http://localhost:5173/funds

#### 测试项目
- [ ] 页面正常加载，无报错
- [ ] 可用定金余额正常显示
- [ ] 资金流水列表正常显示
- [ ] 切换流水类型Tab（全部/充值/提现）正常
- [ ] 下拉刷新正常
- [ ] 上拉加载更多正常
- [ ] 银行卡列表正常显示
- [ ] 点击"充值"弹窗正常
- [ ] 点击"提现"弹窗正常

#### 预期结果
```javascript
// 控制台不应该出现
❌ "加载资金流水失败: TypeError: Cannot read properties of undefined (reading 'list')"

// 应该正常显示
✅ 资金流水列表
✅ 银行卡列表
```

---

### 2. 交易页面测试

**URL**: http://localhost:5173/trade

#### 测试项目
- [ ] 页面正常加载，无崩溃
- [ ] "可用定金"余额正常显示（不是NaN或undefined）
- [ ] 当前价格正常显示
- [ ] 输入买入价格和数量，自动计算总金额和所需定金
- [ ] 切换到"卖出"Tab正常
- [ ] 表单验证正常
- [ ] 可以正常提交订单（需要支付密码）

#### 预期结果
```javascript
// 控制台不应该出现
❌ "Uncaught TypeError: Cannot read properties of undefined (reading 'available_amount')"

// 应该正常显示
✅ 可用定金: ¥10,000.00
✅ 当前价格
✅ 计算的总金额
```

---

### 3. 订单页面测试

**URL**: http://localhost:5173/orders

#### 测试项目
- [ ] 页面正常加载
- [ ] 订单列表正常显示
- [ ] 切换Tab（全部/持仓中/已平仓）正常
- [ ] 订单详情显示完整
- [ ] 盈亏计算正常
- [ ] 下拉刷新正常

#### 预期结果
- ✅ 订单列表正常显示
- ✅ 状态筛选正常

---

### 4. 银行卡管理测试

**URL**: http://localhost:5173/bank-cards

#### 测试项目
- [ ] 页面正常加载
- [ ] 银行卡列表正常显示（卡号脱敏）
- [ ] 点击"添加银行卡"弹窗
- [ ] 填写表单（需要支付密码）
- [ ] 成功添加银行卡
- [ ] 删除银行卡功能正常

#### 预期结果
- ✅ 银行卡列表显示正常
- ✅ 添加功能正常（如已设置支付密码）
- ✅ 删除功能正常

---

### 5. 通知页面测试

**URL**: http://localhost:5173/notifications

#### 测试项目
- [ ] 页面正常加载
- [ ] 通知列表正常显示
- [ ] 未读通知高亮显示
- [ ] 点击单条通知标记已读
- [ ] 点击"全部已读"按钮正常
- [ ] 已读通知样式变化

#### 预期结果
- ✅ 通知列表显示
- ✅ 标记已读功能正常

---

### 6. 管理员用户管理测试 ⚠️ 重要

**URL**: http://localhost:5173/admin/users

**测试账号**: 使用管理员账号 `13900000000`

#### 测试项目
- [ ] 页面正常加载
- [ ] 用户列表正常显示
- [ ] 搜索功能正常
- [ ] 切换Tab（全部/待审核/已激活/已禁用）正常
- [ ] **点击"通过"按钮审核用户** ← 关键功能
- [ ] **点击"拒绝"按钮，输入原因** ← 关键功能
- [ ] 审核后用户状态更新
- [ ] 审核后有成功提示

#### 预期结果
```javascript
// API调用应该成功
POST /api/v1/users/:id/approve
Body: { action: "approve", note: "" }

// 响应
✅ { message: "用户 13800000001 审核通过" }

// 不应该出现
❌ 400 Bad Request
❌ "请求参数错误"
```

---

### 7. 管理员充值审核测试 ⚠️ 重要

**URL**: http://localhost:5173/admin/deposits

**前置条件**: 先用普通用户提交充值申请

#### 测试项目
- [ ] 页面正常加载
- [ ] 待审核充值列表显示
- [ ] 切换Tab（待审核/已通过/已拒绝）正常
- [ ] **点击"通过"按钮审核充值** ← 关键功能
- [ ] **点击"拒绝"按钮，输入原因** ← 关键功能
- [ ] 审核后状态更新
- [ ] 审核后余额变化（如果通过）

#### 预期结果
```javascript
// API调用
POST /api/v1/deposits/:id/review
Body: { action: "approve", note: "" }

// 不应该出现
❌ 400 Bad Request
❌ "请求参数错误"
```

---

### 8. 管理员提现审核测试 ⚠️ 重要

**URL**: http://localhost:5173/admin/withdraws

**前置条件**: 先用普通用户提交提现申请

#### 测试项目
- [ ] 页面正常加载
- [ ] 待审核提现列表显示
- [ ] 切换Tab（待审核/已通过/已拒绝）正常
- [ ] **点击"通过"按钮审核提现** ← 关键功能
- [ ] **点击"拒绝"按钮，输入原因** ← 关键功能
- [ ] 审核后状态更新

#### 预期结果
```javascript
// API调用
POST /api/v1/withdraws/:id/review
Body: { action: "approve", note: "" }

// 不应该出现
❌ 400 Bad Request
❌ "请求参数错误"
```

---

### 9. 管理员系统配置测试

**URL**: http://localhost:5173/admin/config

#### 测试项目
- [ ] 页面正常加载
- [ ] 配置项正常显示
- [ ] 修改配置值
- [ ] 点击"保存配置"按钮
- [ ] 保存成功提示

#### 预期结果
```javascript
// API调用
POST /api/v1/configs/batch
Body: { "min_order_amount": "100", ... }

// 响应
✅ { data: { updated: 10 }, message: "配置更新成功" }
```

---

### 10. 关于页面测试

**URL**: http://localhost:5173/about

#### 测试项目
- [ ] 页面正常加载
- [ ] 公司信息正常显示
- [ ] 联系方式正常显示

---

## 🐛 已修复的错误对照

| 页面 | 修复前错误 | 修复后状态 |
|------|-----------|-----------|
| `/funds` | `Cannot read properties of undefined (reading 'list')` | ✅ 正常显示 |
| `/trade` | `Cannot read properties of undefined (reading 'available_amount')` | ✅ 正常显示 |
| `/admin/users` | 审核请求400错误 | ✅ 审核成功 |
| `/admin/deposits` | 审核请求400错误 | ✅ 审核成功 |
| `/admin/withdraws` | 审核请求400错误 | ✅ 审核成功 |

---

## 📋 快速测试命令

### 测试后端健康状态
```powershell
curl http://localhost:8080/healthz
```

### 测试登录
```powershell
$body = @{
    phone = "13900000000"
    password = "admin123"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body $body -ContentType "application/json"
$response
```

### 测试用户信息
```powershell
$token = "YOUR_ACCESS_TOKEN"
$headers = @{
    "Authorization" = "Bearer $token"
}

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/user/profile" -Headers $headers
```

---

## ✅ 测试通过标准

### 基础功能（必须全部通过）
- ✅ 所有页面正常加载，无JavaScript错误
- ✅ 数据列表正常显示
- ✅ 下拉刷新和加载更多正常
- ✅ 表单提交正常
- ✅ 错误提示友好

### 管理员功能（核心功能）
- ✅ 用户审核功能正常
- ✅ 充值审核功能正常
- ✅ 提现审核功能正常
- ✅ 系统配置保存正常

### 用户体验
- ✅ 页面响应速度快
- ✅ 无明显卡顿
- ✅ 提示信息清晰
- ✅ 操作流程顺畅

---

## 🚨 如果发现问题

### 1. 检查浏览器控制台
- 按F12打开开发者工具
- 查看Console面板的错误信息
- 查看Network面板的请求响应

### 2. 检查后端日志
- 查看后端控制台输出
- 确认API请求是否到达
- 查看返回的错误信息

### 3. 清除缓存
```javascript
// 在浏览器控制台执行
localStorage.clear()
sessionStorage.clear()
location.reload()
```

### 4. 重启服务
```powershell
# 重启后端
cd backend
go run cmd/server/main.go

# 重启前端
cd frontend
npm run dev
```

---

**开始测试前请确保**:
1. ✅ 后端服务运行在 http://localhost:8080
2. ✅ 前端服务运行在 http://localhost:5173
3. ✅ 数据库文件存在且可访问
4. ✅ 浏览器已清除缓存

**测试完成后请反馈**:
- 哪些功能正常 ✅
- 哪些功能有问题 ❌
- 具体的错误信息
