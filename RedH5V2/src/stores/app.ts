import { i18n } from '@/plugins/i18n'
import { getInitialLocale, persistLocale, type Locale } from '@/utils/locale'

export type { Locale } from '@/utils/locale'
export type ThemeMode = 'dark' | 'light'

const THEME_STORAGE_KEY = 'ppmx_theme'
const DARK_THEME_COLOR = '#060608'
const LIGHT_THEME_COLOR = '#ffffff'

function isThemeMode(value: unknown): value is ThemeMode {
  return value === 'dark' || value === 'light'
}

function readPersistedTheme(): ThemeMode {
  if (typeof window === 'undefined') return 'dark'

  const storedTheme = window.localStorage.getItem(THEME_STORAGE_KEY)
  if (isThemeMode(storedTheme)) return storedTheme

  const datasetTheme = document.documentElement.dataset.theme
  return isThemeMode(datasetTheme) ? datasetTheme : 'dark'
}

function applyDocumentTheme(value: ThemeMode) {
  if (typeof document === 'undefined') return

  document.documentElement.dataset.theme = value
  document.documentElement.style.colorScheme = value

  const themeColor = value === 'light' ? LIGHT_THEME_COLOR : DARK_THEME_COLOR
  const meta = document.querySelector<HTMLMetaElement>('meta[name="theme-color"]')
  if (meta) {
    meta.content = themeColor
  }
}

export const useAppStore = defineStore('app', () => {
  const locale = ref<Locale>(getInitialLocale())
  const theme = ref<ThemeMode>(readPersistedTheme())

  function setLocale(value: Locale) {
    locale.value = value
    i18n.global.locale.value = value
    persistLocale(value)
  }

  function setTheme(value: ThemeMode) {
    theme.value = value
    applyDocumentTheme(value)

    if (typeof window !== 'undefined') {
      window.localStorage.setItem(THEME_STORAGE_KEY, value)
    }
  }

  function toggleTheme() {
    setTheme(theme.value === 'light' ? 'dark' : 'light')
  }

  function initTheme() {
    setTheme(readPersistedTheme())
  }

  return {
    locale,
    theme,
    setLocale,
    setTheme,
    toggleTheme,
    initTheme,
  }
})
