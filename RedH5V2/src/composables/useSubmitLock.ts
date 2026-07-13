import { ref } from 'vue'
import type { OperationLockOptions } from '@/types/business'

export function useSubmitLock<TArgs extends unknown[], TResult>(
  action: (...args: TArgs) => Promise<TResult>,
  options: OperationLockOptions = {},
) {
  const locked = ref(false)
  const error = ref<unknown>(null)

  async function run(...args: TArgs) {
    if (locked.value) {
      throw new Error(options.lockedMessage ?? 'Operation is already running')
    }

    const startedAt = Date.now()
    locked.value = true
    error.value = null

    try {
      return await action(...args)
    } catch (currentError) {
      error.value = currentError
      throw currentError
    } finally {
      const elapsed = Date.now() - startedAt
      const delay = Math.max((options.minLockMs ?? 0) - elapsed, 0)

      if (delay > 0) {
        await new Promise(resolve => setTimeout(resolve, delay))
      }

      locked.value = false
    }
  }

  return {
    locked,
    error,
    run,
  }
}
