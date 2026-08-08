import { computed, ref } from 'vue'
import type { AppHomeGameData, AppHomeGameItem } from '@/api/game'
import { getAppGameHome } from '@/api/game'
import type { LocalizedText, PpmxGame, PpmxHomeCategory } from '../types'

const text = (value: string): LocalizedText => ({
  es: value,
  en: value,
  zh: value,
})

const categoryMeta: Record<string, Omit<PpmxHomeCategory, 'games'>> = {
  hot: {
    id: 'hot',
    icon: 'fa-fire',
    target: 'casino',
    name: { es: 'Populares', en: 'Popular', zh: '热门' },
  },
  slots: {
    id: 'slots',
    icon: 'fa-dice',
    target: 'casino',
    name: { es: 'Tragamonedas', en: 'Slots', zh: '老虎机' },
  },
  casino: {
    id: 'casino',
    icon: 'fa-tower-broadcast',
    target: 'casino',
    name: { es: 'Casino en vivo', en: 'Live Casino', zh: '真人娱乐场' },
  },
  blockchain: {
    id: 'blockchain',
    icon: 'fa-link',
    target: 'casino',
    name: { es: 'Blockchain', en: 'Blockchain', zh: '区块链' },
  },
  fishing: {
    id: 'fishing',
    icon: 'fa-fish-fins',
    target: 'casino',
    name: { es: 'Pesca', en: 'Fishing', zh: '捕鱼' },
  },
  rummy: {
    id: 'rummy',
    icon: 'fa-chess',
    target: 'casino',
    name: { es: 'Rummy', en: 'Rummy', zh: '纸牌' },
  },
  sports: {
    id: 'sports',
    icon: 'fa-futbol',
    target: 'sports',
    name: { es: 'Deportes', en: 'Sports', zh: '体育' },
  },
  mini: {
    id: 'mini',
    icon: 'fa-gamepad',
    target: 'casino',
    name: { es: 'Mini juegos', en: 'Mini Games', zh: '小游戏' },
  },
  lottery: {
    id: 'lottery',
    icon: 'fa-ticket',
    target: 'casino',
    name: { es: 'Lotería', en: 'Lottery', zh: '彩票' },
  },
}

const preferredCategoryOrder = [
  'hot',
  'slots',
  'casino',
  'blockchain',
  'fishing',
  'rummy',
  'sports',
  'mini',
  'lottery',
]

function normalizeCategoryCode(code?: string) {
  return code?.trim().toLowerCase() || 'other'
}

function getCategoryMeta(code: string): Omit<PpmxHomeCategory, 'games'> {
  const normalized = normalizeCategoryCode(code)
  const meta = categoryMeta[normalized]
  if (meta) return meta

  const label = normalized
    .split(/[-_\s]+/)
    .filter(Boolean)
    .map(part => `${part.charAt(0).toUpperCase()}${part.slice(1)}`)
    .join(' ') || 'Games'

  return {
    id: normalized,
    icon: 'fa-dice',
    target: 'casino',
    name: text(label),
  }
}

function getDynamicCategoryCodes(data: AppHomeGameData) {
  const availableCodes = Object.keys(data).filter(code => (data[code]?.length ?? 0) > 0)
  const preferredCodes = preferredCategoryOrder.filter(code => availableCodes.includes(code))
  const extraCodes = availableCodes
    .filter(code => !preferredCodes.includes(code))
    .sort((a, b) => a.localeCompare(b))

  return [...preferredCodes, ...extraCodes]
}

export function mapAppHomeGame(game: AppHomeGameItem, index = 0): PpmxGame {
  const name = game.gameName || `Game ${game.gameId}`
  const manufacturer = game.manufacturer?.trim()
  const categoryCode = normalizeCategoryCode(game.categoryCode)
  const brand = manufacturer?.toUpperCase() || categoryCode.toUpperCase() || 'GAME'

  return {
    id: `game-${game.gameId || `${categoryCode}-${index}`}`,
    gameId: game.gameId,
    categoryCode,
    fakeOnlineCount: game.fakeOnlineCount,
    cover: game.gameIcon || game.horizontalImage,
    imageId: 11 + (Math.abs(Number(game.gameId || index)) % 90),
    name: text(name),
    brand,
    subtitle: brand,
    hot: categoryCode === 'hot' || game.showIndex === 1,
    isNew: game.showIndex === 2,
  }
}

export function createDynamicHomeCategories(data: AppHomeGameData): PpmxHomeCategory[] {
  return getDynamicCategoryCodes(data).map((code) => {
    const meta = getCategoryMeta(code)

    return {
      ...meta,
      games: (data[code] || []).map(mapAppHomeGame),
    }
  })
}

export function usePpmxGameHome() {
  const loading = ref(false)
  const error = ref<Error | null>(null)
  const categories = ref<PpmxHomeCategory[]>([])
  const loaded = ref(false)

  const categoryIds = computed(() => categories.value.map(category => category.id))
  const empty = computed(() => loaded.value && !loading.value && categories.value.length === 0)

  async function loadGameHome() {
    if (loading.value) return

    loading.value = true
    error.value = null

    try {
      const data = await getAppGameHome()
      categories.value = createDynamicHomeCategories(data || {})
    } catch (caught) {
      error.value = caught instanceof Error ? caught : new Error('Failed to load games')
      categories.value = []
    } finally {
      loaded.value = true
      loading.value = false
    }
  }

  return {
    loading,
    error,
    empty,
    loaded,
    categories,
    categoryIds,
    loadGameHome,
  }
}
