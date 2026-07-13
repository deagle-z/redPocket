import type { TrackEventPayload, TrackEventType } from '@/types/business'

const QUEUE_KEY = 'track_event_queue'
const MAX_QUEUE_SIZE = 100

type TrackInput = Omit<
  TrackEventPayload,
  'route' | 'timestamp' | 'releaseVersion'
> & {
  route?: string
  timestamp?: number
  releaseVersion?: string
}

function getEndpoint() {
  return import.meta.env.VITE_TRACKING_ENDPOINT
}

function getRoute() {
  if (typeof window === 'undefined') return '/'
  return `${window.location.pathname}${window.location.search}${window.location.hash}`
}

function readQueue() {
  try {
    return JSON.parse(localStorage.getItem(QUEUE_KEY) || '[]') as TrackEventPayload[]
  } catch {
    return []
  }
}

function writeQueue(events: TrackEventPayload[]) {
  localStorage.setItem(QUEUE_KEY, JSON.stringify(events.slice(-MAX_QUEUE_SIZE)))
}

function queueEvent(event: TrackEventPayload) {
  if (typeof localStorage === 'undefined') return
  writeQueue([...readQueue(), event])
}

function normalizeEvent(input: TrackInput): TrackEventPayload {
  return {
    route: getRoute(),
    timestamp: Date.now(),
    releaseVersion: import.meta.env.VITE_RELEASE_VERSION,
    ...input,
  }
}

async function sendEvents(events: TrackEventPayload[]) {
  const endpoint = getEndpoint()

  if (!endpoint) {
    console.info('[track]', events)
    return true
  }

  const body = JSON.stringify({ events })

  if (navigator.sendBeacon) {
    const sent = navigator.sendBeacon(
      endpoint,
      new Blob([body], { type: 'application/json' }),
    )

    if (sent) return true
  }

  const response = await fetch(endpoint, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body,
    keepalive: true,
  })

  return response.ok
}

export async function trackEvent(input: TrackInput) {
  const event = normalizeEvent(input)

  try {
    const sent = await sendEvents([event])
    if (!sent) queueEvent(event)
  } catch {
    queueEvent(event)
  }
}

export function trackPageView(route?: string) {
  return trackEvent({
    eventName: 'page_view',
    eventType: 'page_view',
    route,
  })
}

export function trackApiError(payload: Record<string, unknown>) {
  return trackEvent({
    eventName: 'api_error',
    eventType: 'api_error',
    payload,
  })
}

export async function flushTrackQueue() {
  if (typeof localStorage === 'undefined') return
  const queue = readQueue()
  if (queue.length === 0) return

  try {
    const sent = await sendEvents(queue)
    if (sent) writeQueue([])
  } catch {
    writeQueue(queue)
  }
}

export function initTracker() {
  void flushTrackQueue()

  window.addEventListener('error', event => {
    void trackEvent({
      eventName: 'frontend_error',
      eventType: 'frontend_error',
      payload: {
        message: event.message,
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
      },
    })
  })

  window.addEventListener('unhandledrejection', event => {
    void trackEvent({
      eventName: 'frontend_error',
      eventType: 'frontend_error',
      payload: {
        reason:
          event.reason instanceof Error ? event.reason.message : String(event.reason),
      },
    })
  })

  window.setTimeout(() => {
    const root = document.querySelector('#app')
    if (!root || root.childElementCount === 0) {
      void trackEvent({
        eventName: 'white_screen',
        eventType: 'white_screen' as TrackEventType,
      })
    }
  }, 3000)
}
