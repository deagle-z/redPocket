import { describe, expect, it } from 'vitest'

const { existsSync, readFileSync } = await import('node:' + 'fs')

describe('PP.PE component path', () => {
  it('keeps PP.PE home under the page-private route module path', () => {
    const root = new URL('../..', import.meta.url)
    const pagePrivateHome = new URL('./src/pages/ppmx-home/PpmxHome.vue', root)
    const oldComponentFolder = new URL('./src/components/ppmx-home', root)
    const oldFeatureFolder = new URL(`./src/${'features'}/ppmx-home`, root)
    const pageSource = readFileSync(new URL('../pages/(home)/index.vue', import.meta.url), 'utf8') as string
    const viteConfig = readFileSync(new URL('../../vite.config.ts', import.meta.url), 'utf8') as string

    expect(existsSync(pagePrivateHome)).toBe(true)
    expect(existsSync(oldComponentFolder)).toBe(false)
    expect(existsSync(oldFeatureFolder)).toBe(false)
    expect(pageSource).toContain('../ppmx-home/PpmxHome.vue')
    expect(pageSource).not.toContain(`@/${'features'}/ppmx-home`)
    expect(pageSource).not.toContain('@/components/ppmx-home')
    expect(viteConfig).toContain('src/pages/ppmx-home/**')
  })
})
