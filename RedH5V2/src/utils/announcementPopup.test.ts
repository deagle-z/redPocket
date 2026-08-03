import { describe, expect, it } from 'vitest'

import type { AppPopupAnnouncement } from '@/api/banner'
import { filterPopupAnnouncements } from './announcementPopup'

function createAnnouncement(
  id: number,
  sort: number,
  overrides: Partial<AppPopupAnnouncement> = {},
): AppPopupAnnouncement {
  return {
    id,
    tenantId: 1,
    bannerName: `Announcement ${id}`,
    bannerCode: null,
    position: 'popup',
    platform: 'h5',
    bannerType: 'popup',
    jumpType: 'none',
    displayType: 'popup',
    openMode: 'current',
    sort,
    status: 1,
    startTime: null,
    endTime: null,
    languageCode: 'es-MX',
    countryCode: 'MX',
    title: null,
    subTitle: null,
    description: null,
    buttonText: null,
    imageUrl: '',
    thumbUrl: null,
    bgImageUrl: null,
    iconUrl: null,
    videoUrl: null,
    jumpValue: null,
    textColor: null,
    buttonColor: null,
    bgColor: null,
    ...overrides,
  }
}

describe('announcement popup state', () => {
  it('keeps active popup items in backend order without local seen filtering', () => {
    const announcements = [
      createAnnouncement(4, 20),
      createAnnouncement(3, 10),
      createAnnouncement(2, 10),
      createAnnouncement(1, 0, { status: 0 }),
      createAnnouncement(5, 0, { position: 'home' }),
    ]

    expect(
      filterPopupAnnouncements(announcements).map(
        (announcement) => announcement.id,
      ),
    ).toEqual([3, 2, 4])
  })
})
