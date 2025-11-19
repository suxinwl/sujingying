# 前后端API对齐检查报告

**检查时间**: 2025-11-18  
**状态**: 发现多处不对齐问题需要修复

---

## 🚨 严重问题（需立即修复）

### 1. 管理员用户管理路由错误

**问题**: 前端所有 `/admin/users` 路由不存在

**前端配置** (错误):
```javascript
ADMIN_USERS: '/api/v1/admin/users',
ADMIN_USER_DETAIL: '/api/v1/admin/users/:id',
ADMIN_USER_APPROVE: '/api/v1/admin/users/:id/approve',
```

**后端实际路由**:
```go
GET  /api/v1/users
GET  /api/v1/users/:id
POST /api/v1/users/:id/approve
```

**修复方案**: 修改 `frontend/src/config/api.js`
```javascript
ADMIN_USERS: '/api/v1/users',
ADMIN_USER_DETAIL: '/api/v1/users/:id',
ADMIN_USER_APPROVE: '/api/v1/users/:id/approve',
```

---

### 2. 资金流水路由错误

**前端配置** (错误):
```javascript
FUND_FLOW: '/api/v1/fund-flow',
```

**后端实际路由**:
```go
GET /api/v1/fund-logs
```

**修复方案**:
```javascript
FUND_FLOW: '/api/v1/fund-logs',
```

---

### 3. 订单创建路由不一致

**前端配置** (错误):
```javascript
ORDER_BUY: '/api/v1/orders/buy',
ORDER_SELL: '/api/v1/orders/sell',
```

**后端实际**: 统一使用 `POST /api/v1/orders`，通过 `direction` 字段区分

**修复方案**:
```javascript
ORDER_CREATE: '/api/v1/orders',
// 删除 ORDER_BUY 和 ORDER_SELL
```

---

## ⚠️ 缺失的后端接口

### 1. 用户信息管理

- ❌ `PUT /api/v1/user/profile` - 更新用户信息
- ❌ `POST /api/v1/user/password` - 修改密码

**建议**: 在 `auth.go` 中添加

### 2. 持仓管理

- ❌ `GET /api/v1/positions` - 持仓列表
- ❌ `GET /api/v1/positions/:id` - 持仓详情

**建议**: 使用订单接口筛选或新增 positions 路由

---

## ✅ 已对齐的接口

### 认证 (100%)
- ✅ POST /auth/login
- ✅ POST /auth/register
- ✅ POST /auth/refresh
- ✅ POST /auth/logout
- ✅ GET /user/profile

### 银行卡 (100%)
- ✅ GET /bank-cards
- ✅ POST /bank-cards
- ✅ DELETE /bank-cards/:id

### 配置 (100%)
- ✅ GET /configs
- ✅ POST /configs/batch

### 充值提现 (100%)
- ✅ GET /deposits
- ✅ POST /deposits
- ✅ GET /deposits/pending
- ✅ POST /deposits/:id/review
- ✅ GET /withdraws
- ✅ POST /withdraws
- ✅ GET /withdraws/pending
- ✅ POST /withdraws/:id/review

---

## 📋 完整修复清单

创建文件: `frontend/src/config/api-fixes.txt`

需要修改 `frontend/src/config/api.js`:

1. 管理员相关（3处）
2. 资金流水（1处）
3. 订单创建（2处）

总计需要修复: **6处API配置错误**

