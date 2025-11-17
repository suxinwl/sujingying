import { createRouter, createWebHistory } from 'vue-router'
import Quotes from '../pages/Quotes.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/quotes' },
    { path: '/quotes', component: Quotes }
  ]
})

export default router
