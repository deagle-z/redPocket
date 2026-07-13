import { describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { useSubmitLock } from './useSubmitLock'

describe('useSubmitLock', () => {
  it('prevents concurrent submit calls', async () => {
    let resolveFirst!: (value: string) => void
    const action = vi.fn(
      () =>
        new Promise<string>(resolve => {
          resolveFirst = resolve
        }),
    )
    const { locked, run } = useSubmitLock(action)

    const first = run()
    const second = run()

    expect(locked.value).toBe(true)
    expect(action).toHaveBeenCalledTimes(1)
    await expect(second).rejects.toThrow('Operation is already running')

    resolveFirst('ok')
    await expect(first).resolves.toBe('ok')
    await nextTick()
    expect(locked.value).toBe(false)
  })
})
