import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const source = readFileSync(new URL('./PpmxSkeleton.vue', import.meta.url), 'utf8') as string

describe('PpmxSkeleton source contract', () => {
  it('exposes PP.BET skeleton variants and accessible busy status', () => {
    for (const variant of [
      'page',
      'card-list',
      'game-grid',
      'form',
      'modal',
      'checkin',
      'play',
      'inline',
      'home-games',
      'profile-page',
      'bank-page',
      'recharge-page',
      'team-page',
      'casino-page',
      'records-page',
      'wheel-page',
      'user-profile-page',
      'checkin-modal',
      'support-modal',
      'withdraw-modal',
      'commission-withdraw-modal',
      'bank-account-list',
      'casino-more',
      'play-launch',
      'play-frame',
    ]) {
      expect(source).toContain(`'${variant}'`)
    }

    expect(source).toContain('role="status"')
    expect(source).toContain('aria-busy="true"')
    expect(source).toContain(':data-variant="variant"')
  })

  it('renders semantic page and modal skeleton layouts instead of only generic blocks', () => {
    expect(source).toContain("const semanticPageVariants = new Set<SkeletonVariant>")
    expect(source).toContain("const semanticModalVariants = new Set<SkeletonVariant>")
    expect(source).toContain('isSemanticPageVariant')
    expect(source).toContain('isSemanticModalVariant')
    expect(source).toContain('ppmx-skeleton__page-shell')
    expect(source).toContain('ppmx-skeleton__modal-shell')
    expect(source).toContain('ppmx-skeleton__metric-row')
    expect(source).toContain('ppmx-skeleton__wheel-stage')
    expect(source).toContain('ppmx-skeleton__support-channel')
    expect(source).toContain('ppmx-skeleton__withdraw-balance-compact')
  })

  it('uses a local shimmer animation without external dependencies', () => {
    expect(source).toContain('ppmx-skeleton-shimmer')
    expect(source).toContain('@keyframes ppmxSkeletonShimmer')
    expect(source).toContain('prefers-reduced-motion')
    expect(source).not.toContain('van-skeleton')
  })

  it('has a dedicated light palette for the wheel page skeleton', () => {
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--wheel-page')
    expect(source).toContain('--skeleton-border: rgba(255, 26, 26, 0.13);')
    expect(source).toContain('--skeleton-panel-strong: rgba(255, 255, 255, 0.96);')
    expect(source).toContain('--skeleton-sheen: rgba(255, 255, 255, 0.72);')
  })

  it('has a dedicated light palette for casino game list skeletons', () => {
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--casino-page,')
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--casino-more')
    expect(source).toContain('--skeleton-base: rgba(226, 31, 51, 0.032);')
    expect(source).toContain('--skeleton-strong: rgba(226, 31, 51, 0.072);')
    expect(source).toContain('--skeleton-panel: rgba(255, 255, 255, 0.88);')
    expect(source).toContain('--skeleton-border: rgba(226, 31, 51, 0.105);')
    expect(source).toContain('--skeleton-gold: rgba(255, 215, 0, 0.12);')
  })

  it('has dedicated light palettes for account and payment page skeletons', () => {
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--home-games,')
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--user-profile-page')
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--bank-page,')
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--bank-account-list,')
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--recharge-page')
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--team-page,')
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--records-page')
    expect(source).toContain('--skeleton-border: rgba(226, 31, 51, 0.105);')
    expect(source).toContain('--skeleton-panel-strong: rgba(255, 255, 255, 0.96);')
  })

  it('has dedicated light palettes for support and withdraw modal skeletons', () => {
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--support-modal,')
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--withdraw-modal,')
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--commission-withdraw-modal')
    expect(source).toContain('--skeleton-border: rgba(226, 31, 51, 0.12);')
    expect(source).toContain('--skeleton-panel-strong: rgba(255, 255, 255, 0.97);')
  })

  it('has dedicated light palettes for play launch and frame skeletons', () => {
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--play-launch,')
    expect(source).toContain(':global(html[data-theme="light"]) .ppmx-skeleton--play-frame')
    expect(source).toContain('--skeleton-base: rgba(25, 27, 34, 0.06);')
    expect(source).toContain('--skeleton-panel: rgba(255, 255, 255, 0.84);')
    expect(source).toContain('--skeleton-sheen: rgba(255, 255, 255, 0.76);')
  })

  it('adds breathing room above the home game list skeleton', () => {
    expect(source).toContain('.ppmx-skeleton--home-games {')
    expect(source).toContain('margin-top: 12px;')
  })

  it('renders play launch and frame skeletons as an immersive game room', () => {
    expect(source).toContain("variant === 'play-launch' || variant === 'play-frame'")
    expect(source).toContain('ppmx-skeleton__play-room')
    expect(source).toContain('ppmx-skeleton__play-room-bar')
    expect(source).toContain('ppmx-skeleton__play-room-status')
    expect(source).toContain('.ppmx-skeleton--play-launch .ppmx-skeleton__modal-shell')
    expect(source).toContain('.ppmx-skeleton--play-frame .ppmx-skeleton__modal-shell')
  })
})
