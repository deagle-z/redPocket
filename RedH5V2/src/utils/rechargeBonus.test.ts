import { describe, expect, it } from 'vitest'
import {
  calculateRechargeBonusAmount,
  calculateRechargeBonusRate,
  rechargeBonusAmountByNumber,
  rechargeBonusRateByNumber,
} from './rechargeBonus'

describe('recharge bonus rules', () => {
  it('default tiers increase from 8% to 18%', () => {
    expect(calculateRechargeBonusRate(100, true)).toBe(0.08)
    expect(calculateRechargeBonusRate(100, false)).toBe(0.09)
    expect(rechargeBonusRateByNumber(100, 1)).toBe(0.08)
    expect(rechargeBonusRateByNumber(100, 2)).toBe(0.09)
    expect(rechargeBonusRateByNumber(100, 3)).toBe(0.1)
    expect(rechargeBonusRateByNumber(100, 5)).toBe(0.12)
    expect(rechargeBonusRateByNumber(100, 11)).toBe(0.18)
    expect(rechargeBonusRateByNumber(100, 12)).toBe(0.18)
  })

  it('respects configured 11-tier rates', () => {
    const rates = [8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18]
    expect(rechargeBonusRateByNumber(100, 1, rates)).toBe(0.08)
    expect(rechargeBonusRateByNumber(100, 2, rates)).toBe(0.09)
    expect(rechargeBonusRateByNumber(100, 3, rates)).toBe(0.1)
    expect(rechargeBonusRateByNumber(100, 9, rates)).toBe(0.16)
  })

  it('gifts at 50 and above, but not below 50', () => {
    expect(rechargeBonusRateByNumber(50, 1)).toBe(0.08)
    expect(rechargeBonusAmountByNumber(50, 1)).toBe(4)
    expect(calculateRechargeBonusAmount(50, true)).toBe(4)
    expect(rechargeBonusRateByNumber(49.99, 1)).toBe(0)
    expect(rechargeBonusAmountByNumber(49.99, 1)).toBe(0)
  })

  it('rounds bonus amount to two decimals', () => {
    expect(rechargeBonusAmountByNumber(1000, 1)).toBe(80)
    expect(rechargeBonusAmountByNumber(1234.56, 2)).toBe(111.11)
  })
})
