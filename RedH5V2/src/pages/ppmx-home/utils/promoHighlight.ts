export function escapePromoHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

export function highlightPromoText(text: string): string {
  return escapePromoHtml(text)
    .replace(
      /([+-]?\d[\d.,]*\s?%)/g,
      '<span class="ppmx-promo-highlight ppmx-promo-highlight--percent">$1</span>',
    )
    .replace(/(\$\s?[\d.,]+)/g, '<span class="ppmx-promo-highlight">$1</span>')
    .replace(
      /(\d[\d.,]*\s?(?:美元|dollars?|USD|días?|days?|day|天|veces|倍|giros?|次))/gi,
      '<span class="ppmx-promo-highlight">$1</span>',
    )
}
