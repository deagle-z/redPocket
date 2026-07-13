import type { RouteMeta } from 'vue-router'

export type AppShellKind = 'default' | 'ppmx' | 'none'

export function resolveAppShell(meta: Pick<RouteMeta, 'shell'>): AppShellKind {
  return meta.shell || 'default'
}
