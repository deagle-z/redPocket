import coinSvg from '@/assets/svg/coin.svg'

export const CURRENCY_SYMBOL = coinSvg
export const CURRENCY_CODE = 'GOLD'

interface FormatCurrencyOptions {
  signed?: boolean
  spaceBetweenSymbolAndAmount?: boolean
  fractionDigits?: number
}

// Internal text marker — used by CoinAmount / CurrencyText to detect and replace with the coin icon
const _symbolChar = '\u0E3F'

export function truncate2(value: number) {
  const raw = Number(value || 0)
  const safe = Number.isFinite(raw) ? raw : 0
  const scaled = safe * 100
  const adjusted = scaled > 0 ? scaled + 1e-9 : scaled < 0 ? scaled - 1e-9 : scaled
  return Math.trunc(adjusted) / 100
}

export function formatCurrency(value: number, options: FormatCurrencyOptions = {}) {
  const {
    signed = false,
    spaceBetweenSymbolAndAmount = false,
    fractionDigits = 2,
  } = options

  const safe = truncate2(value)
  const displayFractionDigits = Math.min(Math.max(Math.trunc(fractionDigits), 0), 2)
  const absText = Math.abs(safe).toFixed(displayFractionDigits)
  const signText = signed ? (safe >= 0 ? '+' : '-') : (safe < 0 ? '-' : '')
  const space = spaceBetweenSymbolAndAmount ? ' ' : ''
  return `${signText}${_symbolChar}${space}${absText}`
}

export function formatCurrencyPlain(value: number, fractionDigits = 2) {
  const safe = truncate2(value)
  const displayFractionDigits = Math.min(Math.max(Math.trunc(fractionDigits), 0), 2)
  return Math.abs(safe).toFixed(displayFractionDigits)
}
