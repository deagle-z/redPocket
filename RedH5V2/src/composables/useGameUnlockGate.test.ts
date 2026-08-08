import { beforeEach, describe, expect, it, vi } from 'vitest'

import { getRechargeCount } from '@/api/user'
import { getLocalRechargeCount, setLocalRechargeCount } from '@/utils/rechargeCount'
import { useGameUnlockGate } from './useGameUnlockGate'

vi.mock('@/api/user', () => ({
  getRechargeCount: vi.fn(),
}))

vi.mock('@/utils/rechargeCount', () => ({
  getLocalRechargeCount: vi.fn(),
  setLocalRechargeCount: vi.fn(),
}))

describe('useGameUnlockGate', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getLocalRechargeCount).mockReturnValue(0)
  })

  it('allows play without recharge when the restriction is disabled', async () => {
    vi.mocked(getRechargeCount).mockResolvedValue({
      rechargeCount: 0,
      rebateTransferred: false,
      rechargeGameUnlockEnabled: false,
    })

    const gate = useGameUnlockGate()

    await expect(gate.ensureUnlocked()).resolves.toBe(true)
    expect(gate.unlockOpen.value).toBe(false)
    expect(setLocalRechargeCount).not.toHaveBeenCalled()
  })

  it('keeps unqualified users locked when the restriction is enabled', async () => {
    vi.mocked(getRechargeCount).mockResolvedValue({
      rechargeCount: 0,
      rebateTransferred: false,
      rechargeGameUnlockEnabled: true,
    })

    const gate = useGameUnlockGate()

    await expect(gate.ensureUnlocked()).resolves.toBe(false)
    expect(gate.unlockOpen.value).toBe(true)
  })

  it('caches a successful recharge qualification', async () => {
    vi.mocked(getRechargeCount).mockResolvedValue({
      rechargeCount: 2,
      rebateTransferred: false,
      rechargeGameUnlockEnabled: true,
    })

    const gate = useGameUnlockGate()

    await expect(gate.ensureUnlocked()).resolves.toBe(true)
    expect(setLocalRechargeCount).toHaveBeenCalledWith(2)
  })
})
