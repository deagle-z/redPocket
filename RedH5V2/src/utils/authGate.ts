import type { RouteMeta } from 'vue-router'

export const PUBLIC_TABBAR_ROUTES = ['/', '/wheel', '/team', '/promo', '/profile'] as const
export const PUBLIC_AUTH_ROUTES = ['/login', '/register'] as const
export const PUBLIC_INFO_ROUTES = ['/responsible'] as const

type RouteLike = {
  path: string
  meta: Pick<RouteMeta, 'requiresAuth'>
}

function normalizeRoutePath(path: string) {
  const cleanPath = path.split('?')[0]?.replace(/\/+$/, '') || '/'

  return cleanPath === '' ? '/' : cleanPath
}

export function isPublicTabbarRoute(path: string) {
  return PUBLIC_TABBAR_ROUTES.includes(normalizeRoutePath(path) as (typeof PUBLIC_TABBAR_ROUTES)[number])
}

export function isPublicAuthRoute(path: string) {
  return PUBLIC_AUTH_ROUTES.includes(normalizeRoutePath(path) as (typeof PUBLIC_AUTH_ROUTES)[number])
}

export function isPublicInfoRoute(path: string) {
  return PUBLIC_INFO_ROUTES.includes(normalizeRoutePath(path) as (typeof PUBLIC_INFO_ROUTES)[number])
}

export function shouldRequireRouteLogin(route: RouteLike) {
  // Explicit metadata wins: `requiresAuth: false` marks a route as public
  // (e.g. forgot-password at /sec-password, which must be reachable without login).
  if (route.meta.requiresAuth === true) return true
  if (route.meta.requiresAuth === false) return false
  if (isPublicTabbarRoute(route.path)) return false
  if (isPublicAuthRoute(route.path)) return false
  if (isPublicInfoRoute(route.path)) return false

  return true
}
