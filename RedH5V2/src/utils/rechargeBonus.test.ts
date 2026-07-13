import { describe, expect, it } from 'vitest'
import {
  calculateRechargeBonusAmount,
  calculateRechargeBonusRate,
  rechargeBonusAmountByNumber,
  rechargeBonusRateByNumber,
} from './rechargeBonus'

describe('recharge bonus rules', () => {
  it('default tiers: first 18%, others 10%', () => {
    expect(calculateRechargeBonusRate(100, true)).toBe(0.18)
    expect(calculateRechargeBonusRate(100, false)).toBe(0.1)
    expect(rechargeBonusRateByNumber(100, 1)).toBe(0.18)
    expect(rechargeBonusRateByNumber(100, 2)).toBe(0.1)
    expect(rechargeBonusRateByNumber(100, 3)).toBe(0.1)
    expect(rechargeBonusRateByNumber(100, 5)).toBe(0.1)
  })

  it('respects configured per-tier rates [first,second,third,rest]', () => {
    const rates = [20, 15, 12, 8]
    expect(rechargeBonusRateByNumber(100, 1, rates)).toBe(0.2)
    expect(rechargeBonusRateByNumber(100, 2, rates)).toBe(0.15)
    expect(rechargeBonusRateByNumber(100, 3, rates)).toBe(0.12)
    expect(rechargeBonusRateByNumber(100, 9, rates)).toBe(0.08)
  })

  it('gifts at 50 and above, but not below 50', () => {
    expect(rechargeBonusRateByNumber(50, 1)).toBe(0.18)
    expect(rechargeBonusAmountByNumber(50, 1)).toBe(9)
    expect(calculateRechargeBonusAmount(50, true)).toBe(9)
    expect(rechargeBonusRateByNumber(49.99, 1)).toBe(0)
    expect(rechargeBonusAmountByNumber(49.99, 1)).toBe(0)
  })

  it('rounds bonus amount to two decimals', () => {
    expect(rechargeBonusAmountByNumber(1000, 1)).toBe(180)
    expect(rechargeBonusAmountByNumber(1234.56, 2)).toBe(123.46)
  })
})
