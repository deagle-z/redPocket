<script setup lang="ts">
import { showToast } from 'vant'
import { launchAppGame } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import AppRetryState from '@/components/AppRetryState.vue'
import { safeBack } from '@/utils/navigation'

const IFRAME_LOAD_TIMEOUT = 15000

const route = useRoute()
const router = useRouter()
const { locale } = useI18n()

const loading = ref(false)
const frameLoading = ref(false)
const loadError = ref('')
const gameUrl = ref('')
const gameTitle = computed(() => String(route.query.title || 'Game'))
const gameId = computed(() => Number(route.query.gameId || 0))
const launchLanguage = computed(() => String(locale.value || 'en-US').split('-')[0] || 'en')
let frameTimer: ReturnType<typeof setTimeout> | undefined

function goBack() {
  safeBack(router)
}

function clearFrameTimer() {
  if (frameTimer) {
    clearTimeout(frameTimer)
    frameTimer = undefined
  }
}

function startFrameTimer() {
  clearFrameTimer()
  frameLoading.value = true
  frameTimer = setTimeout(() => {
    frameLoading.value = false
    loadError.value = 'Game loading timed out. Please reload.'
  }, IFRAME_LOAD_TIMEOUT)
}

async function loadGame() {
  if (!Number.isFinite(gameId.value) || gameId.value <= 0) {
    showToast('Invalid game')
    loadError.value = 'Invalid game'
    return
  }

  try {
    loading.value = true
    frameLoading.value = false
    loadError.value = ''
    gameUrl.value = ''
    clearFrameTimer()
    const { data } = await launchAppGame({
      gameId: gameId.value,
      language: launchLanguage.value,
    })
    if (!data?.url) {
      showToast('Game link is empty')
      loadError.value = 'Game link is empty'
      return
    }
    gameUrl.value = data.url
    startFrameTimer()
  }
  catch {
    gameUrl.value = ''
    loadError.value = 'Unable to launch game. Please try again.'
  }
  finally {
    loading.value = false
  }
}

function handleFrameLoad() {
  clearFrameTimer()
  frameLoading.value = false
  loadError.value = ''
}

onMounted(() => {
  void loadGame()
})

watch(gameId, () => {
  void loadGame()
})

onBeforeUnmount(() => {
  clearFrameTimer()
})
</script>

<template>
  <div class="game-play-page">
    <AppPageHeader :title="gameTitle" @back="goBack" />

    <main class="game-play-body">
      <div v-if="loading" class="game-play-loading">
        <van-loading color="#ffd98b" size="28px" vertical>
          Loading
        </van-loading>
      </div>

      <div v-else-if="gameUrl && !loadError" class="game-play-frame-wrap">
        <div v-if="frameLoading" class="game-play-frame-loading">
          <van-loading color="#ffd98b" size="24px" vertical>
            Opening game
          </van-loading>
        </div>
        <iframe
          class="game-play-frame"
          :src="gameUrl"
          title="game"
          allow="autoplay; fullscreen; clipboard-read; clipboard-write"
          referrerpolicy="no-referrer-when-downgrade"
          @load="handleFrameLoad"
        />
      </div>

      <AppRetryState
        v-else
        class="game-play-retry"
        title="Unable to load game"
        :text="loadError || 'The game could not be opened. Please try again.'"
        action-text="Retry"
        :min-height="240"
        @retry="loadGame"
      />
    </main>
  </div>
</template>

<style scoped>
.game-play-page {
  min-height: 100vh;
  height: 100vh;
  height: 100dvh;
  display: flex;
  flex-direction: column;
  background:
    radial-gradient(circle at 20% 10%, rgba(212, 175, 55, 0.16), transparent 30%),
    linear-gradient(180deg, var(--color-game-bg-start) 0%, var(--color-game-bg-end) 100%);
  color: var(--color-game-text);
  overflow: hidden;
}

.game-play-body {
  min-height: 0;
  flex: 1;
  display: flex;
}

.game-play-loading,
.game-play-retry {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 12px;
}

.game-play-frame-wrap {
  position: relative;
  flex: 1;
  min-width: 0;
  min-height: 0;
}

.game-play-frame-loading {
  position: absolute;
  inset: 0;
  z-index: 2;
  background: radial-gradient(circle at 50% 30%, rgba(212, 175, 55, 0.12), transparent 30%), rgba(18, 0, 0, 0.86);
  color: var(--color-game-text);
  display: flex;
  align-items: center;
  justify-content: center;
}

.game-play-frame {
  width: 100%;
  height: 100%;
  flex: 1;
  border: 0;
  background: #120000;
}
</style>
