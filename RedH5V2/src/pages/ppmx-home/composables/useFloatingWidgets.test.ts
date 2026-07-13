import { afterEach, describe, expect, it, vi } from 'vitest'

import { constrainWidgetPosition, useFloatingWidgets } from './useFloatingWidgets'

function stubViewport(width: number, height: number) {
  vi.stubGlobal('window', {
    innerHeight: height,
    innerWidth: width,
    clearTimeout: globalThis.clearTimeout,
    setTimeout: globalThis.setTimeout,
  })
}

describe('useFloatingWidgets', () => {
  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('hides widgets without mutating the previous hidden set', () => {
    const floating = useFloatingWidgets()
    const before = floating.hiddenWidgetIds.value

    floating.hideWidget('checkin')

    expect(before.has('checkin')).toBe(false)
    expect(floating.isWidgetVisible('checkin')).toBe(false)
    expect(floating.isWidgetVisible('social')).toBe(true)
  })

  it('keeps dragged widgets inside the viewport safe area', () => {
    expect(
      constrainWidgetPosition({
        height: 640,
        width: 390,
        position: { x: 360, y: 610 },
        safeArea: 16,
        widgetHeight: 82,
        widgetWidth: 74,
      }),
    ).toEqual({ x: 300, y: 542 })
  })

  it('places the default widget stack on the right side around the middle-lower viewport', () => {
    stubViewport(390, 844)
    const floating = useFloatingWidgets()

    expect(floating.getWidgetStyle('checkin', 0)).toEqual({ left: '298px', top: '361px' })
    expect(floating.getWidgetStyle('social', 1)).toEqual({ left: '298px', top: '449px' })
    expect(floating.getWidgetStyle('download', 2)).toEqual({ left: '298px', top: '537px' })
  })

  it('hides widgets while scrolling and restores them after the scroll settles', () => {
    vi.useFakeTimers()
    stubViewport(390, 844)
    const floating = useFloatingWidgets()

    floating.handlePageScroll()

    expect(floating.isScrollHiding.value).toBe(true)
    vi.advanceTimersByTime(259)
    expect(floating.isScrollHiding.value).toBe(true)
    vi.advanceTimersByTime(1)
    expect(floating.isScrollHiding.value).toBe(false)
  })

  it('does not start the scroll hide animation while a widget is being dragged', () => {
    vi.useFakeTimers()
    stubViewport(390, 844)
    const floating = useFloatingWidgets()

    floating.startWidgetDrag('checkin', { clientX: 320, clientY: 390 }, 0)
    floating.handlePageScroll()

    expect(floating.isScrollHiding.value).toBe(false)
  })
})
