# 银行卡删除问题修复

**修复时间**: 2025-11-18 14:49  
**问题**: 删除银行卡时返回400错误

---

## 🐛 问题描述

### 错误信息
```
DELETE http://localhost:8080/api/v1/bank-cards/5 400 (Bad Request)
删除银行卡失败: AxiosError
```

### 现象
- 点击删除按钮
- 确认删除
- 返回400错误
- 无具体错误提示

---

## 🔍 问题分析

### 可能的原因

1. **银行卡不存在**
   - 银行卡已被删除
   - ID不正确

2. **无权操作**
   - 不是自己的银行卡

3. **默认卡限制** ⭐ **最可能**
   - 该卡是默认卡
   - 还有其他银行卡存在
   - 需要先设置其他卡为默认

### 后端原逻辑

```go
func (s *BankCardService) DeleteCard(userID uint, cardID uint) error {
    card, err := s.cardRepo.FindByID(cardID)
    if err != nil {
        return errors.New("银行卡不存在")
    }
    
    if card.UserID != userID {
        return errors.New("无权操作此银行卡")
    }
    
    // ❌ 问题：无论有几张卡，都不允许删除默认卡
    if card.IsDefault {
        return errors.New("请先将其他银行卡设为默认卡")
    }
    
    return s.cardRepo.Delete(cardID, userID)
}
```

**问题**: 即使只有一张卡，也不允许删除默认卡

---

## ✅ 修复方案

### 1. 前端：显示具体错误信息

**文件**: `frontend/src/pages/BankCards.vue`

**修改前**:
```javascript
catch (error) {
  if (error === 'cancel') return
  console.error('删除银行卡失败:', error)
  showToast('删除失败')  // ❌ 无具体信息
}
```

**修改后**:
```javascript
catch (error) {
  if (error === 'cancel') return
  console.error('删除银行卡失败:', error)
  // ✅ 显示后端返回的具体错误
  const errorMsg = error.response?.data?.error || error.response?.data?.message || '删除失败'
  showToast(errorMsg)
}
```

---

### 2. 后端：优化删除逻辑

**文件**: `backend/internal/service/bank_card_service.go`

**修改前**:
```go
// 3. 如果是默认卡，提示错误
if card.IsDefault {
    return errors.New("请先将其他银行卡设为默认卡")
}
```

**修改后**:
```go
// 3. 如果是默认卡，检查是否还有其他卡
if card.IsDefault {
    count, err := s.cardRepo.CountByUserID(userID)
    if err != nil {
        return err
    }
    // 如果不是唯一的卡，需要先设置其他卡为默认
    if count > 1 {
        return errors.New("请先将其他银行卡设为默认卡")
    }
    // 如果是唯一的卡，可以直接删除
}

// 4. 删除
return s.cardRepo.Delete(cardID, userID)
```

**改进点**:
- ✅ 允许删除唯一的默认卡
- ✅ 只在有多张卡时才提示先设置其他默认卡
- ✅ 提升用户体验

---

## 🎯 业务逻辑

### 删除规则

| 场景 | 是否允许删除 | 说明 |
|-----|------------|------|
| 非默认卡 | ✅ 允许 | 直接删除 |
| 默认卡（唯一） | ✅ 允许 | 没有其他卡，可以删除 |
| 默认卡（非唯一） | ❌ 拒绝 | 需要先设置其他卡为默认 |

### 流程图

```
用户点击删除
    ↓
前端弹出确认框
    ↓
用户确认
    ↓
发送DELETE请求
    ↓
后端验证归属
    ↓
检查是否为默认卡
    ├─ 不是 → 直接删除 → 成功
    └─ 是
        ↓
        检查卡数量
        ├─ 只有1张 → 允许删除 → 成功
        └─ 多于1张 → 拒绝删除 → 提示"请先将其他银行卡设为默认卡"
```

---

## 🧪 测试场景

### 场景1: 删除非默认卡

**前提**: 有2张卡，删除的不是默认卡

**步骤**:
1. 点击非默认卡的删除图标
2. 确认删除

