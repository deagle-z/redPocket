import { beforeEach, describe, expect, it, vi } from 'vitest'

const fingerprintMock = vi.hoisted(() => ({
  load: vi.fn(),
}))

vi.mock('@fingerprintjs/fingerprintjs', () => ({
  default: {
    load: fingerprintMock.load,
  },
}))

function installWindowStorage(initial: Record<string, string> = {}) {
  const data = new Map<string, string>(Object.entries(initial))
  const localStorage = {
    get length() {
      return data.size
    },
    clear: vi.fn(() => data.clear()),
    getItem: vi.fn((key: string) => data.get(key) ?? null),
    key: vi.fn((index: number) => Array.from(data.keys())[index] ?? null),
    removeItem: vi.fn((key: string) => data.delete(key)),
    setItem: vi.fn((key: string, value: string) => data.set(key, value)),
  }
  const crypto = {
    randomUUID: vi.fn(() => 'fallback-uuid'),
  }

  vi.stubGlobal('window', { localStorage, crypto })

  return { localStorage, crypto }
}

async function loadModule() {
  vi.resetModules()
  return import('./deviceFingerprint')
}

describe('device fingerprint utility', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    vi.clearAllMocks()
    fingerprintMock.load.mockReset()
  })

  it('reuses the cached browser fingerprint without recalculating it', async () => {
    installWindowStorage({ ppmx_device_fingerprint_v1: 'cached-device-id' })
    const { getDeviceFingerprint } = await loadModule()

    await expect(getDeviceFingerprint()).resolves.toBe('cached-device-id')
    expect(fingerprintMock.load).not.toHaveBeenCalled()
  })

  it('stores the FingerprintJS visitor id for later registration submits', async () => {
    const { localStorage } = installWindowStorage()
    fingerprintMock.load.mockResolvedValue({
      get: vi.fn(async () => ({ visitorId: 'visitor-device-id' })),
    })
    const { DEVICE_FINGERPRINT_STORAGE_KEY, getDeviceFingerprint } = await loadModule()

    await expect(getDeviceFingerprint()).resolves.toBe('visitor-device-id')
    expect(localStorage.getItem(DEVICE_FINGERPRINT_STORAGE_KEY)).toBe('visitor-device-id')
  })

  it('uses a stable local fallback when FingerprintJS fails', async () => {
    const { localStorage } = installWindowStorage()
    fingerprintMock.load.mockRejectedValue(new Error('fingerprint unavailable'))
    const { DEVICE_FINGERPRINT_STORAGE_KEY, getDeviceFingerprint } = await loadModule()

    await expect(getDeviceFingerprint()).resolves.toBe('local-fallback-uuid')
    expect(localStorage.getItem(DEVICE_FINGERPRINT_STORAGE_KEY)).toBe('local-fallback-uuid')
  })
})
