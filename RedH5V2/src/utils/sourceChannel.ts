export const INVITE_CODE_KEY = 'invite_code'
export const SOURCE_CHANNEL_CODE_KEY = 'source_channel_code'

type RouteQueryLike = Record<string, unknown>

function getBrowserStorage() {
  if (typeof localStorage === 'undefined') return null
  return localStorage
}

function readQueryValue(query: RouteQueryLike, keys: string[]) {
  for (const key of keys) {
    const value = query[key]
    const firstValue = Array.isArray(value) ? value[0] : value

    if (typeof firstValue === 'string' && firstValue.trim()) {
      return firstValue
    }
  }

  return ''
}

export function normalizeInviteCode(value?: string | null) {
  return value?.trim() ?? ''
}

export function normalizeSourceChannelCode(value?: string | null) {
  return value?.trim().toUpperCase() ?? ''
}

export function setInviteCode(value?: string | null) {
  const code = normalizeInviteCode(value)
  const storage = getBrowserStorage()

  if (code && storage) storage.setItem(INVITE_CODE_KEY, code)

  return code
}

export function getStoredInviteCode() {
  return normalizeInviteCode(getBrowserStorage()?.getItem(INVITE_CODE_KEY))
}

export function captureInviteCode(query: RouteQueryLike) {
  return setInviteCode(readQueryValue(query, ['c', 'inviteCode', 'ref']))
}

export function setSourceChannelCode(value?: string | null) {
  const code = normalizeSourceChannelCode(value)
  const storage = getBrowserStorage()

  if (code && storage) storage.setItem(SOURCE_CHANNEL_CODE_KEY, code)

  return code
}

export function getSourceChannelCode() {
  return normalizeSourceChannelCode(getBrowserStorage()?.getItem(SOURCE_CHANNEL_CODE_KEY))
}

export function captureSourceChannelCode(query: RouteQueryLike) {
  return setSourceChannelCode(
    readQueryValue(query, ['sc', 'sourceChannelCode', 'channelCode']),
  )
}
