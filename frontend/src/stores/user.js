/**
 * 用户状态管理
 * 
 * @module stores/user
 * @description 管理用户认证状态、用户信息、权限等
 * @author 速金盈技术团队
 * @date 2025-11-18
 */

import { defineStore } from 'pinia'
import request from '../utils/request'
import { API_ENDPOINTS } from '../config/api'

/**
 * 用户Store
 * 
 * @typedef {Object} UserState
 * @property {Object|null} userInfo - 用户信息对象
 * @property {string} token - 访问令牌
 * @property {string} refreshToken - 刷新令牌
 */

export const useUserStore = defineStore('user', {
  /**
   * 状态定义
   * 
   * @returns {UserState}
   */
  state: () => ({
    /** @type {Object|null} 用户信息 */
    userInfo: null,
    
    /** @type {string} 访问令牌 */
    token: localStorage.getItem('access_token') || '',
    
    /** @type {string} 刷新令牌 */
    refreshToken: localStorage.getItem('refresh_token') || ''
  }),
  
  /**
   * 计算属性
   */
  getters: {
    /**
     * 是否已登录
     * 
     * @param {UserState} state - 状态对象
     * @returns {boolean}
     */
    isLogin: (state) => !!state.token,
    
    /**
     * 用户角色
     * 
     * @param {UserState} state - 状态对象
     * @returns {string} customer|sales|support|super_admin
     */
    role: (state) => state.userInfo?.role || '',
    
    /**
     * 是否为客户角色
     * 
     * @param {UserState} state - 状态对象
     * @returns {boolean}
     */
    isCustomer: (state) => state.userInfo?.role === 'customer',
    
    /**
     * 是否为销售角色
     * 
     * @param {UserState} state - 状态对象
     * @returns {boolean}
     */
    isSales: (state) => state.userInfo?.role === 'sales',
    
    /**
     * 是否为客服角色
     * 
     * @param {UserState} state - 状态对象
     * @returns {boolean}
     */
    isSupport: (state) => state.userInfo?.role === 'support',
    
    /**
     * 是否为管理员角色
     * 
     * @param {UserState} state - 状态对象
     * @returns {boolean}
     */
    isAdmin: (state) => state.userInfo?.role === 'super_admin',
    
    /**
     * 用户ID
     * 
     * @param {UserState} state - 状态对象
     * @returns {number}
     */
    userId: (state) => state.userInfo?.id || 0
  },
  
  /**
   * 操作方法
   */
  actions: {
    /**
     * 用户登录
     * 
     * @async
     * @param {Object} credentials - 登录凭证
     * @param {string} credentials.username - 用户名
     * @param {string} credentials.password - 密码
     * @returns {Promise<Object>} 登录响应数据
     * @description 提交登录请求，保存token，并获取用户信息
     */
    async login(credentials) {
      // 响应拦截器已统一处理data字段
      const data = await request.post(API_ENDPOINTS.LOGIN, credentials)
      
      // 保存token到状态和localStorage
      this.token = data.access_token
      this.refreshToken = data.refresh_token
      localStorage.setItem('access_token', data.access_token)
      localStorage.setItem('refresh_token', data.refresh_token)
      
      // 获取用户信息
      await this.getUserInfo()
      
      return data
    },
    
    /**
     * 用户注册
     * 
     * @async
     * @param {Object} userData - 注册信息
     * @param {string} userData.username - 用户名
     * @param {string} userData.password - 密码
     * @param {string} userData.real_name - 真实姓名
     * @param {string} userData.phone - 手机号
     * @param {string} userData.invite_code - 邀请码
     * @returns {Promise<Object>} 注册响应数据
     * @description 提交注册请求，需等待审核
     */
    async register(userData) {
      return await request.post(API_ENDPOINTS.REGISTER, userData)
    },
    
    /**
     * 获取用户信息
     * 
     * @async
     * @returns {Promise<Object>} 用户信息
     * @description 从API获取当前登录用户的详细信息
     */
    async getUserInfo() {
      const data = await request.get(API_ENDPOINTS.USER_PROFILE)
      this.userInfo = data
      return data
    },
    
    /**
     * 更新用户信息
     * 
     * @async
     * @param {Object} userData - 要更新的用户信息
     * @returns {Promise<Object>} 更新后的用户信息
     * @description 更新用户个人信息，并合并到本地状态
     */
    async updateUserInfo(userData) {
      const data = await request.put(API_ENDPOINTS.USER_UPDATE, userData)
      this.userInfo = { ...this.userInfo, ...data }
      return data
    },
    
    /**
     * 修改密码
     * 
     * @async
     * @param {Object} passwordData - 密码数据
     * @param {string} passwordData.old_password - 旧密码
     * @param {string} passwordData.new_password - 新密码
     * @returns {Promise<Object>} 修改结果
     * @description 提交密码修改请求
     */
    async changePassword(passwordData) {
      return await request.post(API_ENDPOINTS.CHANGE_PASSWORD, passwordData)
    },
    
    /**
     * 退出登录
     * 
     * @async
     * @returns {Promise<void>}
     * @description 调用退出API，清除本地token和用户信息
     */
    async logout() {
      try {
        // 调用后端退出接口
        await request.post(API_ENDPOINTS.LOGOUT)
      } catch (error) {
        console.error('退出登录失败:', error)
      }
      
      // 清除本地状态
      this.token = ''
      this.refreshToken = ''
      this.userInfo = null
      
      // 清除localStorage
      localStorage.clear()
    }
  }
})
