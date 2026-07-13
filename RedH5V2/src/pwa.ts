import { showConfirmDialog } from 'vant'
import { ref } from 'vue'
import { registerSW } from 'virtual:pwa-register'
import { trackEvent } from '@/utils/tracker'

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed'; platform: string }>
}

export const canInstallPwa = ref(false)
const installPromptEvent = ref<BeforeInstallPromptEvent | null>(null)

export const updateServiceWorker = registerSW({
  immediate: true,
  onNeedRefresh() {
    void trackEvent({
      eventName: 'pwa_update_available',
      eventType: 'pwa',
    })

    showConfirmDialog({
      title: '发现新版本',
      message: '是否立即刷新并使用最新版本？',
    }).then(() => {
      updateServiceWorker(true)
    })
  },
  onOfflineReady() {
    void trackEvent({
      eventName: 'pwa_offline_ready',
      eventType: 'pwa',
    })
  },
})

window.addEventListener('beforeinstallprompt', event => {
  event.preventDefault()
  installPromptEvent.value = event as BeforeInstallPromptEvent
  canInstallPwa.value = true
})

window.addEventListener('appinstalled', () => {
  canInstallPwa.value = false
  installPromptEvent.value = null
  void trackEvent({
    eventName: 'pwa_installed',
    eventType: 'pwa',
  })
})

export async function promptInstallPwa() {
  const promptEvent = installPromptEvent.value
  if (!promptEvent) return false

  await promptEvent.prompt()
  const choice = await promptEvent.userChoice
  canInstallPwa.value = false
  installPromptEvent.value = null

  void trackEvent({
    eventName: 'pwa_install_prompt',
    eventType: 'pwa',
    payload: {
      outcome: choice.outcome,
      platform: choice.platform,
    },
  })

  return choice.outcome === 'accepted'
}
