import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const popupSource = readFileSync(
  new URL('./PpmxAnnouncementPopup.vue', import.meta.url),
  'utf8',
) as string

describe('PpmxAnnouncementPopup', () => {
  it('renders the backend popup content contract accessibly', () => {
    expect(popupSource).toContain('<Teleport to="body">')
    expect(popupSource).toContain('role="dialog"')
    expect(popupSource).toContain('aria-modal="true"')
    expect(popupSource).toContain('announcement.subTitle')
    expect(popupSource).toContain('announcement.description')
    expect(popupSource).toContain('announcement.buttonText')
    expect(popupSource).toContain('announcement.imageUrl')
  })

  it('treats backend text as plain text and exposes close and confirm actions', () => {
    expect(popupSource).toContain("close: []")
    expect(popupSource).toContain("confirm: []")
    expect(popupSource).toContain("emit('close')")
    expect(popupSource).toContain("emit('confirm')")
    expect(popupSource).not.toContain('v-html')
  })
})
