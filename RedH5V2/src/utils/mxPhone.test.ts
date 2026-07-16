import { describe, expect, it } from 'vitest'
import {
  buildMxPhoneWithDialCode,
  isValidMxNationalPhone,
  normalizeMxNationalPhone,
  onlyPhoneDigits,
} from './mxPhone'

describe('mxPhone utilities', () => {
  it('validates Mexico national mobile numbers', () => {
    expect(isValidMxNationalPhone('5512345678')).toBe(true)
    expect(isValidMxNationalPhone('1551234567')).toBe(false)
    expect(isValidMxNationalPhone('551234567')).toBe(false)
    expect(isValidMxNationalPhone('525512345678')).toBe(false)
  })

  it('normalizes dial-code prefixed input', () => {
    expect(onlyPhoneDigits('+52 55 1234 5678')).toBe('525512345678')
    expect(normalizeMxNationalPhone('+52 55 1234 5678')).toBe('5512345678')
    expect(buildMxPhoneWithDialCode('+52 55 1234 5678')).toBe('525512345678')
  })
})
