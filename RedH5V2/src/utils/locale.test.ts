import { describe, expect, it } from 'vitest'

import { getInitialLocale, normalizeLocale, resolveBrowserLocale } from './locale'

describe('locale utilities', () => {
  it('normalizes supported browser language codes', () => {
    expect(normalizeLocale('zh')).toBe('zh-CN')
    expect(normalizeLocale('zh-Hans-CN')).toBe('zh-CN')
    expect(normalizeLocale('en')).toBe('en-US')
    expect(normalizeLocale('en-GB')).toBe('en-US')
    expect(normalizeLocale('es')).toBe('es-PE')
    expect(normalizeLocale('es-PE')).toBe('es-PE')
  })

  it('resolves the first supported non-Chinese browser language', () => {
    expect(resolveBrowserLocale(['fr-FR', 'en-US'])).toBe('en-US')
    expect(resolveBrowserLocale(['pt-BR', 'es-ES', 'en-US'])).toBe('es-PE')
    expect(resolveBrowserLocale(['zh-CN', 'en-US'])).toBe('en-US')
  })

  it('falls back to Mexican Spanish for unsupported or Chinese-only browser languages', () => {
    expect(resolveBrowserLocale(['fr-FR', 'ja-JP'])).toBe('es-PE')
    expect(resolveBrowserLocale(['zh-CN', 'zh-Hans-CN'])).toBe('es-PE')
    expect(resolveBrowserLocale([])).toBe('es-PE')
  })

  it('defaults initial locale to Mexican Spanish when no manual locale is stored', () => {
    expect(getInitialLocale(['zh-CN'])).toBe('es-PE')
  })
})
