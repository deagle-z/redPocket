import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const modalSource = readFileSync(new URL('./PpmxModal.vue', import.meta.url), 'utf8') as string
const baseModalSource = readFileSync(new URL('../../../components/PpmxBaseModal.vue', import.meta.url), 'utf8') as string
const authSource = readFileSync(new URL('./PpmxAuthModal.vue', import.meta.url), 'utf8') as string

describe('PpmxModal', () => {
  it('provides the reusable PP.PE modal shell contract', () => {
    expect(modalSource).toContain("defineOptions({ name: 'PpmxModal' })")
    expect(modalSource).toContain('PpmxBaseModal')
    expect(baseModalSource).toContain("defineOptions({ name: 'PpmxBaseModal' })")
    expect(baseModalSource).toContain('<Teleport to="body">')
    expect(baseModalSource).toContain('role="dialog"')
    expect(baseModalSource).toContain('aria-modal="true"')
    expect(modalSource).toContain('closeOnBackdrop')
    expect(modalSource).toContain('cardClass')
    expect(baseModalSource).toContain("'update:modelValue'")
    expect(baseModalSource).toContain('close: []')
  })

  it('is used by the auth modal instead of duplicating the shell markup', () => {
    expect(authSource).toContain("import PpmxModal from './PpmxModal.vue'")
    expect(authSource).toContain('<PpmxModal')
    expect(authSource).toContain('v-model="modalOpen"')
    expect(authSource).toContain('card-class="ppmx-auth-card"')
    expect(authSource).not.toContain('<Teleport to="body">')
    expect(authSource).not.toContain('class="ppmx-auth-backdrop"')
  })
})
