import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  captureFacebookPixelId,
  getFacebookPixelId,
  trackCompleteRegistration,
  trackPurchase,
} from './facebookPixel'

function installBrowserGlobals() {
  const data = new Map<string, string>()
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
  const head = {
    appendChild: vi.fn(),
  }
  const document = {
    getElementById: vi.fn(() => null),
    createElement: vi.fn(() => ({ async: false, id: '', src: '' })),
    head,
  }
  const window = {}

  vi.stubGlobal('localStorage', localStorage)
  vi.stubGlobal('document', document)
  vi.stubGlobal('window', window)

  return { document, head, window }
}

describe('facebook pixel utilities', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
  })

  it('uses fbId query value before env fallback', () => {
    installBrowserGlobals()

    expect(captureFacebookPixelId({ fbId: '123456' })).toBe('123456')
    expect(getFacebookPixelId()).toBe('123456')
  })

  it('does not throw when pixel is not configured', () => {
    vi.unstubAllGlobals()

    expect(() => trackCompleteRegistration('event-1')).not.toThrow()
    expect(trackCompleteRegistration('event-1')).toBe(false)
  })

  it('tracks CompleteRegistration with event id when configured', () => {
    const globals = installBrowserGlobals()
    captureFacebookPixelId({ fbId: '123456' })

    expect(trackCompleteRegistration('register-event-1')).toBe(true)
    expect(globals.head.appendChild).toHaveBeenCalledTimes(1)
    expect(globals.window).toHaveProperty('fbq')
  })

  it('tracks Purchase with order number and event id when configured', () => {
    const globals = installBrowserGlobals()
    captureFacebookPixelId({ fbId: '123456' })

    expect(trackPurchase({
      orderNo: 'RC1001',
      amount: 200,
      currency: 'USD',
      eventId: 'purchase-event-1',
    })).toBe(true)

    const fbq = (globals.window as { fbq: { queue?: unknown[] } }).fbq
    expect(fbq.queue).toContainEqual([
      'track',
      'Purchase',
      {
        value: 200,
        currency: 'USD',
        content_name: 'PP.BET Recharge',
        orderNo: 'RC1001',
      },
      { eventID: 'purchase-event-1' },
    ])
  })
})
