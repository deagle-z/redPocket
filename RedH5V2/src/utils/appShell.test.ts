import { describe, expect, it } from 'vitest'
import { resolveAppShell } from './appShell'

describe('resolveAppShell', () => {
  it('uses the default shell when route meta does not request a shell', () => {
    expect(resolveAppShell({})).toBe('default')
  })

  it('accepts the dedicated PP.MX shell for exact homepage replicas', () => {
    expect(resolveAppShell({ shell: 'ppmx' })).toBe('ppmx')
  })

  it('accepts shell bypass for full-screen runtime pages', () => {
    expect(resolveAppShell({ shell: 'none' })).toBe('none')
  })
})
