import { computed } from 'vue'

import { i18n } from '@/plugins/i18n'
import { getInitialLocale, persistLocale, type Locale } from '@/utils/locale'

import type { LocaleCode, LocalizedText } from '../types'

export function ppmxLocaleToAppLocale(locale: LocaleCode): Locale {
  if (locale === 'zh') return 'zh-CN'
  if (locale === 'en') return 'en-US'
  return 'es-US'
}

export function resolvePpmxLocale(locale: string): LocaleCode {
  const normalizedLocale = locale.toLowerCase()

  if (normalizedLocale.startsWith('zh')) return 'zh'
  if (normalizedLocale.startsWith('en')) return 'en'
  return 'es'
}

export function usePpmxLocale(defaultLocale: LocaleCode | string = getInitialLocale()) {
  const defaultAppLocale = ppmxLocaleToAppLocale(resolvePpmxLocale(defaultLocale))
  if (i18n.global.locale.value !== defaultAppLocale) {
    i18n.global.locale.value = defaultAppLocale
  }

  const homeLocale = computed<LocaleCode>(() => resolvePpmxLocale(i18n.global.locale.value))

  function setHomeLocale(locale: LocaleCode) {
    const appLocale = ppmxLocaleToAppLocale(locale)
    i18n.global.locale.value = appLocale
    persistLocale(appLocale)
  }

  function pickText(text: LocalizedText) {
    return text[homeLocale.value]
  }

  return {
    homeLocale,
    pickText,
    setHomeLocale,
  }
}
