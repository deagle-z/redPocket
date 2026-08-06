import { beforeEach, describe, expect, it, vi } from 'vitest'

import { post } from '@/utils/http'
import { getPopupAnnouncements } from './banner'

vi.mock('@/utils/http', () => ({
  post: vi.fn(async () => ({ popup: [] })),
}))

describe('popup announcement API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('requests only the H5 popup position for the configured market', async () => {
    await getPopupAnnouncements('es-PE')

    expect(post).toHaveBeenCalledWith(
      '/v1/app/banners',
      {
        platform: 'h5',
        position: 'popup',
        lang: 'es-PE',
        countryCode: 'PE',
      },
      {
        meta: {
          auth: false,
          dedupe: true,
        },
      },
    )
  })
})
