import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { initIOSKeyboardViewport } from './iosKeyboardViewport'

type ListenerMap = Record<string, EventListener[]>

function createTarget(listeners: ListenerMap) {
  return {
    addEventListener: vi.fn((type: string, listener: EventListener) => {
      listeners[type] = [...(listeners[type] ?? []), listener]
    }),
    removeEventListener: vi.fn((type: string, listener: EventListener) => {
      listeners[type] = (listeners[type] ?? []).filter(item => item !== listener)
    }),
  }
}

describe('initIOSKeyboardViewport', () => {
  const originalWindow = globalThis.window
  const originalDocument = globalThis.document
  const originalNavigator = globalThis.navigator

  let viewportListeners: ListenerMap
  let windowListeners: ListenerMap
  let documentListeners: ListenerMap
  let classNames: Set<string>
  let styleValues: Record<string, string>
  let activeElement: Element | null
  let visualViewport: { height: number, offsetTop: number } & ReturnType<typeof createTarget>

  beforeEach(() => {
    viewportListeners = {}
    windowListeners = {}
    documentListeners = {}
    classNames = new Set()
    styleValues = {}
    activeElement = { tagName: 'INPUT' } as Element
    visualViewport = {
      ...createTarget(viewportListeners),
      height: 844,
      offsetTop: 0,
    }

    const documentElement = {
      style: {
        setProperty: vi.fn((key: string, value: string) => {
          styleValues[key] = value
        }),
        removeProperty: vi.fn((key: string) => {
          delete styleValues[key]
        }),
      },
      classList: {
        toggle: vi.fn((name: string, force?: boolean) => {
          if (force) classNames.add(name)
          else classNames.delete(name)
        }),
        remove: vi.fn((name: string) => {
          classNames.delete(name)
        }),
      },
    }
    const documentTarget = createTarget(documentListeners)

    vi.stubGlobal('window', {
      ...createTarget(windowListeners),
      innerHeight: 844,
      visualViewport,
      requestAnimationFrame: (callback: FrameRequestCallback) => {
        callback(0)
        return 1
      },
      cancelAnimationFrame: vi.fn(),
    })
    vi.stubGlobal('document', {
      ...documentTarget,
      documentElement,
      get activeElement() {
        return activeElement
      },
    })
    vi.stubGlobal('navigator', {
      userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)',
      maxTouchPoints: 5,
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    if (originalWindow) vi.stubGlobal('window', originalWindow)
    if (originalDocument) vi.stubGlobal('document', originalDocument)
    if (originalNavigator) vi.stubGlobal('navigator', originalNavigator)
  })

  it('updates visual viewport variables and keyboard class while an input is focused', () => {
    const cleanup = initIOSKeyboardViewport()

    expect(styleValues['--app-visual-height']).toBe('844px')
    expect(styleValues['--app-keyboard-inset']).toBe('0px')
    expect(classNames.has('is-ios-keyboard-open')).toBe(false)

    visualViewport.height = 520
    viewportListeners.resize[0](new Event('resize'))

    expect(styleValues['--app-visual-height']).toBe('520px')
    expect(styleValues['--app-keyboard-inset']).toBe('324px')
    expect(classNames.has('is-ios-keyboard-open')).toBe(true)

    cleanup()
    expect(styleValues['--app-visual-height']).toBeUndefined()
    expect(styleValues['--app-keyboard-inset']).toBeUndefined()
    expect(classNames.has('is-ios-keyboard-open')).toBe(false)
  })

  it('does not keep keyboard class when focus leaves editable controls', () => {
    initIOSKeyboardViewport()
    visualViewport.height = 520
    viewportListeners.resize[0](new Event('resize'))

    activeElement = { tagName: 'DIV' } as Element
    documentListeners.focusout[0](new Event('focusout'))

    expect(classNames.has('is-ios-keyboard-open')).toBe(false)
  })
})
