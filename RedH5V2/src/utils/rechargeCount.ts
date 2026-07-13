import { getTtlStorage, setTtlStorage } from './storage'

const RECHARGE_COUNT_KEY = 'ppmx_recharge_count'

export function getLocalRechargeCount() {
  return getTtlStorage<number>(RECHARGE_COUNT_KEY) ?? 0
}

export function setLocalRechargeCount(count: number) {
  setTtlStorage(RECHARGE_COUNT_KEY, count)
}

export function incrementLocalRechargeCount() {
  const next = getLocalRechargeCount() + 1
  setLocalRechargeCount(next)

  return next
}
