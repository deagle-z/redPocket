import { describe, expect, it } from 'vitest'
import { buildIosWebClipProfile } from './iosWebClip'

describe('iOS Web Clip profile', () => {
  it('builds a removable PP.PE Web Clip mobileconfig with escaped label and URL', () => {
    const profile = buildIosWebClipProfile({
      iconBase64: 'base64-icon',
      label: 'PP.PE & Casino',
      url: 'https://example.com/?a=1&b=<2>',
    })

    expect(profile).toContain('<key>PayloadType</key>')
    expect(profile).toContain('<string>Configuration</string>')
    expect(profile).toContain('<string>com.apple.webClip.managed</string>')
    expect(profile).toContain('<key>Label</key>')
    expect(profile).toContain('<string>PP.PE &amp; Casino</string>')
    expect(profile).toContain('<key>URL</key>')
    expect(profile).toContain('<string>https://example.com/?a=1&amp;b=&lt;2&gt;</string>')
    expect(profile).toContain('<key>IsRemovable</key>')
    expect(profile).toContain('<true/>')
    expect(profile).toContain('<key>Precomposed</key>')
    expect(profile).toContain('<key>Icon</key>')
    expect(profile).toContain('<data>base64-icon</data>')
  })
})
