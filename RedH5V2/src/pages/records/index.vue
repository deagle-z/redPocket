<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import {
  getAppUserBetRecordList,
  getRechargeOrderHistory,
  getWithdrawOrderHistory,
} from '@/api/user'
import type {
  AppOrderHistoryItem,
  AppUserBetRecordItem,
  AppUserBetRecordReq,
  WithdrawOrderHistoryItem,
} from '@/api/user'
import { formatMoney, toCent } from '@/utils/money'
import PpmxMoneyText from '@/components/PpmxMoneyText.vue'
import PpmxSegmentedOptions from '@/components/PpmxSegmentedOptions.vue'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '../ppmx-home/types'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'

definePage({
  name: 'records',
  meta: {
    titleKey: 'records.title',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: true,
  },
})

type RecordKind = 'deposit' | 'withdraw' | 'bet' | 'win'
type RecordStatus =
  | 'done'
  | 'processing'
  | 'pending'
  | 'won'
  | 'lost'
  | 'failed'
  | 'cancelled'
  | 'closed'
  | 'refunded'
  | 'returned'
  | 'unknown'
type RecordsPeriod = 'today' | '3d' | '7d' | '1mo'
type RecordsType = 'deposit' | 'withdraw' | 'bet'
type RecordsCategory = 'all' | 'slots' | 'casino' | 'blockchain' | 'fishing' | 'rummy' | 'sports' | 'mini' | 'lottery'

interface RecordsOption<T extends string> {
  value: T
  label: LocalizedText
}

interface RecordsLedgerItem {
  id: string
  time: string
  dateKey: string
  kind: RecordKind
  status: RecordStatus
  statusLabel: LocalizedText
  amount: number
  label: LocalizedText
  category?: RecordsCategory
  rawTime?: string
}

interface RecordStatusMeta {
  status: RecordStatus
  label: LocalizedText
}

const pageSize = 10
const { pickText } = usePpmxLocale()
const recordsText = {
  eyebrow: { es: 'Historial de movimientos', en: 'Transaction history', zh: '流水记录' },
  title: { es: 'Historial', en: 'History', zh: '流水记录' },
  sub: {
    es: 'Filtra por periodo, tipo y categoría de juego.',
    en: 'Filter by period, type and game category.',
    zh: '按时间、类型与游戏类目筛选。',
  },
  period: { es: 'Periodo', en: 'Period', zh: '时间' },
  type: { es: 'Tipo', en: 'Type', zh: '类型' },
  category: { es: 'Categoría de juego', en: 'Game category', zh: '游戏类目' },
  amount: { es: 'Monto', en: 'Amount', zh: '金额' },
  net: { es: 'Neto', en: 'Net', zh: '净额' },
  count: { es: 'Registros', en: 'Records', zh: '记录' },
  empty: {
    es: 'Sin movimientos en este periodo.',
    en: 'No activity in this period.',
    zh: '该时间段暂无记录。',
  },
  loadingTitle: { es: 'Cargando historial', en: 'Loading records', zh: '正在加载流水' },
  loadingMessage: { es: 'Consultando movimientos recientes.', en: 'Fetching recent activity.', zh: '正在查询最近记录。' },
  errorTitle: { es: 'No se pudo cargar', en: 'Could not load records', zh: '流水加载失败' },
  errorMessage: { es: 'Revisa tu conexión e inténtalo de nuevo.', en: 'Check your connection and retry.', zh: '请检查网络后重试。' },
  retry: { es: 'Reintentar', en: 'Retry', zh: '重试' },
  operation: { es: 'operación', en: 'operation', zh: '笔' },
  operations: { es: 'operaciones', en: 'operations', zh: '笔' },
} satisfies Record<string, LocalizedText>

