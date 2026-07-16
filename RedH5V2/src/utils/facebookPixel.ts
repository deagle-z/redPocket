import { APP_CURRENCY } from '@/config/market'

export const FACEBOOK_PIXEL_ID_KEY = 'facebook_pixel_id'

declare global {
  interface Window {
    fbq?: FacebookPixelQueue
    _fbq?: FacebookPixelQueue
  }
}

type FacebookPixelQueue = {
  (...args: unknown[]): void
  callMethod?: (...args: unknown[]) => void
  queue?: unknown[]
  loaded?: boolean
  version?: string
  push?: FacebookPixelQueue
}

type RouteQueryLike = Record<string, unknown>

const PIXEL_SCRIPT_ID = 'facebook-pixel-sdk'

function getBrowserStorage() {
  if (typeof localStorage === 'undefined') return null
  return localStorage
}

function readQueryValue(query: RouteQueryLike, key: string) {
  const value = query[key]
  const firstValue = Array.isArray(value) ? value[0] : value

  return typeof firstValue === 'string' ? firstValue.trim() : ''
}

export function normalizeFacebookPixelId(value?: string | null) {
  return value?.trim() ?? ''
}

export function captureFacebookPixelId(query: RouteQueryLike) {
  const pixelId = normalizeFacebookPixelId(readQueryValue(query, 'fbId'))
  const storage = getBrowserStorage()

  if (pixelId && storage) storage.setItem(FACEBOOK_PIXEL_ID_KEY, pixelId)

  return pixelId
}

export function getFacebookPixelId() {
  return (
    normalizeFacebookPixelId(getBrowserStorage()?.getItem(FACEBOOK_PIXEL_ID_KEY)) ||
    normalizeFacebookPixelId(import.meta.env.VITE_FACEBOOK_PIXEL_ID)
  )
}

function ensureFacebookQueue() {
  if (typeof window === 'undefined') return null
  if (window.fbq) return window.fbq

  const fbq = function facebookPixelQueue(...args: unknown[]) {
    if (fbq.callMethod) {
      fbq.callMethod(...args)
      return
    }

    fbq.queue?.push(args)
  } as FacebookPixelQueue

  fbq.queue = []
  fbq.loaded = true
  fbq.version = '2.0'
  fbq.push = fbq

  window.fbq = fbq
  window._fbq = fbq

  return fbq
}

function injectFacebookPixelScript() {
  if (typeof document === 'undefined') return
  if (document.getElementById(PIXEL_SCRIPT_ID)) return

  const script = document.createElement('script')
  script.id = PIXEL_SCRIPT_ID
  script.async = true
  script.src = 'https://connect.facebook.net/en_US/fbevents.js'
  document.head.appendChild(script)
}

export function initFacebookPixel() {
  const pixelId = getFacebookPixelId()
  if (!pixelId) return ''

  const fbq = ensureFacebookQueue()
  if (!fbq) return pixelId

  injectFacebookPixelScript()
  fbq('init', pixelId)
  fbq('track', 'PageView')

  return pixelId
}

export function trackFacebookPixelEvent(
  eventName: string,
  params: Record<string, unknown> = {},
  eventId?: string,
) {
  const pixelId = initFacebookPixel()
  if (!pixelId || typeof window === 'undefined' || !window.fbq) return false

  if (eventId) {
    window.fbq('track', eventName, params, { eventID: eventId })
  } else {
    window.fbq('track', eventName, params)
  }

  return true
}

export function trackCompleteRegistration(eventId?: string) {
  return trackFacebookPixelEvent('CompleteRegistration', {
    content_name: 'PP.MX Register',
    status: 'completed',
  }, eventId)
}

export interface FacebookPurchasePayload {
  orderNo: string
  amount: number
  currency?: string
  eventId?: string
}

export function trackPurchase(payload: FacebookPurchasePayload) {
  const orderNo = String(payload.orderNo || '').trim()
  const amount = Number(payload.amount || 0)

  if (!orderNo || !Number.isFinite(amount) || amount <= 0) return false

  return trackFacebookPixelEvent('Purchase', {
    value: amount,
    currency: payload.currency || APP_CURRENCY,
    content_name: 'PP.MX Recharge',
    orderNo,
  }, payload.eventId)
}
