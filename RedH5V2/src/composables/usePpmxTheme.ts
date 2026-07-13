import type { ThemeMode } from '@/stores/app'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'

export function usePpmxTheme() {
  const appStore = useAppStore()
  const { t } = useI18n()

  const theme = computed(() => appStore.theme)
  const isDark = computed(() => theme.value === 'dark')
  const isLight = computed(() => theme.value === 'light')
  const themeLabel = computed(() => t(`theme.${theme.value}`))
  const themeToggleLabel = computed(() => t('theme.toggle'))
  const themeIconClass = computed(() => (isLight.value ? 'fa-sun' : 'fa-moon'))

  function setTheme(value: ThemeMode) {
    appStore.setTheme(value)
  }

  function toggleTheme() {
    appStore.toggleTheme()
  }

  return {
    theme,
    isDark,
    isLight,
    themeLabel,
    themeToggleLabel,
    themeIconClass,
    setTheme,
    toggleTheme,
  }
}
