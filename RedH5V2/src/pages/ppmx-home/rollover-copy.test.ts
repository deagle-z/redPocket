import { describe, expect, it } from 'vitest'

import { messages } from '../../locales'
import { ppmxBanners } from './data'
import { ppmxPromoItems, ppmxWheelRules } from './pageData'

const oldRolloverPattern = /15\s*[xX]\b|15\s*×|15\s*倍流水|15倍流水|15\s*倍/

describe('PP.BET rollover copy', () => {
  it('uses 10x rollover wording across home, promo, and locale copy', () => {
    const copy = JSON.stringify({
      messages,
      ppmxBanners,
      ppmxPromoItems,
      ppmxWheelRules,
    })

    expect(copy).not.toMatch(oldRolloverPattern)
    expect(copy).toContain('10x')
    expect(copy).toContain('10 倍')
  })
})
