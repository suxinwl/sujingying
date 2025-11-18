# 前端代码规范

## 1. 文件命名规范

### Vue组件
- 使用 **PascalCase** (大驼峰命名)
- 文件名应具有描述性

```
✅ 正确:
- Login.vue
- UserProfile.vue
- OrderList.vue

❌ 错误:
- login.vue
- user-profile.vue
- orderlist.vue
```

### JavaScript文件
- 使用 **camelCase** (小驼峰命名)
- Store文件使用具体名称

```
✅ 正确:
- request.js
- helpers.js
- user.js (store)

❌ 错误:
- Request.js
- Helpers.js
```

### 样式文件
- 使用 **kebab-case** (短横线命名)

```
✅ 正确:
- common.css
- user-profile.css
```

---

## 2. 变量命名规范

### 普通变量
- 使用 **camelCase**
- 具有描述性

```javascript
✅ 正确:
const userName = 'John'
const isActive = true
const currentPrice = 500

❌ 错误:
const user_name = 'John'
const IsActive = true
const cp = 500
```

### 常量
- 使用 **UPPER_SNAKE_CASE**

```javascript
✅ 正确:
const API_BASE_URL = 'http://localhost:8080'
const MAX_RETRY_COUNT = 3

❌ 错误:
const apiBaseUrl = 'http://localhost:8080'
const maxRetryCount = 3
```

### 布尔值
- 使用 `is/has/can/should` 前缀

```javascript
✅ 正确:
const isLogin = true
const hasPermission = false
const canEdit = true
const shouldUpdate = false

❌ 错误:
const login = true
const permission = false
```

---

## 3. 函数命名规范

### 普通函数
- 使用 **camelCase**
- 动词开头，描述功能

```javascript
✅ 正确:
function getUserInfo() {}
function calculateTotal() {}
function validateForm() {}

❌ 错误:
function user_info() {}
function total() {}
function validate() {}
```

### 事件处理函数
- 使用 `on` 或 `handle` 前缀

```javascript
✅ 正确:
const onSubmit = () => {}
const onClick = () => {}
const handleChange = () => {}

❌ 错误:
const submit = () => {}
const click = () => {}
```

### 生命周期钩子
- 按顺序排列
- 添加注释说明

```javascript
/**
 * 组件挂载时执行
 */
onMounted(() => {
  loadData()
})

/**
 * 组件卸载时执行
 */
onUnmounted(() => {
  cleanup()
})
```

---

## 4. 注释规范

### 文件注释
每个文件顶部必须包含文件说明：

```javascript
/**
 * 用户状态管理
 * 
 * @module stores/user
 * @description 管理用户认证状态、用户信息、权限等
 * @author 速金盈技术团队
 * @date 2025-11-18
 */
```

### 函数注释
使用 JSDoc 格式：

```javascript
/**
 * 获取用户信息
 * 
 * @async
 * @param {number} userId - 用户ID
 * @returns {Promise<Object>} 用户信息对象
 * @throws {Error} 当用户不存在时抛出错误
 * @description 从API获取指定用户的详细信息
 * 
 * @example
 * const user = await getUserInfo(123)
 * console.log(user.name)
 */
async function getUserInfo(userId) {
  // 实现代码
}
```

### 变量注释
使用 JSDoc 类型注解：

```javascript
/** @type {string} 用户名 */
const username = 'admin'

/** @type {import('vue').Ref<number>} 计数器 */
const count = ref(0)

/** @type {{name: string, age: number}} 用户对象 */
const user = {
  name: 'John',
  age: 30
}
```

### 行内注释
- 复杂逻辑必须添加注释
- 注释应解释"为什么"而不是"是什么"

```javascript
✅ 正确:
// 防止重复提交，添加防抖
const debouncedSubmit = debounce(onSubmit, 300)

// 由于后端返回时间戳，需要转换为Date对象
const date = new Date(timestamp * 1000)

❌ 错误:
// 创建变量
const user = {}

// 调用函数
getData()
```

### 代码块注释
使用分隔符组织代码：

```javascript
// ==================== 响应式数据 ====================
const loading = ref(false)
const data = ref([])

// ==================== 计算属性 ====================
const total = computed(() => data.value.length)

// ==================== 事件处理 ====================
const onSubmit = async () => {
  // ...
}

// ==================== 生命周期 ====================
onMounted(() => {
  // ...
})
```

---

## 5. Vue组件规范

### 组件结构顺序

```vue
<template>
  <!-- 模板 -->
</template>

<script setup>
/**
 * 组件说明注释
 */

// 1. 导入
import { ref } from 'vue'

// 2. Props定义
const props = defineProps({})

// 3. Emits定义
const emit = defineEmits(['update'])

// 4. 响应式数据
const loading = ref(false)

// 5. 计算属性
const total = computed(() => {})

// 6. 方法
const onSubmit = () => {}

// 7. 生命周期
onMounted(() => {})
</script>

<style scoped>
/* 样式 */
</style>
```

### Props定义
- 必须定义类型
- 添加验证和默认值
- 添加注释

