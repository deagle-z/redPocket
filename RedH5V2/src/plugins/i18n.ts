import { createI18n } from 'vue-i18n'
import { messages } from '@/locales'
import { getInitialLocale } from '@/utils/locale'

export const i18n = createI18n({
  legacy: false,
  locale: getInitialLocale(),
  fallbackLocale: 'en-US',
  messages,
})
