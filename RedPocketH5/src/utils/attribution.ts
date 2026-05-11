import { STORAGE_TOKEN_KEY } from '@/stores/mutation-type'
import { getFacebookPixelId } from '@/utils/facebook-pixel'
import { getSourceChannelCode } from '@/utils/source-channel'

const VISITOR_ID_STORAGE_KEY = 'attribution_visitor_id'
const SESSION_ID_STORAGE_KEY = 'attribution_session_id'
const EVENT_SENT_STORAGE_PREFIX = 'attribution_event_sent_'
const ATTRIBUTION_ENDPOINT = '/api/v1/app/attribution/event'

export interface AttributionEventPayload {
  eventName: string
  thirdPartyEventId?: string
  pixelId?: string
  sourceChannelCode?: string
  visitorId?: string
  sessionId?: string
  pageUrl?: string
  referrer?: string
  metadata?: Record<string, unknown>
}

function createId(prefix: string) {
  const id = globalThis.crypto?.randomUUID?.()
    || `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 10)}`
  return `${prefix}_${id.replace(/-/g, '')}`
}

function getVisitorId() {
  const savedId = localStorage.getItem(VISITOR_ID_STORAGE_KEY)
  if (savedId)
    return savedId

  const visitorId = createId('v')
  localStorage.setItem(VISITOR_ID_STORAGE_KEY, visitorId)
  return visitorId
}

function getSessionId() {
  const savedId = sessionStorage.getItem(SESSION_ID_STORAGE_KEY)
  if (savedId)
    return savedId

  const sessionId = createId('s')
  sessionStorage.setItem(SESSION_ID_STORAGE_KEY, sessionId)
  return sessionId
}

function getAttributionEndpoint() {
  const baseURL = String(import.meta.env.VITE_APP_API_BASE_URL || '').replace(/\/$/, '')
  return `${baseURL}${ATTRIBUTION_ENDPOINT}`
}

function normalizeEventName(eventName: string) {
  return String(eventName || '').trim()
}

function createThirdPartyEventId(eventName: string, seed?: string | number) {
  const normalizedEventName = normalizeEventName(eventName).replace(/[^A-Za-z0-9_.-]/g, '_') || 'event'
  const normalizedSeed = String(seed ?? '').trim().replace(/[^A-Za-z0-9_.-]/g, '_')
  if (normalizedSeed)
    return `${normalizedEventName}_${normalizedSeed}`.slice(0, 128)
  return createId(normalizedEventName).slice(0, 128)
}

function buildAttributionPayload(payload: AttributionEventPayload) {
  const eventName = normalizeEventName(payload.eventName)
  return {
    eventName,
    thirdPartyEventId: payload.thirdPartyEventId || createThirdPartyEventId(eventName),
    pixelId: payload.pixelId || getFacebookPixelId(),
    sourceChannelCode: payload.sourceChannelCode || getSourceChannelCode(),
    visitorId: payload.visitorId || getVisitorId(),
    sessionId: payload.sessionId || getSessionId(),
    pageUrl: payload.pageUrl || window.location.href,
    referrer: payload.referrer || document.referrer || '',
    metadata: payload.metadata || {},
  }
}

function trackAttributionEvent(payload: AttributionEventPayload) {
  const body = buildAttributionPayload(payload)
  if (!body.eventName)
    return
  const sentStorageKey = body.thirdPartyEventId ? `${EVENT_SENT_STORAGE_PREFIX}${body.thirdPartyEventId}` : ''
  if (sentStorageKey && localStorage.getItem(sentStorageKey) === '1')
    return
  if (sentStorageKey)
    localStorage.setItem(sentStorageKey, '1')

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  const token = localStorage.getItem(STORAGE_TOKEN_KEY)
  if (token)
    headers.Authorization = token

  fetch(getAttributionEndpoint(), {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
    keepalive: true,
  })
    .catch(() => {
      if (sentStorageKey)
        localStorage.removeItem(sentStorageKey)
    })
}

function trackPageView(path?: string) {
  trackAttributionEvent({
    eventName: 'page_view',
    pageUrl: window.location.href,
    metadata: {
      path: path || window.location.pathname,
      title: document.title,
    },
  })
}

export {
  createThirdPartyEventId,
  getSessionId,
  getVisitorId,
  trackAttributionEvent,
  trackPageView,
}
