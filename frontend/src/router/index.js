import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

// 布局组件
import Layout from '../pages/Layout.vue'

// 页面组件
import Login from '../pages/Login.vue'
import Register from '../pages/Register.vue'
import Home from '../pages/Home.vue'
import Quotes from '../pages/Quotes.vue'
import Trade from '../pages/Trade.vue'
import Positions from '../pages/Positions.vue'
import Funds from '../pages/Funds.vue'
import Mine from '../pages/Mine.vue'
import Orders from '../pages/Orders.vue'
import OrderDetail from '../pages/OrderDetail.vue'
import BankCards from '../pages/BankCards.vue'
import Notifications from '../pages/Notifications.vue'
import About from '../pages/About.vue'
import InviteCodes from '../pages/InviteCodes.vue'

// 管理员页面
import AdminUsers from '../pages/admin/Users.vue'
import AdminDeposits from '../pages/admin/Deposits.vue'
import AdminWithdraws from '../pages/admin/Withdraws.vue'
import AdminConfig from '../pages/admin/Config.vue'
import PaymentSettings from '../pages/admin/PaymentSettings.vue'
import PlatformAddresses from '../pages/admin/PlatformAddresses.vue'
import AdminAnnouncements from '../pages/admin/Announcements.vue'
import AdminSales from '../pages/admin/Sales.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // 登录注册
    { path: '/login', component: Login, meta: { guest: true } },
    { path: '/register', component: Register, meta: { guest: true } },
    
    // 主应用
    {
      path: '/',
      component: Layout,
      redirect: '/quotes',
      meta: { requiresAuth: true },
      children: [
        { path: 'quotes', component: Quotes },
        { path: 'home', component: Home },
        { path: 'trade', component: Trade },
        { path: 'positions', component: Positions },
        { path: 'funds', component: Funds },
        { path: 'mine', component: Mine }
      ]
    },
    
    // 独立页面（需要登录）
    { 
      path: '/orders', 
      component: Orders, 
      meta: { requiresAuth: true } 
    },
    {
      path: '/orders/:id',
      component: OrderDetail,
      meta: { requiresAuth: true }
    },
    { 
      path: '/bank-cards', 
      component: BankCards, 
      meta: { requiresAuth: true } 
    },
    { 
      path: '/notifications', 
      component: Notifications, 
      meta: { requiresAuth: true } 
    },
    { 
      path: '/about', 
      component: About, 
      meta: { requiresAuth: true } 
    },
    {
      path: '/invite-codes',
      component: InviteCodes,
      meta: { requiresAuth: true }
    },
    
    // 管理员页面
    { 
      path: '/admin/users', 
      component: AdminUsers, 
      meta: { requiresAuth: true, requiresAdmin: true } 
    },
    { 
      path: '/admin/announcements', 
      component: AdminAnnouncements, 
      meta: { requiresAuth: true, requiresAdmin: true } 
    },
    { 
      path: '/admin/sales', 
      component: AdminSales, 
      meta: { requiresAuth: true, requiresAdmin: true } 
    },
    { 
      path: '/admin/deposits', 
      component: AdminDeposits, 
      meta: { requiresAuth: true, requiresAdmin: true } 
    },
    { 
      path: '/admin/withdraws', 
      component: AdminWithdraws, 
      meta: { requiresAuth: true, requiresAdmin: true } 
    },
    { 
      path: '/admin/config', 
      component: AdminConfig, 
      meta: { requiresAuth: true, requiresAdmin: true } 
    },
    { 
      path: '/admin/platform-addresses', 
      component: PlatformAddresses, 
      meta: { requiresAuth: true, requiresAdmin: true } 
    },
    { 
      path: '/admin/payment-settings', 
      component: PaymentSettings, 
      meta: { requiresAuth: true, requiresAdmin: true } 
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  // 需要登录的页面
  if (to.meta.requiresAuth && !userStore.isLogin) {
    next('/login')
    return
  }
  
  // 需要管理员权限的页面
  if (to.meta.requiresAdmin) {
    const role = userStore.userInfo?.role
    if (role !== 'super_admin' && role !== 'support') {
      next('/')
      return
    }
  }
  
  // 游客页面（已登录用户不能访问）
  if (to.meta.guest && userStore.isLogin) {
    next('/')
    return
  }
  
  next()
})

export default router
