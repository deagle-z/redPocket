export const RECHARGE_WALLET_TYPE_VALUES = ['balance', 'sport'] as const

export type RechargeWalletType = (typeof RECHARGE_WALLET_TYPE_VALUES)[number]

export const DEFAULT_RECHARGE_WALLET_TYPE: RechargeWalletType = 'balance'

export function isRechargeWalletType(value: unknown): value is RechargeWalletType {
  return RECHARGE_WALLET_TYPE_VALUES.includes(value as RechargeWalletType)
}