const periodOptions: Array<RecordsOption<RecordsPeriod>> = [
  { value: 'today', label: { es: 'Hoy', en: 'Today', zh: '今天' } },
  { value: '3d', label: { es: '3 días', en: '3 days', zh: '三天' } },
  { value: '7d', label: { es: '7 días', en: '7 days', zh: '七天' } },
  { value: '1mo', label: { es: '1 mes', en: '1 month', zh: '一个月' } },
]
const typeOptions: Array<RecordsOption<RecordsType>> = [
  { value: 'deposit', label: { es: 'Depósitos', en: 'Deposits', zh: '充值' } },
  { value: 'withdraw', label: { es: 'Retiros', en: 'Withdrawals', zh: '提现' } },
  { value: 'bet', label: { es: 'Apuestas', en: 'Bets', zh: '下注' } },
]
const categoryOptions: Array<RecordsOption<RecordsCategory>> = [
  { value: 'all', label: { es: 'Todas', en: 'All', zh: '全部' } },
  { value: 'slots', label: { es: 'Tragamonedas', en: 'Slots', zh: '老虎机' } },
  { value: 'casino', label: { es: 'Casino en vivo', en: 'Live Casino', zh: '真人娱乐场' } },
  { value: 'blockchain', label: { es: 'Blockchain', en: 'Blockchain', zh: '区块链' } },
  { value: 'fishing', label: { es: 'Pesca', en: 'Fishing', zh: '捕鱼' } },
  { value: 'rummy', label: { es: 'Rummy', en: 'Rummy', zh: '纸牌' } },
  { value: 'sports', label: { es: 'Deportes', en: 'Sports', zh: '体育' } },
  { value: 'mini', label: { es: 'Mini juegos', en: 'Mini Games', zh: '小游戏' } },
  { value: 'lottery', label: { es: 'Sorteos', en: 'Lottery', zh: '彩票' } },
]

const selectedPeriod = ref<RecordsPeriod>('7d')
const selectedType = ref<RecordsType>('deposit')
const selectedCategory = ref<RecordsCategory>('all')
const records = ref<RecordsLedgerItem[]>([])
const recordTotal = ref(0)
const loading = ref(false)
const loadError = ref(false)

const filteredRecords = computed(() => records.value)
const summary = computed(() => {
  const amountTotal = records.value.reduce((total, item) => total + Math.abs(item.amount), 0)
  const netTotal = records.value.reduce((total, item) => total + item.amount, 0)

  return {
    amount: amountTotal,
    net: netTotal,
    count: recordTotal.value || records.value.length,
  }
})
const groupedRecords = computed(() => {
  const groups = new Map<string, RecordsLedgerItem[]>()

  records.value.forEach((item) => {
    groups.set(item.dateKey, [...(groups.get(item.dateKey) || []), item])
  })

  return [...groups.entries()].map(([dateKey, items]) => ({ dateKey, items }))
})
const periodSegmentOptions = computed(() =>
  periodOptions.map(option => ({
    label: pickText(option.label),
    value: option.value,
  })),
)
const typeSegmentOptions = computed(() =>
  typeOptions.map(option => ({
    label: pickText(option.label),
    value: option.value,
  })),
)
const categorySegmentOptions = computed(() =>
  categoryOptions.map(option => ({
    label: pickText(option.label),
    value: option.value,
  })),
)

function getPeriodRange(period: RecordsPeriod) {
  const end = new Date()
  const start = new Date(end)
  const days = {
    today: 1,
    '3d': 3,
    '7d': 7,
    '1mo': 30,
  }[period]

  start.setHours(0, 0, 0, 0)
  start.setDate(start.getDate() - (days - 1))

  return {
    startTime: Math.floor(start.getTime() / 1000),
    endTime: Math.floor(end.getTime() / 1000),
  }
}

function toMoneyCent(value: number | string | undefined | null) {
  const numeric = Number(value || 0)
  if (!Number.isFinite(numeric)) return 0

  return toCent(numeric.toFixed(2))
}

