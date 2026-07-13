interface TtlStoragePayload<T> {
  value: T
  expiresAt: number | null
}

type StorageArea = 'local' | 'session'

function now() {
  return Date.now()
}

function getStorage(area: StorageArea = 'local') {
  if (typeof window === 'undefined') return null

  return area === 'session' ? window.sessionStorage : window.localStorage
}

function setStorageValue<T>(area: StorageArea, key: string, value: T, ttlMs?: number) {
  const storage = getStorage(area)
  if (!storage) return

  const payload: TtlStoragePayload<T> = {
    value,
    expiresAt: ttlMs ? now() + ttlMs : null,
  }

  storage.setItem(key, JSON.stringify(payload))
}

function getStorageValue<T>(area: StorageArea, key: string): T | null {
  const storage = getStorage(area)
  if (!storage) return null

  const raw = storage.getItem(key)
  if (!raw) return null

  try {
    const payload = JSON.parse(raw) as TtlStoragePayload<T>

    if (payload.expiresAt && payload.expiresAt <= now()) {
      storage.removeItem(key)
      return null
    }

    return payload.value
  } catch {
    storage.removeItem(key)
    return null
  }
}

export function setTtlStorage<T>(key: string, value: T, ttlMs?: number) {
  setStorageValue('local', key, value, ttlMs)
}

export function getTtlStorage<T>(key: string): T | null {
  return getStorageValue<T>('local', key)
}

export function removeTtlStorage(key: string) {
  getStorage()?.removeItem(key)
}

export function setSessionTtlStorage<T>(key: string, value: T, ttlMs?: number) {
  setStorageValue('session', key, value, ttlMs)
}

export function getSessionTtlStorage<T>(key: string): T | null {
  return getStorageValue<T>('session', key)
}

export function removeSessionTtlStorage(key: string) {
  getStorage('session')?.removeItem(key)
}

export function clearExpiredTtlStorage(prefix?: string) {
  const storage = getStorage()
  if (!storage) return

  for (let index = storage.length - 1; index >= 0; index -= 1) {
    const key = storage.key(index)
    if (!key) continue
    if (prefix && !key.startsWith(prefix)) continue

    const raw = storage.getItem(key)
    if (!raw) continue

    try {
      const payload = JSON.parse(raw) as TtlStoragePayload<unknown>

      if (payload.expiresAt && payload.expiresAt <= now()) {
        storage.removeItem(key)
      }
    } catch {
      // Normal localStorage entries are not TTL payloads.
    }
  }
}
