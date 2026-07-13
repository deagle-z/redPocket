<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { showToast } from 'vant'
import type { AppHomeGameItem } from '@/api/game'
import { getAppGameCategoryList } from '@/api/game'
import { getPpmxFallbackGradient } from '../ppmx-home/assets'
import { mapAppHomeGame } from '../ppmx-home/composables/usePpmxGameHome'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { PpmxGame } from '../ppmx-home/types'
import AppState from '@/components/AppState.vue'
import PpmxSkeleton from '@/components/PpmxSkeleton.vue'
import PpmxGameTile from '../ppmx-home/components/PpmxGameTile.vue'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'
import PpmxUnlockModal from '@/components/PpmxUnlockModal.vue'

definePage({
  name: 'casino',
  meta: {
    titleKey: 'ppmx.casino.title',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: true,
  },
})

type CasinoCategoryId = 'slots' | 'casino' | 'blockchain' | 'fishing' | 'rummy' | 'sports' | 'mini' | 'lottery'
type CasinoSortMode = 'popular' | 'rtp' | 'new' | 'az'

interface CasinoCategory {
  id: CasinoCategoryId
  icon: string
  labelKey: string
  categoryCode: string
}

const PAGE_SIZE = 24
const SEARCH_DELAY = 320
const FORCE_CASINO_SKELETON = false

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const { pickText } = usePpmxLocale()
const { unlockOpen, ensureUnlocked } = useGameUnlockGate()

const casinoCategories: CasinoCategory[] = [
  { id: 'slots', icon: 'fa-coins', labelKey: 'ppmx.casino.catSlots', categoryCode: 'slots' },
  { id: 'casino', icon: 'fa-tower-broadcast', labelKey: 'ppmx.casino.catCasino', categoryCode: 'casino' },
  { id: 'blockchain', icon: 'fa-link', labelKey: 'ppmx.casino.catBlockchain', categoryCode: 'blockchain' },
  { id: 'fishing', icon: 'fa-fish-fins', labelKey: 'ppmx.casino.catFishing', categoryCode: 'fishing' },
  { id: 'rummy', icon: 'fa-chess', labelKey: 'ppmx.casino.catRummy', categoryCode: 'rummy' },
  { id: 'sports', icon: 'fa-futbol', labelKey: 'ppmx.casino.catSports', categoryCode: 'sports' },
  { id: 'mini', icon: 'fa-gamepad', labelKey: 'ppmx.casino.catMini', categoryCode: 'mini' },
  { id: 'lottery', icon: 'fa-ticket', labelKey: 'ppmx.casino.catLottery', categoryCode: 'lottery' },
]

const sortOptions: Array<{ value: CasinoSortMode; labelKey: string }> = [
  { value: 'popular', labelKey: 'ppmx.casino.sortPopular' },
  { value: 'rtp', labelKey: 'ppmx.casino.sortRtp' },
  { value: 'new', labelKey: 'ppmx.casino.sortNew' },
  { value: 'az', labelKey: 'ppmx.casino.sortAz' },
]

const pageChromeRef = ref<InstanceType<typeof PpmxPageChrome> | null>(null)
const activeCategoryId = ref<CasinoCategoryId>(resolveInitialCategory())
const selectedProvider = ref('')
const searchKeyword = ref('')
const sortMode = ref<CasinoSortMode>('popular')
const sortMenuOpen = ref(false)
const games = ref<AppHomeGameItem[]>([])
const providers = ref<string[]>([])
const currentPage = ref(0)
const total = ref(0)
const loading = ref(FORCE_CASINO_SKELETON)
const loadingMore = ref(false)
const loadFailed = ref(false)
const searchTimer = ref<number | undefined>()
const casinoControlsRef = ref<HTMLElement | null>(null)
let requestSeq = 0

