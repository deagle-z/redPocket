import { describe, expect, it } from 'vitest'
import {
  buildPePhoneWithDialCode,
  isValidPeNationalPhone,
  normalizePeNationalPhone,
  onlyPhoneDigits,
} from './pePhone'

describe('pePhone utilities', () => {
  it('validates Peru national mobile numbers', () => {
    expect(isValidPeNationalPhone('912345678')).toBe(true)
    expect(isValidPeNationalPhone('812345678')).toBe(false)
    expect(isValidPeNationalPhone('91234567')).toBe(false)
    expect(isValidPeNationalPhone('51912345678')).toBe(false)
  })

  it('normalizes dial-code prefixed input', () => {
    expect(onlyPhoneDigits('+51 912 345 678')).toBe('51912345678')
    expect(normalizePeNationalPhone('+51 912 345 678')).toBe('912345678')
    expect(buildPePhoneWithDialCode('+51 912 345 678')).toBe('51912345678')
  })
})
