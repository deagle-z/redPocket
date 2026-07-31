import { describe, expect, it } from 'vitest'

import type { AppPopupAnnouncement } from '@/api/banner'
import {
  filterUnseenPopupAnnouncements,
  getAnnouncementPopupSeenKey,
  getMillisecondsUntilNextDay,
} from './announcementPopup'

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
  it('scopes the seen key to the user and local calendar day', () => {
    expect(
      getAnnouncementPopupSeenKey(
        38,
        new Date(2026, 6, 30, 14, 20, 0),
      ),
    ).toBe('announcement_popup_seen:38:2026-07-30')
  })

  it('expires at the next local calendar day', () => {
    expect(
      getMillisecondsUntilNextDay(new Date(2026, 6, 30, 23, 59, 0)),
    ).toBe(60_000)
  })

  it('keeps only active unseen popup items in backend order', () => {
    const announcements = [
      createAnnouncement(4, 20),
      createAnnouncement(3, 10),
      createAnnouncement(2, 10),
      createAnnouncement(1, 0, { status: 0 }),
      createAnnouncement(5, 0, { position: 'home' }),
    ]

    expect(
      filterUnseenPopupAnnouncements(announcements, [2]).map(
        (announcement) => announcement.id,
      ),
    ).toEqual([3, 4])
  })
})
