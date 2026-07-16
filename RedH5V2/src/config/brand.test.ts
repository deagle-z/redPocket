import { describe, expect, it } from 'vitest'

import { APP_BRAND } from './brand'

const { readFileSync } = await import('node:' + 'fs')

const activeAppFiles = [
  new URL('../../index.html', import.meta.url),
  new URL('../../package.json', import.meta.url),
  new URL('../../vite.config.ts', import.meta.url),
  new URL('../components/AppShell.vue', import.meta.url),
  new URL('../components/AppTabbar.vue', import.meta.url),
  new URL('../pages/ppmx-home/PpmxHome.vue', import.meta.url),
  new URL('../locales/es-MX.ts', import.meta.url),
  new URL('../locales/en-US.ts', import.meta.url),
  new URL('../locales/zh-CN.ts', import.meta.url),
  new URL('../pages/profile/index.vue', import.meta.url),
  new URL('../pages/wheel/index.vue', import.meta.url),
  new URL('../pages/team/index.vue', import.meta.url),
]

const activeSource = activeAppFiles
  .map((file) => readFileSync(file, 'utf8'))
  .join('\n')

describe('PP.MX brand foundation', () => {
  it('exposes the project brand contract', () => {
    expect(APP_BRAND.name).toBe('PP.MX')
    expect(APP_BRAND.shortName).toBe('PP.MX')
    expect(APP_BRAND.mark).toBe('PP')
    expect(APP_BRAND.themeColor).toBe('#060608')
  })

  it('removes previous RedH5 brand copy from active app files', () => {
    const forbiddenCopy = [
      'RedH5V2',
      'RedH5',
      'Lucky Money',
      'Red Packet',
      'To C Console',
      'Mobile H5 app',
      'red-h5-v2',
      '红包大厅',
      '普通红包',
      '扫雷红包',
      '红包场次',
    ]

    for (const text of forbiddenCopy) {
      expect(activeSource).not.toContain(text)
    }
  })

  it('renders a branded initialization loading screen before Vue mounts', () => {
    const indexSource = readFileSync(new URL('../../index.html', import.meta.url), 'utf8') as string
    const mainSource = readFileSync(new URL('../main.ts', import.meta.url), 'utf8') as string

    expect(indexSource).toContain('id="app-loading"')
    expect(indexSource).toContain('linear-gradient(180deg, #060608 0%, #08080b 58%, #050507 100%)')
    expect(indexSource).toContain('#app-loading::before')
    expect(indexSource).toContain('.app-loading__panel')
    expect(indexSource).toContain('%BASE_URL%images/ppmx/logo.png')
    expect(indexSource).toContain('app-loading__brand">PP<span>.</span>MX')
    expect(indexSource).toContain('app-loading-spin')
    expect(indexSource).not.toContain('backdrop-filter: blur(22px) saturate(150%)')
    expect(mainSource).toContain('removeInitialLoading')
    expect(mainSource).toContain("document.getElementById('app-loading')")
    expect(mainSource).toContain("loader.classList.add('is-leaving')")
  })
})
