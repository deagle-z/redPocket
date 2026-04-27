import { STORAGE_TOKEN_KEY } from '@/stores/mutation-type'
import { getSourceChannelCode } from '@/utils/source-channel'

const VISITOR_ID_STORAGE_KEY = 'attribution_visitor_id'
const SESSION_ID_STORAGE_KEY = 'attribution_session_id'
const ATTRIBUTION_ENDPOINT = '/api/v1/app/attribution/event'

export interface AttributionEventPayload {
  eventName: string
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

function buildAttributionPayload(payload: AttributionEventPayload) {
  return {
    eventName: normalizeEventName(payload.eventName),
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
  }).catch(() => {})
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
  getSessionId,
  getVisitorId,
  trackAttributionEvent,
  trackPageView,
}
