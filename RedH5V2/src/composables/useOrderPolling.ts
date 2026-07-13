import { onScopeDispose, ref } from 'vue'
import type {
  PollingOptions,
  TransactionSnapshot,
  TransactionStatus,
} from '@/types/business'

const DEFAULT_TERMINAL_STATUSES: TransactionStatus[] = [
  'success',
  'failed',
  'canceled',
  'expired',
  'reviewing',
]

export function useOrderPolling<T extends TransactionSnapshot>(
  fetchSnapshot: () => Promise<T>,
  options: PollingOptions = {},
) {
  const current = ref<T | null>(null)
  const polling = ref(false)
  const attempts = ref(0)
  const error = ref<unknown>(null)
  const intervalMs = options.intervalMs ?? 1500
  const maxAttempts = options.maxAttempts ?? 20
  const terminalStatuses =
    options.terminalStatuses ?? DEFAULT_TERMINAL_STATUSES
  let timer: ReturnType<typeof setTimeout> | null = null
  let resolvePolling: ((value: T) => void) | null = null
  let rejectPolling: ((reason?: unknown) => void) | null = null

  function stop() {
    polling.value = false

    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  async function tick() {
    attempts.value += 1

    try {
      const snapshot = await fetchSnapshot()
      current.value = snapshot

      if (terminalStatuses.includes(snapshot.status)) {
        stop()
        resolvePolling?.(snapshot)
        return
      }

      if (attempts.value >= maxAttempts) {
        stop()
        const timeoutError = new Error('Polling attempts exceeded')
        error.value = timeoutError
        rejectPolling?.(timeoutError)
        return
      }

      timer = setTimeout(tick, intervalMs)
    } catch (currentError) {
      stop()
      error.value = currentError
      rejectPolling?.(currentError)
    }
  }

  function start() {
    stop()
    attempts.value = 0
    error.value = null
    polling.value = true

    const promise = new Promise<T>((resolve, reject) => {
      resolvePolling = resolve
      rejectPolling = reject
    })

    if (options.immediate ?? true) {
      void tick()
    } else {
      timer = setTimeout(tick, intervalMs)
    }

    return promise
  }

  onScopeDispose(stop)

  return {
    current,
    polling,
    attempts,
    error,
    start,
    stop,
  }
}
