import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8') as string

describe('crypto payment page', () => {
  it('supports configured USDT, BTC, and ETH payment options', () => {
    expect(source).toContain("normalizedNetwork === 'TRC20'")
    expect(source).toContain("normalizedNetwork === 'BITCOIN'")
    expect(source).toContain("normalizedNetwork === 'ETHEREUM'")
    expect(source).toContain("option.token.toUpperCase() === 'BTC'")
    expect(source).toContain("option.token.toUpperCase() === 'ETH'")
    expect(source).toContain("selectedToken === 'BTC'")
    expect(source).toContain("selectedToken === 'ETH'")
    expect(source).not.toContain('Actualmente solo está disponible USDT')
  })

  it('uses backend QR content and copies the numeric amount without its ticker', () => {
    expect(source).toContain('cryptoPayment.value?.qrContent')
    expect(source).not.toContain('cryptoPayment.value?.qrContent || cryptoPayment.value?.receiveAddress')
    expect(source).toContain("copyValue(payAmountValue)")
    expect(source).not.toContain("copyValue(payAmount)")
  })

  it('uses the server platform amount and blocks inconsistent quotes', () => {
    expect(source).toContain('platformAmountInCents(result.platformAmount)')
    expect(source).toContain('platformAmount.value = quotedCents / 100')
    expect(source).toContain('platformAmountsMatch(amount, platformAmount.value)')
    expect(source).toContain('result.options.some(option => !platformAmountsMatch(option.platformAmount, platformAmount.value))')
  })

  it('restores pending orders and resumes polling after page restoration', () => {
    expect(source).toContain("PENDING_ORDER_STORAGE_PREFIX = 'h5_crypto_pending_order_v1'")
    expect(source).toContain('restorePersistedPendingOrder()')
    expect(source).toContain("window.addEventListener('pageshow', handlePageShow)")
    expect(source).toContain('await resumePendingPayment()')
    expect(source).toContain('must request a manual adjustment')
    expect(source).toContain('必须联系客服人工补单')
  })
})
