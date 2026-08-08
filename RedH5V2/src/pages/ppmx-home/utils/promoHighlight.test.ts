import { describe, expect, it } from 'vitest'

import { highlightPromoText } from './promoHighlight'

describe('PP.PE promo text highlighting', () => {
  it('wraps promo money and numeric terms with the reference highlight spans', () => {
    expect(highlightPromoText('注册立即赠送 3 比索')).toBe(
      '注册立即赠送 <span class="ppmx-promo-highlight">3 比索</span>',
    )
    expect(highlightPromoText('Recarga y recibe hasta 30% de regalo')).toBe(
      'Recarga y recibe hasta <span class="ppmx-promo-highlight ppmx-promo-highlight--percent">30%</span> de regalo',
    )
  })

  it('escapes text before applying highlight markup', () => {
    expect(highlightPromoText('<script>S/38</script>')).toBe(
      '&lt;script&gt;<span class="ppmx-promo-highlight">S/38</span>&lt;/script&gt;',
    )
  })
})
