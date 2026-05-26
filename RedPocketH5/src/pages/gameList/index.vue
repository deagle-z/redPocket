<script setup lang="ts">
import { showToast } from 'vant'
import type { AppHomeGameItem } from '@/api/user'
import { getAppGameCategoryList } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { getGameProviderIcon, getGameProviderLabel, normalizeGameProviderCode } from '@/constants/gameProviders'
import { safeBack } from '@/utils/navigation'

const PAGE_SIZE = 20
const SEARCH_DELAY = 300

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const games = ref<AppHomeGameItem[]>([])
const providers = ref<string[]>([])
const selectedProvider = ref('')
const searchText = ref('')
const currentPage = ref(0)
const total = ref(0)
const searchTimer = ref<ReturnType<typeof setTimeout> | null>(null)
let requestSeq = 0

const categoryCode = computed(() => String(route.query.categoryCode || 'slots'))
const categoryLabel = computed(() => {
  const text = categoryCode.value.replace(/[_-]+/g, ' ')
  return text.replace(/\b\w/g, char => char.toUpperCase())
})
const pageCount = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))
const hasGames = computed(() => games.value.length > 0)
const headerTitle = computed(() => {
  const provider = selectedProvider.value ? getGameProviderLabel(selectedProvider.value) : ''
  return provider ? `${provider}-${categoryLabel.value}` : categoryLabel.value
})
const visiblePages = computed(() => {
  const count = pageCount.value
  if (count <= 3)
    return Array.from({ length: count }, (_, index) => index)
  const start = Math.min(Math.max(currentPage.value - 1, 0), count - 3)
  return [start, start + 1, start + 2]
})

function buildRequest(page = currentPage.value) {
  const payload: {
    categoryCode: string
    gameName: string
    currentPage: number
    pageSize: number
    thirdGameCategory?: string
  } = {
    categoryCode: categoryCode.value,
    gameName: searchText.value.trim(),
    currentPage: page,
    pageSize: PAGE_SIZE,
  }
  if (selectedProvider.value)
    payload.thirdGameCategory = selectedProvider.value
  return payload
}

async function loadGames(page = currentPage.value) {
  const seq = ++requestSeq
  try {
    loading.value = true
    const { data } = await getAppGameCategoryList(buildRequest(page))
    if (seq !== requestSeq)
      return

    const nextProviders = (data?.thirdGameCategories || []).map(normalizeGameProviderCode).filter(Boolean)
    if (nextProviders.length > 0)
      providers.value = nextProviders

    selectedProvider.value = normalizeGameProviderCode(data?.thirdGameCategory || selectedProvider.value || nextProviders[0])
    games.value = data?.list || []
    total.value = Number(data?.total || 0)
    currentPage.value = Number(data?.currentPage ?? page)
  }
  catch {
    if (seq === requestSeq)
      games.value = []
  }
  finally {
    if (seq === requestSeq)
      loading.value = false
  }
}

function resetAndLoad() {
  currentPage.value = 0
  void loadGames(0)
}

function handleProviderClick(provider: string) {
  const code = normalizeGameProviderCode(provider)
  if (!code || code === selectedProvider.value)
    return
  selectedProvider.value = code
  resetAndLoad()
}

function handleSearchInput() {
  if (searchTimer.value)
    clearTimeout(searchTimer.value)
  searchTimer.value = setTimeout(resetAndLoad, SEARCH_DELAY)
}

function goPage(page: number) {
  if (page < 0 || page >= pageCount.value || page === currentPage.value)
    return
  currentPage.value = page
  void loadGames(page)
}

function playGame(game: AppHomeGameItem) {
  if (!game.gameId) {
    showToast('Invalid game')
    return
  }
  router.push({
    path: '/gamePlay',
    query: {
      gameId: String(game.gameId),
      title: game.gameName,
    },
  })
}

function goBack() {
  safeBack(router)
}

function gameImage(game: AppHomeGameItem) {
  return game.gameIcon || game.horizontalImage || ''
}

watch(categoryCode, () => {
  selectedProvider.value = ''
  searchText.value = ''
  currentPage.value = 0
  total.value = 0
  games.value = []
  providers.value = []
  void loadGames(0)
})

onMounted(() => {
  void loadGames(0)
})

onBeforeUnmount(() => {
  if (searchTimer.value)
    clearTimeout(searchTimer.value)
})
</script>

