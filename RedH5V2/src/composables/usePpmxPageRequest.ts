import { shallowRef } from 'vue'
import type { PageStateStatus } from '@/types/business'
import { usePageState } from './usePageState'

interface PpmxPageRequestOptions<T> {
  emptyMessage?: string
  errorStatus?: PageStateStatus
  isEmpty?: (data: T) => boolean
  loadingMessage?: string
}

export function usePpmxPageRequest<T>(
  loader: () => Promise<T>,
  options: PpmxPageRequestOptions<T> = {},
) {
  const state = usePageState()
  const data = shallowRef<T | null>(null)

  async function run() {
    state.startLoading(options.loadingMessage)

    try {
      const result = await loader()
      data.value = result

      if (options.isEmpty?.(result)) {
        state.setEmpty(options.emptyMessage)
      } else {
        state.setReady()
      }

      return result
    } catch (error) {
      state.setError(error, options.errorStatus)
      return null
    }
  }

  return {
    ...state,
    data,
    run,
  }
}
