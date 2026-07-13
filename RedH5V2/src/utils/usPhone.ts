export const US_COUNTRY_CODE = 'US' as const
export const US_DIAL_CODE = '1'

export type UnitedStatesCountryCode = typeof US_COUNTRY_CODE

const US_NATIONAL_PHONE_RE = /^[2-9]\d{2}[2-9]\d{6}$/

export function onlyPhoneDigits(value: string) {
  return value.replace(/\D+/g, '')
}

export function isValidUsNationalPhone(value: string) {
  return US_NATIONAL_PHONE_RE.test(value)
}

export function normalizeUsNationalPhone(value: string) {
  const digits = onlyPhoneDigits(value)

  if (digits.startsWith(US_DIAL_CODE)) {
    const withoutDialCode = digits.slice(US_DIAL_CODE.length)

    if (isValidUsNationalPhone(withoutDialCode)) {
      return withoutDialCode
    }
  }

  return digits
}

export function buildUsPhoneWithDialCode(value: string) {
  return `${US_DIAL_CODE}${normalizeUsNationalPhone(value)}`
}
