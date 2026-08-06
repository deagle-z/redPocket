import {
  APP_COUNTRY_CODE,
  APP_DIAL_CODE,
  type AppCountryCode,
} from '@/config/market'

export const PE_COUNTRY_CODE = APP_COUNTRY_CODE
export const PE_DIAL_CODE = APP_DIAL_CODE

export type PeruCountryCode = AppCountryCode

const PE_NATIONAL_PHONE_RE = /^9\d{8}$/

export function onlyPhoneDigits(value: string) {
  return String(value || '').replace(/\D+/g, '')
}

export function isValidPeNationalPhone(value: string) {
  return PE_NATIONAL_PHONE_RE.test(onlyPhoneDigits(value))
}

export function normalizePeNationalPhone(value: string) {
  const digits = onlyPhoneDigits(value)

  if (digits.startsWith(PE_DIAL_CODE)) {
    const withoutDialCode = digits.slice(PE_DIAL_CODE.length)
    if (isValidPeNationalPhone(withoutDialCode)) {
      return withoutDialCode
    }
  }

  return digits
}

export function buildPePhoneWithDialCode(value: string) {
  return `${PE_DIAL_CODE}${normalizePeNationalPhone(value)}`
}
