import type { AppPopupAnnouncement } from '@/api/banner'
import { getTtlStorage, setTtlStorage } from '@/utils/storage'

export const ANNOUNCEMENT_POPUP_SEEN_PREFIX = 'announcement_popup_seen'

function getLocalDateKey(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')

  return `${year}-${month}-${day}`
}

export function getAnnouncementPopupSeenKey(userId: number | string, date = new Date()) {
  return `${ANNOUNCEMENT_POPUP_SEEN_PREFIX}:${userId}:${getLocalDateKey(date)}`
}

export function getMillisecondsUntilNextDay(date = new Date()) {
  const nextDay = new Date(
    date.getFullYear(),
    date.getMonth(),
    date.getDate() + 1,
  )

  return Math.max(1, nextDay.getTime() - date.getTime())
}

export function getSeenAnnouncementPopupIds(userId: number | string, date = new Date()) {
  const storedIds = getTtlStorage<unknown>(
    getAnnouncementPopupSeenKey(userId, date),
  )

  if (!Array.isArray(storedIds)) return []

  return storedIds.filter(
    (id): id is number => typeof id === 'number' && Number.isInteger(id),
  )
}

export function markAnnouncementPopupSeen(
  userId: number | string,
  announcementId: number,
  date = new Date(),
) {
  const seenIds = getSeenAnnouncementPopupIds(userId, date)

  if (!seenIds.includes(announcementId)) {
    seenIds.push(announcementId)
  }

  setTtlStorage(
    getAnnouncementPopupSeenKey(userId, date),
    seenIds,
    getMillisecondsUntilNextDay(date),
  )
}

export function filterUnseenPopupAnnouncements(
  announcements: AppPopupAnnouncement[],
  seenIds: number[],
) {
  const seenIdSet = new Set(seenIds)

  return announcements
    .filter(
      (announcement) =>
        announcement.status === 1 &&
        announcement.position === 'popup' &&
        !seenIdSet.has(announcement.id),
    )
    .sort((left, right) => left.sort - right.sort || right.id - left.id)
}
