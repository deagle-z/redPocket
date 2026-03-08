export const CURRENCY_SYMBOL = '฿'
export const CURRENCY_CODE = 'THB'

interface FormatCurrencyOptions {
  signed?: boolean
  spaceBetweenSymbolAndAmount?: boolean
  fractionDigits?: number
}

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
  return `${signText}${CURRENCY_SYMBOL}${space}${absText}`
}
