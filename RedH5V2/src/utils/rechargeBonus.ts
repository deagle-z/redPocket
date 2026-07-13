// Minimum deposit (and gift threshold): deposits of less than this earn no bonus.
// Backend: orderAmount < 50 -> no gift（即 ≥50 就送）。
export const RECHARGE_BONUS_MIN_AMOUNT = 50

// 默认各档位赠送比例(%)：首充,二充,三充,第4次及以后（与后端 recharge_v2_gift_rates 默认一致）
export const DEFAULT_RECHARGE_GIFT_RATES = [18, 10, 10, 10]

function pickRate(rates: number[] | undefined, rechargeNumber: number): number {
  const list =
    Array.isArray(rates) && rates.length >= 4 ? rates : DEFAULT_RECHARGE_GIFT_RATES
  let idx = Math.floor(Number(rechargeNumber)) - 1
  if (!Number.isFinite(idx) || idx < 0) idx = 0
  if (idx > 3) idx = 3
  const pct = Number(list[idx])
  return Number.isFinite(pct) && pct > 0 ? pct : 0
}

/**
 * 按「该用户第几次充值」返回赠送比例(小数，如 0.18)。
 * @param rechargeNumber 1=首充 2=二充 3=三充 ≥4=后续
 * @param rates 各档百分比 [首充,二充,三充,后续]，不传用默认
 * @param minAmount 赠送门槛(满此金额即送)，不传用默认 50
 */
export function rechargeBonusRateByNumber(
  amount: number,
  rechargeNumber: number,
  rates?: number[],
  minAmount: number = RECHARGE_BONUS_MIN_AMOUNT,
): number {
  if (!Number.isFinite(amount) || amount < minAmount) return 0
  return pickRate(rates, rechargeNumber) / 100
}

export function rechargeBonusAmountByNumber(
  amount: number,
  rechargeNumber: number,
  rates?: number[],
  minAmount?: number,
): number {
  if (!Number.isFinite(amount) || amount <= 0) return 0
  return (
    Math.round(amount * rechargeBonusRateByNumber(amount, rechargeNumber, rates, minAmount) * 100) /
    100
  )
}

// ===== 兼容旧的「首充/非首充」二元接口（默认比例下与档位口径一致） =====
export function calculateRechargeBonusRate(amount: number, isFirstRecharge: boolean) {
  return rechargeBonusRateByNumber(amount, isFirstRecharge ? 1 : 2)
}

export function calculateRechargeBonusAmount(amount: number, isFirstRecharge: boolean) {
  return rechargeBonusAmountByNumber(amount, isFirstRecharge ? 1 : 2)
}
