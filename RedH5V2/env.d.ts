/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string
  readonly VITE_APP_API_BASE_URL?: string
  readonly VITE_API_PROXY_TARGET?: string
  readonly VITE_CDN_BASE_URL?: string
  readonly VITE_APP_BASE: string
  readonly VITE_TRACKING_ENDPOINT?: string
  readonly VITE_RELEASE_VERSION: string
  readonly VITE_FORCE_UPDATE_URL?: string
  readonly VITE_FACEBOOK_PIXEL_ID?: string
  readonly VITE_WS_URL?: string
  readonly VITE_APP_WS_URL?: string
  readonly VITE_APP_WS_UID?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
