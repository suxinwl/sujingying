# 速金盈 - 前端应用

基于 Vue 3 + Vite + Vant 4 构建的黄金交易平台前端应用。

## 技术栈

- **Vue 3** - 渐进式JavaScript框架
- **Vite** - 下一代前端构建工具
- **Vant 4** - 轻量、可靠的移动端组件库
- **Vue Router** - Vue.js 官方路由管理器
- **Pinia** - Vue 的状态管理库
- **Axios** - HTTP客户端
- **Day.js** - 轻量级日期处理库

## 功能模块

### 用户功能
- ✅ 用户注册/登录
- ✅ 个人信息管理
- ✅ 修改密码

### 交易功能
- ✅ 实时行情展示（WebSocket）
- ✅ 买入/卖出下单
- ✅ 订单管理
- ✅ 持仓管理
- ✅ 盈亏计算

### 资金功能
- ✅ 余额查询
- ✅ 充值申请
- ✅ 提现申请
- ✅ 资金流水
- ✅ 银行卡管理

### 销售功能
- ✅ 邀请码生成
- ✅ 客户管理
- ✅ 提成记录

### 管理功能
- ✅ 用户审核
- ✅ 充值审核
- ✅ 提现审核
- ✅ 系统配置

## 项目结构

```
frontend/
├── public/             # 静态资源
├── src/
│   ├── assets/         # 资源文件
│   ├── config/         # 配置文件
│   │   ├── api.js      # API配置
│   │   └── websocket.js # WebSocket配置
│   ├── pages/          # 页面组件
│   │   ├── Layout.vue  # 主布局
│   │   ├── Login.vue   # 登录
│   │   ├── Register.vue # 注册
│   │   ├── Home.vue    # 首页
│   │   ├── Trade.vue   # 交易
│   │   ├── Positions.vue # 持仓
│   │   ├── Funds.vue   # 资金
│   │   └── Mine.vue    # 我的
│   ├── router/         # 路由配置
│   │   └── index.js
│   ├── stores/         # 状态管理
│   │   ├── user.js     # 用户状态
│   │   └── quote.js    # 行情状态
│   ├── utils/          # 工具函数
│   │   ├── request.js  # HTTP请求
│   │   └── helpers.js  # 辅助函数
│   ├── App.vue         # 根组件
│   └── main.js         # 入口文件
├── index.html          # HTML模板
├── package.json        # 项目配置
├── vite.config.js      # Vite配置
└── README.md           # 项目说明

```

## 开发指南

### 环境要求

- Node.js >= 16.0.0
- npm >= 7.0.0

### 安装依赖

```bash
cd frontend
npm install
```

### 开发运行

```bash
npm run dev
```

访问: http://localhost:5173

### 生产构建

```bash
npm run build
```

构建产物在 `dist` 目录。

### 预览构建

```bash
npm run preview
```

## 环境变量

创建 `.env.development` 和 `.env.production` 文件：

```bash
# 开发环境
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080

# 生产环境
VITE_API_BASE_URL=https://api.suxinying.com
VITE_WS_BASE_URL=wss://api.suxinying.com
```

## API接口

### 认证相关
- `POST /api/v1/auth/login` - 登录
- `POST /api/v1/auth/register` - 注册
- `POST /api/v1/auth/refresh` - 刷新Token
- `POST /api/v1/auth/logout` - 登出

### 用户相关
- `GET /api/v1/users/profile` - 获取用户信息
- `PUT /api/v1/users/profile` - 更新用户信息
- `POST /api/v1/users/password` - 修改密码

### 订单相关
- `GET /api/v1/orders` - 订单列表
- `POST /api/v1/orders/buy` - 买入下单
- `POST /api/v1/orders/sell` - 卖出下单
- `GET /api/v1/orders/:id` - 订单详情

### 持仓相关
- `GET /api/v1/positions` - 持仓列表
- `GET /api/v1/positions/:id` - 持仓详情

### 资金相关
- `GET /api/v1/deposits` - 充值记录
- `POST /api/v1/deposits` - 申请充值
- `GET /api/v1/withdraws` - 提现记录
- `POST /api/v1/withdraws` - 申请提现
- `GET /api/v1/fund-flow` - 资金流水

### WebSocket
- `ws://localhost:8080/ws/quote` - 行情推送
- `ws://localhost:8080/ws/notification` - 通知推送

## 代码规范

### 组件命名
- 使用 PascalCase: `UserProfile.vue`
- 组件名称应该具有描述性

### 变量命名
- 使用 camelCase: `userName`, `isActive`
- 布尔值使用 `is/has/can` 前缀

### 函数命名
- 使用 camelCase: `getUserInfo`, `handleSubmit`
- 事件处理函数使用 `on` 前缀: `onSubmit`, `onClick`

### CSS类命名
- 使用 kebab-case: `user-card`, `order-list`

## 状态管理

使用 Pinia 进行状态管理：

```javascript
// stores/user.js
import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', {
  state: () => ({
    userInfo: null,
    token: ''
  }),
  
  getters: {
    isLogin: (state) => !!state.token
  },
  
  actions: {
    async login(credentials) {
      // 登录逻辑
    }
  }
})
```

## 路由守卫

```javascript
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth && !userStore.isLogin) {
    next('/login')
    return
  }
  
  next()
})
```

## WebSocket使用

```javascript
// 连接行情WebSocket
const quoteStore = useQuoteStore()
quoteStore.connectWebSocket()

// 监听价格变化
watch(() => quoteStore.currentPrice, (newPrice) => {
  console.log('价格更新:', newPrice)
})
```

## 常见问题

### 1. 跨域问题

在 `vite.config.js` 中配置代理：

```javascript
export default {
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
}
```

### 2. WebSocket连接失败

检查后端服务是否启动，WebSocket地址是否正确。

### 3. Token过期

系统会自动使用 refresh_token 刷新 access_token。

## 部署

### Nginx配置

```nginx
server {
    listen 80;
    server_name suxinying.com;
    root /var/www/suxinying/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## 联系方式

- 项目地址: https://github.com/suxinwl/sujingying
- 技术支持: 速金盈技术团队

## License

MIT License
