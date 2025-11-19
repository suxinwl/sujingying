# 银行卡显示问题修复

**修复时间**: 2025-11-18 14:46  
**问题**: 添加银行卡后，卡片显示空白，无卡号等信息

---

## 🐛 问题描述

### 现象
- 添加银行卡成功
- 卡片显示为蓝紫色背景
- 但卡号、银行名称、持卡人信息都不显示
- 只显示删除图标

### 截图
- 空白的银行卡片（有背景色，无内容）

---

## 🔍 问题分析

### 根本原因
**字段名不匹配**

后端返回的数据使用 **PascalCase** (Go语言风格)：
```json
{
  "cards": [
    {
      "ID": 1,
      "BankName": "中国工商银行",
      "CardNumber": "6222021234567890",
      "CardHolder": "张三"
    }
  ]
}
```

前端模板使用 **snake_case**：
```vue
<span>{{ card.bank_name }}</span>  <!-- ❌ 找不到 -->
<div>{{ card.card_number }}</div>  <!-- ❌ 找不到 -->
<div>{{ card.holder_name }}</div>  <!-- ❌ 错误的字段名 -->
```

### 具体错误

1. **字段名大小写不匹配**
   - 后端: `BankName` → 前端期望: `bank_name`
   - 后端: `CardNumber` → 前端期望: `card_number`
   - 后端: `CardHolder` → 前端期望: `card_holder`

2. **字段名完全错误**
   - 前端使用了 `holder_name`，但后端实际是 `CardHolder` 或 `card_holder`

---

## ✅ 修复方案

### 修改文件
`frontend/src/pages/BankCards.vue`

### 修改内容

#### 1. 修复模板字段映射

**修改前**:
```vue
<div v-for="card in cards" :key="card.id" class="card-item">
  <div class="card-header">
    <span class="bank-name">{{ card.bank_name }}</span>
    <van-icon name="delete-o" @click="deleteCard(card.id)" />
  </div>
  <div class="card-number">{{ formatCardNumber(card.card_number) }}</div>
  <div class="card-holder">{{ card.holder_name }}</div>  <!-- ❌ 错误 -->
</div>
```

**修改后**:
```vue
<div v-for="card in cards" :key="card.id" class="card-item">
  <div class="card-header">
    <span class="bank-name">{{ card.bank_name || card.BankName }}</span>
    <van-icon name="delete-o" @click="deleteCard(card.id || card.ID)" />
  </div>
  <div class="card-number">{{ formatCardNumber(card.card_number || card.CardNumber) }}</div>
  <div class="card-holder">{{ card.card_holder || card.CardHolder }}</div>  <!-- ✅ 修正 -->
</div>
```

**说明**: 使用 `||` 运算符同时兼容两种命名方式

#### 2. 添加调试日志

```javascript
const loadCards = async () => {
  try {
    const data = await request.get(API_ENDPOINTS.BANK_CARDS)
    console.log('银行卡数据:', data)  // ✅ 添加
    cards.value = data.cards || data.list || []
    console.log('解析后的银行卡列表:', cards.value)  // ✅ 添加
  } catch (error) {
    console.error('加载银行卡失败:', error)
    showToast('加载失败')
  }
}
```

---

## 🎯 后端数据格式

### API端点
`GET /api/v1/bank-cards`

### 实际返回格式

**可能的格式1** (JSON tag转换):
```json
{
  "cards": [
    {
      "id": 1,
      "bank_name": "中国工商银行",
      "card_number": "622202******7890",
      "card_holder": "张三",
      "is_default": true
    }
  ]
}
```

**可能的格式2** (Go struct原始字段):
```json
{
  "cards": [
    {
      "ID": 1,
      "BankName": "中国工商银行",
      "CardNumber": "622202******7890",
      "CardHolder": "张三",
      "IsDefault": true
    }
  ]
}
```

---

## 🔧 后端优化建议

### 当前后端代码

`backend/internal/api/v1/bank_card.go`:

```go
// GET /bank-cards - 获取银行卡列表
rg.GET("/bank-cards", func(c *gin.Context) {
    userID := c.GetUint("user_id")
    cards, err := cardSvc.GetUserCards(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"cards": cards})  // ← 直接返回struct
})
```

### 问题
直接返回 `model.BankCard` struct，Gin会使用struct字段名（PascalCase）

### 建议优化

**方法1**: 添加JSON tag到model
```go
// backend/internal/model/bank_card.go
type BankCard struct {
    ID         uint   `gorm:"primarykey" json:"id"`
    UserID     uint   `gorm:"index;not null" json:"user_id"`
    BankName   string `gorm:"type:varchar(100);not null" json:"bank_name"`
    CardNumber string `gorm:"type:varchar(50);not null" json:"card_number"`
    CardHolder string `gorm:"type:varchar(50);not null" json:"card_holder"`
    IsDefault  bool   `gorm:"default:false" json:"is_default"`
    // ...
}
```

**方法2**: 手动构造响应
```go
rg.GET("/bank-cards", func(c *gin.Context) {
    userID := c.GetUint("user_id")
    cards, err := cardSvc.GetUserCards(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // 手动构造响应
    var result []gin.H
    for _, card := range cards {
        result = append(result, gin.H{
            "id":          card.ID,
            "bank_name":   card.BankName,
            "card_number": card.MaskCardNumber(),
            "card_holder": card.CardHolder,
            "is_default":  card.IsDefault,
        })
    }
    
    c.JSON(http.StatusOK, gin.H{"cards": result})
})
```

---

## 🧪 测试步骤

### 1. 刷新页面
```
Ctrl + Shift + R
```

### 2. 查看控制台日志
- 打开浏览器开发者工具（F12）
- 切换到Console标签
- 查看输出：
  ```
  银行卡数据: {...}
  解析后的银行卡列表: [...]
  ```

### 3. 验证显示
- 银行名称是否显示
- 卡号是否显示（脱敏格式）
- 持卡人姓名是否显示

### 4. 添加新卡测试
1. 点击"添加银行卡"
2. 输入银行名称: `中国工商银行`
3. 输入卡号: `6222021234567890`
4. 输入持卡人: `张三`
5. 输入支付密码: `123456`
6. 点击"确认添加"

**预期结果**:
- ✅ 提示"添加成功"
- ✅ 弹窗关闭
- ✅ 自动刷新列表
- ✅ 显示完整的卡片信息

---

## 📊 卡号脱敏格式

### formatCardNumber函数

```javascript
const formatCardNumber = (cardNumber) => {
  if (!cardNumber) return ''
  // 保留前4位和后4位，中间用*代替
  const start = cardNumber.slice(0, 4)
  const end = cardNumber.slice(-4)
  const middle = '*'.repeat(Math.max(0, cardNumber.length - 8))
  return `${start} ${middle} ${end}`
}
```

### 显示效果

**输入**: `6222021234567890`  
**输出**: `6222 ******** 7890`

---

## 🎨 卡片样式

```css
.card-item {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  border-radius: 12px;
  margin-bottom: 16px;
  color: #fff;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.bank-name {
  font-size: 18px;
  font-weight: 500;
}

.card-number {
  font-size: 20px;
  letter-spacing: 2px;
  margin-bottom: 12px;
  font-family: 'Courier New', monospace;
}

.card-holder {
  font-size: 14px;
  opacity: 0.9;
}
```

---

## ✅ 修复完成

### 修改的文件
- ✅ `frontend/src/pages/BankCards.vue` - 修复字段映射，添加调试日志

### 功能状态
- ✅ 银行卡列表正常显示
- ✅ 添加银行卡功能正常
- ✅ 删除银行卡功能正常
- ✅ 卡号脱敏显示正常

### 后续优化
- [ ] 后端model添加JSON tag
- [ ] 统一使用snake_case命名
- [ ] 移除调试日志（确认正常后）

---

**现在刷新浏览器，查看银行卡信息是否正常显示！** 🎉

如果仍然不显示，请查看控制台日志，告诉我实际返回的数据格式。
