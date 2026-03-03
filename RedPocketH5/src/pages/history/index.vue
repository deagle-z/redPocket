<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { getLuckyAppHistory } from '@/api/user'

type RangeKey = 'today' | 'week' | 'month' | 'custom'
type TxType = 'send' | 'grab'
type TxResult = 'loss' | 'win'
type MenuKey = 'type' | 'result'

const DEFAULT_AVATAR = 'https://game.luckypacket.me/images/avatar-placeholder.png'

interface TxItem {
  id: string
  avatar: string
  title: string
  sub: string
  time: string
  amount: number
  badge: string
  type: TxType
  result: TxResult
}

const router = useRouter()
const activeRange = ref<RangeKey>('custom')
const showRangePopup = ref(false)
const showDatePicker = ref(false)
const editingField = ref<'start' | 'end'>('start')
const pickerValues = ref<string[]>([])
const activeMenu = ref<MenuKey | null>(null)
const filter = reactive<{
  type: 'all' | 'send' | 'grab'
  result: 'all' | 'win' | 'loss'
}>({
  type: 'all',
  result: 'all',
})

const typeOptions = [
  { label: '全部类型', value: 'all' as const },
  { label: '我发出的', value: 'send' as const },
  { label: '我抢到的', value: 'grab' as const },
]

const resultOptions = [
  { label: '全部结果', value: 'all' as const },
  { label: '盈利/中奖', value: 'win' as const },
  { label: '亏损/中雷', value: 'loss' as const },
]

const typeLabel = computed(() => typeOptions.find(item => item.value === filter.type)?.label || '全部类型')
const resultLabel = computed(() => resultOptions.find(item => item.value === filter.result)?.label || '全部结果')

