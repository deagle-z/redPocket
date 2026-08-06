import { describe, expect, it } from 'vitest'
import {
  isPublicAuthRoute,
  isPublicInfoRoute,
  isPublicTabbarRoute,
  shouldRequireRouteLogin,
} from './authGate'

describe('PP.PE auth gate', () => {
  it('keeps tabbar routes public', () => {
    for (const path of ['/', '/wheel', '/team', '/promo', '/profile']) {
      expect(isPublicTabbarRoute(path)).toBe(true)
      expect(shouldRequireRouteLogin({ path, meta: {} })).toBe(false)
    }
  })

  it('keeps login and register routes public when they exist', () => {
    expect(isPublicAuthRoute('/login')).toBe(true)
    expect(isPublicAuthRoute('/register')).toBe(true)
    expect(shouldRequireRouteLogin({ path: '/login', meta: {} })).toBe(false)
    expect(shouldRequireRouteLogin({ path: '/register', meta: {} })).toBe(false)
  })

  it('keeps public compliance information routes accessible before login', () => {
    expect(isPublicInfoRoute('/responsible')).toBe(true)
    expect(shouldRequireRouteLogin({ path: '/responsible', meta: {} })).toBe(false)
  })

  it('requires login for every non-public route by default', () => {
    expect(shouldRequireRouteLogin({ path: '/play', meta: {} })).toBe(true)
    expect(shouldRequireRouteLogin({ path: '/wallet/records', meta: {} })).toBe(true)
  })

  it('still honors explicit requiresAuth metadata on public routes', () => {
    expect(shouldRequireRouteLogin({ path: '/profile', meta: { requiresAuth: true } })).toBe(true)
  })

  it('treats an explicit requiresAuth:false route as public (e.g. forgot password)', () => {
    expect(shouldRequireRouteLogin({ path: '/sec-password', meta: { requiresAuth: false } })).toBe(false)
  })
})
