import {
  getTtlStorage,
  removeTtlStorage,
  setTtlStorage,
} from '@/utils/storage'

export function useTtlLocalStorage<T>(
  key: string,
  initialValue: T,
  ttlMs?: number,
) {
  const stored = getTtlStorage<T>(key)
  const state = ref<T>(stored ?? initialValue)

  watch(
    state,
    value => {
      setTtlStorage(key, value, ttlMs)
    },
    { deep: true },
  )

  function remove() {
    removeTtlStorage(key)
    state.value = initialValue
  }

  return {
    state,
    remove,
  }
}
