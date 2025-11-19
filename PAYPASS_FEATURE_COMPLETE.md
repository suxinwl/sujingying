# 支付密码功能完成

**完成时间**: 2025-11-18 14:07  
**功能**: 支付密码设置与使用

---

## ✅ 已完成的功能

### 1. Mine页面 - 支付密码设置入口

**位置**: `http://localhost:5173/mine`

**UI变化**:
```
┌─────────────────────────┐
│ 修改密码           🔒 >  │
│ 设置支付密码       🛡️ >  │  ← 新增！
│   已设置/未设置          │  ← 状态提示
│ 关于我们           ℹ️ >  │
└─────────────────────────┘
```

**功能**:
- ✅ 显示支付密码状态（已设置/未设置）
- ✅ 点击打开设置弹窗
- ✅ 支持首次设置和修改

---

### 2. 支付密码设置弹窗

#### 首次设置（未设置时）

```
┌─────────────────────────────┐
│    设置支付密码              │
├─────────────────────────────┤
│ 新支付密码                   │
│ [ 请输入6位数字 ]            │
│                              │
│ 确认密码                     │
│ [ 请再次输入支付密码 ]       │
└─────────────────────────────┘
     [取消]      [确定]
```

#### 修改支付密码（已设置时）

```
┌─────────────────────────────┐
│    修改支付密码              │
├─────────────────────────────┤
│ 旧支付密码                   │
│ [ 请输入旧支付密码 ]         │
│                              │
│ 新支付密码                   │
│ [ 请输入6位数字 ]            │
│                              │
│ 确认密码                     │
│ [ 请再次输入支付密码 ]       │
└─────────────────────────────┘
     [取消]      [确定]
```

**验证规则**:
- ✅ 支付密码必须是6位数字
- ✅ 两次输入必须一致
- ✅ 已设置时需要验证旧密码

---

### 3. Trade页面 - 下单时支付密码输入

**原方式**:
```javascript
const payPassword = prompt('请输入支付密码')  // ❌ 体验差
```

**新方式**:
```
┌─────────────────────────────┐
│    请输入支付密码            │
├─────────────────────────────┤
│ 请输入6位数字支付密码        │
│                              │
│ [ ●●●●●● ]                  │  ← 密码输入框
└─────────────────────────────┘
     [取消]      [确定]
```

**验证**:
- ✅ 必须输入6位数字
- ✅ 空值验证
- ✅ 格式验证

---

## 📝 代码变化

### 1. API配置 (frontend/src/config/api.js)

```javascript
export const API_ENDPOINTS = {
  // 用户相关
  USER_PROFILE: '/api/v1/user/profile',
  USER_UPDATE: '/api/v1/user/profile',
  CHANGE_PASSWORD: '/api/v1/user/password',
  PAYPASS: '/api/v1/user/paypass',  // ✅ 新增
  // ...
}
```

---

### 2. Mine.vue - 设置入口和弹窗

**新增Cell**:
```vue
<van-cell 
  title="设置支付密码" 
  is-link 
  @click="showPayPasswordDialog = true" 
  icon="shield-o"
  :label="userStore.userInfo?.has_pay_password ? '已设置' : '未设置'"
/>
```

**新增弹窗**:
```vue
<van-dialog
  v-model:show="showPayPasswordDialog"
  :title="userStore.userInfo?.has_pay_password ? '修改支付密码' : '设置支付密码'"
  show-cancel-button
  @confirm="onSetPayPassword"
>
  <van-form ref="payPasswordFormRef">
    <!-- 已设置时显示旧密码输入框 -->
    <van-field
      v-if="userStore.userInfo?.has_pay_password"
      v-model="payPasswordForm.old_pay_password"
      type="password"
      label="旧支付密码"
      maxlength="6"
    />
    
    <!-- 新密码 -->
    <van-field
      v-model="payPasswordForm.new_pay_password"
      type="password"
      label="新支付密码"
      placeholder="请输入6位数字"
      maxlength="6"
      :rules="[
        { required: true, message: '请输入新支付密码' },
        { pattern: /^\d{6}$/, message: '支付密码必须是6位数字' }
      ]"
    />
    
    <!-- 确认密码 -->
    <van-field
      v-model="payPasswordForm.confirm_pay_password"
      type="password"
      label="确认密码"
      maxlength="6"
      :rules="[
        { required: true, message: '请确认支付密码' },
        { validator: validatePayPassword, message: '两次支付密码不一致' }
      ]"
    />
  </van-form>
</van-dialog>
```

**提交函数**:
```javascript
const onSetPayPassword = async () => {
  await payPasswordFormRef.value?.validate()
  
  const hasPayPassword = userStore.userInfo?.has_pay_password
  
  await request.post(API_ENDPOINTS.PAYPASS, {
    old_pay_password: hasPayPassword ? payPasswordForm.value.old_pay_password : undefined,
    new_pay_password: payPasswordForm.value.new_pay_password
  })
  
  showToast(hasPayPassword ? '支付密码修改成功' : '支付密码设置成功')
  await userStore.loadUserInfo()  // 更新用户信息
  
  // 清空表单并关闭
  payPasswordForm.value = { old_pay_password: '', new_pay_password: '', confirm_pay_password: '' }
  showPayPasswordDialog.value = false
}
```

