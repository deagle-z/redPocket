import axios from 'axios'
import type {
  AxiosError,
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  InternalAxiosRequestConfig,
  Method,
} from 'axios'
import { showFailToast } from 'vant'
import { i18n } from '@/plugins/i18n'
import { trackApiError } from '@/utils/tracker'
import { notifyAuthExpired } from '@/utils/authEvents'
import { clearAuthToken, getAuthToken } from '@/utils/authToken'

export type ApiCode = number | string

export interface ApiEnvelope<T = unknown> {
  code: ApiCode
  data: T
  msg?: string
  message?: string
  error?: string
  success?: boolean
}

export interface RequestMeta {
  /**
   * Whether to inject Authorization header. Enabled by default.
   */
  auth?: boolean
  /**
   * Whether to show a Vant error toast for HTTP/business failures.
   * Enabled by default.
   */
  showError?: boolean
  /**
   * Return the backend envelope instead of envelope.data.
   */
  returnEnvelope?: boolean
  /**
   * Return the full Axios response.
   */
  returnResponse?: boolean
  /**
   * Cancel an existing pending request with the same method/url/params/data.
   * Disabled by default to avoid surprising callers.
   */
  dedupe?: boolean
  /**
   * Retry count for network errors, timeouts, and 5xx responses.
   */
  retry?: number
  /**
   * Initial retry delay in milliseconds. Delay grows linearly by attempt.
   */
  retryDelay?: number
  /**
   * Override the request key used by dedupe.
   */
  requestKey?: string
}

export interface HttpRequestConfig<D = unknown> extends AxiosRequestConfig<D> {
  meta?: RequestMeta
}

export interface UploadFileOptions {
  file: File | Blob
  fieldName?: string
  data?: Record<string, string | Blob>
  config?: HttpRequestConfig<FormData>
}

export class HttpError<T = unknown> extends Error {
  code?: ApiCode
  status?: number
  data?: T
  response?: AxiosResponse
  isBusinessError: boolean
  isCanceled: boolean

  constructor(options: {
    message: string
    code?: ApiCode
    status?: number
    data?: T
    response?: AxiosResponse
    isBusinessError?: boolean
    isCanceled?: boolean
  }) {
    super(options.message)
    this.name = 'HttpError'
    this.code = options.code
    this.status = options.status
    this.data = options.data
    this.response = options.response
    this.isBusinessError = Boolean(options.isBusinessError)
    this.isCanceled = Boolean(options.isCanceled)
  }
}

const DEFAULT_TIMEOUT = 15000
const DEFAULT_RETRY_DELAY = 300
const SUCCESS_CODES: ApiCode[] = [0, 200, '0', '200']

const pendingControllers = new Map<string, AbortController>()

const http: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: DEFAULT_TIMEOUT,
})

function getMeta(config: HttpRequestConfig): Required<
  Pick<RequestMeta, 'auth' | 'showError' | 'dedupe' | 'retry' | 'retryDelay'>
> &
  RequestMeta {
  return {
    auth: config.meta?.auth ?? true,
    showError: config.meta?.showError ?? true,
    dedupe: config.meta?.dedupe ?? false,
    retry: config.meta?.retry ?? 0,
    retryDelay: config.meta?.retryDelay ?? DEFAULT_RETRY_DELAY,
    returnEnvelope: config.meta?.returnEnvelope,
    returnResponse: config.meta?.returnResponse,
    requestKey: config.meta?.requestKey,
  }
}

function isApiEnvelope(value: unknown): value is ApiEnvelope {
  if (!value || typeof value !== 'object') return false

  const record = value as Record<string, unknown>

  return 'code' in record || 'success' in record
}

function isSuccessEnvelope(envelope: ApiEnvelope) {
  if (typeof envelope.success === 'boolean') {
    return envelope.success
  }

  return SUCCESS_CODES.includes(envelope.code)
}

