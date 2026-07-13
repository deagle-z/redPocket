import { trackEvent } from '@/utils/tracker'
import { getFacebookPixelId } from '@/utils/facebookPixel'
import { getSourceChannelCode } from '@/utils/sourceChannel'

const VISITOR_ID_KEY = 'visitor_id'
const SESSION_ID_KEY = 'session_id'
const ATTRIBUTION_PATH = '/v1/app/attribution/event'

export interface AttributionEventInput {
  eventName: string
  eventId?: string
  metadata?: Record<string, unknown>
}

function createId(prefix: string) {
  const random =
    typeof crypto !== 'undefined' && crypto.randomUUID
      ? crypto.randomUUID()
      : `${Date.now()}-${Math.random().toString(16).slice(2)}`

  return `${prefix}_${random}`
}

function getLocalStorageValue(key: string, fallbackPrefix: string) {
  if (typeof localStorage === 'undefined') return createId(fallbackPrefix)

  const existing = localStorage.getItem(key)
  if (existing) return existing

  const value = createId(fallbackPrefix)
  localStorage.setItem(key, value)

  return value
}

function getSessionStorageValue(key: string, fallbackPrefix: string) {
  if (typeof sessionStorage === 'undefined') return createId(fallbackPrefix)

  const existing = sessionStorage.getItem(key)
  if (existing) return existing

  const value = createId(fallbackPrefix)
  sessionStorage.setItem(key, value)

  return value
}

function getPageUrl() {
  if (typeof window === 'undefined') return ''
  return window.location.href
}

function getReferrer() {
  if (typeof document === 'undefined') return ''
  return document.referrer
}

function getAttributionEndpoint() {
  const apiBase = import.meta.env.VITE_API_BASE_URL || '/api'
  return `${apiBase.replace(/\/$/, '')}${ATTRIBUTION_PATH}`
}

export function getVisitorId() {
  return getLocalStorageValue(VISITOR_ID_KEY, 'visitor')
}

export function getSessionId() {
  return getSessionStorageValue(SESSION_ID_KEY, 'session')
}

export function createThirdPartyEventId(eventName: string) {
  return `${eventName}_${Date.now()}_${Math.random().toString(16).slice(2)}`
}

export async function trackAttributionEvent(input: AttributionEventInput) {
  const eventId = input.eventId || createThirdPartyEventId(input.eventName)
  const payload = {
    eventName: input.eventName,
    eventId,
    pixelId: getFacebookPixelId(),
    sourceChannelCode: getSourceChannelCode(),
    visitorId: getVisitorId(),
    sessionId: getSessionId(),
    pageUrl: getPageUrl(),
    referrer: getReferrer(),
    metadata: input.metadata ?? {},
    timestamp: Date.now(),
  }

  void trackEvent({
    eventName: input.eventName,
    eventType: 'conversion',
    payload,
  })

  if (typeof fetch === 'undefined') return payload

  try {
    await fetch(getAttributionEndpoint(), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
      keepalive: true,
    })
  } catch {
    // Attribution failure must never block login or registration.
  }

  return payload
}