<template>
  <div class="game-list-page">
    <AppPageHeader :title="headerTitle" @back="goBack" />

    <div class="game-search">
      <input
        v-model="searchText"
        type="search"
        placeholder="Introduce el nombre del juego para buscar"
        @input="handleSearchInput"
      >
      <van-icon name="search" />
    </div>

    <main class="game-list-layout">
      <aside class="provider-tabs">
        <template v-if="loading && providers.length === 0">
          <div v-for="idx in 8" :key="`provider-skeleton-${idx}`" class="provider-tab provider-tab--skeleton">
            <van-skeleton-image />
            <van-skeleton title :row="0" />
          </div>
        </template>

        <button
          v-for="provider in providers"
          v-else
          :key="provider"
          type="button"
          class="provider-tab"
          :class="{ 'provider-tab--active': provider === selectedProvider }"
          @click="handleProviderClick(provider)"
        >
          <img v-if="getGameProviderIcon(provider)" :src="getGameProviderIcon(provider)" :alt="getGameProviderLabel(provider)">
          <span v-else class="provider-tab__fallback">{{ getGameProviderLabel(provider).slice(0, 2) }}</span>
          <p>{{ getGameProviderLabel(provider) }}</p>
        </button>
      </aside>

      <section class="game-list-content">
        <div v-if="loading && !hasGames" class="game-card-grid">
          <div v-for="idx in 12" :key="`game-skeleton-${idx}`" class="game-card game-card--skeleton">
            <van-skeleton-image />
          </div>
        </div>

        <AppEmpty v-else-if="!hasGames" text="暂无游戏" :min-height="220" />

        <div v-else class="game-card-grid">
          <button
            v-for="game in games"
            :key="game.gameId"
            type="button"
            class="game-card"
            @click="playGame(game)"
          >
            <img v-if="gameImage(game)" :src="gameImage(game)" :alt="game.gameName">
            <span v-else class="game-card__fallback">{{ getGameProviderLabel(game.manufacturer).slice(0, 2) }}</span>
            <strong>{{ game.gameName }}</strong>
          </button>
        </div>
      </section>
    </main>

    <nav v-if="pageCount > 1" class="floating-pager" aria-label="pagination">
      <button type="button" :disabled="currentPage <= 0" @click="goPage(currentPage - 1)">
        <van-icon name="arrow-left" />
      </button>
      <button
        v-for="page in visiblePages"
        :key="page"
        type="button"
        :class="{ 'floating-pager__page--active': page === currentPage }"
        @click="goPage(page)"
      >
        {{ page + 1 }}
      </button>
      <span v-if="visiblePages[visiblePages.length - 1] < pageCount - 1">...</span>
      <button type="button" :disabled="currentPage >= pageCount - 1" @click="goPage(currentPage + 1)">
        <van-icon name="arrow" />
      </button>
    </nav>
  </div>
</template>

<style scoped>
.game-list-page {
  height: 100vh;
  height: 100dvh;
  padding: 0 12px 0;
  box-sizing: border-box;
  overflow: hidden;
  color: #f9e8c6;
  display: flex;
  flex-direction: column;
  background:
    radial-gradient(circle at 16% 18%, rgba(212, 175, 55, 0.16), transparent 28%),
    radial-gradient(circle at 82% 90%, rgba(255, 120, 0, 0.12), transparent 28%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.035) 18px,
      rgba(212, 175, 55, 0.035) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #230000 58%, #120000 100%);
}

.game-list-page::before {
  content: '';
  position: fixed;
  inset: 44px 0 0;
  pointer-events: none;
  background:
    radial-gradient(circle at 12% 42%, rgba(255, 217, 139, 0.08), transparent 20%),
    linear-gradient(180deg, rgba(0, 0, 0, 0), rgba(0, 0, 0, 0.18));
}

.game-search {
  position: relative;
  flex: 0 0 46px;
  z-index: 1;
  height: 46px;
  margin: 10px 2px 14px;
  border: 1px solid rgba(255, 218, 127, 0.18);
  border-radius: 999px;
  background: rgba(77, 24, 23, 0.76);
  box-shadow:
    inset 0 1px 0 rgba(255, 245, 210, 0.08),
    0 12px 24px rgba(20, 0, 0, 0.28);
}

.game-search input {
  width: 100%;
  height: 100%;
  padding: 0 44px 0 18px;
  color: #fff3de;
  font-size: 15px;
}

.game-search input::placeholder {
  color: rgba(255, 232, 193, 0.6);
}

.game-search :deep(.van-icon) {
  position: absolute;
  top: 50%;
  right: 16px;
  color: rgba(255, 236, 191, 0.86);
  font-size: 22px;
  transform: translateY(-50%);
}

.game-list-layout {
  position: relative;
  flex: 1 1 auto;
  z-index: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 94px minmax(0, 1fr);
  gap: 12px;
  overflow: hidden;
}

.provider-tabs {
  min-height: 0;
  height: 100%;
  max-height: none;
  overflow: auto;
  overscroll-behavior: contain;
  scrollbar-width: none;
}

.provider-tabs::-webkit-scrollbar {
  display: none;
}

