import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const supportSource = readFileSync(new URL('./PpmxSupportModal.vue', import.meta.url), 'utf8') as string
const chromeSource = readFileSync(
  new URL('../pages/ppmx-home/components/PpmxPageChrome.vue', import.meta.url),
  'utf8',
) as string
const profileSource = readFileSync(new URL('../pages/profile/index.vue', import.meta.url), 'utf8') as string
const cssSource = readFileSync(new URL('../assets/styles/pp-mx-home.css', import.meta.url), 'utf8') as string
const userApiSource = readFileSync(new URL('../api/user.ts', import.meta.url), 'utf8') as string

describe('PpmxSupportModal', () => {
  it('matches the PP.PE reference support modal contract and loads channels from API', () => {
    expect(supportSource).toContain("defineOptions({ name: 'PpmxSupportModal' })")
    expect(supportSource).toContain('<Teleport to="body">')
    expect(supportSource).toContain('id="supportModal"')
    expect(supportSource).toContain('role="dialog"')
    expect(supportSource).toContain('aria-modal="true"')
    expect(supportSource).toContain('fa-headset')
    expect(supportSource).toContain("import { getTenantServiceLinks } from '@/api/user'")
    expect(supportSource).toContain('function loadServiceLinks()')
    expect(supportSource).toContain('await getTenantServiceLinks()')
    expect(supportSource).toContain('normalizeServiceLinks')
    expect(supportSource).toContain('normalizeExternalUrl')
    expect(supportSource).toContain("linkKey: 'tgServiceUrl'")
    expect(supportSource).toContain("linkKey: 'wsServiceUrl'")
    expect(supportSource).toContain('ppmx-support-channel')
    expect(supportSource).not.toContain('https://t.me/ppmx_soporte')
    expect(supportSource).not.toContain('https://wa.me/5215512345678')
  })

  it('shows loading, error, empty, and retry states instead of falling back to static links', () => {
    expect(supportSource).toContain('<AppState')
    expect(supportSource).toContain('status="loading"')
    expect(supportSource).toContain('status="error"')
    expect(supportSource).toContain('status="empty"')
    expect(supportSource).toContain('@retry="loadServiceLinks"')
    expect(supportSource).toContain('loaded && supportChannels.length === 0')
    expect(supportSource).toContain('serviceLinks.value = normalizeServiceLinks()')
  })

  it('adds the tenant service links API wrapper without an /api prefix', () => {
    expect(userApiSource).toContain('export interface TenantServiceLinks')
    expect(userApiSource).toContain('export function getTenantServiceLinks()')
    expect(userApiSource).toContain("get<TenantServiceLinks>('/v1/app/domain/serviceLinks'")
    expect(userApiSource).toContain('meta: { showError: false }')
    expect(userApiSource).not.toContain("'/api/v1/app/domain/serviceLinks'")
  })

  it('is mounted globally by the PP.PE page chrome and opened from support targets', () => {
    expect(chromeSource).toContain("import PpmxSupportModal from '@/components/PpmxSupportModal.vue'")
    expect(chromeSource).toContain('supportModalOpen')
    expect(chromeSource).toContain('function openSupport()')
    expect(chromeSource).toContain("target === 'help' || target === 'support'")
    expect(chromeSource).toContain('openSupport,')
    expect(chromeSource).toContain('<PpmxSupportModal v-model="supportModalOpen"')
    expect(chromeSource).toContain('ppmx-top-strip__support-button')
  })

  it('opens the support modal from the account support hub item', () => {
    expect(profileSource).toContain("item.id === 'support'")
    expect(profileSource).toContain('pageChromeRef.value?.openSupport()')
    expect(profileSource).not.toContain("item.id === 'support') {\n    showUpcoming")
  })

  it('keeps top strip support visually clickable without changing the existing layout', () => {
    expect(cssSource).toContain('.ppmx-top-strip__support-button')
    expect(cssSource).toContain('.ppmx-top-strip__support-button:hover')
    expect(cssSource).toContain('.ppmx-top-strip__support-button i')
  })
})
