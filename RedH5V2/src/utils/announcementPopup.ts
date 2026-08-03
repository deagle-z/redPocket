import type { AppPopupAnnouncement } from '@/api/banner'

export function filterPopupAnnouncements(announcements: AppPopupAnnouncement[]) {
  return announcements
    .filter(
      (announcement) =>
        announcement.status === 1 &&
        announcement.position === 'popup',
    )
    .sort((left, right) => left.sort - right.sort || right.id - left.id)
}
