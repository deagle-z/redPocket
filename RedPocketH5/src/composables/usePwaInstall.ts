import { useLocalStorage } from '@vueuse/core'
import { STORAGE_TOKEN_KEY } from '@/stores/mutation-type'

const DISMISS_KEY = 'pwa_install_dismissed_at'
const LOGIN_PROMPT_DELAY = 5000

export function usePwaInstall() {
  const deferredPrompt = ref<any>(null)
  const isInstallable = ref(false)
  const showDialog = ref(false)
  const dismissedAt = useLocalStorage<number | null>(DISMISS_KEY, null)
  const accessToken = useLocalStorage<string | null>(STORAGE_TOKEN_KEY, '')
  const isLoginPromptReady = ref(false)
  let loginPromptTimer: ReturnType<typeof setTimeout> | undefined

  function shouldShow() {
    if (!dismissedAt.value)
      return true
    const days = (Date.now() - dismissedAt.value) / (1000 * 60 * 60 * 24)
    return days >= 7
  }

  function tryShowDialog() {
    if (isInstallable.value && isLoginPromptReady.value && shouldShow())
      showDialog.value = true
  }

  function clearLoginPromptTimer() {
    if (!loginPromptTimer)
      return
    clearTimeout(loginPromptTimer)
    loginPromptTimer = undefined
  }

  function scheduleLoginPrompt(token: string | null) {
    clearLoginPromptTimer()
    isLoginPromptReady.value = false

    if (!token) {
      showDialog.value = false
      return
    }

    loginPromptTimer = setTimeout(() => {
      isLoginPromptReady.value = true
      tryShowDialog()
    }, LOGIN_PROMPT_DELAY)
  }

  function onBeforeInstallPrompt(e: Event) {
    e.preventDefault()
    deferredPrompt.value = e
    isInstallable.value = true
    tryShowDialog()
  }

  onMounted(() => {
    window.addEventListener('beforeinstallprompt', onBeforeInstallPrompt)
  })

  onUnmounted(() => {
    window.removeEventListener('beforeinstallprompt', onBeforeInstallPrompt)
    clearLoginPromptTimer()
  })

  watch(accessToken, token => scheduleLoginPrompt(token), { immediate: true })

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
