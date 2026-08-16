<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { launchGscSportGame } from '@/api/game'

definePage({
  name: 'gsc-sport-test',
  meta: {
    titleKey: 'gscSportTest.title',
    shell: 'none',
    tabbar: false,
    requiresAuth: true,
  },
})

const IFRAME_LOAD_TIMEOUT = 15000

const loading = ref(false)
const frameLoading = ref(false)
const launchError = ref('')
const gameUrl = ref('')
const gameContent = ref('')
let frameTimer: ReturnType<typeof setTimeout> | undefined

const hasGame = computed(() => Boolean(gameUrl.value || gameContent.value))

function clearFrameTimer() {
  if (!frameTimer) return
  clearTimeout(frameTimer)
  frameTimer = undefined
}

function startFrameTimer() {
  clearFrameTimer()
  if (!gameUrl.value) return

  frameLoading.value = true
  frameTimer = setTimeout(() => {
    frameLoading.value = false
    launchError.value = 'GSC game iframe load timeout'
  }, IFRAME_LOAD_TIMEOUT)
}

function handleFrameLoad() {
  clearFrameTimer()
  frameLoading.value = false
}

async function launchGscSport() {
  if (loading.value) return

  clearFrameTimer()
  loading.value = true
  frameLoading.value = false
  launchError.value = ''
  gameUrl.value = ''
  gameContent.value = ''

  try {
    const data = await launchGscSportGame()
    const url = String(data?.url || '').trim()
    const content = String(data?.content || '').trim()
    if (!url && !content) {
      launchError.value = 'GSC launch response empty'
      return
    }

    gameUrl.value = url
    gameContent.value = content
    if (url) startFrameTimer()
  } catch (error) {
    launchError.value = error instanceof Error ? error.message : 'GSC launch failed'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void launchGscSport()
})

onBeforeUnmount(() => {
  clearFrameTimer()
})
</script>

<template>
  <main class="gsc-sport-test-page">
    <iframe
      v-if="hasGame"
      class="gsc-sport-test-frame"
      :src="gameUrl || undefined"
      :srcdoc="gameUrl ? undefined : gameContent"
      title="GSC Sport"
      allow="autoplay; fullscreen; clipboard-read; clipboard-write"
      referrerpolicy="no-referrer-when-downgrade"
      @load="handleFrameLoad"
    />

    <div v-if="loading || frameLoading || launchError" class="gsc-sport-test-overlay" :class="{ 'is-error': launchError }">
      <i v-if="!launchError" class="fa-solid fa-spinner fa-spin" />
      <i v-else class="fa-solid fa-triangle-exclamation" />
      <span>{{ launchError || 'Loading GSC Sport...' }}</span>
    </div>
  </main>
</template>

<style scoped>
.gsc-sport-test-page {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: #050507;
}

.gsc-sport-test-frame {
  display: block;
  width: 100vw;
  height: 100vh;
  border: 0;
  background: #050507;
}

.gsc-sport-test-overlay {
  position: fixed;
  inset: 0;
  z-index: 2;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 12px;
  padding: 24px;
  color: #fff;
  background: #050507;
  text-align: center;
  font-size: 14px;
  font-weight: 800;
}

.gsc-sport-test-overlay i {
  color: #f5c950;
  font-size: 28px;
}

.gsc-sport-test-overlay.is-error i,
.gsc-sport-test-overlay.is-error span {
  color: #ff5b68;
}
</style>
