import { describe, expect, it } from 'vitest'

import { resolvePpmxLocale, usePpmxLocale } from './usePpmxLocale'

describe('usePpmxLocale', () => {
  it('resolves global app locale to the reference home locale', () => {
    expect(resolvePpmxLocale('zh-CN')).toBe('zh')
    expect(resolvePpmxLocale('en-US')).toBe('en')
    expect(resolvePpmxLocale('es-MX')).toBe('es')
    expect(resolvePpmxLocale('fr-FR')).toBe('es')
  })

  it('switches between the reference home locales', () => {
    const locale = usePpmxLocale('zh-CN')

    expect(locale.pickText({ en: 'Home', es: 'Inicio', zh: '首页' })).toBe('首页')

    locale.setHomeLocale('en')
    expect(locale.pickText({ en: 'Home', es: 'Inicio', zh: '首页' })).toBe('Home')

    locale.setHomeLocale('zh')
    expect(locale.pickText({ en: 'Home', es: 'Inicio', zh: '首页' })).toBe('首页')
  })

  it('shares locale changes across PP.MX components', () => {
    const headerLocale = usePpmxLocale('es-MX')
    const pageLocale = usePpmxLocale('es-MX')

    headerLocale.setHomeLocale('en')

    expect(pageLocale.pickText({ en: 'Home', es: 'Inicio', zh: '首页' })).toBe('Home')
  })
})
