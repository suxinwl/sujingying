import { defineStore } from 'pinia'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'

export const useUserStore = defineStore('user', {
  state: () => ({
    userInfo: null,
    token: localStorage.getItem('access_token') || '',
    refreshToken: localStorage.getItem('refresh_token') || ''
  }),
  
  getters: {
    isLogin: (state) => !!state.token,
    role: (state) => state.userInfo?.role || '',
    isCustomer: (state) => state.userInfo?.role === 'customer',
    isSales: (state) => state.userInfo?.role === 'sales',
    isSupport: (state) => state.userInfo?.role === 'support',
    isAdmin: (state) => state.userInfo?.role === 'super_admin',
    userId: (state) => state.userInfo?.id || 0
  },
  
  actions: {
    // 登录
    async login(credentials) {
      const { data } = await request.post(API_ENDPOINTS.LOGIN, credentials)
      this.token = data.access_token
      this.refreshToken = data.refresh_token
      localStorage.setItem('access_token', data.access_token)
      localStorage.setItem('refresh_token', data.refresh_token)
      
      // 获取用户信息
      await this.getUserInfo()
      return data
    },
    
    // 注册
    async register(userData) {
      const { data } = await request.post(API_ENDPOINTS.REGISTER, userData)
      return data
    },
    
    // 获取用户信息
    async getUserInfo() {
      const { data } = await request.get(API_ENDPOINTS.USER_PROFILE)
      this.userInfo = data
      return data
    },
    
    // 更新用户信息
    async updateUserInfo(userData) {
      const { data } = await request.put(API_ENDPOINTS.USER_UPDATE, userData)
      this.userInfo = { ...this.userInfo, ...data }
      return data
    },
    
    // 修改密码
    async changePassword(passwordData) {
      const { data } = await request.post(API_ENDPOINTS.CHANGE_PASSWORD, passwordData)
      return data
    },
    
    // 退出登录
    async logout() {
      try {
        await request.post(API_ENDPOINTS.LOGOUT)
      } catch (error) {
        console.error('退出登录失败:', error)
      }
      this.token = ''
      this.refreshToken = ''
      this.userInfo = null
      localStorage.clear()
    }
  }
})