const activeCategory = computed(() =>
  casinoCategories.find(category => category.id === activeCategoryId.value) || casinoCategories[0],
)
const hasLoadedGames = computed(() => !FORCE_CASINO_SKELETON && games.value.length > 0)
const hasMore = computed(() => games.value.length < total.value)
const visibleGames = computed(() => sortGameItems(games.value).map(mapAppHomeGame))
const activeSortLabel = computed(() =>
  t(sortOptions.find(option => option.value === sortMode.value)?.labelKey || 'ppmx.casino.sortPopular'),
)

function resolveInitialCategory(): CasinoCategoryId {
  const queryCode = String(route.query.categoryCode || route.query.category || '').toLowerCase()
  const matched = casinoCategories.find(category =>
    category.id === queryCode || category.categoryCode === queryCode,
  )

  return matched?.id || 'slots'
}

function buildRequest(page: number) {
  const payload: {
    currentPage: number
    pageSize: number
    categoryCode?: string
    thirdGameCategory?: string
    gameName?: string
  } = {
    currentPage: page,
    pageSize: PAGE_SIZE,
  }

  payload.categoryCode = activeCategory.value.categoryCode
  if (selectedProvider.value) payload.thirdGameCategory = selectedProvider.value

  const keyword = searchKeyword.value.trim()
  if (keyword) payload.gameName = keyword

  return payload
}

function sortGameItems(items: AppHomeGameItem[]) {
  const sorted = [...items]

  if (sortMode.value === 'az') {
    sorted.sort((a, b) => (a.gameName || '').localeCompare(b.gameName || ''))
    return sorted
  }

  if (sortMode.value === 'new') {
    sorted.sort((a, b) => Number(b.showIndex === 2) - Number(a.showIndex === 2) || Number(a.sort ?? 0) - Number(b.sort ?? 0))
    return sorted
  }

  sorted.sort((a, b) => Number(a.sort ?? 0) - Number(b.sort ?? 0) || Number(a.gameId ?? 0) - Number(b.gameId ?? 0))
  return sorted
}

const PROVIDER_PRIORITY = ['PG', 'PP', 'JILI', 'JDB', 'CQ9', 'KA']

function normalizeProviders(values?: string[]) {
  const list = [...new Set((values || []).map(v => v.trim()).filter(Boolean))]
  return list.sort((a, b) => {
    const ai = PROVIDER_PRIORITY.indexOf(a.toUpperCase())
    const bi = PROVIDER_PRIORITY.indexOf(b.toUpperCase())
    if (ai !== -1 && bi !== -1) return ai - bi
    if (ai !== -1) return -1
    if (bi !== -1) return 1
    return 0
  })
}

async function loadGames(page = 0, append = false) {
  if (FORCE_CASINO_SKELETON) {
    loading.value = true
    loadingMore.value = false
    loadFailed.value = false
    games.value = []
    total.value = 0
    return
  }

  const seq = ++requestSeq
  const requestedProvider = selectedProvider.value

  try {
    loading.value = !append
    loadingMore.value = append
    loadFailed.value = false

    const data = await getAppGameCategoryList(buildRequest(page))
    if (seq !== requestSeq) return

    const nextList = data?.list || []
    games.value = append ? [...games.value, ...nextList] : nextList
    total.value = Number(data?.total || 0)
    currentPage.value = Number(data?.currentPage ?? page)

    if (!append) {
      const nextProviders = normalizeProviders(data?.thirdGameCategories)
      providers.value = nextProviders

      const shouldRefetchWithDefaultProvider = nextProviders.length > 0
        && (!requestedProvider || !nextProviders.includes(requestedProvider))

      if (shouldRefetchWithDefaultProvider) {
        selectedProvider.value = nextProviders[0]
        void loadGames(0)
        return
      }
    }
  } catch {
    if (seq !== requestSeq) return

    if (!append) {
      games.value = []
      total.value = 0
      providers.value = []
    }
    loadFailed.value = true
  } finally {
    if (seq === requestSeq) {
      loading.value = false
      loadingMore.value = false
    }
  }
}

