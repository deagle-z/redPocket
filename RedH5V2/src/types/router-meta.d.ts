import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    titleKey?: string
    requiresAuth?: boolean
    keepAlive?: boolean
    tabbar?: boolean
    shell?: 'default' | 'ppmx' | 'none'
  }
}
