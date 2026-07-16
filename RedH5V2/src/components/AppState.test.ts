import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const source = readFileSync(new URL('./AppState.vue', import.meta.url), 'utf8') as string

describe('AppState loading presentation', () => {
  it('uses the PP.MX skeleton for loading states', () => {
    expect(source).toContain("import PpmxSkeleton from '@/components/PpmxSkeleton.vue'")
    expect(source).toContain("import type { SkeletonVariant } from '@/components/PpmxSkeleton.vue'")
    expect(source).toContain('skeletonVariant?: SkeletonVariant')
    expect(source).toContain('status === \'loading\'')
    expect(source).toContain(':variant="skeletonVariant || \'page\'"')
    expect(source).toContain(':aria-label="title || t(preset.titleKey)"')
    expect(source).not.toContain('<van-loading')
  })
})
