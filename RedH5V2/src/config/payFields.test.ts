import { createI18n } from 'vue-i18n'
import { describe, expect, it } from 'vitest'

import { messages } from '@/locales'
import { buildRechargeFields } from './payFields'

describe('recharge pay fields', () => {
  it.each([
    ['es-PE', ['Tipo de documento', 'Número de documento', 'Nombre del titular']],
    ['en-US', ['Document type', 'Document number', 'Holder name']],
    ['zh-CN', ['证件类型', '证件号码', '证件姓名']],
  ] as const)('builds localized labels for %s', (locale, labels) => {
    const i18n = createI18n({
      legacy: false,
      locale,
      fallbackLocale: 'es-PE',
      messages,
    })

    const fields = buildRechargeFields(i18n.global.t)

    expect(fields.map(field => field.fieldKey)).toEqual(['identityType', 'identity', 'identityName'])
    expect(fields.map(field => field.fieldLabel)).toEqual(labels)
  })
})
