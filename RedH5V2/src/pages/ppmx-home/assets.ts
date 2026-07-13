export const PPMX_LOCAL_IMAGE_ROOT = 'images/ppmx'

export const PPMX_FALLBACK_GRADIENTS = [
  'linear-gradient(135deg,#c8102e,#7a0f24)',
  'linear-gradient(135deg,#138a52,#0c5132)',
  'linear-gradient(135deg,#e6a417,#9a6a06)',
  'linear-gradient(135deg,#6d28d9,#3b0f70)',
  'linear-gradient(135deg,#0ea5e9,#075985)',
] as const

export function ppmxAsset(path: string) {
  const base = import.meta.env.BASE_URL || '/'
  const normalizedBase = base.endsWith('/') ? base : `${base}/`
  const normalizedPath = path.replace(/^\/+/, '')

  return `${normalizedBase}${PPMX_LOCAL_IMAGE_ROOT}/${normalizedPath}`
}

export function getPpmxGameImage(imageId: number) {
  return ppmxAsset(`games/${imageId}.jpg`)
}

export function getPpmxFallbackGradient(imageId = 0) {
  return PPMX_FALLBACK_GRADIENTS[Math.abs(imageId) % PPMX_FALLBACK_GRADIENTS.length]
}

export function getPpmxGameTileBackground(imageId: number) {
  return {
    backgroundImage: `url('${getPpmxGameImage(imageId)}'), ${getPpmxFallbackGradient(imageId)}`,
  }
}
