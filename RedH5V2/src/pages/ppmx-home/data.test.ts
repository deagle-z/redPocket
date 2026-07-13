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

describe('PP.BET home reference data', () => {
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
    expect(ppmxNotices[0].text.zh).toContain('38 美元')
    expect(ppmxNotices[2].text.zh).toContain('最高 30%返利')
    expect(ppmxBanners[0].amountHtml?.es).toContain('38')
    expect(ppmxBanners[2].amountHtml?.es).toContain('188,888')
    expect(ppmxBanners[3].id).toBe('invita')
    expect(ppmxBanners[3].sub.zh).toContain('每位有效用户最高可享充值金额50%佣金')
    expect(ppmxBanners[3].amountHtml?.zh).toContain('返佣')
    expect(ppmxBanners[4].id).toBe('recarga')
    expect(ppmxBanners[4].title.zh).toContain('最高30%')
    expect(ppmxBanners[4].title.zh).not.toContain('+25%')
    expect(ppmxBanners[4].title.es).toContain('hasta 30%')
    expect(ppmxBanners[4].amountHtml?.zh).toContain('最高')
    expect(ppmxBanners[4].amountHtml?.zh).not.toContain('+')
    expect(ppmxNotices[2].text.zh).toContain('最高 30%返利')
    expect(ppmxNotices[2].text.zh).not.toContain('+25%')
    expect(ppmxBanners[4].benefitHtml?.zh).toContain('最高 30%返利')
    expect(ppmxBanners[6].sub.zh).toContain('每达到 10000 流水')
    expect(ppmxBanners[6].sub.zh).toContain('单次充值金额≧$200')
    expect(ppmxBanners[6].sub.en).toContain('Each single deposit of ≥$200 grants one free spin')
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
    expect(ppmxSocialItems.map((item) => item.url)).toEqual([
      'https://www.facebook.com/profile.php?id=61591107827237',
      'https://www.instagram.com/ppus/',
      'https://t.me/ppus',
      'https://www.tiktok.com/@pp.us',
    ])
  })
})
