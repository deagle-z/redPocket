import { describe, expect, it } from 'vitest'

import { pickMostVisibleCategory } from './useCategoryScrollSpy'

const { readFileSync } = await import('node:' + 'fs')
const scrollSpySource = readFileSync(new URL('./useCategoryScrollSpy.ts', import.meta.url), 'utf8') as string

describe('useCategoryScrollSpy', () => {
  it('selects the category with the highest visible ratio', () => {
    expect(
      pickMostVisibleCategory(
        [
          { id: 'slots', ratio: 0.2 },
          { id: 'pesca', ratio: 0.72 },
          { id: 'deportes', ratio: 0.45 },
        ],
        'slots',
      ),
    ).toBe('pesca')
  })

  it('keeps the current category when no section is visible', () => {
    expect(pickMostVisibleCategory([], 'populares')).toBe('populares')
  })

  it('supports dynamic category id lists', () => {
    expect(scrollSpySource).toContain('categoryIds: string[] | Ref<string[]>')
    expect(scrollSpySource).toContain('toValue(categoryIds)')
    expect(scrollSpySource).toContain('observer.disconnect()')
  })
})
