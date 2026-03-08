import type { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import axios from 'axios'
import { showToast } from 'vant'
import { STORAGE_TOKEN_KEY } from '@/stores/mutation-type'
import { i18n } from '@/utils/i18n'

// 这里是用于设定请求后端时，所用的 Token KEY
// 可以根据自己的需要修改，常见的如 Access-Token，Authorization
// 需要注意的是，请尽量保证使用中横线`-` 来作为分隔符，
// 避免被 nginx 等负载均衡器丢弃了自定义的请求头
export const REQUEST_TOKEN_KEY = 'Authorization'

// 创建 axios 实例
const request = axios.create({
  // API 请求的默认前缀
  baseURL: import.meta.env.VITE_APP_API_BASE_URL,
  timeout: 6000, // 请求超时时间
})

export type RequestError = AxiosError<{
  code?: number
  msg?: string
  success?: boolean
  data?: any
  message?: string
  result?: any
  errorMessage?: string
}>

function tr(key: string, fallback: string) {
  const value = i18n.global.t(key)
  return typeof value === 'string' ? value : fallback
}

function getErrorMessage(payload?: Record<string, any>, fallback = tr('common.requestFailed', 'Request failed')) {
  return payload?.message || payload?.msg || payload?.errorMessage || fallback
}

function normalizeErrorPayload(payload: unknown): Record<string, any> {
  if (!payload)
    return {}

  if (typeof payload === 'string') {
    try {
      const parsed = JSON.parse(payload)
      return parsed && typeof parsed === 'object' ? parsed as Record<string, any> : { message: payload }
    }
    catch {
      return { message: payload }
    }
  }

  if (typeof payload === 'object')
    return payload as Record<string, any>

  return { message: String(payload) }
}

// 异常拦截处理器
function errorHandler(error: RequestError): Promise<any> {
  if (error.response) {
    const { data, status, statusText } = error.response
    const payload = normalizeErrorPayload(data)
    const message = getErrorMessage(payload, statusText || tr('common.requestFailed', 'Request failed'))

    // 403 无权限
    if (status === 403) {
      showToast(message)
    }
    // 401 未登录/未授权
    if (status === 401 && payload.result && payload.result.isLogin) {
      showToast(tr('common.authFailed', 'Authorization verification failed'))
      // 如果你需要直接跳转登录页面
      // location.replace(loginRoutePath)
    }
    if (status === 400) {
      showToast(getErrorMessage(payload, tr('common.badRequest', 'Invalid request parameters')))
    }
    if (status !== 400 && status !== 401 && status !== 403) {
      showToast(message)
    }
  }
  else {
    showToast(error.message || tr('common.networkError', 'Network error, please try again later'))
  }
  return Promise.reject(error)
}

// 请求拦截器
function requestHandler(config: InternalAxiosRequestConfig): InternalAxiosRequestConfig | Promise<InternalAxiosRequestConfig> {
  const savedToken = localStorage.getItem(STORAGE_TOKEN_KEY)
  // 如果 token 存在
  // 让每个请求携带自定义 token, 请根据实际情况修改
  if (savedToken)
    config.headers[REQUEST_TOKEN_KEY] = savedToken

  return config
}

// Add a request interceptor
request.interceptors.request.use(requestHandler, errorHandler)

// 响应拦截器
function responseHandler(response: AxiosResponse): any {
  const payload = response.data as Record<string, any>

  if (payload && typeof payload === 'object') {
    const hasSuccess = typeof payload.success === 'boolean'
    const hasCode = typeof payload.code === 'number'

    const isBizError = (hasSuccess && payload.success === false)
      || (hasCode && ![0, 200].includes(payload.code))

    if (isBizError) {
      const message = getErrorMessage(payload)
      showToast(message)
      return Promise.reject(new Error(message))
    }
  }

  return payload
}

// Add a response interceptor
request.interceptors.response.use(responseHandler, errorHandler)

interface RequestInstance extends AxiosInstance {
  <T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  <T = any>(config: AxiosRequestConfig): Promise<T>
  get: <T = any>(url: string, config?: AxiosRequestConfig) => Promise<T>
  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) => Promise<T>
  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) => Promise<T>
  delete: <T = any>(url: string, config?: AxiosRequestConfig) => Promise<T>
}

export default request as unknown as RequestInstance
