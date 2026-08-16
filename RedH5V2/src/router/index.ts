import { createRouter, createWebHashHistory } from 'vue-router'
import { watch } from 'vue'
import { handleHotUpdate, routes } from 'vue-router/auto-routes'
import { i18n } from '@/plugins/i18n'
import { APP_BRAND } from '@/config/brand'
import { formatDocumentTitle, resolveRouteTitle } from '@/router/documentTitle'
import { shouldRequireRouteLogin } from '@/utils/authGate'
import { trackAttributionPageView } from '@/utils/attribution'
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

function updateDocumentTitle() {
  const routeTitle = resolveRouteTitle(
    router.currentRoute.value,
    key => i18n.global.t(key),
    key => i18n.global.te(key),
    APP_BRAND.name,
  )
  document.title = formatDocumentTitle(routeTitle, APP_BRAND.name)
}

router.afterEach(to => {
  const routeTitle = resolveRouteTitle(
    to,
    key => i18n.global.t(key),
    key => i18n.global.te(key),
    APP_BRAND.name,
  )
  document.title = formatDocumentTitle(routeTitle, APP_BRAND.name)
  void trackPageView(to.fullPath)
  void trackAttributionPageView(to.fullPath)
})

watch(i18n.global.locale, updateDocumentTitle)

if (import.meta.hot) {
  handleHotUpdate(router)
}
