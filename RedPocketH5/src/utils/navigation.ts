import type { Router } from 'vue-router'

export function safeBack(router: Router, fallback = '/') {
  if (typeof window !== 'undefined' && window.history.state?.back) {
    window.history.back()
    return
  }

  router.replace(fallback)
}
