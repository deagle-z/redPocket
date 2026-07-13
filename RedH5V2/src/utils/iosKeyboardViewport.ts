const VISUAL_HEIGHT_VAR = '--app-visual-height'
const KEYBOARD_INSET_VAR = '--app-keyboard-inset'
const KEYBOARD_OPEN_CLASS = 'is-ios-keyboard-open'
const KEYBOARD_OPEN_THRESHOLD = 80

function isEditableElement(target: Element | null): boolean {
  if (!target) return false

  const tagName = target.tagName.toLowerCase()

  return tagName === 'input' || tagName === 'textarea' || tagName === 'select' || target.hasAttribute('contenteditable')
}

function getViewportHeight(): number {
  const visualViewportHeight = window.visualViewport?.height
  const fallbackHeight = window.innerHeight || document.documentElement.clientHeight || 0

  return Math.round(visualViewportHeight || fallbackHeight)
}

function getKeyboardInset(visualHeight: number): number {
  const layoutHeight = window.innerHeight || visualHeight
  const offsetTop = window.visualViewport?.offsetTop || 0

  return Math.max(0, Math.round(layoutHeight - visualHeight - offsetTop))
}

export function initIOSKeyboardViewport(): () => void {
  if (typeof window === 'undefined' || typeof document === 'undefined') {
    return () => {}
  }

  const root = document.documentElement
  const visualViewport = window.visualViewport
  let animationFrame = 0
  let editableFocused = isEditableElement(document.activeElement)

  const applyViewportState = () => {
    const visualHeight = getViewportHeight()
    const keyboardInset = getKeyboardInset(visualHeight)
    const keyboardOpen = editableFocused && keyboardInset > KEYBOARD_OPEN_THRESHOLD

    root.style.setProperty(VISUAL_HEIGHT_VAR, `${visualHeight}px`)
    root.style.setProperty(KEYBOARD_INSET_VAR, `${keyboardInset}px`)
    root.classList.toggle(KEYBOARD_OPEN_CLASS, keyboardOpen)
  }

  const scheduleViewportState = () => {
    if (animationFrame) window.cancelAnimationFrame(animationFrame)

    animationFrame = window.requestAnimationFrame(() => {
      animationFrame = 0
      applyViewportState()
    })
  }

  const onFocusIn = () => {
    editableFocused = isEditableElement(document.activeElement)
    scheduleViewportState()
  }

  const onFocusOut = () => {
    editableFocused = false
    scheduleViewportState()
  }

  applyViewportState()
  visualViewport?.addEventListener('resize', scheduleViewportState)
  visualViewport?.addEventListener('scroll', scheduleViewportState)
  window.addEventListener('orientationchange', scheduleViewportState)
  window.addEventListener('resize', scheduleViewportState)
  document.addEventListener('focusin', onFocusIn)
  document.addEventListener('focusout', onFocusOut)
  document.addEventListener('visibilitychange', scheduleViewportState)

  return () => {
    if (animationFrame) window.cancelAnimationFrame(animationFrame)

    visualViewport?.removeEventListener('resize', scheduleViewportState)
    visualViewport?.removeEventListener('scroll', scheduleViewportState)
    window.removeEventListener('orientationchange', scheduleViewportState)
    window.removeEventListener('resize', scheduleViewportState)
    document.removeEventListener('focusin', onFocusIn)
    document.removeEventListener('focusout', onFocusOut)
    document.removeEventListener('visibilitychange', scheduleViewportState)
    root.style.removeProperty(VISUAL_HEIGHT_VAR)
    root.style.removeProperty(KEYBOARD_INSET_VAR)
    root.classList.remove(KEYBOARD_OPEN_CLASS)
  }
}
