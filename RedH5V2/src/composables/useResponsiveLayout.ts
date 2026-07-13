import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import type { DeviceKind } from '@/types/business'

const PC_MEDIA_QUERY = '(min-width: 1024px)'
const PC_BREAKPOINT = 1024

function readPcViewport() {
  return (
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia(PC_MEDIA_QUERY).matches
  )
}

const isPcViewport = ref(readPcViewport())
let mediaQuery: MediaQueryList | null = null
let consumers = 0

function syncViewport(event?: MediaQueryListEvent) {
  isPcViewport.value = event?.matches ?? mediaQuery?.matches ?? readPcViewport()
}

function bindMediaQuery() {
  if (
    typeof window === 'undefined' ||
    typeof window.matchMedia !== 'function' ||
    mediaQuery
  ) {
    return
  }

  mediaQuery = window.matchMedia(PC_MEDIA_QUERY)
  syncViewport()

  if (typeof mediaQuery.addEventListener === 'function') {
    mediaQuery.addEventListener('change', syncViewport)
    return
  }

  mediaQuery.addListener(syncViewport)
}

function unbindMediaQuery() {
  if (!mediaQuery) return

  if (typeof mediaQuery.removeEventListener === 'function') {
    mediaQuery.removeEventListener('change', syncViewport)
  } else {
    mediaQuery.removeListener(syncViewport)
  }

  mediaQuery = null
}

export function useResponsiveLayout() {
  onMounted(() => {
    consumers += 1
    bindMediaQuery()
  })

  onBeforeUnmount(() => {
    consumers = Math.max(0, consumers - 1)

    if (consumers === 0) {
      unbindMediaQuery()
    }
  })

  const isPc = computed(() => isPcViewport.value)
  const isMobile = computed(() => !isPc.value)
  const deviceKind = computed<DeviceKind>(() => (isPc.value ? 'pc' : 'mobile'))

  return {
    breakpoint: PC_BREAKPOINT,
    deviceKind,
    isMobile,
    isPc,
  }
}