function resetAndLoad() {
  currentPage.value = 0
  total.value = 0
  games.value = []
  void loadGames(0)
}

function chooseCategory(category: CasinoCategoryId) {
  if (activeCategoryId.value === category) return

  activeCategoryId.value = category
  selectedProvider.value = ''
  providers.value = []
  void router.replace({
    path: route.path,
    query: {
      ...route.query,
      categoryCode: activeCategory.value.categoryCode,
    },
  })
  resetAndLoad()
}

function chooseProvider(provider: string) {
  if (selectedProvider.value === provider) return

  selectedProvider.value = provider
  resetAndLoad()
}

function handleSearchInput() {
  if (searchTimer.value) window.clearTimeout(searchTimer.value)
  searchTimer.value = window.setTimeout(resetAndLoad, SEARCH_DELAY)
}

function chooseSortMode(value: CasinoSortMode) {
  sortMode.value = value
  sortMenuOpen.value = false
}

function loadMoreGames() {
  if (loading.value || loadingMore.value || !hasMore.value) return

  void loadGames(currentPage.value + 1, true)
}

function onDocumentPointerDown(event: PointerEvent) {
  if (!sortMenuOpen.value) return

  const target = event.target
  if (target instanceof Node && casinoControlsRef.value?.contains(target)) return

  sortMenuOpen.value = false
}

function getGameTileStyle(game: PpmxGame, index = 0) {
  const imageId = game.imageId || index

  if (game.cover) {
    return {
      backgroundImage: `url("${game.cover.replace(/"/g, '%22')}"), ${getPpmxFallbackGradient(imageId)}`,
    }
  }

  return {
    backgroundImage: getPpmxFallbackGradient(imageId),
  }
}

async function openGame(game: PpmxGame) {
  if (!pageChromeRef.value?.requireRegisterAuth()) return

  if (!game.gameId) {
    showToast(t('ppmx.casino.playUnavailable'))
    return
  }

  if (!(await ensureUnlocked())) return

  void router.push({
    path: '/play',
    query: {
      gameId: String(game.gameId),
      title: pickText(game.name),
    },
  })
}

function onUnlockActivate() {
  void router.push('/recharge')
}

watch(sortMode, () => {
  games.value = sortGameItems(games.value)
})

onMounted(() => {
  document.addEventListener('pointerdown', onDocumentPointerDown)
  void loadGames(0)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocumentPointerDown)
  if (searchTimer.value) window.clearTimeout(searchTimer.value)
})
</script>

