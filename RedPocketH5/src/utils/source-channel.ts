const SOURCE_CHANNEL_STORAGE_KEY = 'source_channel_code'

function normalizeSourceChannelCode(value: unknown): string {
  const raw = Array.isArray(value) ? value[0] : value
  return String(raw || '').trim().toUpperCase()
}

function getSourceChannelCode() {
  return normalizeSourceChannelCode(localStorage.getItem(SOURCE_CHANNEL_STORAGE_KEY))
}

function setSourceChannelCode(channelCode: unknown) {
  const normalizedCode = normalizeSourceChannelCode(channelCode)
  if (!normalizedCode)
    return ''

  localStorage.setItem(SOURCE_CHANNEL_STORAGE_KEY, normalizedCode)
  return normalizedCode
}

function clearSourceChannelCode() {
  localStorage.removeItem(SOURCE_CHANNEL_STORAGE_KEY)
}

function captureSourceChannelCode(query: Record<string, unknown>) {
  return setSourceChannelCode(
    query.sc
    || query.sourceChannelCode
    || query.channelCode,
  )
}

export {
  SOURCE_CHANNEL_STORAGE_KEY,
  captureSourceChannelCode,
  clearSourceChannelCode,
  getSourceChannelCode,
  normalizeSourceChannelCode,
  setSourceChannelCode,
}