.provider-tab {
  width: 100%;
  min-height: 78px;
  margin-bottom: 10px;
  padding: 8px 6px;
  border: 1px solid rgba(255, 222, 154, 0.12);
  border-radius: 8px;
  background: rgba(70, 24, 32, 0.84);
  color: rgba(255, 232, 193, 0.72);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 5px;
}

.provider-tab--active {
  border-color: rgba(255, 245, 195, 0.72);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.32), transparent 44%),
    linear-gradient(135deg, #ffe78a 0%, #ffb300 52%, #d57200 100%);
  color: #3a0800;
  box-shadow:
    0 10px 20px rgba(128, 45, 0, 0.34),
    inset 0 1px 0 rgba(255, 255, 255, 0.62);
}

.provider-tab img,
.provider-tab__fallback {
  width: 52px;
  height: 36px;
  object-fit: contain;
}

.provider-tab__fallback {
  border-radius: 10px;
  background: rgba(31, 0, 0, 0.48);
  color: #ffd98b;
  display: grid;
  place-items: center;
  font-size: 16px;
  font-weight: 900;
}

.provider-tab p {
  width: 100%;
  margin: 0;
  overflow: hidden;
  font-size: 12px;
  line-height: 1;
  font-weight: 800;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.provider-tab--skeleton :deep(.van-skeleton-image) {
  width: 48px;
  height: 32px;
  border-radius: 8px;
}

.provider-tab--skeleton :deep(.van-skeleton__title) {
  width: 42px;
  height: 10px;
  margin: 0;
}

.game-list-content {
  min-width: 0;
  min-height: 0;
  height: 100%;
  padding-bottom: 92px;
  box-sizing: border-box;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: none;
}

.game-list-content::-webkit-scrollbar {
  display: none;
}

.game-card-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px 10px;
}

.game-card {
  position: relative;
  width: 100%;
  min-width: 0;
  aspect-ratio: 0.76;
  border-radius: 10px;
  overflow: hidden;
  background: linear-gradient(165deg, #7f1b10 0%, #320000 100%);
  box-shadow:
    0 12px 20px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 248, 214, 0.18);
}

.game-card::after {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(circle at 82% 14%, rgba(255, 255, 255, 0.22), transparent 16%),
    linear-gradient(180deg, rgba(0, 0, 0, 0) 50%, rgba(46, 0, 0, 0.72) 100%);
}

.game-card img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.game-card strong {
  position: absolute;
  left: 7px;
  right: 7px;
  bottom: 8px;
  z-index: 1;
  overflow: hidden;
  color: #fff8ea;
  font-size: 13px;
  line-height: 1.12;
  font-weight: 900;
  text-align: left;
  text-shadow: 0 2px 5px rgba(74, 0, 0, 0.74);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.game-card__fallback {
  width: 100%;
  height: 100%;
  color: #ffd98b;
  display: grid;
  place-items: center;
  font-size: 24px;
  font-weight: 900;
}

.game-card--skeleton {
  display: flex;
}

.game-card--skeleton :deep(.van-skeleton-image) {
  width: 100%;
  height: 100%;
  border-radius: 10px;
}

.floating-pager {
  position: fixed;
  left: 50%;
  bottom: 16px;
  z-index: 10;
  transform: translateX(-50%);
  min-height: 48px;
  padding: 5px;
  border: 1px solid rgba(255, 218, 127, 0.22);
  border-radius: 14px;
  background: rgba(31, 0, 0, 0.82);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  gap: 5px;
  box-shadow: 0 12px 26px rgba(0, 0, 0, 0.34);
}

.floating-pager button,
.floating-pager span {
  min-width: 38px;
  height: 38px;
  border-radius: 10px;
  color: #ffe8c1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 17px;
  font-weight: 800;
}

.floating-pager button {
  border: 1px solid rgba(255, 218, 127, 0.2);
  background: rgba(57, 0, 0, 0.52);
}

.floating-pager button:disabled {
  opacity: 0.42;
}

.floating-pager__page--active {
  border-color: rgba(255, 245, 195, 0.78) !important;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.32), transparent 44%),
    linear-gradient(135deg, #ffe78a 0%, #ffb300 52%, #d57200 100%) !important;
  color: #3a0800 !important;
}

@media (max-width: 360px) {
  .game-list-page {
    padding-right: 10px;
    padding-left: 10px;
  }

  .game-list-layout {
    grid-template-columns: 82px minmax(0, 1fr);
    gap: 9px;
  }

  .provider-tabs {
    width: 82px;
  }

  .game-list-content {
    min-width: 0;
  }

  .provider-tab {
    min-height: 72px;
  }

  .provider-tab img,
  .provider-tab__fallback {
    width: 46px;
  }

  .game-card-grid {
    gap: 9px 7px;
  }

  .game-card strong {
    font-size: 11px;
  }
}
</style>

<route lang="json5">
{
  name: 'GameList'
}
</route>
