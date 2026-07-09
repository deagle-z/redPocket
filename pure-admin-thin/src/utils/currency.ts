export const BACKSTAGE_DISPLAY_CURRENCY = "USD";

export function withBackstageDisplayCurrency(value: string | number) {
  return `${value} ${BACKSTAGE_DISPLAY_CURRENCY}`;
}
