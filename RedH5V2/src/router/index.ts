import { createRouter, createWebHashHistory } from 'vue-router'
import { handleHotUpdate, routes } from 'vue-router/auto-routes'
import { i18n } from '@/plugins/i18n'
import { APP_BRAND } from '@/config/brand'
import { shouldRequireRouteLogin } from '@/utils/authGate'
import { trackPageView } from '@/utils/tracker'

export const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach(to => {
  const userStore = useUserStore()

  if (shouldRequireRouteLogin(to) && !userStore.isLogin) {
    return {
      path: '/profile',
      query: {
        auth: 'register',
        redirect: to.fullPath,
      },
    }
  }

  return true
})

router.afterEach(to => {
  const routeTitle = to.meta.titleKey
    ? i18n.global.t(to.meta.titleKey)
    : String(to.meta.title || APP_BRAND.name)
  document.title = routeTitle === APP_BRAND.name ? APP_BRAND.name : `${routeTitle} | ${APP_BRAND.name}`
  void trackPageView(to.fullPath)
})

if (import.meta.hot) {
  handleHotUpdate(router)
}
