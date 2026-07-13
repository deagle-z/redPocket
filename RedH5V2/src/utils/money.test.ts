import { describe, expect, it } from 'vitest'
import {
  formatMoney,
  formatSignedMoney,
  fromCent,
  maskBalance,
  toCent,
  toDisplayCents,
} from './money'

describe('money utilities', () => {
  it('formats cents with fixed precision and thousands separators', () => {
    expect(formatMoney(0)).toBe('0.00')
    expect(formatMoney(123)).toBe('1.23')
    expect(formatMoney(123456789)).toBe('1,234,567.89')
    expect(formatMoney(-501)).toBe('-5.01')
  })

  it('converts yuan values to cents without floating point drift', () => {
    expect(toCent(1)).toBe(100)
    expect(toCent('0.1')).toBe(10)
    expect(toCent('0.29')).toBe(29)
    expect(toCent('1234567.89')).toBe(123456789)
    expect(toCent('-5.01')).toBe(-501)
  })

  it('normalizes API display money values to cents', () => {
    expect(toDisplayCents(undefined)).toBe(0)
    expect(toDisplayCents(null)).toBe(0)
    expect(toDisplayCents('')).toBe(0)
    expect(toDisplayCents(1)).toBe(100)
    expect(toDisplayCents('1')).toBe(100)
    expect(toDisplayCents('1.25')).toBe(125)
    expect(toDisplayCents('abc')).toBe(0)
  })

  it('converts cents back to yuan number values', () => {
    expect(fromCent(0)).toBe(0)
    expect(fromCent(123)).toBe(1.23)
    expect(fromCent(-501)).toBe(-5.01)
  })

  it('formats signed money with explicit plus sign for positive values', () => {
    expect(formatSignedMoney(123)).toBe('+1.23')
    expect(formatSignedMoney(-123)).toBe('-1.23')
    expect(formatSignedMoney(0)).toBe('0.00')
  })

  it('masks balances while preserving rough magnitude', () => {
    expect(maskBalance(0)).toBe('***')
    expect(maskBalance(123456)).toBe('1,***.**')
  })

  it('rejects invalid yuan input', () => {
    expect(() => toCent('abc')).toThrow('Invalid money value')
    expect(() => toCent('1.234')).toThrow('Invalid money value')
  })
})
