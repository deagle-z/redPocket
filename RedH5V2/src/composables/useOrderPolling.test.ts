import { describe, expect, it, vi } from 'vitest'
import { useOrderPolling } from './useOrderPolling'
import type { TransactionSnapshot } from '@/types/business'

describe('useOrderPolling', () => {
  it('stops after a terminal transaction status', async () => {
    vi.useFakeTimers()

    const snapshots: TransactionSnapshot[] = [
      { id: '1', status: 'processing' },
      { id: '1', status: 'success' },
    ]
    const fetchSnapshot = vi.fn(() => Promise.resolve(snapshots.shift()!))
    const { current, start, stop } = useOrderPolling(fetchSnapshot, {
      intervalMs: 100,
      maxAttempts: 5,
    })

    const polling = start()
    await vi.runOnlyPendingTimersAsync()
    await vi.advanceTimersByTimeAsync(100)

    await expect(polling).resolves.toMatchObject({ status: 'success' })
    expect(current.value?.status).toBe('success')
    expect(fetchSnapshot).toHaveBeenCalledTimes(2)

    stop()
    vi.useRealTimers()
  })
})
