import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const appSource = readFileSync(new URL('./App.vue', import.meta.url), 'utf8') as string
const themeSource = readFileSync(new URL('./assets/styles/theme.css', import.meta.url), 'utf8') as string

describe('App PWA install prompt', () => {
  it('keeps the close control inside the main capsule without nesting buttons', () => {
    expect(appSource).toContain('class="install-pwa-button"')
    expect(appSource).toContain('<div\n      class="install-pwa-button__main"')
    expect(appSource).toContain('class="install-pwa-button__install"')
    expect(appSource).toContain('class="install-pwa-button__close"')
    expect(appSource).not.toContain('<button\n      class="install-pwa-button__main"')

    const mainIndex = appSource.indexOf('class="install-pwa-button__main"')
    const installIndex = appSource.indexOf('class="install-pwa-button__install"')
    const closeIndex = appSource.indexOf('class="install-pwa-button__close"')
    const mainCloseIndex = appSource.indexOf('</div>', mainIndex)

    expect(mainIndex).toBeGreaterThan(-1)
    expect(installIndex).toBeGreaterThan(mainIndex)
    expect(closeIndex).toBeGreaterThan(installIndex)
    expect(closeIndex).toBeLessThan(mainCloseIndex)
  })

  it('defines polished capsule motion with reduced-motion fallback', () => {
    expect(themeSource).toContain('.install-pwa-button__install')
    expect(themeSource).toContain('@keyframes installPwaEnter')
    expect(themeSource).toContain('@keyframes installPwaSheen')
    expect(themeSource).toContain('prefers-reduced-motion')
  })

  it('dismisses the install prompt only until the next local day', () => {
    expect(appSource).toContain("import { getTtlStorage, setTtlStorage } from '@/utils/storage'")
    expect(appSource).toContain('function getMsUntilTomorrow()')
    expect(appSource).toContain('tomorrow.setHours(24, 0, 0, 0)')
    expect(appSource).toContain('setTtlStorage(INSTALL_PWA_DISMISSED_KEY, true, getMsUntilTomorrow())')
    expect(appSource).toContain('getTtlStorage<boolean>(INSTALL_PWA_DISMISSED_KEY) === true')
    expect(appSource).not.toContain("window.localStorage.setItem(INSTALL_PWA_DISMISSED_KEY, '1')")
  })
})
