/**
 * HTTP请求工具
 * 
 * @module utils/request
 * @description 基于axios封装的HTTP客户端，包含请求/响应拦截器和错误处理
 * @author 速金盈技术团队
 * @date 2025-11-18
 */

import axios from 'axios'
import { showToast, showFailToast } from 'vant'
import { API_BASE_URL } from '../config/api'
import router from '../router'

/**
 * 创建axios实例
 * 
 * @constant
 * @type {import('axios').AxiosInstance}
 */
const request = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000
})

/**
 * 请求拦截器
 * 
 * @description 在请求发送前添加认证token
 */
request.interceptors.request.use(
  config => {
    // 从localStorage获取token
    const token = localStorage.getItem('access_token')
    
    // 如果token存在，添加到请求头
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    
    return config
  },
  error => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

/**
 * 响应拦截器
 * 
 * @description 处理响应数据和错误，包含自动token刷新逻辑
 */
request.interceptors.response.use(
  response => {
    const res = response.data
    
    // 如果是文件下载，直接返回完整响应
    if (response.config.responseType === 'blob') {
      return response
    }
    
    // 统一处理响应格式
    // 后端统一返回 {data: {...}}
    // 如果响应有data字段，直接返回data内容
    // 如果没有data字段，返回整个响应（兼容旧格式）
    if (res && typeof res === 'object' && 'data' in res) {
      return res.data
    }
    
    // 兼容旧格式（直接返回数据）
    return res
  },
  async error => {
    console.error('响应错误:', error)
    
    if (!error.response) {
      showFailToast('网络错误，请检查网络连接')
      return Promise.reject(error)
    }
    
    const { status, data } = error.response
    
    switch (status) {
      case 400:
        showFailToast(data.error || '请求参数错误')
        break
      case 401:
        // Token过期，尝试刷新
        const refreshToken = localStorage.getItem('refresh_token')
        if (refreshToken && !error.config._retry) {
          error.config._retry = true
          try {
            const response = await axios.post(
              `${API_BASE_URL}/api/v1/auth/refresh`,
              { refresh_token: refreshToken }
            )
            // 后端统一返回 {data: {access_token, refresh_token}}
            const tokenData = response.data.data || response.data
            const { access_token } = tokenData
            localStorage.setItem('access_token', access_token)
            error.config.headers.Authorization = `Bearer ${access_token}`
            return request(error.config)
          } catch (refreshError) {
            // 刷新失败，清除token并跳转登录
            localStorage.clear()
            router.push('/login')
            showFailToast('登录已过期，请重新登录')
            return Promise.reject(refreshError)
          }
        } else {
          localStorage.clear()
          router.push('/login')
          showFailToast('登录已过期，请重新登录')
        }
        break
      case 403:
        showFailToast('没有权限访问')
        break
      case 404:
        showFailToast('请求的资源不存在')
        break
      case 500:
        showFailToast('服务器错误，请稍后重试')
        break
      default:
        showFailToast(data.error || '请求失败')
    }
    
    return Promise.reject(error)
  }
)

export default request
