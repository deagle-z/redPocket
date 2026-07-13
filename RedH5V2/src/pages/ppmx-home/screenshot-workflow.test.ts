import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')

describe('PP.BET screenshot workflow', () => {
  it('captures the required reference and Vue home viewports', () => {
    const script = readFileSync(
      new URL('../../../scripts/capture-ppmx-screenshots.mjs', import.meta.url),
      'utf8',
    ) as string

    for (const viewport of ['375x812', '390x844', '430x932', '768x1024', '1024x768', '1440x900']) {
      expect(script).toContain(viewport)
    }

    expect(script).toContain('pp-mx-apple-v2.reference.html')
    expect(script).toContain('http://127.0.0.1:4173/')
  })
})
