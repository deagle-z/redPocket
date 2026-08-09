export type LocaleCode = 'es' | 'en' | 'zh'
export type PpmxAuthMode = 'login' | 'register'
export type PpmxRouteTarget =
  | 'home'
  | 'wheel'
  | 'team'
  | 'profile'
  | 'records'
  | 'security'
  | 'vip'
  | 'account'
  | 'promo'
  | 'casino'
  | 'download'
  | 'guide'
  | 'help'
  | 'responsible'

export interface LocalizedText {
  es: string
  en: string
  zh: string
}

export interface PpmxNavItem {
  id: string
  icon: string
  label: LocalizedText
  target: string
}

export interface PpmxNotice {
  id: string
  icon: string
  text: LocalizedText
  target?: string
}

export interface PpmxBanner {
  id: string
  tag: LocalizedText
  title: LocalizedText
  sub: LocalizedText
  cta: LocalizedText
  background: string
  image?: string
  backgroundPosition?: string
  amountHtml?: LocalizedText
  benefitHtml?: LocalizedText
  valid?: LocalizedText
  howTo?: LocalizedText
  terms?: LocalizedText
  thumbs?: number[]
}

export interface PpmxPromoItem {
  id: string
  tag: LocalizedText
  title: LocalizedText
  sub: LocalizedText
  howTo: LocalizedText
  valid: LocalizedText
  terms: LocalizedText
  tone: 'red' | 'gold' | 'green' | 'violet' | 'blue' | 'pink'
  backgroundImage?: string
  backgroundPosition?: string
}

export interface PpmxGame {
  id: string
  name: LocalizedText
  imageId?: number
  cover?: string
  icon?: string
  brand?: string
  subtitle?: string
  gameId?: number
  categoryCode?: string
  fakeOnlineCount?: number
  hot?: boolean
  isNew?: boolean
}

export interface PpmxHomeCategory {
  id: string
  icon: string
  name: LocalizedText
  games: PpmxGame[]
  target?: string
  wheel?: boolean
}

export interface PpmxTrustItem {
  id: string
  icon: string
  title: LocalizedText
  description: LocalizedText
}

export interface PpmxFooterLink {
  id: string
  label: LocalizedText
  target: string
}

export interface PpmxFooterColumn {
  id: string
  title: LocalizedText
  links: PpmxFooterLink[]
}

export interface PpmxSocialItem {
  id: string
  icon: string
  label: string
  target: string
  tone: 'instagram' | 'facebook' | 'x' | 'youtube' | 'telegram' | 'tiktok' | 'default'
}

export interface PpmxFloatingWidget {
  id: string
  icon: string
  label: LocalizedText
  target: string
}

export interface PpmxCategoryBrand {
  id: string
  name: LocalizedText
  icon?: string
  games: PpmxGame[]
}

export interface PpmxWheelPrize {
  amount: number
  color: string
  gold?: boolean
  weight: number
}

export interface PpmxTeamKpi {
  id: string
  icon: string
  tone: 'gold' | 'green' | 'red' | 'blue'
  value: string
  label: LocalizedText
}

export interface PpmxAccountHubItem {
  id: string
  icon: string
  tone: 'green' | 'blue' | 'gold' | 'red' | 'teal' | 'violet'
  label: LocalizedText
  description: LocalizedText
  target?: PpmxRouteTarget | string
}

export interface PpmxProfileField {
  id: string
  label: LocalizedText
  model: 'firstName' | 'email' | 'phone'
  type?: 'email' | 'tel' | 'text'
}

export interface PpmxDownloadFeature {
  id: string
  icon: string
  title: LocalizedText
  description: LocalizedText
}

export interface PpmxDownloadStep {
  id: string
  title: LocalizedText
  description: LocalizedText
}

export interface PpmxGuideCard {
  id: string
  icon: string
  tagIcon: string
  tone: 'green' | 'gold' | 'blue' | 'violet' | 'red'
  title: LocalizedText
  description: LocalizedText
  tag: LocalizedText
}
