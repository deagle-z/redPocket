import { describe, expect, it } from 'vitest'
import {
  ppmxBanners,
  ppmxCategoryBrands,
  ppmxDrawerItems,
  ppmxFloatingWidgets,
  ppmxFooterColumns,
  ppmxHomeCategories,
  ppmxNavItems,
  ppmxNotices,
  ppmxSocialItems,
  ppmxTrustItems,
} from './data'

describe('PP.PE home reference data', () => {
  it('contains the reference home shell navigation data', () => {
    expect(ppmxNavItems.map((item) => item.id)).toEqual([
      'home',
      'wheel',
      'team',
      'promo',
      'account',
    ])
    expect(ppmxDrawerItems.map((item) => item.id)).toEqual([
      'home',
      'promo',
      'account',
      'help',
      'responsible',
    ])
  })

  it('contains the reference notice and hero banner data', () => {
    expect(ppmxNotices).toHaveLength(5)
    expect(ppmxNotices[0].target).toBe('registro')

    expect(ppmxBanners.map((banner) => banner.id)).toEqual([
      'registro',
      'checkin',
      'vip',
      'invita',
      'recarga',
      'download',
      'wheel',
    ])
    expect(ppmxBanners.every((banner) => banner.image?.startsWith('/images/ppmx/promos/'))).toBe(true)
    expect(ppmxBanners.every((banner) => banner.background.includes('url("/images/ppmx/promos/'))).toBe(true)
    expect(ppmxNotices[0].text.zh).toContain('3 比索')
    expect(ppmxNotices[2].text.zh).toContain('最高 18%返利')
    expect(ppmxBanners[0].amountHtml?.es).toContain('3')
    expect(ppmxBanners[2].amountHtml?.es).toContain('188,888')
    expect(ppmxBanners[3].id).toBe('invita')
    expect(ppmxBanners[3].sub.zh).toContain('每位有效用户最高可享充值金额50%佣金')
    expect(ppmxBanners[3].amountHtml?.zh).toContain('返佣')
    expect(ppmxBanners[4].id).toBe('recarga')
    expect(ppmxBanners[4].title.zh).toContain('最高18%')
    expect(ppmxBanners[4].title.zh).not.toContain('+25%')
    expect(ppmxBanners[4].sub.zh).toBe('首次充值享8%返利，二次充值享9%返利，三次充值享10%返利，第四次及以上充值享18%返利。充值的次数越多享受的返利越高！')
    expect(ppmxBanners[4].title.es).toContain('hasta 18%')
    expect(ppmxBanners[4].title.en).toContain('up to 18%')
    expect(ppmxBanners[4].sub.es).toBe('La primera recarga recibe un 8% de reembolso, la segunda un 9%, la tercera un 10% y desde la cuarta un 18%. Cuantas más recargas hagas, mayor será el reembolso.')
    expect(ppmxBanners[4].sub.en).toBe('The first deposit gets an 8% rebate, the second 9%, the third 10%, and the fourth and subsequent deposits 18%. The more times you deposit, the higher your rebate.')
    expect(ppmxBanners[4].amountHtml?.zh).toContain('最高')
    expect(ppmxBanners[4].amountHtml?.zh).not.toContain('+')
    expect(ppmxNotices[2].text.zh).toContain('最高 18%返利')
    expect(ppmxNotices[2].text.zh).not.toContain('+25%')
    expect(ppmxBanners[4].benefitHtml?.zh).toContain('最高 18%返利')
    expect(ppmxBanners[6].sub.zh).toContain('每达到 10000 流水')
    expect(ppmxBanners[6].sub.zh).toContain('单次充值金额≧S/200')
    expect(ppmxBanners[6].sub.en).toContain('Each single deposit of ≥S/200 grants one free spin')
  })

  it('contains the reference lobby categories and brand catalog', () => {
    expect(ppmxHomeCategories.map((category) => category.id)).toEqual([
      'populares',
      'slots',
      'pesca',
      'deportes',
      'envivo',
    ])
    expect(ppmxHomeCategories.every((category) => category.games.length === 8)).toBe(true)
    expect(ppmxHomeCategories[0].games[0]).toMatchObject({
      id: 'sweet-bonanza',
      hot: true,
      imageId: 11,
    })
    expect(Object.keys(ppmxCategoryBrands)).toEqual([
      'slots',
      'pesca',
      'deportes',
      'envivo',
    ])
    expect(ppmxCategoryBrands.slots[0].id).toBe('pragmatic')
  })

  it('contains trust and floating widget data from the reference home', () => {
    expect(ppmxTrustItems.map((item) => item.id)).toEqual([
      'instant-payments',
      'ssl',
      'support',
      'license',
    ])
    expect(ppmxFloatingWidgets.map((widget) => widget.id)).toEqual([
      'checkin',
      'social',
      'download',
    ])
  })

  it('contains the reference footer link groups and social channels', () => {
    expect(ppmxFooterColumns.map((column) => column.id)).toEqual([
      'play',
      'support',
      'company',
    ])
    expect(ppmxFooterColumns[0].title.es).toBe('Juega')
    expect(ppmxFooterColumns[1].links.map((link) => link.id)).toEqual([
      'help',
      'terms',
      'privacy',
      'responsible',
    ])
    expect(ppmxSocialItems.map((item) => item.icon)).toEqual([
      'fa-facebook',
      'fa-instagram',
      'fa-telegram',
      'fa-tiktok',
    ])
    expect(ppmxSocialItems.map((item) => item.target)).toEqual([
      'facebook',
      'instagram',
      'telegram',
      'tiktok',
    ])
  })
})
