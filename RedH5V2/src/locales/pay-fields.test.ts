import { createI18n } from 'vue-i18n'
import { describe, expect, it } from 'vitest'

import { messages } from './index'

describe('pay field translations', () => {
  it.each([
    ['en-US', 'you@example.com'],
    ['es-PE', 'tu@correo.com'],
    ['zh-CN', 'you@example.com'],
  ] as const)('compiles the email placeholder for %s', (locale, expected) => {
    const i18n = createI18n({
      legacy: false,
      locale,
      fallbackLocale: 'es-PE',
      messages,
    })

    expect(i18n.global.t('payFields.emailPlaceholder')).toBe(expected)
  })
})
