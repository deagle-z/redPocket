import type { MoneyValue } from '@/types/business'

const MONEY_PATTERN = /^-?\d+(\.\d{1,2})?$/

function normalizeMoneyInput(value: MoneyValue) {
  const raw = String(value).trim().replace(/,/g, '')

  if (!MONEY_PATTERN.test(raw)) {
    throw new Error('Invalid money value')
  }

  return raw
}

export function toCent(value: MoneyValue) {
  const raw = normalizeMoneyInput(value)
  const negative = raw.startsWith('-')
  const unsigned = negative ? raw.slice(1) : raw
  const [yuan = '0', decimal = ''] = unsigned.split('.')
  const cents = `${decimal}00`.slice(0, 2)
  const result = Number(yuan) * 100 + Number(cents)

  return negative ? -result : result
}

export function toDisplayCents(value: MoneyValue | null | undefined) {
  if (value === undefined || value === null || value === '') return 0

  try {
    return toCent(value)
  } catch {
    return 0
  }
}

export function fromCent(value: number) {
  return value / 100
}

export function formatMoney(value: number, options?: { currency?: string }) {
  const sign = value < 0 ? '-' : ''
  const absolute = Math.abs(Math.trunc(value))
  const yuan = Math.floor(absolute / 100)
  const cents = String(absolute % 100).padStart(2, '0')
  const integer = yuan.toLocaleString('en-US')
  const formatted = `${sign}${integer}.${cents}`

  return options?.currency ? `${options.currency}${formatted}` : formatted
}

export function formatSignedMoney(value: number) {
  if (value > 0) return `+${formatMoney(value)}`
  return formatMoney(value)
}

export function maskBalance(value: number) {
  if (value === 0) return '***'

  const [integer] = formatMoney(value).split('.')
  const first = integer.split(',')[0]

  return `${first},***.**`
}