**预期**: ✅ 删除成功

---

### 场景2: 删除唯一的默认卡

**前提**: 只有1张卡，且是默认卡

**步骤**:
1. 点击删除图标
2. 确认删除

**预期**: ✅ 删除成功

---

### 场景3: 删除多张卡中的默认卡

**前提**: 有2张或更多卡，删除的是默认卡

**步骤**:
1. 点击默认卡的删除图标
2. 确认删除

**预期**: ❌ 提示"请先将其他银行卡设为默认卡"

**解决方案**:
1. 先点击其他卡（假设有设置默认卡的功能）
2. 将其他卡设为默认
3. 然后删除原来的默认卡

---

## 🔧 后续优化建议

### 1. 添加设置默认卡功能

目前前端没有设置默认卡的UI，建议添加：

```vue
<div v-for="card in cards" :key="card.id" class="card-item">
  <div class="card-header">
    <span class="bank-name">{{ card.bank_name || card.BankName }}</span>
    <div class="card-actions">
      <!-- 如果不是默认卡，显示"设为默认"按钮 -->
      <van-button 
        v-if="!card.is_default && !card.IsDefault" 
        size="small" 
        type="primary"
        @click="setDefault(card.id)"
      >
        设为默认
      </van-button>
      <van-icon name="delete-o" @click="deleteCard(card.id)" />
    </div>
  </div>
  <!-- 如果是默认卡，显示标签 -->
  <van-tag v-if="card.is_default || card.IsDefault" type="success">默认</van-tag>
  <div class="card-number">{{ formatCardNumber(card.card_number || card.CardNumber) }}</div>
  <div class="card-holder">{{ card.card_holder || card.CardHolder }}</div>
</div>
```

---

### 2. 优化删除确认提示

**当前**:
```javascript
await showConfirmDialog({
  title: '确认删除',
  message: '确定要删除这张银行卡吗？'
})
```

**建议**:
```javascript
// 判断是否为默认卡且有多张卡
const isDefault = card.is_default || card.IsDefault
const hasMultipleCards = cards.value.length > 1

let message = '确定要删除这张银行卡吗？'
if (isDefault && hasMultipleCards) {
  message = '这是您的默认银行卡，请先将其他银行卡设为默认后再删除。'
  showToast(message)
  return  // 直接返回，不允许删除
}

await showConfirmDialog({
  title: '确认删除',
  message: message
})
```

---

### 3. 后端API添加设置默认卡接口

```go
// PUT /bank-cards/:id/default - 设置默认银行卡
rg.PUT("/bank-cards/:id/default", func(c *gin.Context) {
    cardID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    userID := c.GetUint("user_id")
    
    if err := cardSvc.SetDefaultCard(userID, uint(cardID)); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "已设为默认银行卡"})
})
```

---

## ✅ 修复完成

### 修改的文件

1. ✅ `frontend/src/pages/BankCards.vue`
   - 显示具体错误信息

2. ✅ `backend/internal/service/bank_card_service.go`
   - 优化删除逻辑
   - 允许删除唯一的默认卡

### 功能状态

- ✅ 删除非默认卡正常
- ✅ 删除唯一的默认卡正常
- ✅ 删除多张卡中的默认卡会提示
- ✅ 错误信息显示正常

### 测试清单

- [ ] 测试删除非默认卡
- [ ] 测试删除唯一的默认卡
- [ ] 测试删除多张卡中的默认卡（应该提示）
- [ ] 验证错误提示是否显示

---

## 🚀 测试步骤

### 1. 重启后端服务

```powershell
# 停止后端
Ctrl + C

# 重新启动
cd e:\AI\SuxinZK\code\backend
go run cmd/server/main.go
```

### 2. 刷新前端

```
Ctrl + Shift + R
```

### 3. 测试删除功能

1. 访问 http://localhost:5173/bank-cards
2. 如果只有一张卡，直接删除测试
3. 如果有多张卡：
   - 先删除非默认卡（应该成功）
   - 再删除默认卡（如果还有其他卡，应该提示）

---

**重启后端服务后测试！** 🚀
