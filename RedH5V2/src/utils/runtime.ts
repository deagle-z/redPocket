import { reactive } from 'vue'
import type { AppRuntimeConfig } from '@/types/business'

export const runtimeState = reactive<{
  loading: boolean
  config: AppRuntimeConfig
  error: string
}>({
  loading: false,
  config: {},
  error: '',
})

function compareVersion(current: string, minimum: string) {
  const currentParts = current.split('.').map(Number)
  const minimumParts = minimum.split('.').map(Number)
  const length = Math.max(currentParts.length, minimumParts.length)

  for (let index = 0; index < length; index += 1) {
    const currentPart = currentParts[index] ?? 0
    const minimumPart = minimumParts[index] ?? 0

    if (currentPart > minimumPart) return 1
    if (currentPart < minimumPart) return -1
  }

  return 0
}

export async function getRuntimeConfig() {
  const url = import.meta.env.VITE_FORCE_UPDATE_URL

  if (!url) return runtimeState.config

  runtimeState.loading = true
  runtimeState.error = ''

  try {
    const response = await fetch(url, {
      cache: 'no-store',
      headers: {
        Accept: 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`Runtime config request failed: ${response.status}`)
    }

    runtimeState.config = (await response.json()) as AppRuntimeConfig
  } catch (error) {
    runtimeState.error = error instanceof Error ? error.message : '运行配置加载失败'
    console.warn(runtimeState.error)
  } finally {
    runtimeState.loading = false
  }

  return runtimeState.config
}

export function checkForceUpdate(config = runtimeState.config) {
  if (config.forceUpdate) return true
  if (!config.minVersion) return false

  return compareVersion(import.meta.env.VITE_RELEASE_VERSION, config.minVersion) < 0
}

export async function initRuntimeConfig() {
  const config = await getRuntimeConfig()
  return {
    config,
    forceUpdate: checkForceUpdate(config),
  }
}
