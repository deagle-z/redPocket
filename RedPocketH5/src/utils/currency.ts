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

export function formatCurrency(value: number, options: FormatCurrencyOptions = {}) {
  const {
    signed = false,
    spaceBetweenSymbolAndAmount = false,
    fractionDigits = 2,
  } = options

  const raw = Number(value || 0)
  const safe = Number.isFinite(raw) ? raw : 0
  const absText = Math.abs(safe).toFixed(fractionDigits)
  const signText = signed ? (safe >= 0 ? '+' : '-') : (safe < 0 ? '-' : '')
  const space = spaceBetweenSymbolAndAmount ? ' ' : ''
  return `${signText}${_symbolChar}${space}${absText}`
}
