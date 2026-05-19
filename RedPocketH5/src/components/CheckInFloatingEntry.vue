<script setup lang="ts">
const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const FLOAT_POSITION_KEY = 'checkin_floating_entry_position'
const EDGE_PADDING = 8
const DEFAULT_RIGHT = 12
const DEFAULT_BOTTOM = 86
const DRAG_THRESHOLD = 4

const entryRef = ref<HTMLButtonElement>()
const floatingPosition = ref<{ x: number, y: number } | null>(null)
const dragging = ref(false)
const hasDragged = ref(false)
const dragStart = {
  pointerId: 0,
  pointerX: 0,
  pointerY: 0,
  x: 0,
  y: 0,
}

const showEntry = computed(() => route.path !== '/checkin')
const floatingStyle = computed(() => {
  if (!floatingPosition.value)
    return undefined

  return {
    left: `${floatingPosition.value.x}px`,
    top: `${floatingPosition.value.y}px`,
    right: 'auto',
    bottom: 'auto',
  }
})

function goCheckIn() {
  if (hasDragged.value) {
    hasDragged.value = false
    return
  }

  router.push('/checkin')
}

function getEntrySize() {
  const rect = entryRef.value?.getBoundingClientRect()
  return {
    width: rect?.width || 54,
    height: rect?.height || 42,
  }
}

function clampPosition(x: number, y: number) {
  const { width, height } = getEntrySize()
  return {
    x: Math.min(Math.max(EDGE_PADDING, x), window.innerWidth - width - EDGE_PADDING),
    y: Math.min(Math.max(EDGE_PADDING, y), window.innerHeight - height - EDGE_PADDING),
  }
}

function resolveDefaultPosition() {
  const { width, height } = getEntrySize()
  return clampPosition(
    window.innerWidth - width - DEFAULT_RIGHT,
    window.innerHeight - height - DEFAULT_BOTTOM,
  )
}

function readStoredPosition() {
  try {
    const saved = JSON.parse(localStorage.getItem(FLOAT_POSITION_KEY) || '')
    if (Number.isFinite(saved?.x) && Number.isFinite(saved?.y))
      return clampPosition(Number(saved.x), Number(saved.y))
  }
  catch {}

  return resolveDefaultPosition()
}

function saveFloatingPosition() {
  if (!floatingPosition.value)
    return
  localStorage.setItem(FLOAT_POSITION_KEY, JSON.stringify(floatingPosition.value))
}

function updateFloatingPosition(x: number, y: number) {
  floatingPosition.value = clampPosition(x, y)
}

function handlePointerDown(event: PointerEvent) {
  if (!entryRef.value)
    return

  const current = floatingPosition.value || resolveDefaultPosition()
  floatingPosition.value = current
  dragging.value = true
  hasDragged.value = false
  dragStart.pointerId = event.pointerId
  dragStart.pointerX = event.clientX
  dragStart.pointerY = event.clientY
  dragStart.x = current.x
  dragStart.y = current.y
  entryRef.value.setPointerCapture(event.pointerId)
}

function handlePointerMove(event: PointerEvent) {
  if (!dragging.value || event.pointerId !== dragStart.pointerId)
    return

  const diffX = event.clientX - dragStart.pointerX
  const diffY = event.clientY - dragStart.pointerY
  if (Math.hypot(diffX, diffY) > DRAG_THRESHOLD)
    hasDragged.value = true
  updateFloatingPosition(dragStart.x + diffX, dragStart.y + diffY)
  event.preventDefault()
}

function handlePointerUp(event: PointerEvent) {
  if (!dragging.value || event.pointerId !== dragStart.pointerId)
    return

  dragging.value = false
  saveFloatingPosition()
  entryRef.value?.releasePointerCapture(event.pointerId)
}

function handleResize() {
  const next = floatingPosition.value
    ? clampPosition(floatingPosition.value.x, floatingPosition.value.y)
    : resolveDefaultPosition()
  floatingPosition.value = next
  saveFloatingPosition()
}

onMounted(() => {
  requestAnimationFrame(() => {
    floatingPosition.value = readStoredPosition()
  })
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <transition name="checkin-float">
    <button
      v-if="showEntry"
      ref="entryRef"
      type="button"
      class="checkin-floating-entry"
      :class="{ dragging }"
      :style="floatingStyle"
      :aria-label="t('checkInPage.floatLabel')"
      @pointerdown="handlePointerDown"
      @pointermove="handlePointerMove"
      @pointerup="handlePointerUp"
      @pointercancel="handlePointerUp"
      @click="goCheckIn"
    >
      <span class="entry-icon" aria-hidden="true">
        <van-icon name="calendar-o" />
      </span>
      <span class="entry-text">{{ t('checkInPage.floatText') }}</span>
    </button>
  </transition>
</template>

<style scoped>
.checkin-floating-entry {
  position: fixed;
  right: 12px;
  bottom: calc(86px + env(safe-area-inset-bottom));
  z-index: 120;
  min-width: 54px;
  height: 42px;
  border: 0;
  border-radius: 14px;
  padding: 5px 8px 6px;
  color: #ffe9a6;
  background:
    linear-gradient(180deg, rgba(255, 232, 151, 0.08), rgba(255, 232, 151, 0)),
    linear-gradient(145deg, rgba(126, 18, 12, 0.96), rgba(72, 4, 2, 0.96));
  box-shadow:
    0 8px 18px rgba(25, 0, 0, 0.28),
    0 0 0 1px rgba(222, 174, 75, 0.32),
    inset 0 1px 0 rgba(255, 236, 166, 0.12);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1px;
  overflow: hidden;
  transform: translateZ(0);
  touch-action: none;
  user-select: none;
  cursor: grab;
}

.checkin-floating-entry.dragging {
  cursor: grabbing;
  opacity: 0.94;
}

.checkin-floating-entry::after {
  content: '';
  position: absolute;
  inset: 1px;
  border-radius: 13px;
  border: 1px solid rgba(255, 241, 164, 0.08);
  pointer-events: none;
}

.checkin-floating-entry:active {
  transform: translateY(1px) scale(0.98);
}

.entry-icon {
  position: relative;
  z-index: 1;
  width: 18px;
  height: 18px;
  border-radius: 6px;
  background: rgba(219, 164, 62, 0.18);
  color: #f2c76f;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  line-height: 1;
}

.entry-text {
  position: relative;
  z-index: 1;
  font-size: 10px;
  line-height: 1.1;
  font-weight: 800;
  letter-spacing: 0;
  text-shadow: none;
  max-width: 42px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.checkin-float-enter-active,
.checkin-float-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.checkin-float-enter-from,
.checkin-float-leave-to {
  opacity: 0;
  transform: translateY(8px) scale(0.92);
}

@media (max-width: 360px) {
  .checkin-floating-entry {
    right: 8px;
    min-width: 50px;
    height: 40px;
  }

  .entry-text {
    font-size: 10px;
    max-width: 40px;
  }
}
</style>
