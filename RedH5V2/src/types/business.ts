export type MoneyValue = number | string

export type DeviceKind = 'mobile' | 'pc'

export type TransactionStatus =
  | 'pending'
  | 'processing'
  | 'success'
  | 'failed'
  | 'canceled'
  | 'expired'
  | 'reviewing'

export interface TransactionSnapshot {
  id: string
  status: TransactionStatus
  amountCent?: number
  message?: string
  updatedAt?: string
}

export interface OperationLockOptions {
  lockedMessage?: string
  minLockMs?: number
}

export interface PollingOptions {
  intervalMs?: number
  maxAttempts?: number
  immediate?: boolean
  terminalStatuses?: TransactionStatus[]
}

export type PageStateStatus =
  | 'idle'
  | 'loading'
  | 'ready'
  | 'empty'
  | 'network'
  | 'rateLimit'
  | 'maintenance'
  | 'fallback'
  | 'error'

export interface AppRuntimeConfig {
  maintenance?: boolean
  maintenanceMessage?: string
  minVersion?: string
  forceUpdate?: boolean
  forceUpdateMessage?: string
  updateUrl?: string
  grayRelease?: boolean
  grayRatio?: number
  disabledFeatures?: string[]
}

export type TrackEventType =
  | 'page_view'
  | 'conversion'
  | 'transaction'
  | 'api_error'
  | 'frontend_error'
  | 'white_screen'
  | 'pwa'

export interface TrackEventPayload {
  eventName: string
  eventType: TrackEventType
  route: string
  timestamp: number
  releaseVersion: string
  userId?: string | number
  payload?: Record<string, unknown>
}
