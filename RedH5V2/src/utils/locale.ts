import { getTtlStorage, setTtlStorage } from './storage'

export const LOCALE_STORAGE_KEY = 'locale'
export const SUPPORTED_LOCALES = ['es-MX', 'en-US', 'zh-CN'] as const
export type Locale = (typeof SUPPORTED_LOCALES)[number]

const DEFAULT_LOCALE: Locale = 'es-MX'

export function isSupportedLocale(value: unknown): value is Locale {
  return typeof value === 'string' && SUPPORTED_LOCALES.includes(value as Locale)
}

export function normalizeLocale(value: string | null | undefined): Locale | null {
  if (!value) return null

  const normalized = value.trim().replace('_', '-').toLowerCase()
  if (!normalized) return null

  if (normalized === 'zh' || normalized.startsWith('zh-')) return 'zh-CN'
  if (normalized === 'en' || normalized.startsWith('en-')) return 'en-US'
  if (normalized === 'es' || normalized.startsWith('es-')) return 'es-MX'

  return null
}

export function resolveBrowserLocale(languages?: readonly string[]): Locale {
  const browserLanguages =
    languages ??
    (typeof navigator === 'undefined'
      ? []
      : navigator.languages?.length
        ? navigator.languages
        : [navigator.language])

  for (const language of browserLanguages) {
    const locale = normalizeLocale(language)
    if (locale === 'zh-CN') continue
    if (locale) return locale
  }

  return DEFAULT_LOCALE
}

export function getStoredLocale(): Locale | null {
  const storedLocale = getTtlStorage<Locale>(LOCALE_STORAGE_KEY)
  return isSupportedLocale(storedLocale) ? storedLocale : null
}

export function getInitialLocale(languages?: readonly string[]): Locale {
  return getStoredLocale() ?? resolveBrowserLocale(languages)
}

export function persistLocale(locale: Locale) {
  setTtlStorage(LOCALE_STORAGE_KEY, locale)
}
