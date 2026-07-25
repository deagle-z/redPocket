import { describe, expect, it } from 'vitest'
import {
  DEFAULT_RECHARGE_WALLET_TYPE,
  RECHARGE_WALLET_TYPE_VALUES,
  isRechargeWalletType,
} from './walletType'

describe('recharge wallet type', () => {
  it('defaults to the balance wallet and only accepts supported API values', () => {
    expect(DEFAULT_RECHARGE_WALLET_TYPE).toBe('balance')
    expect(RECHARGE_WALLET_TYPE_VALUES).toEqual(['balance', 'sport'])
    expect(isRechargeWalletType('balance')).toBe(true)
    expect(isRechargeWalletType('sport')).toBe(true)
    expect(isRechargeWalletType('casino')).toBe(false)
  })
})
