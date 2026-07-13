const IOS_PROFILE_FILE_NAME = 'ppmx-ios.mobileconfig'
const IOS_PROFILE_ICON_PATH = 'pwa-192x192.png'
const IOS_PROFILE_LABEL = 'PP.BET'
const IOS_PROFILE_IDENTIFIER = 'com.ppmx.webclip'
const IOS_PROFILE_MIME_TYPE = 'application/x-apple-aspen-config'

export interface IosWebClipProfileOptions {
  iconBase64: string
  label: string
  url: string
}

function escapeXml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;')
}

function profileUuid(lastSegment: string) {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID().toUpperCase()
  }

  return `8F7E6D5C-4B3A-4921-8C7D-${lastSegment}`
}

function normalizeBaseUrl() {
  const base = import.meta.env.BASE_URL || '/'
  return new URL(base, window.location.origin).toString()
}

function assetUrl(path: string) {
  return new URL(path, normalizeBaseUrl()).toString()
}

function arrayBufferToBase64(buffer: ArrayBuffer) {
  const bytes = new Uint8Array(buffer)
  let binary = ''

  for (const byte of bytes) {
    binary += String.fromCharCode(byte)
  }

  return btoa(binary)
}

async function loadIconBase64() {
  const response = await fetch(assetUrl(IOS_PROFILE_ICON_PATH))

  if (!response.ok) {
    throw new Error('Unable to load iOS Web Clip icon')
  }

  return arrayBufferToBase64(await response.arrayBuffer())
}

function downloadTextFile(fileName: string, content: string, type: string) {
  const blob = new Blob([content], { type })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')

  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

export function buildIosWebClipProfile(options: IosWebClipProfileOptions) {
  const label = escapeXml(options.label)
  const url = escapeXml(options.url)
  const iconBase64 = options.iconBase64.trim()
  const profileUuidValue = profileUuid('000000000001')
  const webClipUuidValue = profileUuid('000000000002')

  return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>PayloadContent</key>
  <array>
    <dict>
      <key>FullScreen</key>
      <true/>
      <key>Icon</key>
      <data>${iconBase64}</data>
      <key>IsRemovable</key>
      <true/>
      <key>Label</key>
      <string>${label}</string>
      <key>PayloadDescription</key>
      <string>Install the PP.BET Web Clip on the Home Screen.</string>
      <key>PayloadDisplayName</key>
      <string>${label}</string>
      <key>PayloadIdentifier</key>
      <string>${IOS_PROFILE_IDENTIFIER}.webclip</string>
      <key>PayloadType</key>
      <string>com.apple.webClip.managed</string>
      <key>PayloadUUID</key>
      <string>${webClipUuidValue}</string>
      <key>PayloadVersion</key>
      <integer>1</integer>
      <key>Precomposed</key>
      <true/>
      <key>URL</key>
      <string>${url}</string>
    </dict>
  </array>
  <key>PayloadDescription</key>
  <string>Add PP.BET to the iOS Home Screen.</string>
  <key>PayloadDisplayName</key>
  <string>${label}</string>
  <key>PayloadIdentifier</key>
  <string>${IOS_PROFILE_IDENTIFIER}.profile</string>
  <key>PayloadOrganization</key>
  <string>PP.BET</string>
  <key>PayloadRemovalDisallowed</key>
  <false/>
  <key>PayloadType</key>
  <string>Configuration</string>
  <key>PayloadUUID</key>
  <string>${profileUuidValue}</string>
  <key>PayloadVersion</key>
  <integer>1</integer>
</dict>
</plist>
`
}

export async function downloadIosWebClipProfile(fileName = IOS_PROFILE_FILE_NAME) {
  if (typeof window === 'undefined' || typeof document === 'undefined') {
    throw new Error('iOS Web Clip profile download requires a browser')
  }

  const profile = buildIosWebClipProfile({
    iconBase64: await loadIconBase64(),
    label: IOS_PROFILE_LABEL,
    url: normalizeBaseUrl(),
  })

  downloadTextFile(fileName, profile, IOS_PROFILE_MIME_TYPE)
}
