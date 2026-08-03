import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/utils/tracker', () => ({
  trackEvent: vi.fn(),
}))

vi.mock('@/utils/facebookPixel', () => ({
  getFacebookPixelId: () => 'pixel-1',
}))

vi.mock('@/utils/sourceChannel', () => ({
  getSourceChannelCode: () => 'channel-1',
}))

import { trackAttributionPageView } from '@/utils/attribution'

function createStorage() {
  const values = new Map<string, string>()
  return {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => values.set(key, value),
    removeItem: (key: string) => values.delete(key),
    clear: () => values.clear(),
    key: (index: number) => [...values.keys()][index] ?? null,
    get length() {
      return values.size
    },
  }
}

describe('trackAttributionPageView', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createStorage())
    vi.stubGlobal('sessionStorage', createStorage())
    vi.stubGlobal('window', {
      location: { href: 'https://promo.example.com/#/games' },
    })
    vi.stubGlobal('document', {
      referrer: 'https://source.example/',
      title: 'Games',
    })
    vi.stubGlobal('crypto', {
      randomUUID: vi.fn()
        .mockReturnValueOnce('visitor-id')
        .mockReturnValueOnce('session-id'),
    })
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true }))
  })

  it('posts one backend page_view with visitor and route data', async () => {
    await trackAttributionPageView('/games')

    expect(fetch).toHaveBeenCalledTimes(1)
    const [url, options] = vi.mocked(fetch).mock.calls[0]!
    expect(url).toBe('/api/v1/app/attribution/event')
    expect(options?.method).toBe('POST')
    expect(JSON.parse(String(options?.body))).toMatchObject({
      eventName: 'page_view',
      visitorId: 'visitor_visitor-id',
      sessionId: 'session_session-id',
      pageUrl: 'https://promo.example.com/#/games',
      referrer: 'https://source.example/',
      metadata: {
        path: '/games',
        title: 'Games',
      },
    })
  })
})
