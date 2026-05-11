export type AuthCountryCode = 'MX' | 'ID' | 'BR'

const phoneRules: Record<AuthCountryCode, RegExp> = {
  MX: /^[2-9]\d{9}$/,
  ID: /^0?8\d{8,11}$/,
  BR: /^(?:[1-9]{2}\d{8}|[1-9]{2}9\d{8})$/,
}

const dialCodeMap: Record<AuthCountryCode, string> = {
  MX: '52',
  ID: '62',
  BR: '55',
}

export function onlyPhoneDigits(rawPhone: string) {
  return rawPhone.replace(/\D+/g, '')
}

export function isValidNationalPhone(country: AuthCountryCode, phone: string) {
  return phoneRules[country].test(phone)
}

export function normalizeNationalPhone(country: AuthCountryCode, rawPhone: string) {
  const digits = onlyPhoneDigits(rawPhone)
  const dialCode = dialCodeMap[country]
  const phoneWithoutDialCode = dialCode && digits.startsWith(dialCode)
    ? digits.slice(dialCode.length)
    : ''
  const nationalPhone = phoneWithoutDialCode && isValidNationalPhone(country, phoneWithoutDialCode)
    ? phoneWithoutDialCode
    : digits

  if (country === 'ID')
    return nationalPhone.replace(/^0+/, '')

  return nationalPhone
}

export function buildPhoneWithDialCode(country: AuthCountryCode, rawPhone: string) {
  return `${dialCodeMap[country]}${normalizeNationalPhone(country, rawPhone)}`
}