<template>
  <PpmxPageChrome ref="pageChromeRef">
    <main class="ppmx-subpage ppmx-casino-page">
      <section class="ppmx-subpage__wide ppmx-casino-stack">
        <header class="ppmx-subpage-head ppmx-casino-head">
          <div class="eyebrow">
            {{ t('ppmx.casino.eyebrow') }}
          </div>
          <h1>{{ t('ppmx.casino.title') }}</h1>
          <p>{{ t('ppmx.casino.description') }}</p>
        </header>

        <nav class="ppmx-casino-tabs" :aria-label="t('ppmx.casino.categoryAria')">
          <button
            v-for="category in casinoCategories"
            :key="category.id"
            class="ppmx-casino-tab"
            :class="{ 'is-active': activeCategoryId === category.id }"
            type="button"
            @click="chooseCategory(category.id)"
          >
            <i class="fa-solid" :class="category.icon" />
            <span>{{ t(category.labelKey) }}</span>
          </button>
        </nav>

        <div ref="casinoControlsRef" class="ppmx-casino-controls">
          <label class="ppmx-casino-search">
            <i class="fa-solid fa-magnifying-glass" />
            <input
              v-model="searchKeyword"
              type="search"
              :placeholder="t('ppmx.casino.search')"
              @input="handleSearchInput"
            >
          </label>

          <div class="ppmx-casino-sort" :class="{ 'is-open': sortMenuOpen }">
            <span>{{ t('ppmx.casino.sortLabel') }}</span>
            <button
              class="ppmx-casino-sort__trigger"
              type="button"
              :aria-label="t('ppmx.casino.sortLabel')"
              aria-haspopup="listbox"
              :aria-expanded="sortMenuOpen"
              @click="sortMenuOpen = !sortMenuOpen"
            >
              <strong>{{ activeSortLabel }}</strong>
              <i class="fa-solid fa-chevron-down" />
            </button>
            <div v-if="sortMenuOpen" class="ppmx-casino-sort__menu" role="listbox">
              <button
                v-for="option in sortOptions"
                :key="option.value"
                class="ppmx-casino-sort__option"
                :class="{ 'is-active': sortMode === option.value }"
                type="button"
                role="option"
                :aria-selected="sortMode === option.value"
                @click="chooseSortMode(option.value)"
              >
                <span>{{ t(option.labelKey) }}</span>
                <i v-if="sortMode === option.value" class="fa-solid fa-check" />
              </button>
            </div>
          </div>
        </div>

        <div v-if="providers.length > 0" class="ppmx-casino-providers" :aria-label="t('ppmx.casino.providerAria')">
          <button
            v-for="provider in providers"
            :key="provider"
            class="ppmx-casino-provider"
            :class="{ 'is-active': selectedProvider === provider }"
            type="button"
            @click="chooseProvider(provider)"
          >
            {{ provider.toUpperCase() }}
          </button>
        </div>

        <AppState
          v-if="loading && !hasLoadedGames"
          class="ppmx-casino-state"
          status="loading"
          :title="t('ppmx.casino.loadingTitle')"
          :message="t('ppmx.casino.loadingMessage')"
          skeleton-variant="casino-page"
        />

        <AppState
          v-else-if="loadFailed && !hasLoadedGames"
          class="ppmx-casino-state"
          status="error"
          :title="t('ppmx.casino.errorTitle')"
          :message="t('ppmx.casino.errorMessage')"
          :action-text="t('ppmx.casino.retry')"
          @retry="resetAndLoad"
        />

        <AppState
          v-else-if="visibleGames.length === 0"
          class="ppmx-casino-state"
          status="empty"
          :title="t('ppmx.casino.emptyTitle')"
          :message="t('ppmx.casino.empty')"
          :action-text="t('ppmx.casino.retry')"
          @retry="resetAndLoad"
        />

        <template v-else>
          <div id="casinoGrid" class="ppmx-casino-grid" :class="{ 'is-loading': loading }">
            <PpmxGameTile
              v-for="game, gameIndex in visibleGames"
              :key="game.id"
              @click="openGame(game)"
            >
              <span class="ppmx-htile-thumb" :style="getGameTileStyle(game, gameIndex)">
                <span v-if="game.hot" class="ppmx-htile-badge ppmx-htile-badge--hot">HOT</span>
                <span v-if="game.isNew" class="ppmx-htile-badge ppmx-htile-badge--new">NEW</span>
              </span>
              <span class="ppmx-htile-name">{{ pickText(game.name) }}</span>
              <span v-if="game.subtitle" class="ppmx-htile-provider">{{ game.subtitle }}</span>
            </PpmxGameTile>
          </div>

          <PpmxSkeleton
            v-if="loadingMore"
            variant="casino-more"
            :items="6"
            compact
            :aria-label="t('ppmx.casino.loadingTitle')"
          />

          <div class="ppmx-casino-more">
            <button
              class="ppmx-btn-ghost"
              type="button"
              :disabled="loadingMore || !hasMore"
              @click="loadMoreGames"
            >
              <span>{{ hasMore ? t('ppmx.casino.more') : t('ppmx.casino.noMore') }}</span>
            </button>
          </div>
        </template>
      </section>
    </main>

    <PpmxUnlockModal v-model="unlockOpen" @activate="onUnlockActivate" />
  </PpmxPageChrome>
</template>
