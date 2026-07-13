import { computed, ref } from 'vue'
import type { PageStateStatus } from '@/types/business'
import { i18n } from '@/plugins/i18n'

const t = i18n.global.t

export function usePageState(initialStatus: PageStateStatus = 'idle') {
  const status = ref<PageStateStatus>(initialStatus)
  const message = ref('')
  const detail = ref('')
  const error = ref<unknown>(null)

  const loading = computed(() => status.value === 'loading')
  const ready = computed(() => status.value === 'ready')
  const hasError = computed(() =>
    ['network', 'rateLimit', 'maintenance', 'fallback', 'error'].includes(
      status.value,
    ),
  )

  function startLoading(currentMessage = t('common.loading')) {
    status.value = 'loading'
    message.value = currentMessage
    detail.value = ''
    error.value = null
  }

  function setReady() {
    status.value = 'ready'
    message.value = ''
    detail.value = ''
    error.value = null
  }

  function setEmpty(currentMessage = t('state.empty.title')) {
    status.value = 'empty'
    message.value = currentMessage
    detail.value = ''
    error.value = null
  }

  function setError(
    currentError: unknown,
    currentStatus: PageStateStatus = 'error',
  ) {
    status.value = currentStatus
    error.value = currentError
    message.value =
      currentError instanceof Error ? currentError.message : t('error.requestFailed')
    detail.value = ''
  }

  return {
    status,
    message,
    detail,
    error,
    loading,
    ready,
    hasError,
    startLoading,
    setReady,
    setEmpty,
    setError,
  }
}
