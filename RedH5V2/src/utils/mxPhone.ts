import {
  APP_COUNTRY_CODE,
  APP_DIAL_CODE,
  type AppCountryCode,
} from '@/config/market'

export const MX_COUNTRY_CODE = APP_COUNTRY_CODE
export const MX_DIAL_CODE = APP_DIAL_CODE

export type MexicoCountryCode = AppCountryCode

const MX_NATIONAL_PHONE_RE = /^[2-9]\d{9}$/

export function onlyPhoneDigits(value: string) {
  return String(value || '').replace(/\D+/g, '')
}

export function isValidMxNationalPhone(value: string) {
  return MX_NATIONAL_PHONE_RE.test(onlyPhoneDigits(value))
}

export function normalizeMxNationalPhone(value: string) {
  const digits = onlyPhoneDigits(value)

  if (digits.startsWith(MX_DIAL_CODE)) {
    const withoutDialCode = digits.slice(MX_DIAL_CODE.length)
    if (isValidMxNationalPhone(withoutDialCode)) {
      return withoutDialCode
    }
  }

  return digits
}

export function buildMxPhoneWithDialCode(value: string) {
  return `${MX_DIAL_CODE}${normalizeMxNationalPhone(value)}`
}
