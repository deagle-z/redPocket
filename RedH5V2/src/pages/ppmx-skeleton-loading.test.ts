import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')

function read(path: string) {
  return readFileSync(new URL(`../../${path}`, import.meta.url), 'utf8') as string
}

describe('PP.PE loading skeleton coverage', () => {
  it('uses skeletons for home lobby and check-in data loading', () => {
    const source = read('src/pages/ppmx-home/PpmxHome.vue')

    expect(source).toContain("import PpmxSkeleton from '@/components/PpmxSkeleton.vue'")
    expect(source).toContain('variant="home-games"')
    expect(source).toContain('variant="checkin-modal"')
    expect(source).not.toContain('ppmx-game-loading')
    expect(source).not.toContain('fa-circle-notch fa-spin" />\n              <span>{{ pickText(checkinText.loading) }}</span>')
  })

  it('uses page-specific skeleton variants for AppState-backed page loading', () => {
    const bank = read('src/pages/bank/index.vue')
    const recharge = read('src/pages/recharge/index.vue')
    const profile = read('src/pages/profile/index.vue')
    const team = read('src/pages/team/index.vue')
    const casino = read('src/pages/casino/index.vue')
    const records = read('src/pages/records/index.vue')
    const wheel = read('src/pages/wheel/index.vue')
    const userProfile = read('src/pages/user-profile/index.vue')

    expect(bank).toContain('skeleton-variant="bank-page"')
    expect(recharge).toContain('skeleton-variant="recharge-page"')
    expect(profile).toContain('skeleton-variant="profile-page"')
    expect(team).toContain('skeleton-variant="team-page"')
    expect(casino).toContain('skeleton-variant="casino-page"')
    expect(records).toContain('skeleton-variant="records-page"')
    expect(wheel).toContain('skeleton-variant="wheel-page"')
    expect(userProfile).toContain('skeleton-variant="user-profile-page"')
  })

  it('uses semantic skeletons for bank, profile, team, casino, support, and play content loading', () => {
    const bank = read('src/pages/bank/index.vue')
    const profile = read('src/pages/profile/index.vue')
    const team = read('src/pages/team/index.vue')
    const casino = read('src/pages/casino/index.vue')
    const play = read('src/pages/play/index.vue')
    const support = read('src/components/PpmxSupportModal.vue')

    expect(bank).toContain('PpmxSkeleton')
    expect(bank).toContain('variant="bank-account-list"')
    expect(bank).not.toContain('<van-loading')
    expect(profile).toContain('variant="withdraw-modal"')
    expect(team).toContain('id="teamV2TaskHistoryModal"')
    expect(team).toContain('class="ppmx-team-task-history-state"')
    expect(casino).toContain('variant="casino-more"')
    expect(casino).toContain('v-if="loadingMore"')
    expect(play).toContain('variant="play-launch"')
    expect(play).toContain('variant="play-frame"')
    expect(support).toContain('skeleton-variant="support-modal"')
    expect(play).not.toContain('ppmx-play-loader__ring')
  })

  it('keeps action button loading indicators outside skeleton replacement', () => {
    const button = read('src/components/PpmxButton.vue')
    const dialog = read('src/components/PpmxDialog.vue')
    const profile = read('src/pages/user-profile/index.vue')
    const home = read('src/pages/ppmx-home/PpmxHome.vue')

    expect(button).toContain('fa-circle-notch fa-spin')
    expect(dialog).toContain('confirmLoading')
    expect(dialog).toContain('fa-spinner fa-spin')
    expect(profile).toContain("uploadingAvatar ? 'fa-spinner fa-spin' : 'fa-camera'")
    expect(home).toContain('v-if="checkinSubmitting" class="fa-solid fa-circle-notch fa-spin"')
  })
})
