import { describe, expect, it } from 'vitest'
import { messages } from '@/locales'
import { formatDocumentTitle, resolveRouteTitle } from './documentTitle'

type MessageTree = Record<string, unknown>

function routeTitleKeys(): string[] {
  const pageSources = import.meta.glob('../pages/**/*.vue', {
    eager: true,
    import: 'default',
    query: '?raw',
  }) as Record<string, string>
  const keys = new Set<string>()

  for (const source of Object.values(pageSources)) {
    for (const match of source.matchAll(/titleKey\s*:\s*['"]([^'"]+)['"]/g)) {
      keys.add(match[1])
    }
  }

  return [...keys].sort()
}

function messageAtPath(tree: MessageTree, path: string): unknown {
  return path.split('.').reduce<unknown>((current, segment) => {
    if (!current || typeof current !== 'object') return undefined
    return (current as MessageTree)[segment]
  }, tree)
}

describe('route document titles', () => {
  it('defines every route title key in every locale', () => {
    const keys = routeTitleKeys()
    const invalidTitles: string[] = []

    for (const [locale, tree] of Object.entries(messages)) {
      for (const key of keys) {
        const value = messageAtPath(tree as MessageTree, key)
        if (typeof value !== 'string' || !value.trim() || value === key) {
          invalidTitles.push(`${locale}:${key}`)
        }
      }
    }

    expect(invalidTitles, 'missing, empty, or leaked route titles').toEqual([])
  })

  it('uses the real game name for play routes', () => {
    const title = resolveRouteTitle(
      {
        name: 'play',
        query: { title: '  Monkey Warrior  ' },
        meta: { titleKey: 'play.title' },
      },
      key => key === 'play.title' ? 'Game' : key,
      key => key === 'play.title',
      'PP.MX',
    )

    expect(formatDocumentTitle(title, 'PP.MX')).toBe('Monkey Warrior | PP.MX')
  })

  it('never exposes a missing title key', () => {
    const title = resolveRouteTitle(
      {
        name: 'unknown',
        query: {},
        meta: { titleKey: 'missing.title' },
      },
      key => key,
      () => false,
      'PP.MX',
    )

    expect(formatDocumentTitle(title, 'PP.MX')).toBe('PP.MX')
  })
})
