import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

// 布局组件
import Layout from '../pages/Layout.vue'

// 页面组件
import Login from '../pages/Login.vue'
import Register from '../pages/Register.vue'
import Home from '../pages/Home.vue'
import Trade from '../pages/Trade.vue'
import Positions from '../pages/Positions.vue'
import Funds from '../pages/Funds.vue'
import Mine from '../pages/Mine.vue'

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
      redirect: '/home',
      meta: { requiresAuth: true },
      children: [
        { path: 'home', component: Home },
        { path: 'trade', component: Trade },
        { path: 'positions', component: Positions },
        { path: 'funds', component: Funds },
        { path: 'mine', component: Mine }
      ]
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
  
  // 游客页面（已登录用户不能访问）
  if (to.meta.guest && userStore.isLogin) {
    next('/')
    return
  }
  
  next()
})

export default router
