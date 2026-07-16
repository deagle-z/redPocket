import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAppStore } from './app'

function installThemeDom(initialTheme?: string) {
  const data = new Map<string, string>()
  if (initialTheme) data.set('ppmx_theme', initialTheme)

  const storage = {
    get length() {
      return data.size
    },
    clear: vi.fn(() => data.clear()),
    getItem: vi.fn((key: string) => data.get(key) ?? null),
    key: vi.fn((index: number) => Array.from(data.keys())[index] ?? null),
    removeItem: vi.fn((key: string) => data.delete(key)),
    setItem: vi.fn((key: string, value: string) => data.set(key, value)),
  }
  const meta = { content: '#060608' }
  const documentElement = {
    dataset: {} as Record<string, string>,
    style: {} as Record<string, string>,
  }

  vi.stubGlobal('window', { localStorage: storage })
  vi.stubGlobal('document', {
    documentElement,
    querySelector: vi.fn((selector: string) => selector === 'meta[name="theme-color"]' ? meta : null),
  })

  return { documentElement, meta, storage }
}

describe('app store theme', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    vi.clearAllMocks()
    setActivePinia(createPinia())
  })

  it('defaults to dark when there is no persisted PP.MX theme', () => {
    const { documentElement, meta, storage } = installThemeDom()
    const store = useAppStore()

    expect(store.theme).toBe('dark')

    store.initTheme()

    expect(documentElement.dataset.theme).toBe('dark')
    expect(documentElement.style.colorScheme).toBe('dark')
    expect(meta.content).toBe('#060608')
    expect(storage.getItem('ppmx_theme')).toBe('dark')
  })

  it('keeps a valid persisted light theme and updates browser theme color', () => {
    const { documentElement, meta, storage } = installThemeDom('light')
    const store = useAppStore()

    store.initTheme()

    expect(store.theme).toBe('light')
    expect(documentElement.dataset.theme).toBe('light')
    expect(documentElement.style.colorScheme).toBe('light')
    expect(meta.content).toBe('#ffffff')
    expect(storage.getItem('ppmx_theme')).toBe('light')
  })

  it('falls back to dark for invalid persisted values and can toggle theme', () => {
    const { documentElement, meta, storage } = installThemeDom('system')
    const store = useAppStore()

    store.initTheme()
    expect(store.theme).toBe('dark')

    store.toggleTheme()

    expect(store.theme).toBe('light')
    expect(documentElement.dataset.theme).toBe('light')
    expect(meta.content).toBe('#ffffff')
    expect(storage.getItem('ppmx_theme')).toBe('light')
  })
})
