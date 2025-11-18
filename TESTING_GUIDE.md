# 登录测试指南

## 🔍 当前问题

收到 **401 Unauthorized** 错误，可能的原因：

1. ✅ CORS配置正确
2. ✅ 后端API运行
3. ❓ 用户输入的账号/密码
4. ❓ 用户账号状态
5. ❓ 发送的数据格式

---

## 🧪 测试步骤

### 第一步：检查浏览器控制台

打开浏览器开发者工具 (F12)，查看Console标签：

1. **查找"登录数据:"日志**
   - 应该显示: `登录数据: {username: '...', password: '...'}`
   - 确认username和password都有值

2. **查看网络请求**
   - 切换到Network标签
   - 点击登录按钮
   - 找到 `auth/login` 请求
   - 点击查看 **Request Payload**

**预期看到**:
```json
{
  "username": "13800000001",
  "password": "123456"
}
```

---

### 第二步：测试正确的账号

#### 测试账号信息
```
手机号: 13800000001
密码: 123456
状态: active
```

#### 在登录页面输入
- **用户名框**: `13800000001`
- **密码框**: `123456`

⚠️ **注意**：
- 手机号必须完整输入（11位）
- 密码区分大小写
- 不要有空格

---

### 第三步：检查后端日志

查看后端终端输出，寻找相关错误：

```bash
# 可能的错误信息
❌ "用户名或密码错误"
❌ "您的账户还在审核中"
❌ "您的账户已被禁用"
❌ "账户已被锁定"
```

---

### 第四步：验证数据库用户状态

运行以下命令检查用户：

```bash
cd backend
go run scripts/activate_users.go
```

**应该看到**:
```
✅ Activated X users

Users in database:
  ID: 6, Phone: 13800000001, Role: customer, Status: active
```

如果Status不是`active`，用户将无法登录。

---

## 🔧 常见问题修复

### 问题1：用户状态是pending

**症状**: 401错误，后端日志"您的账户还在审核中"

**解决**:
```bash
cd backend
go run scripts/activate_users.go
```

### 问题2：密码错误

**症状**: 401错误，后端日志"用户名或密码错误"

**原因**: 
- 密码不对
- 用户不存在
- 输入了错误的手机号

**检查**:
1. 确认手机号是 `13800000001`
2. 确认密码是 `123456`
3. 检查是否有拼写错误

### 问题3：表单值为空

**症状**: 控制台显示 `{username: '', password: ''}`

**原因**: 表单没有正确绑定或Vant组件未加载

**解决**:
1. 刷新浏览器 (Ctrl+F5)
2. 清除浏览器缓存
3. 确认前端服务正常运行

### 问题4：后端未启动

**症状**: 
- Network Error
- ERR_CONNECTION_REFUSED

**解决**:
```bash
cd backend
go run cmd/server/main.go
```

---

## 📝 完整测试流程

### 1. 确认服务运行
```bash
# 检查后端 (应该看到进程)
netstat -ano | findstr :8080

# 检查前端 (应该看到进程)
netstat -ano | findstr :5175
```

### 2. 激活测试用户
```bash
cd backend
go run scripts/activate_users.go
```

### 3. 测试API
```bash
# 使用PowerShell测试
cd E:\AI\SuxinZK\code
.\test_login.ps1
```

**预期输出**:
```
✅ Login Success!
Token: eyJhbGci...
```

### 4. 浏览器测试
1. 打开 http://localhost:5175
2. 打开浏览器开发者工具 (F12)
3. 输入账号密码
4. 点击登录
5. 查看Console和Network标签

---

## 🐛 调试技巧

### 查看完整错误信息

在浏览器Console中，展开错误对象：

```javascript
// 点击错误对象查看
AxiosError {
  response: {
    status: 401,
    data: {
      error: "具体的错误信息"  // ← 这里是关键
    }
  }
}
```

### 查看后端日志

后端会记录详细的失败原因：

```bash
[GIN] 2025/11/18 - 10:50:00 | 401 |   POST "/api/v1/auth/login"
# 上面这行后面会有具体的错误日志
```

### 手动测试API

使用curl或PowerShell测试：

```powershell
# PowerShell
$body = @{
    username = "13800000001"
    password = "123456"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
    -Method POST `
    -Headers @{"Content-Type" = "application/json"} `
    -Body $body
```

---

## ✅ 成功标志

登录成功时应该看到：

### 浏览器Console
```
登录数据: {username: '13800000001', password: '123456'}
✅ 跳转到首页
```

### Network标签
```
Status: 200 OK
Response:
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "user": {...}
}
```

### LocalStorage
```
access_token: "eyJ..."
refresh_token: "eyJ..."
```

### 页面行为
- ✅ 显示Toast "登录成功"
- ✅ 自动跳转到首页 (/)
- ✅ 看到实时金价
- ✅ 底部导航栏显示

---

## 📞 还是不行？

如果以上步骤都无法解决，请提供：

1. **浏览器Console完整日志** (截图)
2. **Network标签的请求详情** (截图)
3. **后端终端输出** (文本)
4. **用户数据库状态** (运行activate_users.go的输出)

---

**最后更新**: 2025-11-18 10:50
