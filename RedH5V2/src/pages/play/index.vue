<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { showToast } from 'vant'
import { launchAppGame } from '@/api/game'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '../ppmx-home/types'
import PpmxLogo from '../ppmx-home/components/PpmxLogo.vue'
import PpmxSkeleton from '@/components/PpmxSkeleton.vue'
import { useUserStore } from '@/stores/user'

definePage({
  name: 'play',
  meta: {
    titleKey: 'play.title',
    shell: 'none',
    requiresAuth: true,
  },
})

const IFRAME_LOAD_TIMEOUT = 15000

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { locale } = useI18n()
const { pickText } = usePpmxLocale()

const loading = ref(false)
const frameLoading = ref(false)
const loadError = ref('')
const gameUrl = ref('')
let frameTimer: ReturnType<typeof setTimeout> | undefined
let didRequestExitBalanceRefresh = false

const playText = {
  fallbackTitle: { es: 'Juego PP.BET', en: 'PP.BET game', zh: 'PP.BET 游戏' },
  loadingTitle: { es: 'Preparando juego', en: 'Preparing game', zh: '正在准备游戏' },
  openingTitle: { es: 'Abriendo sala segura', en: 'Opening secure room', zh: '正在打开安全游戏房间' },
  secureSession: { es: 'Sesión segura', en: 'Secure session', zh: '安全会话' },
  invalidGame: { es: 'Juego inválido', en: 'Invalid game', zh: '无效游戏' },
  emptyLink: { es: 'El enlace del juego está vacío', en: 'Game link is empty', zh: '游戏链接为空' },
  launchFailed: {
    es: 'No se pudo abrir el juego. Inténtalo nuevamente.',
    en: 'Unable to open the game. Please try again.',
    zh: '游戏打开失败，请重试。',
  },
  timeout: {
    es: 'El juego tardó demasiado en cargar. Recarga la sala.',
    en: 'The game took too long to load. Reload the room.',
    zh: '游戏加载超时，请重新加载。',
  },
  back: { es: 'Volver', en: 'Back', zh: '返回' },
  retry: { es: 'Reintentar', en: 'Retry', zh: '重试' },
  reload: { es: 'Recargar', en: 'Reload', zh: '刷新' },
  openExternal: { es: 'Abrir fuera', en: 'Open outside', zh: '外部打开' },
  externalUnavailable: {
    es: 'Primero abre el juego',
    en: 'Open the game first',
    zh: '请先打开游戏',
  },
} satisfies Record<string, LocalizedText>

const gameId = computed(() => Number(route.query.gameId || 0))
const gameTitle = computed(() => {
  const title = String(route.query.title || '').trim()

  return title || pickText(playText.fallbackTitle)
})
const launchLanguage = computed(() => String(locale.value || 'es-US').split('-')[0] || 'es')

function clearFrameTimer() {
  if (!frameTimer) return
  clearTimeout(frameTimer)
  frameTimer = undefined
}

function startFrameTimer() {
  clearFrameTimer()
  frameLoading.value = true
  frameTimer = setTimeout(() => {
    frameLoading.value = false
    loadError.value = pickText(playText.timeout)
  }, IFRAME_LOAD_TIMEOUT)
}

function refreshUserInfoAfterGameExit() {
  if (didRequestExitBalanceRefresh || !userStore.isLogin) return

  didRequestExitBalanceRefresh = true
  void userStore.loadUserInfo().catch((error) => {
    console.warn('[play] Failed to refresh user balance after game exit.', error)
  })
}

function goBack() {
  refreshUserInfoAfterGameExit()

  if (window.history.length > 1) {
    router.back()
    return
  }

  void router.push('/')
}

async function loadGame() {
  if (!Number.isFinite(gameId.value) || gameId.value <= 0) {
    const message = pickText(playText.invalidGame)
    showToast(message)
    loadError.value = message
    return
  }

  try {
    loading.value = true
    frameLoading.value = false
    loadError.value = ''
    gameUrl.value = ''
    clearFrameTimer()

    const data = await launchAppGame({
      gameId: gameId.value,
      language: launchLanguage.value,
    })

    if (!data?.url) {
      const message = pickText(playText.emptyLink)
      showToast(message)
      loadError.value = message
      return
    }

    gameUrl.value = data.url
    startFrameTimer()
  } catch (error) {
    gameUrl.value = ''
    loadError.value = error instanceof Error ? error.message : pickText(playText.launchFailed)
  } finally {
    loading.value = false
  }
}

function handleFrameLoad() {
  clearFrameTimer()
  frameLoading.value = false
  loadError.value = ''
}

function reloadGame() {
  void loadGame()
}

function openExternalGame() {
  if (!gameUrl.value) {
    showToast(pickText(playText.externalUnavailable))
    return
  }

  window.open(gameUrl.value, '_blank', 'noopener,noreferrer')
}

onMounted(() => {
  void loadGame()
})

watch(
  () => route.query.gameId,
  () => {
    void loadGame()
  },
)

onBeforeUnmount(() => {
  clearFrameTimer()
  refreshUserInfoAfterGameExit()
})
</script>

<template>
  <main class="ppmx-play-page">
    <header class="ppmx-play-topbar">
      <button class="ppmx-play-icon-button" type="button" :aria-label="pickText(playText.back)" @click="goBack">
        <i class="fa-solid fa-chevron-left" />
      </button>

      <PpmxLogo />

      <div class="ppmx-play-title">
        <strong>{{ gameTitle }}</strong>
        <span>
          <i class="fa-solid fa-shield-halved" />
          {{ pickText(playText.secureSession) }}
        </span>
      </div>

      <div class="ppmx-play-actions">
        <button class="ppmx-play-action" type="button" @click="reloadGame">
          <i class="fa-solid fa-rotate-right" />
          <span>{{ pickText(playText.reload) }}</span>
        </button>
        <button class="ppmx-play-action" type="button" @click="openExternalGame">
          <i class="fa-solid fa-up-right-from-square" />
          <span>{{ pickText(playText.openExternal) }}</span>
        </button>
      </div>
    </header>

    <section class="ppmx-play-stage">
      <PpmxSkeleton
        v-if="loading"
        class="ppmx-play-loader"
        variant="play-launch"
        :aria-label="pickText(playText.loadingTitle)"
      />

      <div v-else-if="gameUrl && !loadError" class="ppmx-play-frame-wrap">
        <PpmxSkeleton
          v-if="frameLoading"
          class="ppmx-play-frame-loading"
          variant="play-frame"
          compact
          :aria-label="pickText(playText.openingTitle)"
        />
        <iframe
          class="ppmx-play-frame"
          :src="gameUrl"
          :title="gameTitle"
          allow="autoplay; fullscreen; clipboard-read; clipboard-write"
          referrerpolicy="no-referrer-when-downgrade"
          @load="handleFrameLoad"
        />
      </div>

      <div v-else class="ppmx-play-error">
        <span class="ppmx-play-error__icon">
          <i class="fa-solid fa-triangle-exclamation" />
        </span>
        <h1>{{ pickText(playText.launchFailed) }}</h1>
        <p>{{ loadError || pickText(playText.launchFailed) }}</p>
        <div>
          <button class="ppmx-btn-red" type="button" @click="reloadGame">
            <i class="fa-solid fa-rotate-right" />
            {{ pickText(playText.retry) }}
          </button>
          <button class="ppmx-btn-ghost" type="button" @click="goBack">
            <i class="fa-solid fa-chevron-left" />
            {{ pickText(playText.back) }}
          </button>
        </div>
      </div>
    </section>
  </main>
</template>