function getStringField(record: Record<string, unknown>, key: 'msg' | 'message' | 'error') {
  const value = record[key]

  return typeof value === 'string' && value.trim() ? value.trim() : ''
}

export function getErrorPayloadMessage(payload: unknown) {
  if (typeof payload === 'string') {
    const trimmed = payload.trim()
    if (!trimmed) return ''

    try {
      return getErrorPayloadMessage(JSON.parse(trimmed))
    } catch {
      return trimmed
    }
  }

  if (!payload || typeof payload !== 'object') return ''

  const record = payload as Record<string, unknown>

  return (
    getStringField(record, 'msg') ||
    getStringField(record, 'message') ||
    getStringField(record, 'error')
  )
}

function getEnvelopeMessage(envelope: ApiEnvelope) {
  return getErrorPayloadMessage(envelope) || i18n.global.t('error.requestFailed')
}

function getHttpErrorMessage(error: AxiosError) {
  const backendMessage = getErrorPayloadMessage(error.response?.data)

  if (backendMessage) return backendMessage

  if (error.code === 'ECONNABORTED') return i18n.global.t('error.timeout')
  if (error.response?.status === 401) return i18n.global.t('error.unauthorized')
  if (error.response?.status === 403) return i18n.global.t('error.forbidden')
  if (error.response?.status === 404) return i18n.global.t('error.notFound')
  if (error.response?.status === 429) return i18n.global.t('error.rateLimit')
  if (error.response && error.response.status >= 500) return i18n.global.t('error.server')

  return error.message || i18n.global.t('error.network')
}

function toPlainPayload(data: unknown) {
  if (data instanceof FormData) return '[form-data]'
  if (data instanceof Blob) return '[blob]'
  return data
}

function createRequestKey(config: HttpRequestConfig) {
  if (config.meta?.requestKey) return config.meta.requestKey

  const method = (config.method || 'GET').toUpperCase()
  const url = `${config.baseURL || ''}${config.url || ''}`

  return JSON.stringify({
    method,
    url,
    params: config.params,
    data: toPlainPayload(config.data),
  })
}

function clearPending(key?: string) {
  if (!key) return
  pendingControllers.delete(key)
}

function setupDedupe(config: HttpRequestConfig) {
  const meta = getMeta(config)
  if (!meta.dedupe) return config

  const key = createRequestKey(config)
  const previous = pendingControllers.get(key)

  if (previous) {
    previous.abort()
    pendingControllers.delete(key)
  }

  const controller = new AbortController()
  pendingControllers.set(key, controller)
  config.signal = controller.signal
  config.meta = {
    ...config.meta,
    requestKey: key,
  }

  return config
}

function clearAuthAndRedirect() {
  notifyAuthExpired()
  clearAuthToken()

  if (typeof window === 'undefined') return

  const current = `${window.location.pathname}${window.location.search}${window.location.hash}`
  const loginHash = `#/profile?auth=login&redirect=${encodeURIComponent(current)}`

  if (!window.location.hash.startsWith('#/profile?auth=login')) {
    window.location.hash = loginHash
  }
}

function normalizeAxiosError(error: AxiosError) {
  const isCanceled = axios.isCancel(error) || error.code === 'ERR_CANCELED'
  const normalized = new HttpError({
    message: isCanceled ? '请求已取消' : getHttpErrorMessage(error),
    status: error.response?.status,
    response: error.response,
    data: error.response?.data,
    isCanceled,
  })

  if (error.response?.status === 401) {
    clearAuthAndRedirect()
  }

  void trackApiError({
    url: error.config?.url,
    method: error.config?.method,
    status: error.response?.status,
    code: error.code,
    message: normalized.message,
  })

  return normalized
}

function shouldRetry(error: unknown) {
  if (!(error instanceof HttpError)) return false
  if (error.isBusinessError || error.isCanceled) return false
  if (!error.status) return true

  return error.status >= 500
}

function sleep(ms: number) {
  return new Promise(resolve => window.setTimeout(resolve, ms))
}

