import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const dialogSource = readFileSync(new URL('./PpmxDialog.vue', import.meta.url), 'utf8') as string
const profileSource = readFileSync(new URL('../pages/profile/index.vue', import.meta.url), 'utf8') as string

describe('PpmxDialog', () => {
  it('provides a reusable PP.BET dialog contract', () => {
    expect(dialogSource).toContain("defineOptions({ name: 'PpmxDialog' })")
    expect(dialogSource).toContain('<Teleport to="body">')
    expect(dialogSource).toContain('role="dialog"')
    expect(dialogSource).toContain('aria-modal="true"')
    expect(dialogSource).toContain('variant?:')
    expect(dialogSource).toContain('confirmLoading?:')
    expect(dialogSource).toContain('closeOnBackdrop?:')
    expect(dialogSource).toContain("'update:modelValue'")
    expect(dialogSource).toContain('confirm: []')
    expect(dialogSource).toContain('cancel: []')
    expect(dialogSource).toContain('slot name="actions"')
    expect(dialogSource).toContain('slot name="icon"')
  })

  it('uses PP.BET visual classes and motion instead of Vant dialog markup', () => {
    expect(dialogSource).toContain('.ppmx-dialog__card')
    expect(dialogSource).toContain('.ppmx-dialog__button--confirm')
    expect(dialogSource).toContain('linear-gradient(180deg, #e43b45, #b9152a)')
    expect(dialogSource).toContain('ppmx-dialog-fade')
    expect(dialogSource).not.toContain('van-dialog')
  })

  it('is used by the account promo code and logout confirmation modals', () => {
    expect(profileSource).toContain("import PpmxDialog from '@/components/PpmxDialog.vue'")
    expect(profileSource).toContain('promoModalOpen')
    expect(profileSource).toContain('logoutDialogOpen')
    expect(profileSource).toContain('openLogoutDialog')
    expect(profileSource).toContain('<PpmxDialog')
    expect(profileSource).toContain('id="promoModal"')
    expect(profileSource).toContain('@confirm="submitPromoCode"')
    expect(profileSource).toContain('@confirm="logoutAccount"')
    expect(profileSource).not.toContain('showConfirmDialog')
  })
})
