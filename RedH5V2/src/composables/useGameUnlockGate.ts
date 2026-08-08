import { getRechargeCount } from '@/api/user'
import { getLocalRechargeCount, setLocalRechargeCount } from '@/utils/rechargeCount'

/**
 * Deposit gate for launching real-money games.
 *
 * Unlock conditions:
 * 1. Backend has disabled the recharge unlock restriction.
 * 2. A recharge count is already cached locally.
 * 3. Backend recharge count > 0 (has deposited).
 * 4. The user has transferred commission (rebate) into their balance —
 *    commission transfer unlocks games; gift/bonus balance alone does NOT.
 * 5. Otherwise the unlock (deposit) modal is shown.
 */
export function useGameUnlockGate() {
  const unlockOpen = ref(false)
  const checking = ref(false)

  async function ensureUnlocked(): Promise<boolean> {
    if (getLocalRechargeCount() > 0) return true
    if (checking.value) return false

    try {
      checking.value = true
      const data = await getRechargeCount()
      if (data.rechargeGameUnlockEnabled === false) return true

      const count = Number(data?.rechargeCount || 0)
      if (count > 0) {
        setLocalRechargeCount(count)
        return true
      }
      // 佣金转账过即可玩（赠送余额不算）；本地标记已解锁，避免重复请求
      if (data?.rebateTransferred === true) {
        setLocalRechargeCount(1)
        return true
      }
    } catch {
      // Treat an unconfirmed status as locked and prompt to deposit.
    } finally {
      checking.value = false
    }

    unlockOpen.value = true
    return false
  }

  function closeUnlock() {
    unlockOpen.value = false
  }

  return { unlockOpen, checking, ensureUnlocked, closeUnlock }
}
