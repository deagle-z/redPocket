import { computed, ref } from 'vue'

export interface WidgetPoint {
  x: number
  y: number
}

export interface ConstrainWidgetPositionOptions {
  height: number
  position: WidgetPoint
  safeArea: number
  widgetHeight: number
  widgetWidth: number
  width: number
}

export interface WidgetDragEvent {
  clientX: number
  clientY: number
  preventDefault?: () => void
}

const WIDGET_WIDTH = 74
const WIDGET_HEIGHT = 82
const WIDGET_GAP = 88
const WIDGET_EDGE_SAFE_AREA = 18
const WIDGET_DRAG_SAFE_AREA = 16
const MOBILE_TABBAR_SAFE_AREA = 96
const SCROLL_HIDE_DELAY = 260
const WIDGET_STACK_SIZE = 3

export function constrainWidgetPosition(options: ConstrainWidgetPositionOptions): WidgetPoint {
  const minX = options.safeArea
  const minY = options.safeArea
  const maxX = Math.max(minX, options.width - options.widgetWidth - options.safeArea)
  const maxY = Math.max(minY, options.height - options.widgetHeight - options.safeArea)

  return {
    x: Math.min(Math.max(options.position.x, minX), maxX),
    y: Math.min(Math.max(options.position.y, minY), maxY),
  }
}

export function useFloatingWidgets() {
  const hiddenWidgetIds = ref(new Set<string>())
  const isScrollHiding = ref(false)
  const widgetPositions = ref<Record<string, WidgetPoint>>({})
  const dragState = ref<{
    id: string
    offsetX: number
    offsetY: number
  } | null>(null)
  let scrollHideTimer: ReturnType<typeof setTimeout> | null = null

  const isDragging = computed(() => dragState.value !== null)

  function hideWidget(id: string) {
    hiddenWidgetIds.value = new Set(hiddenWidgetIds.value).add(id)
  }

  function isWidgetVisible(id: string) {
    return !hiddenWidgetIds.value.has(id)
  }

  function getDefaultPosition(index: number): WidgetPoint {
    const viewport = getViewportSize()
    const groupHeight = WIDGET_HEIGHT + WIDGET_GAP * (WIDGET_STACK_SIZE - 1)
    const preferredY = Math.round(viewport.height * 0.58 - groupHeight / 2 + index * WIDGET_GAP)
    const bottomSafeArea = viewport.width < 1024 ? MOBILE_TABBAR_SAFE_AREA : WIDGET_EDGE_SAFE_AREA
    const maxY = Math.max(WIDGET_EDGE_SAFE_AREA, viewport.height - WIDGET_HEIGHT - bottomSafeArea)

    return constrainWidgetPosition({
      ...viewport,
      position: {
        x: viewport.width - WIDGET_WIDTH - WIDGET_EDGE_SAFE_AREA,
        y: Math.min(preferredY, maxY),
      },
      safeArea: WIDGET_EDGE_SAFE_AREA,
      widgetHeight: WIDGET_HEIGHT,
      widgetWidth: WIDGET_WIDTH,
    })
  }

  function getViewportSize() {
    return {
      height: typeof window === 'undefined' ? 640 : window.innerHeight,
      width: typeof window === 'undefined' ? 390 : window.innerWidth,
    }
  }

  function getWidgetPosition(id: string, index = 0) {
    return widgetPositions.value[id] || getDefaultPosition(index)
  }

  function getWidgetStyle(id: string, index = 0) {
    const position = getWidgetPosition(id, index)

    return {
      left: `${position.x}px`,
      top: `${position.y}px`,
    }
  }

  function startWidgetDrag(id: string, event: WidgetDragEvent, index = 0) {
    event.preventDefault?.()
    const position = getWidgetPosition(id, index)

    dragState.value = {
      id,
      offsetX: event.clientX - position.x,
      offsetY: event.clientY - position.y,
    }
  }

  function dragWidget(event: WidgetDragEvent) {
    if (!dragState.value) return

    const viewport = getViewportSize()
    const nextPosition = constrainWidgetPosition({
      ...viewport,
      position: {
        x: event.clientX - dragState.value.offsetX,
        y: event.clientY - dragState.value.offsetY,
      },
      safeArea: WIDGET_DRAG_SAFE_AREA,
      widgetHeight: WIDGET_HEIGHT,
      widgetWidth: WIDGET_WIDTH,
    })

    widgetPositions.value = {
      ...widgetPositions.value,
      [dragState.value.id]: nextPosition,
    }
  }

  function endWidgetDrag() {
    dragState.value = null
  }

  function resetWidgetsToSafeArea() {
    const viewport = getViewportSize()
    widgetPositions.value = Object.fromEntries(
      Object.entries(widgetPositions.value).map(([id, position]) => [
        id,
        constrainWidgetPosition({
          ...viewport,
          position,
          safeArea: WIDGET_DRAG_SAFE_AREA,
          widgetHeight: WIDGET_HEIGHT,
          widgetWidth: WIDGET_WIDTH,
        }),
      ]),
    )
  }

  function clearScrollHideTimer() {
    if (!scrollHideTimer) return
    const clearTimer = typeof window === 'undefined' ? clearTimeout : window.clearTimeout
    clearTimer(scrollHideTimer)
    scrollHideTimer = null
  }

  function handlePageScroll() {
    if (isDragging.value) return

    isScrollHiding.value = true
    clearScrollHideTimer()

    const setTimer = typeof window === 'undefined' ? setTimeout : window.setTimeout
    scrollHideTimer = setTimer(() => {
      isScrollHiding.value = false
      scrollHideTimer = null
    }, SCROLL_HIDE_DELAY)
  }

  function stopScrollWatcher() {
    clearScrollHideTimer()
    isScrollHiding.value = false
  }

  return {
    hiddenWidgetIds,
    isDragging,
    isScrollHiding,
    widgetPositions,
    dragWidget,
    endWidgetDrag,
    getWidgetStyle,
    handlePageScroll,
    hideWidget,
    isWidgetVisible,
    resetWidgetsToSafeArea,
    startWidgetDrag,
    stopScrollWatcher,
  }
}
