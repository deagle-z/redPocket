/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_APP_API_BASE_URL: string
  readonly VITE_APP_PREVIEW: string
  readonly VITE_APP_PUBLIC_PATH?: string
  readonly VITE_TG_BOT_ID?: string
  readonly VITE_WS_URL?: string
  readonly VITE_APP_WS_URL?: string
  readonly VITE_APP_WS_UID?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'

  const component: DefineComponent<Record<string, never>, Record<string, never>, any>
  export default component
}
