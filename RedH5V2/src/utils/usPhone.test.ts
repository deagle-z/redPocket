import { describe, expect, it } from 'vitest'
import {
  buildUsPhoneWithDialCode,
  isValidUsNationalPhone,
  normalizeUsNationalPhone,
  onlyPhoneDigits,
} from './usPhone'

describe('US phone utilities', () => {
  it('validates US national phone numbers', () => {
    expect(isValidUsNationalPhone('2015550123')).toBe(true)
    expect(isValidUsNationalPhone('1015550123')).toBe(false)
    expect(isValidUsNationalPhone('2010550123')).toBe(false)
    expect(isValidUsNationalPhone('12015550123')).toBe(false)
  })

  it('normalizes formatted or dial-code phone input', () => {
    expect(onlyPhoneDigits('+1 (201) 555-0123')).toBe('12015550123')
    expect(normalizeUsNationalPhone('+1 (201) 555-0123')).toBe('2015550123')
    expect(buildUsPhoneWithDialCode('+1 (201) 555-0123')).toBe('12015550123')
  })
})
