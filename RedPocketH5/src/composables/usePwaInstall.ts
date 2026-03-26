import { useLocalStorage } from '@vueuse/core'

const DISMISS_KEY = 'pwa_install_dismissed_at'

export function usePwaInstall() {
  const deferredPrompt = ref<any>(null)
  const isInstallable = ref(false)
  const showDialog = ref(false)
  const dismissedAt = useLocalStorage<number | null>(DISMISS_KEY, null)

  function shouldShow() {
    if (!dismissedAt.value)
      return true
    const days = (Date.now() - dismissedAt.value) / (1000 * 60 * 60 * 24)
    return days >= 7
  }

  function onBeforeInstallPrompt(e: Event) {
    e.preventDefault()
    deferredPrompt.value = e
    isInstallable.value = true
    if (shouldShow())
      showDialog.value = true
  }

  onMounted(() => {
    window.addEventListener('beforeinstallprompt', onBeforeInstallPrompt)
  })

  onUnmounted(() => {
    window.removeEventListener('beforeinstallprompt', onBeforeInstallPrompt)
  })

  async function triggerInstall() {
    showDialog.value = false
    if (!deferredPrompt.value)
      return
    deferredPrompt.value.prompt()
    await deferredPrompt.value.userChoice
    deferredPrompt.value = null
    isInstallable.value = false
  }

  function dismiss() {
    showDialog.value = false
    dismissedAt.value = Date.now()
  }

  return { isInstallable, showDialog, triggerInstall, dismiss }
}