function notifyError(error: unknown, showError: boolean) {
  if (!showError) return
  if (error instanceof HttpError && error.isCanceled) return

  const message = error instanceof Error ? error.message : i18n.global.t('error.requestFailed')
  showFailToast(message)
}

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const typedConfig = config as HttpRequestConfig
  const meta = getMeta(typedConfig)

  if (meta.auth) {
    const token = getAuthToken()

    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
  }

  config.headers['X-Requested-With'] = 'XMLHttpRequest'

  return setupDedupe(typedConfig) as InternalAxiosRequestConfig
})

http.interceptors.response.use(
  response => {
    clearPending((response.config as HttpRequestConfig).meta?.requestKey)
    return response
  },
  (error: AxiosError) => {
    clearPending((error.config as HttpRequestConfig | undefined)?.meta?.requestKey)
    return Promise.reject(normalizeAxiosError(error))
  },
)

function handleResponse<T>(
  response: AxiosResponse<ApiEnvelope<T> | T>,
  config: HttpRequestConfig,
) {
  const body = response.data

  if (getMeta(config).returnResponse) {
    return response
  }

  if (isApiEnvelope(body)) {
    if (!isSuccessEnvelope(body)) {
      void trackApiError({
        url: response.config.url,
        method: response.config.method,
        status: response.status,
        code: body.code,
        message: getEnvelopeMessage(body),
        business: true,
      })

      throw new HttpError({
        message: getEnvelopeMessage(body),
        code: body.code,
        status: response.status,
        data: body.data,
        response,
        isBusinessError: true,
      })
    }

    if (getMeta(config).returnEnvelope) {
      return body
    }

    return body.data
  }

  return body
}

async function requestWithRetry<T>(config: HttpRequestConfig): Promise<T> {
  const meta = getMeta(config)
  let attempt = 0

  while (true) {
    try {
      const response = await http.request<ApiEnvelope<T> | T>(config)
      return handleResponse<T>(response, config) as T
    } catch (error) {
      if (attempt >= meta.retry || !shouldRetry(error)) {
        notifyError(error, meta.showError)
        throw error
      }

      attempt += 1
      await sleep(meta.retryDelay * attempt)
    }
  }
}

export function request<T = unknown>(config: HttpRequestConfig): Promise<T> {
  return requestWithRetry<T>(config)
}

export const get = <T = unknown>(url: string, config?: HttpRequestConfig) =>
  request<T>({ ...config, url, method: 'GET' })

export const post = <T = unknown>(
  url: string,
  data?: unknown,
  config?: HttpRequestConfig,
) => request<T>({ ...config, url, data, method: 'POST' })

export const put = <T = unknown>(
  url: string,
  data?: unknown,
  config?: HttpRequestConfig,
) => request<T>({ ...config, url, data, method: 'PUT' })

export const patch = <T = unknown>(
  url: string,
  data?: unknown,
  config?: HttpRequestConfig,
) => request<T>({ ...config, url, data, method: 'PATCH' })

export const del = <T = unknown>(url: string, config?: HttpRequestConfig) =>
  request<T>({ ...config, url, method: 'DELETE' })

export function upload<T = unknown>(url: string, options: UploadFileOptions) {
  const formData = new FormData()
  formData.append(options.fieldName ?? 'file', options.file)

  Object.entries(options.data ?? {}).forEach(([key, value]) => {
    formData.append(key, value)
  })

  return request<T>({
    ...options.config,
    url,
    data: formData,
    method: 'POST',
    headers: {
      ...options.config?.headers,
      'Content-Type': 'multipart/form-data',
    },
  })
}

export function download(
  url: string,
  data?: unknown,
  method: Method = 'GET',
  config?: HttpRequestConfig,
) {
  return request<Blob>({
    ...config,
    url,
    data,
    method,
    responseType: 'blob',
    meta: {
      ...config?.meta,
      returnEnvelope: false,
    },
  })
}

export { http }