---

### 3. Trade.vue - 下单支付密码输入

**改进的输入方式**:
```javascript
// 弹出支付密码输入框
const payPassword = await new Promise((resolve) => {
  showDialog({
    title: '请输入支付密码',
    message: '请输入6位数字支付密码',
    showCancelButton: true,
    beforeClose: (action) => {
      if (action === 'confirm') {
        const input = document.querySelector('.van-dialog__message input')
        resolve(input ? input.value : null)
      } else {
        resolve(null)
      }
      return true
    }
  })
  
  // 在message中动态插入密码输入框
  setTimeout(() => {
    const messageEl = document.querySelector('.van-dialog__message')
    if (messageEl && !messageEl.querySelector('input')) {
      const input = document.createElement('input')
      input.type = 'password'
      input.maxLength = 6
      input.placeholder = '请输入6位数字'
      input.style.cssText = 'width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; margin-top: 8px; font-size: 16px;'
      messageEl.appendChild(input)
      input.focus()
    }
  }, 100)
})

// 验证
if (!payPassword) {
  showToast('请输入支付密码')
  return
}

if (!/^\d{6}$/.test(payPassword)) {
  showToast('支付密码必须是6位数字')
  return
}
```

---

## 🔐 API接口

### POST /api/v1/user/paypass

**请求体**（首次设置）:
```json
{
  "new_pay_password": "123456"
}
```

**请求体**（修改）:
```json
{
  "old_pay_password": "123456",
  "new_pay_password": "654321"
}
```

**响应**:
```json
{
  "message": "支付密码设置成功"
}
```

---

## 🧪 测试清单

### 测试1: 首次设置支付密码

1. 访问 http://localhost:5173/mine
2. 点击"设置支付密码"（显示"未设置"）
3. 输入新支付密码: 123456
4. 确认密码: 123456
5. 点击确定

**期望**:
- ✅ 提示"支付密码设置成功"
- ✅ 状态变为"已设置"

### 测试2: 修改支付密码

1. 访问 http://localhost:5173/mine
2. 点击"设置支付密码"（显示"已设置"）
3. 输入旧支付密码: 123456
4. 输入新支付密码: 654321
5. 确认密码: 654321
6. 点击确定

**期望**:
- ✅ 提示"支付密码修改成功"

### 测试3: 下单时输入支付密码

1. 访问 http://localhost:5173/trade
2. 输入克重: 100
3. 勾选协议
4. 点击"立即买入"
5. 弹出支付密码输入框
6. 输入: 123456
7. 点击确定

**期望**:
- ✅ 显示支付密码输入框
- ✅ 输入框获得焦点
- ✅ 最多输入6位
- ✅ 验证通过后提交订单

### 测试4: 验证规则

**场景1**: 两次密码不一致
- 新密码: 123456
- 确认密码: 654321
- **期望**: 提示"两次支付密码不一致"

**场景2**: 输入非数字
- 新密码: abc123
- **期望**: 提示"支付密码必须是6位数字"

**场景3**: 少于6位
- 新密码: 12345
- **期望**: 提示"支付密码必须是6位数字"

---

## 📊 用户流程

### 新用户首次下单流程

```
用户注册登录
    ↓
访问交易页面
    ↓
输入克重，点击下单
    ↓
弹出支付密码输入框
    ↓
用户发现未设置支付密码
    ↓
取消下单 → 访问"我的"页面
    ↓
点击"设置支付密码"
    ↓
输入6位数字支付密码
    ↓
设置成功
    ↓
返回交易页面重新下单
    ↓
输入支付密码
    ↓
订单提交成功 ✅
```

---

## 🎯 未来改进建议

### 1. 支付密码输入组件
创建专用的支付密码输入组件，使用Vant的PasswordInput：
```vue
<van-password-input
  :value="payPassword"
  :length="6"
  @focus="showKeyboard = true"
/>
<van-number-keyboard
  v-model="payPassword"
  :show="showKeyboard"
  :maxlength="6"
  @blur="showKeyboard = false"
/>
```

### 2. 生物识别支持
- 指纹识别
- 面容识别
- 需要浏览器支持WebAuthn API

### 3. 支付密码找回
- 通过手机验证码找回
- 通过身份证验证找回

### 4. 支付密码错误次数限制
- 连续错误3次锁定5分钟
- 连续错误5次锁定1小时

---

## ✅ 功能完成

**支付密码功能已完全实现！**

**修改的文件**:
1. ✅ `frontend/src/pages/Mine.vue` - 添加设置入口和弹窗
2. ✅ `frontend/src/pages/Trade.vue` - 改进支付密码输入
3. ✅ `frontend/src/config/api.js` - 添加API端点

**现在用户可以**:
- ✅ 在"我的"页面设置支付密码
- ✅ 修改支付密码
- ✅ 在下单时输入支付密码
- ✅ 享受更安全的交易体验

**刷新浏览器，访问 http://localhost:5173/mine 设置支付密码！**