```javascript
/**
 * 组件Props
 */
const props = defineProps({
  /** 用户ID */
  userId: {
    type: Number,
    required: true
  },
  
  /** 是否显示 */
  visible: {
    type: Boolean,
    default: false
  },
  
  /** 用户信息对象 */
  userInfo: {
    type: Object,
    default: () => ({})
  }
})
```

### Emits定义
- 列出所有事件
- 添加注释

```javascript
/**
 * 组件事件
 */
const emit = defineEmits([
  'update',      // 更新事件
  'delete',      // 删除事件
  'close'        // 关闭事件
])
```

---

## 6. Store规范

### State
```javascript
state: () => ({
  /** @type {Object|null} 用户信息 */
  userInfo: null,
  
  /** @type {string} 访问令牌 */
  token: '',
  
  /** @type {boolean} 加载状态 */
  loading: false
})
```

### Getters
```javascript
getters: {
  /**
   * 是否已登录
   * 
   * @param {Object} state - 状态对象
   * @returns {boolean}
   */
  isLogin: (state) => !!state.token
}
```

### Actions
```javascript
actions: {
  /**
   * 用户登录
   * 
   * @async
   * @param {Object} credentials - 登录凭证
   * @returns {Promise<Object>}
   */
  async login(credentials) {
    // 实现
  }
}
```

---

## 7. CSS类命名规范

### BEM命名法
- Block-Element-Modifier
- 使用 **kebab-case**

```css
✅ 正确:
.user-card {}
.user-card__header {}
.user-card__body {}
.user-card--active {}

❌ 错误:
.userCard {}
.user_card_header {}
.UserCard {}
```

### Scoped样式
- 组件样式必须添加 `scoped`
- 全局样式放在单独文件

```vue
<style scoped>
/* 组件内样式 */
.user-card {
  padding: 16px;
}
</style>
```

---

## 8. 导入顺序

按以下顺序导入模块：

```javascript
// 1. Vue核心
import { ref, computed, onMounted } from 'vue'

// 2. 第三方库
import axios from 'axios'
import { showToast } from 'vant'

// 3. 本地模块
import { useUserStore } from '@/stores/user'
import { formatMoney } from '@/utils/helpers'
import MyComponent from '@/components/MyComponent.vue'
```

---

## 9. 错误处理

### Try-Catch
- 异步函数必须包含错误处理
- 提供友好的错误提示

```javascript
✅ 正确:
const loadData = async () => {
  try {
    const { data } = await request.get('/api/data')
    items.value = data.list
  } catch (error) {
    console.error('加载数据失败:', error)
    showFailToast('加载失败，请稍后重试')
  }
}

❌ 错误:
const loadData = async () => {
  const { data } = await request.get('/api/data')
  items.value = data.list
}
```

---

## 10. 代码格式化

### 缩进
- 使用 **2个空格**
- 不使用Tab

### 分号
- 统一不使用分号（可选）
- 或统一使用分号

### 引号
- 统一使用单引号 `'`
- HTML属性使用双引号 `"`

### 对象/数组
- 多行时最后一项添加逗号

```javascript
✅ 正确:
const user = {
  name: 'John',
  age: 30,    // 逗号
}

❌ 错误:
const user = {
  name: 'John',
  age: 30     // 缺少逗号
}
```

---

## 11. 性能优化

### 计算属性
- 使用 `computed` 而不是方法

```javascript
✅ 正确:
const total = computed(() => items.value.length)

❌ 错误:
const getTotal = () => items.value.length
```

### v-if vs v-show
- 切换频繁用 `v-show`
- 初始隐藏用 `v-if`

```vue
<!-- 频繁切换 -->
<div v-show="isVisible">内容</div>

<!-- 条件渲染 -->
<div v-if="hasPermission">内容</div>
```

### 列表渲染
- 必须添加 `:key`
- key使用唯一标识

```vue
✅ 正确:
<div v-for="item in list" :key="item.id">
  {{ item.name }}
</div>

❌ 错误:
<div v-for="(item, index) in list" :key="index">
  {{ item.name }}
</div>
```

---

## 12. TypeScript类型注解（可选）

虽然项目使用JavaScript，但可以使用JSDoc提供类型提示：

```javascript
/**
 * @typedef {Object} User
 * @property {number} id - 用户ID
 * @property {string} name - 用户名
 * @property {string} email - 邮箱
 */

/**
 * @param {User} user - 用户对象
 * @returns {string}
 */
function getUserName(user) {
  return user.name
}
```

---

## 13. Git提交规范

### 提交信息格式
```
type(scope): subject

body

footer
```

### Type类型
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式（不影响代码运行）
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具变动

### 示例
```
feat(login): 添加记住密码功能

- 添加记住密码复选框
- 本地存储用户名
- 自动填充登录表单

Closes #123
```

---

## 14. 检查清单

提交代码前检查：

- [ ] 所有函数都有注释
- [ ] 变量命名符合规范
- [ ] 没有console.log（除了必要的日志）
- [ ] 没有debugger
- [ ] 代码格式统一
- [ ] 没有未使用的变量/导入
- [ ] 错误处理完整
- [ ] 组件结构清晰
- [ ] CSS类命名规范

---

**制定单位**: 速金盈技术团队  
**生效日期**: 2025-11-18  
**版本**: v1.0
