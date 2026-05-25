<script setup lang="ts">
import { showToast } from 'vant'
import { launchAppGame } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { safeBack } from '@/utils/navigation'

const route = useRoute()
const router = useRouter()
const { locale } = useI18n()

const loading = ref(false)
const gameUrl = ref('')
const gameTitle = computed(() => String(route.query.title || 'Game'))
const gameId = computed(() => Number(route.query.gameId || 0))
const launchLanguage = computed(() => String(locale.value || 'en-US').split('-')[0] || 'en')

function goBack() {
  safeBack(router)
}

async function loadGame() {
  if (!Number.isFinite(gameId.value) || gameId.value <= 0) {
    showToast('Invalid game')
    return
  }

  try {
    loading.value = true
    gameUrl.value = ''
    const { data } = await launchAppGame({
      gameId: gameId.value,
      language: launchLanguage.value,
    })
    if (!data?.url) {
      showToast('Game link is empty')
      return
    }
    gameUrl.value = data.url
  }
  catch {
    gameUrl.value = ''
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadGame()
})

watch(gameId, () => {
  void loadGame()
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

      <iframe
        v-else-if="gameUrl"
        class="game-play-frame"
        :src="gameUrl"
        title="game"
        allow="autoplay; fullscreen; clipboard-read; clipboard-write"
        referrerpolicy="no-referrer-when-downgrade"
      />

      <div v-else class="game-play-empty">
        <p>Unable to load game</p>
        <button type="button" @click="loadGame">
          Retry
        </button>
      </div>
    </main>
  </div>
</template>

<style scoped>
.game-play-page {
  min-height: 100vh;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background:
    radial-gradient(circle at 20% 10%, rgba(212, 175, 55, 0.16), transparent 30%),
    linear-gradient(180deg, #3e0000 0%, #190000 100%);
  color: #f9e8c6;
}

.game-play-body {
  min-height: 0;
  flex: 1;
  display: flex;
}

.game-play-loading,
.game-play-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.game-play-empty {
  flex-direction: column;
  gap: 14px;
  color: rgba(255, 232, 193, 0.78);
}

.game-play-empty p {
  margin: 0;
  font-size: 14px;
}

.game-play-empty button {
  min-width: 96px;
  height: 36px;
  border: 1px solid rgba(255, 237, 172, 0.64);
  border-radius: 10px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.32), transparent 44%),
    linear-gradient(135deg, #ffe78a 0%, #ffb300 52%, #d57200 100%);
  color: #3a0800;
  font-size: 14px;
  font-weight: 800;
}

.game-play-frame {
  width: 100%;
  height: 100%;
  flex: 1;
  border: 0;
  background: #120000;
}
</style>