function formatYmd(date: Date) {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function ymdToPickerValues(ymd: string) {
  const [y, m, d] = ymd.split('-')
  return [y || '2026', m || '01', d || '01']
}

function pickerValuesToYmd(values: string[]) {
  const [y = '2026', m = '01', d = '01'] = values
  return `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
}

function getTodayYmd() {
  return formatYmd(new Date())
}

function getPrevMonthFirstDayYmd() {
  const now = new Date()
  return formatYmd(new Date(now.getFullYear(), now.getMonth() - 1, 1))
}

const customRange = reactive({
  start: getPrevMonthFirstDayYmd(),
  end: getTodayYmd(),
})

const tempCustomRange = reactive({
  start: customRange.start,
  end: customRange.end,
})

const allList = ref<TxItem[]>([])
const pageSize = 20
const currentPage = ref(0)
const total = ref(0)
const listLoading = ref(false)
const fetching = ref(false)
const summary = reactive({
  income: 0,
  expense: 0,
  pnl: 0,
})

const displayList = computed(() => allList.value)
const finished = computed(() => total.value > 0 && allList.value.length >= total.value)
const showEmpty = computed(() => !listLoading.value && allList.value.length === 0)

const stats = computed(() => {
  return {
    income: summary.income,
    expense: summary.expense,
    pnl: summary.pnl,
  }
})

const dateTabs = [
  { key: 'today', label: '今日' },
  { key: 'week', label: '本周' },
  { key: 'month', label: '本月' },
  { key: 'custom', label: '自定义' },
] as const

function formatAmount(value: number) {
  const abs = Math.abs(value)
  const sign = value >= 0 ? '+' : '-'
  return `${sign}₱${abs.toFixed(2)}`
}

function goBack() {
  router.back()
}

function getActiveRange() {
  const today = new Date()
  if (activeRange.value === 'today') {
    const d = formatYmd(today)
    return { start: d, end: d }
  }

  if (activeRange.value === 'week') {
    const day = today.getDay() || 7
    const monday = new Date(today)
    monday.setDate(today.getDate() - day + 1)
    return { start: formatYmd(monday), end: formatYmd(today) }
  }

  if (activeRange.value === 'month') {
    const start = new Date(today.getFullYear(), today.getMonth(), 1)
    return { start: formatYmd(start), end: formatYmd(today) }
  }

  return { start: customRange.start, end: customRange.end }
}

function formatTime(raw: string) {
  const d = new Date(raw)
  if (Number.isNaN(d.getTime()))
    return raw || ''
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${hh}:${mm} · ${formatYmd(d)}`
}

function toggleMenu(key: MenuKey) {
  activeMenu.value = activeMenu.value === key ? null : key
}

function closeMenu() {
  activeMenu.value = null
}

function selectType(value: 'all' | 'send' | 'grab') {
  filter.type = value
  closeMenu()
}

function selectResult(value: 'all' | 'win' | 'loss') {
  filter.result = value
  closeMenu()
}

function onDateTabClick(tab: RangeKey) {
  activeRange.value = tab
  if (tab === 'custom')
    openRangePopup()
}

function openRangePopup() {
  tempCustomRange.start = customRange.start
  tempCustomRange.end = customRange.end
  showRangePopup.value = true
}

function cancelRange() {
  showRangePopup.value = false
}

function confirmRange() {
  if (tempCustomRange.start > tempCustomRange.end)
    return
  customRange.start = tempCustomRange.start
  customRange.end = tempCustomRange.end
  showRangePopup.value = false
}

function openDatePicker(field: 'start' | 'end') {
  editingField.value = field
  pickerValues.value = ymdToPickerValues(tempCustomRange[field])
  showDatePicker.value = true
}

function cancelDatePicker() {
  showDatePicker.value = false
}

function confirmDatePicker(payload: any) {
  const values = (payload?.selectedValues || payload) as string[]
  const ymd = pickerValuesToYmd(values)
  tempCustomRange[editingField.value] = ymd
  showDatePicker.value = false
}

function toUnixStart(ymd: string) {
  const d = new Date(`${ymd}T00:00:00`)
  return Math.floor(d.getTime() / 1000)
}

function toUnixEnd(ymd: string) {
  const d = new Date(`${ymd}T23:59:59`)
  return Math.floor(d.getTime() / 1000)
}

function mapActionType() {
  if (filter.type === 'send')
    return 1
  if (filter.type === 'grab')
    return 2
  return 0
}

function mapResultType() {
  if (filter.result === 'win')
    return 1
  if (filter.result === 'loss')
    return 2
  return 0
}

function mapHistoryItem(item: any): TxItem {
  const isSend = Number(item?.actionType) === 1 || item?.recordType === 'send'
  const amount = Number(item?.netProfit || 0)
  const isWin = amount >= 0
  const senderName = item?.senderName || 'User'
  const luckyAmount = Number(item?.luckyAmount || 0)
  const grabAmount = Number(item?.grabAmount || 0)
  const thunder = Number(item?.thunder || 0)

  return {
    id: String(item?.recordId || item?.luckyId || Date.now()),
    avatar: item?.avatar || DEFAULT_AVATAR,
    title: isSend ? `发出红包 ${luckyAmount.toFixed(2)}` : `抢到${senderName}的红包 ₱${grabAmount.toFixed(2)}`,
    sub: isSend ? `雷号: ${thunder}` : `雷号: ${thunder}`,
    time: formatTime(item?.createdAt || ''),
    amount,
    badge: isWin ? '中奖' : '亏损',
    type: isSend ? 'send' : 'grab',
    result: isWin ? 'win' : 'loss',
  }
}

function onAvatarError(event: Event) {
  const el = event?.target as HTMLImageElement | null
  if (!el)
    return
  if (el.src === DEFAULT_AVATAR)
    return
  el.src = DEFAULT_AVATAR
}

async function loadHistory(reset = false) {
  if (fetching.value)
    return

  fetching.value = true
  listLoading.value = true
  const range = getActiveRange()
  const page = reset ? 0 : currentPage.value
  try {
    const { data } = await getLuckyAppHistory({
      currentPage: page,
      pageSize,
      actionType: mapActionType() as 0 | 1 | 2,
      resultType: mapResultType() as 0 | 1 | 2,
      startTime: toUnixStart(range.start),
      endTime: toUnixEnd(range.end),
    })

    const mapped = (data?.list || []).map((item: any) => mapHistoryItem(item))
    if (reset)
      allList.value = mapped
    else
      allList.value = [...allList.value, ...mapped]

    total.value = Number(data?.total || 0)
    currentPage.value = page + 1
    summary.income = Number(data?.totalIncome || 0)
    summary.expense = Number(data?.totalExpense || 0)
    summary.pnl = Number(data?.netProfitLoss || 0)
  }
  finally {
    fetching.value = false
    listLoading.value = false
  }
}

function reloadHistory() {
  currentPage.value = 0
  total.value = 0
  allList.value = []
  void loadHistory(true).catch(() => {})
}

function onLoadMore() {
  if (finished.value)
    return
  void loadHistory(false).catch(() => {})
}

watch(() => [filter.type, filter.result, activeRange.value, customRange.start, customRange.end], () => {
  if (activeRange.value !== 'custom') {
    reloadHistory()
    return
  }
  if (customRange.start && customRange.end)
    reloadHistory()
})

onMounted(() => {
  reloadHistory()
})
</script>

<template>
  <div class="history-page">
    <AppPageHeader title="历史记录" @back="goBack" />

    <div class="filter-wrap">
      <section class="filter-bar card">
        <button type="button" class="filter-btn" :class="{ active: activeMenu === 'type' }" @click="toggleMenu('type')">
          <span>{{ typeLabel }}</span>
          <van-icon :name="activeMenu === 'type' ? 'arrow-up' : 'arrow-down'" />
        </button>
        <button type="button" class="filter-btn" :class="{ active: activeMenu === 'result' }" @click="toggleMenu('result')">
          <span>{{ resultLabel }}</span>
          <van-icon :name="activeMenu === 'result' ? 'arrow-up' : 'arrow-down'" />
        </button>
      </section>

      <div v-if="activeMenu" class="dropdown-panel">
        <template v-if="activeMenu === 'type'">
          <button
            v-for="item in typeOptions"
            :key="item.value"
            type="button"
            class="dropdown-option"
            @click="selectType(item.value)"
          >
            <span>{{ item.label }}</span>
            <van-icon v-if="filter.type === item.value" name="success" />
          </button>
        </template>
        <template v-else>
          <button
            v-for="item in resultOptions"
            :key="item.value"
            type="button"
            class="dropdown-option"
            @click="selectResult(item.value)"
          >
            <span>{{ item.label }}</span>
            <van-icon v-if="filter.result === item.value" name="success" />
          </button>
        </template>
      </div>
      <div v-if="activeMenu" class="dropdown-mask" @click="closeMenu" />
    </div>

    <section class="date-tabs card">
      <button
        v-for="tab in dateTabs"
        :key="tab.key"
        type="button"
        class="date-tab"
        :class="{ active: activeRange === tab.key }"
        @click="onDateTabClick(tab.key)"
      >
        <van-icon v-if="tab.key === 'custom'" name="calendar-o" />
        <span>{{ tab.label }}</span>
      </button>
    </section>

    <section class="stats-row card">
      <div class="stat-item">
        <p class="stat-label">
          总收入
        </p>
        <p class="stat-value income">
          {{ formatAmount(stats.income) }}
        </p>
      </div>
      <div class="stat-item">
        <p class="stat-label">
          总支出
        </p>
        <p class="stat-value expense">
          {{ formatAmount(-stats.expense) }}
        </p>
      </div>
      <div class="stat-item pnl">
        <p class="stat-label">
          净盈亏
        </p>
        <p class="stat-value pnl-text">
          {{ formatAmount(stats.pnl) }}
        </p>
      </div>
    </section>

    <van-list
      v-model:loading="listLoading"
      :finished="finished"
      finished-text="没有更多了"
      @load="onLoadMore"
    >
      <section class="tx-list">
        <article
          v-for="item in displayList"
          :key="item.id"
          class="tx-item"
          :class="item.result"
        >
          <div class="tx-left-avatar">
            <img :src="item.avatar" alt="" class="tx-avatar-img" @error="onAvatarError">
          </div>

          <div class="tx-main">
            <p class="tx-title">
              {{ item.title }}
            </p>
            <p class="tx-sub">
              {{ item.sub }}
            </p>
            <span class="tx-badge" :class="item.result">{{ item.badge }}</span>
          </div>

          <div class="tx-right">
            <p class="tx-time">
              {{ item.time }}
            </p>
            <p class="tx-amount" :class="item.result">
              {{ formatAmount(item.amount) }}
            </p>
          </div>
        </article>
      </section>
    </van-list>

    <section v-if="showEmpty" class="empty-state card">
      <van-icon name="description" />
      <p>暂无历史记录</p>
    </section>

    <van-popup v-model:show="showRangePopup" round position="bottom" class="range-popup">
      <div class="range-header">
        <span class="range-title">选择日期范围</span>
        <div class="range-actions">
          <button type="button" class="range-cancel-btn" @click="cancelRange">
            取消
          </button>
          <button type="button" class="range-confirm-btn" @click="confirmRange">
            确认
          </button>
        </div>
      </div>

      <div class="range-row">
        <span>开始日期</span>
        <div class="range-date-wrap" @click="openDatePicker('start')">
          <span class="range-date-text">{{ tempCustomRange.start }}</span>
          <van-icon name="arrow" />
        </div>
      </div>

      <div class="range-row">
        <span>结束日期</span>
        <div class="range-date-wrap" @click="openDatePicker('end')">
          <span class="range-date-text">{{ tempCustomRange.end }}</span>
          <van-icon name="arrow" />
        </div>
      </div>
    </van-popup>

    <van-popup v-model:show="showDatePicker" round position="bottom">
      <van-date-picker
        v-model="pickerValues"
        title=""
        :show-toolbar="true"
        :columns-type="['year', 'month', 'day']"
        cancel-button-text="取消"
        confirm-button-text="确认"
        @cancel="cancelDatePicker"
        @confirm="confirmDatePicker"
      />
    </van-popup>
  </div>
</template>

<style scoped>
.history-page {
  min-height: 100vh;
  background: var(--bg-secondary);
  padding: 0 10px calc(14px + env(safe-area-inset-bottom));
}

.card {
  background: #fff;
  border-radius: 10px;
}

.filter-bar {
  height: 52px;
  padding: 0 14px;
  display: flex;
  align-items: center;
  justify-content: space-around;
}

.filter-wrap {
  margin-bottom: 8px;
}

.filter-btn {
  border: 0;
  background: transparent;
  color: #1a1a2e;
  font-size: 16px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.filter-btn.active {
  color: var(--color-primary);
  text-decoration: underline;
}

.filter-btn :deep(.van-icon) {
  color: #d1d5db;
  font-size: 14px;
}

.filter-btn.active :deep(.van-icon) {
  color: var(--color-primary);
}

.dropdown-panel {
  background: #fff;
  border-radius: 0 0 10px 10px;
  border-top: 1px solid #eef1f4;
  overflow: hidden;
  position: relative;
  z-index: 20;
}

.dropdown-option {
  height: 56px;
  width: 100%;
  border: 0;
  background: #fff;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #1a1a2e;
  font-size: 16px;
  text-align: left;
}

.dropdown-option + .dropdown-option {
  border-top: 1px solid #f3f4f6;
}

.dropdown-option :deep(.van-icon-success) {
  color: var(--color-primary);
  font-size: 18px;
}

.dropdown-mask {
  height: calc(100vh - 220px);
  background: rgba(0, 0, 0, 0.45);
}

.date-tabs {
  padding: 10px 12px;
  margin-bottom: 8px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.date-tab {
  height: 36px;
  border-radius: 18px;
  border: 1px solid #e0e0e0;
  background: #fff;
  color: #6b7280;
  font-size: 13px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.date-tab.active {
  border-color: #e6a800;
  background: #e6a800;
  color: #fff;
  font-weight: 600;
}

.range-popup {
  min-height: 260px;
}

.range-header {
  height: 72px;
  padding: 0 16px;
  border-bottom: 1px solid #eef1f4;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.range-title {
  color: #111827;
  font-size: 16px;
  font-weight: 600;
}

.range-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.range-cancel-btn,
.range-confirm-btn {
  min-width: 56px;
  height: 36px;
  border-radius: 8px;
  font-size: 16px;
  border: 1px solid #d8dde5;
}

.range-cancel-btn {
  background: #fff;
  color: #111827;
}

.range-confirm-btn {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: #fff;
}

.range-row {
  height: 62px;
  padding: 0 16px;
  border-bottom: 1px solid #f2f4f7;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 16px;
  color: #111827;
}

.range-date-wrap {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.range-date-text {
  font-size: 16px;
  color: #64748b;
}

.range-date-wrap :deep(.van-icon) {
  color: #94a3b8;
  font-size: 16px;
}

.stats-row {
  padding: 10px 12px;
  margin-bottom: 8px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.stat-item {
  height: 64px;
  border-radius: 10px;
  border: 1px solid #e8eaf0;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 4px;
}

.stat-item.pnl {
  background: #fff8e6;
  border-color: #e6a800;
}

.stat-label {
  margin: 0;
  font-size: 11px;
  color: #9ca3af;
}

.stat-value {
  margin: 0;
  font-size: 16px;
  line-height: 1;
  font-weight: 700;
}

.stat-value.income {
  color: var(--color-primary);
}

.stat-value.expense {
  color: #ff4d4f;
}

.stat-value.pnl-text {
  color: #e6a800;
}

.tx-list {
  display: grid;
  gap: 8px;
}

.tx-item {
  border-radius: 12px;
  background: #fff;
  padding: 12px;
  display: grid;
  grid-template-columns: 36px 1fr auto;
  gap: 10px;
  align-items: start;
  border-left: 3px solid var(--color-primary);
}

.tx-item.loss {
  border-left-color: #ff4d4f;
}

.tx-left-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  overflow: hidden;
  background: #eef2f7;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tx-avatar-img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.tx-main {
  min-width: 0;
}

.tx-title {
  margin: 0;
  color: #1a1a2e;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.35;
}

.tx-sub {
  margin: 3px 0 5px;
  color: #9ca3af;
  font-size: 11px;
}

.tx-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 3px 10px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
}

.tx-badge.win {
  background: #fff8e1;
  color: #e6a800;
}

.tx-badge.loss {
  background: #ffebee;
  color: #ff4d4f;
}

.tx-right {
  text-align: right;
}

.tx-time {
  margin: 0;
  color: #9ca3af;
  font-size: 10px;
}

.tx-amount {
  margin: 24px 0 0;
  font-size: 15px;
  line-height: 1;
  font-weight: 700;
}

.tx-amount.win {
  color: var(--color-primary);
}

.tx-amount.loss {
  color: #ff4d4f;
}

.empty-state {
  margin-top: 10px;
  min-height: 180px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #9ca3af;
}

.empty-state :deep(.van-icon) {
  font-size: 18px;
  color: #cbd5e1;
}

.empty-state p {
  margin: 0;
  font-size: 14px;
}
</style>

<route lang="json5">
{
  name: 'History'
}
</route>

