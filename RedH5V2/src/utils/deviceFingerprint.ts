import FingerprintJS from '@fingerprintjs/fingerprintjs'

export const DEVICE_FINGERPRINT_STORAGE_KEY = 'ppmx_device_fingerprint_v1'

let pendingFingerprint: Promise<string> | null = null

function getLocalStorage() {
  if (typeof window === 'undefined') return null

  try {
    return window.localStorage
  } catch {
    return null
  }
}

function normalizeFingerprint(value: string) {
  return value.trim()
}

function getCachedFingerprint() {
  const storage = getLocalStorage()
  if (!storage) return ''

  return normalizeFingerprint(storage.getItem(DEVICE_FINGERPRINT_STORAGE_KEY) ?? '')
}

function cacheFingerprint(value: string) {
  const fingerprint = normalizeFingerprint(value)
  const storage = getLocalStorage()

  if (storage && fingerprint) {
    storage.setItem(DEVICE_FINGERPRINT_STORAGE_KEY, fingerprint)
  }

  return fingerprint
}

function createFallbackFingerprint() {
  const cryptoApi = typeof window !== 'undefined' ? window.crypto : undefined
  const randomId =
    typeof cryptoApi?.randomUUID === 'function'
      ? cryptoApi.randomUUID()
      : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`

  return `local-${randomId}`
}

async function resolveDeviceFingerprint() {
  const cached = getCachedFingerprint()
  if (cached) return cached

  try {
    const agent = await FingerprintJS.load()
    const result = await agent.get()
    const visitorId = normalizeFingerprint(result.visitorId)

    if (visitorId) {
      return cacheFingerprint(visitorId)
    }
  } catch {
    // Browser fingerprinting is a risk signal; registration can still use a stable local fallback.
  }

  return cacheFingerprint(createFallbackFingerprint())
}

export async function getDeviceFingerprint() {
  if (!pendingFingerprint) {
    pendingFingerprint = resolveDeviceFingerprint().finally(() => {
      pendingFingerprint = null
    })
  }

  return pendingFingerprint
}
