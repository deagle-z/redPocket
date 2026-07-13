export const AUTH_EXPIRED_EVENT = 'ppmx-auth-expired'

export function notifyAuthExpired() {
  if (typeof window === 'undefined') return

  window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
}
