export interface RouteTitleSource {
  name?: string | symbol | null
  query: Record<string, unknown>
  meta: {
    title?: unknown
    titleKey?: unknown
  }
}

type TranslateTitle = (key: string) => string
type HasTranslation = (key: string) => boolean

function firstQueryText(value: unknown): string {
  const candidate = Array.isArray(value) ? value[0] : value
  return typeof candidate === 'string' ? candidate.trim() : ''
}

export function resolveRouteTitle(
  route: RouteTitleSource,
  translate: TranslateTitle,
  hasTranslation: HasTranslation,
  brandName: string,
): string {
  if (String(route.name || '') === 'play') {
    const gameTitle = firstQueryText(route.query.title)
    if (gameTitle) return gameTitle
  }

  const titleKey = typeof route.meta.titleKey === 'string' ? route.meta.titleKey.trim() : ''
  if (titleKey) {
    return hasTranslation(titleKey) ? translate(titleKey) : brandName
  }

  const literalTitle = typeof route.meta.title === 'string' ? route.meta.title.trim() : ''
  return literalTitle || brandName
}

export function formatDocumentTitle(routeTitle: string, brandName: string): string {
  return routeTitle === brandName ? brandName : `${routeTitle} | ${brandName}`
}
