import { createI18n } from 'vue-i18n'
import enUS from 'vant/es/locale/lang/en-US'
import esES from 'vant/es/locale/lang/es-ES'
import idID from 'vant/es/locale/lang/id-ID'
import ptBR from 'vant/es/locale/lang/pt-BR'
import zhCN from 'vant/es/locale/lang/zh-CN'
import { Locale } from 'vant'
import type { PickerColumn } from 'vant'
import enUSMessages from '@/locales/en-US.json'
import esMXMessages from '@/locales/es-MX.json'
import idIDMessages from '@/locales/id-ID.json'
import ptBRMessages from '@/locales/pt-BR.json'
import zhCNMessages from '@/locales/zh-CN.json'

const FALLBACK_LOCALE = 'en-US'

const vantLocales = {
  'zh-CN': zhCN,
  'en-US': enUS,
  'pt-BR': ptBR,
  'es-MX': esES,
  'id-ID': idID,
}

const messages = {
  'zh-CN': zhCNMessages,
  'en-US': enUSMessages,
  'pt-BR': ptBRMessages,
  'es-MX': esMXMessages,
  'id-ID': idIDMessages,
}
type SupportedLocale = keyof typeof messages
const supportedLocales = Object.keys(messages) as SupportedLocale[]
const browserMatchLocales = supportedLocales.filter(locale => locale !== 'zh-CN')

export const languageOptions = [

  {
    code: 'US',
    value: 'en-US',
    nativeTextKey: 'login.language.enNative',
    englishTextKey: 'login.language.enEn',
  },
  {
    code: 'BR',
    value: 'pt-BR',
    nativeTextKey: 'login.language.ptNative',
    englishTextKey: 'login.language.ptEn',
  },
  {
    code: 'MX',
    value: 'es-MX',
    nativeTextKey: 'login.language.esNative',
    englishTextKey: 'login.language.esEn',
  },
  {
    code: 'ID',
    value: 'id-ID',
    nativeTextKey: 'login.language.idNative',
    englishTextKey: 'login.language.idEn',
  },
  {
    code: 'CN',
    value: 'zh-CN',
    nativeTextKey: 'login.language.zhNative',
    englishTextKey: 'login.language.zhEn',
  },
]

export const languageColumns: PickerColumn = languageOptions.map(item => ({
  text: item.value,
  value: item.value,
}))

export const i18n = setupI18n()
type I18n = typeof i18n

export const locale = computed({
  get() {
    return i18n.global.locale.value
  },
  set(language: string) {
    setLang(language, i18n)
  },
})

function setupI18n() {
  const locale = getI18nLocale()
  const i18n = createI18n({
    locale,
    fallbackLocale: FALLBACK_LOCALE,
    messages,
    legacy: false,
  })
  setLang(locale, i18n)
  return i18n
}

function setLang(lang: string, i18n: I18n) {
  const normalizedLang = normalizeLocale(lang)

  document.querySelector('html')?.setAttribute('lang', normalizedLang)
  localStorage.setItem('language', normalizedLang)
  i18n.global.locale.value = normalizedLang

  // 设置 vant 组件语言包
  Locale.use(normalizedLang, vantLocales[normalizedLang])
}

function matchSupportedLocale(locale: string | null): SupportedLocale | undefined {
  const normalizedLocale = String(locale || '').trim()
  if (!normalizedLocale)
    return undefined
  return supportedLocales.find(v => v === normalizedLocale || v.indexOf(normalizedLocale) === 0)
}

function matchBrowserLocale(locale: string | null): SupportedLocale | undefined {
  const normalizedLocale = String(locale || '').trim()
  if (!normalizedLocale)
    return undefined
  return browserMatchLocales.find(v => v === normalizedLocale || v.indexOf(normalizedLocale) === 0)
}

function normalizeLocale(locale: string | null): SupportedLocale {
  return matchSupportedLocale(locale) || FALLBACK_LOCALE
}

function getBrowserLocale(): SupportedLocale | undefined {
  if (typeof navigator === 'undefined')
    return undefined

  const browserLocales = [
    ...(navigator.languages || []),
    navigator.language,
  ].filter(Boolean)

  for (const browserLocale of browserLocales) {
    const exactMatch = matchBrowserLocale(browserLocale)
    if (exactMatch)
      return exactMatch

    const language = browserLocale.split('-')[0]
    const languageMatch = browserMatchLocales.find(locale => locale.split('-')[0] === language)
    if (languageMatch)
      return languageMatch
  }

  return undefined
}

// 获取当前语言对应的语言包名称
function getI18nLocale() {
  return matchSupportedLocale(localStorage.getItem('language')) || getBrowserLocale() || FALLBACK_LOCALE
}
