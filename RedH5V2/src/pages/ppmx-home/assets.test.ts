import { describe, expect, it } from 'vitest'
import {
  PPMX_FALLBACK_GRADIENTS,
  PPMX_LOCAL_IMAGE_ROOT,
  getPpmxGameImage,
  getPpmxGameTileBackground,
  ppmxAsset,
} from './assets'
import { ppmxBanners, ppmxHomeCategories } from './data'

const { existsSync, readFileSync } = await import('node:' + 'fs')
const symbolsSource = readFileSync(
  new URL('./components/PpmxSymbols.vue', import.meta.url),
  'utf8',
) as string
const logoSource = readFileSync(
  new URL('./components/PpmxLogo.vue', import.meta.url),
  'utf8',
) as string

describe('PP.MX localized visual assets', () => {
  it('builds public asset URLs from the local PP.MX image root', () => {
    expect(PPMX_LOCAL_IMAGE_ROOT).toBe('images/ppmx')
    expect(ppmxAsset('games/11.jpg')).toBe('/images/ppmx/games/11.jpg')
  })

  it('resolves game image IDs to local files and stable fallback gradients', () => {
    expect(getPpmxGameImage(11)).toBe('/images/ppmx/games/11.jpg')
    expect(PPMX_FALLBACK_GRADIENTS).toHaveLength(5)
    expect(getPpmxGameTileBackground(11)).toEqual({
      backgroundImage: "url('/images/ppmx/games/11.jpg'), linear-gradient(135deg,#138a52,#0c5132)",
    })
  })

  it('keeps home data free of remote visual dependencies', () => {
    const serialized = JSON.stringify({ ppmxBanners, ppmxHomeCategories })
    expect(serialized).not.toContain('images.unsplash.com')
    expect(serialized).not.toContain('http://')
    expect(serialized).not.toContain('https://')
    expect(serialized).not.toContain('data:image')
  })

  it('uses the local PP.MX promo images for the hero carousel', () => {
    expect(ppmxBanners.map((banner) => banner.image)).toEqual([
      '/images/ppmx/promos/registro.jpg',
      '/images/ppmx/promos/checkin.jpg',
      '/images/ppmx/promos/vip.jpg',
      '/images/ppmx/promos/invita.jpg',
      '/images/ppmx/promos/recarga.jpg',
      '/images/ppmx/promos/download.jpg',
      '/images/ppmx/promos/wheel.jpg',
    ])

    for (const banner of ppmxBanners) {
      expect(banner.image).toBeTruthy()
      expect(existsSync(new URL(`../../../public${banner.image}`, import.meta.url))).toBe(true)
      expect(banner.background).toContain('right center / cover')
    }
  })

  it('keeps the registration gift as the first home hero carousel slide', () => {
    expect(ppmxBanners[0]).toMatchObject({
      id: 'registro',
      image: '/images/ppmx/promos/registro.jpg',
    })
    expect(ppmxBanners[0]?.tag.zh).toBe('注册礼金')
    expect(ppmxBanners[0]?.title.zh).toContain('注册立即赠送')
    expect(ppmxBanners[0]?.amountHtml?.zh).toContain('58')
    expect(ppmxBanners[0]?.amountHtml?.zh).toContain('比索')
  })

  it('contains the migrated reference SVG symbols', () => {
    expect(symbolsSource).toContain('id="ppCoin"')
    expect(symbolsSource).toContain('id="ppEmblem"')
    expect(symbolsSource).toContain('radialGradient id="ppcRim"')
    expect(symbolsSource).toContain('linearGradient id="lgRed"')
  })

  it('uses the migrated PP.MX emblem in the logo component', () => {
    expect(logoSource).toContain('href="#ppEmblem"')
    expect(logoSource).toContain('PP')
    expect(logoSource).toContain('MX')
    expect(logoSource).not.toContain('BET')
  })

  it('publishes the PP.MX emblem and wordmark as local image logo assets', () => {
    const emblemUrl = new URL('../../../public/images/ppmx/emblem.svg', import.meta.url)
    const logoUrl = new URL('../../../public/images/ppmx/logo.svg', import.meta.url)
    const indexSource = readFileSync(new URL('../../../index.html', import.meta.url), 'utf8') as string
    const viteSource = readFileSync(new URL('../../../vite.config.ts', import.meta.url), 'utf8') as string

    expect(existsSync(emblemUrl)).toBe(true)
    expect(existsSync(logoUrl)).toBe(true)
    expect(readFileSync(emblemUrl, 'utf8')).toContain('aria-label="PP.MX emblem"')
    expect(readFileSync(logoUrl, 'utf8')).toContain('aria-label="PP.MX"')
    expect(indexSource).toContain('%BASE_URL%images/ppmx/logo.png')
    expect(viteSource).toContain('images/ppmx/emblem.svg')
    expect(viteSource).toContain('images/ppmx/logo.svg')
  })
})
