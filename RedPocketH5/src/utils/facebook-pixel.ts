const FACEBOOK_PIXEL_STORAGE_KEY = 'facebook_pixel_id'
const FACEBOOK_PIXEL_PURCHASE_SENT_PREFIX = 'facebook_pixel_purchase_sent_'
const FACEBOOK_PIXEL_SCRIPT_ID = 'facebook-pixel-script'

type FacebookPixelFunction = {
  (...args: unknown[]): void
  callMethod?: (...args: unknown[]) => void
  queue?: unknown[]
  push?: FacebookPixelFunction
  loaded?: boolean
  version?: string
}

declare global {
  interface Window {
    fbq?: FacebookPixelFunction
    _fbq?: FacebookPixelFunction
  }
}

let initializedPixelId = ''
let pageViewPixelId = ''

function normalizeFacebookPixelId(value: unknown): string {
  const raw = Array.isArray(value) ? value[0] : value
  const pixelId = String(raw || '').trim()
  if (!pixelId || pixelId.length > 64)
    return ''
  return pixelId
}

function getFacebookPixelId() {
  return normalizeFacebookPixelId(localStorage.getItem(FACEBOOK_PIXEL_STORAGE_KEY))
}

function setFacebookPixelId(value: unknown) {
  const pixelId = normalizeFacebookPixelId(value)
  if (!pixelId)
    return ''

  localStorage.setItem(FACEBOOK_PIXEL_STORAGE_KEY, pixelId)
  return pixelId
}

function captureFacebookPixelId(query: Record<string, unknown>) {
  return setFacebookPixelId(query.fbId)
}

function installFacebookPixelScript() {
  if (typeof window === 'undefined' || typeof document === 'undefined')
    return false

  if (!window.fbq) {
    const fbq = function (...args: unknown[]) {
      if (fbq.callMethod)
        fbq.callMethod(...args)
      else
        fbq.queue?.push(args)
    } as FacebookPixelFunction

    fbq.push = fbq
    fbq.loaded = true
    fbq.version = '2.0'
    fbq.queue = []
    window.fbq = fbq
    window._fbq = fbq
  }

  if (!document.getElementById(FACEBOOK_PIXEL_SCRIPT_ID)) {
    const script = document.createElement('script')
    script.id = FACEBOOK_PIXEL_SCRIPT_ID
    script.async = true
    script.src = 'https://connect.facebook.net/en_US/fbevents.js'
    const firstScript = document.getElementsByTagName('script')[0]
    if (firstScript?.parentNode)
      firstScript.parentNode.insertBefore(script, firstScript)
    else
      document.head.appendChild(script)
  }

  return true
}

function initFacebookPixel() {
  const pixelId = getFacebookPixelId()
  if (!pixelId || !installFacebookPixelScript())
    return false

  if (initializedPixelId !== pixelId) {
    window.fbq?.('init', pixelId)
    initializedPixelId = pixelId
  }

  if (pageViewPixelId !== pixelId) {
    window.fbq?.('track', 'PageView')
    pageViewPixelId = pixelId
  }

  return true
}

function trackFacebookPixelEvent(eventName: string, params: Record<string, unknown> = {}) {
  if (!eventName || !initFacebookPixel())
    return false

  window.fbq?.('track', eventName, params)
  return true
}

function getFacebookPurchaseSentKey(orderNo: string) {
  return `${FACEBOOK_PIXEL_PURCHASE_SENT_PREFIX}${orderNo}`
}

function hasTrackedFacebookPurchase(orderNo: string) {
  const normalizedOrderNo = String(orderNo || '').trim()
  if (!normalizedOrderNo)
    return true
  return localStorage.getItem(getFacebookPurchaseSentKey(normalizedOrderNo)) === '1'
}

function trackFirstRechargePurchase(params: {
  orderNo: string
  amount: number
  currency?: string
}) {
  const orderNo = String(params.orderNo || '').trim()
  if (!orderNo || hasTrackedFacebookPurchase(orderNo))
    return false

  const tracked = trackFacebookPixelEvent('Purchase', {
    value: Number(params.amount || 0),
    currency: String(params.currency || 'BRL').trim() || 'BRL',
    order_id: orderNo,
    content_name: 'first_recharge',
  })

  if (tracked)
    localStorage.setItem(getFacebookPurchaseSentKey(orderNo), '1')
  return tracked
}

export {
  FACEBOOK_PIXEL_STORAGE_KEY,
  captureFacebookPixelId,
  getFacebookPixelId,
  initFacebookPixel,
  trackFacebookPixelEvent,
  trackFirstRechargePurchase,
}
