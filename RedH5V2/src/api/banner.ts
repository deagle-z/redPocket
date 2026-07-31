import { APP_COUNTRY_CODE } from '@/config/market'
import { post } from '@/utils/http'

export interface AppPopupAnnouncement {
  id: number
  tenantId: number
  bannerName: string
  bannerCode: string | null
  position: string
  platform: string
  bannerType: string
  jumpType: string
  displayType: string
  openMode: string
  sort: number
  status: number
  startTime: string | null
  endTime: string | null
  languageCode: string
  countryCode: string | null
  title: string | null
  subTitle: string | null
  description: string | null
  buttonText: string | null
  imageUrl: string
  thumbUrl: string | null
  bgImageUrl: string | null
  iconUrl: string | null
  videoUrl: string | null
  jumpValue: string | null
  textColor: string | null
  buttonColor: string | null
  bgColor: string | null
}

export type AppBannerGroups = Record<string, AppPopupAnnouncement[] | undefined>

export function getPopupAnnouncements(lang: string) {
  return post<AppBannerGroups>(
    '/v1/app/banners',
    {
      platform: 'h5',
      position: 'popup',
      lang,
      countryCode: APP_COUNTRY_CODE,
    },
    {
      meta: {
        auth: false,
        dedupe: true,
      },
    },
  )
}