function formatAmount(value: number) {
  return formatMoney(value, { currency: '$' })
}

function formatSignedAmount(value: number) {
  const sign = value > 0 ? '+' : value < 0 ? '-' : ''

  return `${sign}${formatAmount(Math.abs(value))}`
}

function formatTime(raw?: string | null) {
  if (!raw) return '--'

  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw

  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')

  return `${hours}:${minutes}`
}

function formatDateKey(raw?: string | null) {
  if (!raw) return pickText({ es: 'Sin fecha', en: 'No date', zh: '无日期' })

  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw

  const today = new Date()
  const yesterday = new Date()
  yesterday.setDate(today.getDate() - 1)

  if (date.toDateString() === today.toDateString()) return pickText({ es: 'Hoy', en: 'Today', zh: '今天' })
  if (date.toDateString() === yesterday.toDateString()) return pickText({ es: 'Ayer', en: 'Yesterday', zh: '昨天' })

  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function unknownStatusLabel(status: number): LocalizedText {
  return { es: `Estado ${status}`, en: `Status ${status}`, zh: `状态${status}` }
}

function mapRechargeStatus(status: number): RecordStatusMeta {
  switch (status) {
    case 0:
      return { status: 'pending', label: { es: 'Pago pendiente', en: 'Pending payment', zh: '待支付' } }
    case 1:
      return { status: 'done', label: { es: 'Pagado', en: 'Paid', zh: '支付成功' } }
    case 2:
      return { status: 'processing', label: { es: 'Estado 2', en: 'Status 2', zh: '状态2' } }
    case 3:
      return { status: 'failed', label: { es: 'Fallido', en: 'Failed', zh: '失败' } }
    case 4:
      return { status: 'cancelled', label: { es: 'Cancelado', en: 'Cancelled', zh: '取消' } }
    case 5:
      return { status: 'closed', label: { es: 'Cerrado/vencido', en: 'Closed/expired', zh: '关闭/超时' } }
    case 6:
      return { status: 'processing', label: { es: 'Reembolso en curso', en: 'Refunding', zh: '退款中' } }
    case 7:
      return { status: 'refunded', label: { es: 'Reembolsado', en: 'Refunded', zh: '已退款' } }
    default:
      return { status: 'unknown', label: unknownStatusLabel(status) }
  }
}

function mapWithdrawStatus(status: number): RecordStatusMeta {
  switch (status) {
    case 0:
      return { status: 'pending', label: { es: 'En revisión', en: 'Under review', zh: '待审核' } }
    case 1:
      return { status: 'processing', label: { es: 'Por pagar', en: 'Pending payout', zh: '待打款' } }
    case 2:
      return { status: 'processing', label: { es: 'Pagando', en: 'Paying', zh: '打款中' } }
    case 3:
      return { status: 'done', label: { es: 'Retiro exitoso', en: 'Withdrawal successful', zh: '提现成功' } }
    case 4:
      return { status: 'failed', label: { es: 'Fallido', en: 'Failed', zh: '失败' } }
    case 5:
      return { status: 'cancelled', label: { es: 'Cancelado', en: 'Cancelled', zh: '取消' } }
    case 6:
      return { status: 'returned', label: { es: 'Devuelto', en: 'Returned', zh: '退回' } }
    default:
      return { status: 'unknown', label: unknownStatusLabel(status) }
  }
}

function mapRechargeRecord(item: AppOrderHistoryItem): RecordsLedgerItem {
  const time = item.createdAt || item.time
  const statusMeta = mapRechargeStatus(Number(item.status))

  return {
    id: item.orderNo,
    time: formatTime(time),
    dateKey: formatDateKey(time),
    kind: 'deposit',
    status: statusMeta.status,
    statusLabel: statusMeta.label,
    amount: toMoneyCent(item.amount),
    label: { es: 'Depósito', en: 'Deposit', zh: '充值' },
    rawTime: time,
  }
}

function mapWithdrawRecord(item: WithdrawOrderHistoryItem): RecordsLedgerItem {
  const statusMeta = mapWithdrawStatus(Number(item.status))

  return {
    id: item.orderNo,
    time: formatTime(item.createdAt),
    dateKey: formatDateKey(item.createdAt),
    kind: 'withdraw',
    status: statusMeta.status,
    statusLabel: statusMeta.label,
    amount: -toMoneyCent(item.amount),
    label: { es: 'Retiro', en: 'Withdrawal', zh: '提现' },
    rawTime: item.createdAt,
  }
}

function mapBetRecord(item: AppUserBetRecordItem): RecordsLedgerItem {
  const netAmount = Number(item.winAmount || 0) - Number(item.betAmount || 0)
  const isWin = netAmount > 0

  return {
    id: String(item.id || item.roundId || item.traceId),
    time: formatTime(item.date || item.createTime),
    dateKey: formatDateKey(item.date || item.createTime),
    kind: isWin ? 'win' : 'bet',
    status: isWin ? 'won' : 'lost',
    statusLabel: isWin
      ? { es: 'Ganada', en: 'Won', zh: '已中奖' }
      : { es: 'Perdida', en: 'Lost', zh: '未中奖' },
    amount: toMoneyCent(netAmount),
    label: { es: item.gameName || 'Apuesta', en: item.gameName || 'Bet', zh: item.gameName || '下注' },
    category: selectedCategory.value === 'all' ? undefined : selectedCategory.value,
    rawTime: item.date || item.createTime || undefined,
  }
}

async function loadRecords() {
  loading.value = true
  loadError.value = false

  try {
    if (selectedType.value === 'deposit') {
      const data = await getRechargeOrderHistory({ currentPage: 0, pageSize })
      records.value = (data.list || []).map(mapRechargeRecord)
      recordTotal.value = Number(data.total || records.value.length)
    } else if (selectedType.value === 'withdraw') {
      const data = await getWithdrawOrderHistory({ currentPage: 0, pageSize })
      records.value = (data.list || []).map(mapWithdrawRecord)
      recordTotal.value = Number(data.total || records.value.length)
    } else {
      const range = getPeriodRange(selectedPeriod.value)
      const payload: AppUserBetRecordReq = {
        currentPage: 0,
        pageSize,
        ...range,
      }

      if (selectedCategory.value !== 'all') {
        payload.categoryCode = selectedCategory.value
      }

      const data = await getAppUserBetRecordList(payload)
      records.value = (data.list || []).map(mapBetRecord)
      recordTotal.value = Number(data.total || records.value.length)
    }
  } catch {
    records.value = []
    recordTotal.value = 0
    loadError.value = true
  } finally {
    loading.value = false
  }
}

function kindClass(kind: RecordKind) {
  return `is-${kind}`
}

function statusClass(status: RecordStatus) {
  return `is-${status}`
}

function displayStatusLabel(item: RecordsLedgerItem) {
  return pickText(item.statusLabel)
}

function kindIcon(kind: RecordKind) {
  return {
    deposit: 'fa-arrow-down',
    withdraw: 'fa-arrow-up',
    bet: 'fa-receipt',
    win: 'fa-trophy',
  }[kind]
}

function categoryLabel(category?: RecordsCategory) {
  const option = categoryOptions.find(item => item.value === category)

  return option ? pickText(option.label) : ''
}

watch([selectedPeriod, selectedType, selectedCategory], () => {
  void loadRecords()
})

onMounted(() => {
  void loadRecords()
})
</script>

<template>
  <main class="ppmx-page ppmx-subpage ppmx-records-page">
    <PpmxPageChrome>
      <section class="ppmx-subpage__compact ppmx-records-stack">
        <header class="ppmx-subpage-head ppmx-records-head reveal">
          <div class="ppmx-eyebrow">{{ pickText(recordsText.eyebrow) }}</div>
          <h1 class="ppmx-display">{{ pickText(recordsText.title) }}</h1>
          <p>{{ pickText(recordsText.sub) }}</p>
        </header>

        <section class="ppmx-records-filter reveal">
          <div class="ppmx-records-filter__group">
            <p>{{ pickText(recordsText.period) }}</p>
            <PpmxSegmentedOptions
              id="walletTime"
              v-model="selectedPeriod"
              :options="periodSegmentOptions"
            />
          </div>

          <div class="ppmx-records-filter__group">
            <p>{{ pickText(recordsText.type) }}</p>
            <PpmxSegmentedOptions
              id="walletType"
              v-model="selectedType"
              :options="typeSegmentOptions"
            />
          </div>

          <div v-if="selectedType === 'bet'" class="ppmx-records-filter__group">
            <p>{{ pickText(recordsText.category) }}</p>
            <PpmxSegmentedOptions
              id="walletCat"
              v-model="selectedCategory"
              :options="categorySegmentOptions"
            />
          </div>

          <div id="walletSummary" class="ppmx-records-summary">
            <div>
              <PpmxMoneyText :cents="summary.amount" />
              <span>{{ pickText(recordsText.amount) }}</span>
            </div>
            <div>
              <PpmxMoneyText
                :cents="summary.net"
                signed
                :class="{ 'is-deposit': summary.net > 0, 'is-withdraw': summary.net < 0 }"
              />
              <span>{{ pickText(recordsText.net) }}</span>
            </div>
            <div>
              <strong>{{ summary.count }}</strong>
              <span>{{ pickText(recordsText.count) }}</span>
            </div>
          </div>
        </section>

        <AppState
          v-if="loading"
          status="loading"
          :title="pickText(recordsText.loadingTitle)"
          :message="pickText(recordsText.loadingMessage)"
          skeleton-variant="records-page"
        />
        <AppState
          v-else-if="loadError"
          status="error"
          :title="pickText(recordsText.errorTitle)"
          :message="pickText(recordsText.errorMessage)"
          :action-text="pickText(recordsText.retry)"
          @retry="loadRecords"
        />

        <section v-else id="walletList" class="ppmx-records-list reveal">
          <div v-if="!filteredRecords.length" class="ppmx-records-empty">
            <i class="fa-regular fa-folder-open" />
            <span>{{ pickText(recordsText.empty) }}</span>
          </div>

          <div v-for="group in groupedRecords" v-else :key="group.dateKey" class="ppmx-records-day">
            <div class="ppmx-records-day__head">
              <span>{{ group.dateKey }}</span>
              <small>
                {{ group.items.length }}
                {{ pickText(group.items.length === 1 ? recordsText.operation : recordsText.operations) }}
              </small>
            </div>

            <article
              v-for="item in group.items"
              :key="item.id"
              class="ppmx-records-card"
              :class="kindClass(item.kind)"
            >
              <span class="ppmx-records-card__bar" />
              <div class="ppmx-records-card__main">
                <span class="ppmx-records-card__icon" :class="kindClass(item.kind)">
                  <i class="fa-solid" :class="kindIcon(item.kind)" />
                </span>
                <div>
                  <strong>{{ pickText(item.label) }}</strong>
                  <p>
                    <span><i class="fa-regular fa-clock" /> {{ item.time }}</span>
                    <em v-if="selectedType === 'bet' && categoryLabel(item.category)">
                      {{ categoryLabel(item.category) }}
                    </em>
                  </p>
                </div>
              </div>
              <div class="ppmx-records-card__amount">
                <strong :class="kindClass(item.kind)">{{ formatSignedAmount(item.amount) }}</strong>
                <span :class="statusClass(item.status)">
                  <i />
                  {{ displayStatusLabel(item) }}
                </span>
              </div>
            </article>
          </div>
        </section>
      </section>
    </PpmxPageChrome>
  </main>
</template>
