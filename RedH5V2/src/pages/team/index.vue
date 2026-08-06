<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import QRCode from 'qrcode'
import { showToast } from 'vant'
import type { TaskActivityItem, TgInviteRuleConfig, TgInviteStats, WithdrawAccountItem } from '@/api/user'
import {
  claimTaskActivity,
  createRebateWithdrawOrder,
  getCurrentTgInviteRuleConfig,
  getCurrentTgInviteStats,
  getTaskActivityList,
  getTaskActivityRecords,
  getWithdrawAccounts,
  rewardTaskActivity,
  transferRebateToBalance,
} from '@/api/user'
import { APP_COUNTRY_CODE, APP_CURRENCY, APP_CURRENCY_SYMBOL } from '@/config/market'
import { HttpError } from '@/utils/http'
import { formatMoney, toCent } from '@/utils/money'
import PpmxButton from '@/components/PpmxButton.vue'
import PpmxGuestPanel from '@/components/PpmxGuestPanel.vue'
import PpmxSelectInput from '@/components/PpmxSelectInput.vue'
import PpmxWithdrawModal from '@/components/PpmxWithdrawModal.vue'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'

definePage({
  name: 'team',
  meta: {
    titleKey: 'ppmx.team.title',
    shell: 'ppmx',
    tabbar: false,
  },
})

const { t, locale } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const pageChromeRef = ref<InstanceType<typeof PpmxPageChrome> | null>(null)

const teamLoading = ref(false)
const teamLoaded = ref(false)
const teamError = ref('')
const stats = ref<TgInviteStats | null>(null)
const ruleConfig = ref<TgInviteRuleConfig | null>(null)
const inviteQrDataUrl = ref('')
const inviteQrError = ref(false)
const transferModalOpen = ref(false)
const transferAmount = ref('')
const commissionWithdrawModalOpen = ref(false)
const commissionWithdrawAmount = ref('')
const commissionWithdrawAccountId = ref('')
const withdrawAccounts = ref<WithdrawAccountItem[]>([])
const withdrawAccountsLoading = ref(false)
const withdrawAccountsLoaded = ref(false)
const withdrawAccountsError = ref(false)
const taskTier = ref(0)
const taskActivities = ref<TaskActivityItem[]>([])
const taskRecords = ref<TaskActivityItem[]>([])
const taskHistoryOpen = ref(false)
const taskHistoryLoading = ref(false)
const taskHistoryError = ref('')
const taskHistoryLoaded = ref(false)
const taskClaiming = ref(false)
const taskRewardClaiming = ref(false)
const taskCountdownNow = ref(Date.now())
let inviteQrRequestId = 0
let taskCountdownTimer: ReturnType<typeof window.setInterval> | undefined
let taskDeadlineRefreshKey = ''

const CUSTOM_TASK_MIN_INVITE_COUNT = 1
const CUSTOM_TASK_MAX_INVITE_COUNT = 1000
const MIN_COMMISSION_WITHDRAW_AMOUNT = 50

const currentLang = computed(() => String(locale.value || '').trim())
const activeTask = computed(() => taskActivities.value[taskTier.value] ?? taskActivities.value[0] ?? null)
const inviteCode = computed(() => String(stats.value?.inviteCode || '').trim())
const validUsersCount = computed(() => stats.value?.validUsers ?? 0)
const todayValidUsersCount = computed(() => stats.value?.todayValidUsers ?? stats.value?.todayInviteCount ?? 0)
const taskProgress = computed(() => {
  const value = Number(activeTask.value?.progressPercent ?? 0)
  if (!Number.isFinite(value)) return 0

  return Math.max(0, Math.min(100, Math.round(value)))
})
const activeTaskDeadlineMs = computed(() => taskDeadlineMs(activeTask.value))
const activeTaskRemainingSeconds = computed(() => {
  const deadline = activeTaskDeadlineMs.value
  if (!deadline) return 0

  return Math.max(0, Math.ceil((deadline - taskCountdownNow.value) / 1000))
})
const activeTaskRunning = computed(() => {
  const task = activeTask.value
  return Boolean(
    task?.claimed
    && Number(task.status ?? 0) === 0
    && activeTaskRemainingSeconds.value > 0,
  )
})
const activeTaskCountdownText = computed(() => formatTaskCountdown(activeTaskRemainingSeconds.value))

const inviteOrigin = computed(() => {
  const domain = String(ruleConfig.value?.bindDomain || '').trim().replace(/\/+$/, '')
  if (domain) return /^https?:\/\//i.test(domain) ? domain : `https://${domain}`

  return typeof window !== 'undefined' ? window.location.origin : ''
})
const inviteLink = computed(() => {
  if (!inviteCode.value || !inviteOrigin.value) return ''

  return `${inviteOrigin.value}/#/?c=${encodeURIComponent(inviteCode.value)}`
})

const availableCommissionCents = computed(() => toMoneyCents(stats.value?.availableCommission))
const transferAmountValue = computed(() => {
  const normalized = Number(transferAmount.value)

  return Number.isFinite(normalized) && normalized > 0 ? normalized : 0
})
const commissionWithdrawAmountValue = computed(() => {
  const normalized = Number(commissionWithdrawAmount.value)

  return Number.isFinite(normalized) && normalized > 0 ? normalized : 0
})
const commissionWithdrawAmountCents = computed(() => Math.round(commissionWithdrawAmountValue.value * 100))
const commissionWithdrawFeeCents = computed(() => {
  if (commissionWithdrawAmountCents.value <= 0) return 0

  const fixedFeeCents = toMoneyCents(ruleConfig.value?.rebateWithdrawFeeFixed)
  const percentFee = numberValue(ruleConfig.value?.rebateWithdrawFeePercent)
  const percentFeeCents = percentFee > 0
    ? Math.round(commissionWithdrawAmountCents.value * percentFee / 100)
    : 0

  return Math.max(0, fixedFeeCents + percentFeeCents)
})
const commissionWithdrawReceiveCents = computed(() =>
  Math.max(0, commissionWithdrawAmountCents.value - commissionWithdrawFeeCents.value),
)
const rebateWithdrawDisabled = computed(() =>
  Number(stats.value?.rebateWithdrawDisabled ?? userStore.userInfo?.rebateWithdrawDisabled ?? 0) === 1,
)
const selectedCommissionWithdrawAccount = computed(() =>
  withdrawAccounts.value.find(account => String(account.id) === commissionWithdrawAccountId.value) ?? null,
)
const commissionWithdrawAccountOptions = computed(() =>
  withdrawAccounts.value
    .filter(account => account.countryCode === APP_COUNTRY_CODE)
    .map(account => ({
      label: formatWithdrawAccountLabel(account),
      value: String(account.id),
    })),
)
const transferSubmitLock = useSubmitLock(submitTransferOrder, { minLockMs: 600 })
const commissionWithdrawSubmitLock = useSubmitLock(submitCommissionWithdrawOrder, { minLockMs: 600 })
const isTransferSubmitting = computed(() => transferSubmitLock.locked.value)
const isCommissionWithdrawSubmitting = computed(() => commissionWithdrawSubmitLock.locked.value)
const transferSubmitDisabled = computed(() =>
  isTransferSubmitting.value
  || transferAmountValue.value <= 0
  || transferAmountValue.value > availableCommissionCents.value / 100,
)
const commissionWithdrawSubmitDisabled = computed(() =>
  isCommissionWithdrawSubmitting.value
  || withdrawAccountsLoading.value
  || rebateWithdrawDisabled.value
  || commissionWithdrawAmountCents.value < MIN_COMMISSION_WITHDRAW_AMOUNT * 100
  || commissionWithdrawAmountCents.value > availableCommissionCents.value
  || !commissionWithdrawAccountId.value,
)

const teamV2Kpis = computed(() => [
  { id: 'invited', icon: 'fa-user-plus', tone: 'gold', value: formatCount(stats.value?.inviteCount), label: t('ppmx.team.statInviteUsers') },
  { id: 'valid', icon: 'fa-user-check', tone: 'green', value: formatCount(validUsersCount.value), label: t('ppmx.team.statValidUsers') },
  { id: 'today', icon: 'fa-bolt', tone: 'red', value: formatCount(todayValidUsersCount.value), label: t('ppmx.team.statTodayInviteUsers') },
  { id: 'recharge', icon: 'fa-wallet', tone: 'blue', value: formatCount(stats.value?.rechargeUsers), label: t('ppmx.team.statRechargeUsers') },
])

const commissionTiles = computed(() => [
  { id: 'total', value: formatAmount(stats.value?.totalCommission), label: t('ppmx.team.commissionTotal') },
  { id: 'today', value: formatAmount(stats.value?.todayCommission), label: t('ppmx.team.commissionToday') },
  { id: 'pending', value: formatCount(stats.value?.todayRechargeUsers), label: t('ppmx.team.statTodayRechargeUsers') },
])

const rechargeRebateRows = [
  { id: 'first', icon: 'fa-bolt', labelKey: 'rechargeRebateFirst', rate: 10, top: true },
] as const

const ruleItems = computed(() => [
  t('ppmx.team.ruleItem4'),
  t('ppmx.team.ruleItem5'),
  t('ppmx.team.ruleItem6'),
])

function numberValue(value: number | string | undefined | null) {
  const numeric = Number(value ?? 0)

  return Number.isFinite(numeric) ? numeric : 0
}

function formatCount(value: number | string | undefined | null) {
  return Math.trunc(numberValue(value)).toLocaleString('en-US')
}

function toMoneyCents(value: number | string | undefined | null) {
  if (value === undefined || value === null || value === '') return 0

  const normalized = Number(value)
  if (!Number.isFinite(normalized)) return 0

  try {
    return toCent(normalized.toFixed(2))
  } catch {
    return 0
  }
}

function formatAmount(value: number | string | undefined | null) {
  return formatMoney(toMoneyCents(value), { currency: APP_CURRENCY_SYMBOL })
}

function formatMarketMoney(cents: number | { value: number }) {
  const value = typeof cents === 'number' ? cents : cents.value

  return `${formatMoney(value, { currency: APP_CURRENCY_SYMBOL })} ${APP_CURRENCY}`
}

function parseWithdrawAccountData(raw: string) {
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!parsed || typeof parsed !== 'object') return {}

    return Object.fromEntries(
      Object.entries(parsed as Record<string, unknown>).map(([key, value]) => [key, String(value ?? '')]),
    )
  } catch {
    return {}
  }
}

function maskAccountValue(value: string) {
  const cleanValue = value.trim()
  if (!cleanValue) return ''

  const suffix = cleanValue.slice(-4)

  return suffix ? `**** ${suffix}` : cleanValue
}

function formatWithdrawAccountLabel(account: WithdrawAccountItem) {
  const data = parseWithdrawAccountData(account.accountData)
  const bankName = data.bank || data.bankName || data.Bank || data.bankCode || 'PE'
  const accountValue = data.accountNumber
    || data.routingNumber
    || data.accountNo
    || data.cardNo
    || Object.values(data).find(value => /\d{4,}/.test(value))
    || String(account.id)
  const masked = maskAccountValue(accountValue)

  return masked ? `${bankName} · ${masked}` : bankName
}

function normalizeTaskActivities(value: TaskActivityItem[] | undefined | null) {
  const now = Date.now()

  return Array.isArray(value)
    ? value.filter(item => Number(item?.configId) > 0 || Number(item?.recordId) > 0 || isCustomTask(item))
      .map(item => normalizeTaskDeadline(item, now))
    : []
}

function normalizeTaskDeadline(task: TaskActivityItem, now = Date.now()): TaskActivityItem {
  if (!task?.claimed || Number(task.status ?? 0) !== 0) return task
  if (taskDeadlineMs(task)) return task

  const remaining = Number(task.remainingSeconds ?? 0)
  if (!Number.isFinite(remaining) || remaining <= 0) return task

  return {
    ...task,
    deadlineAt: new Date(now + remaining * 1000).toISOString(),
  }
}

function isCustomTask(task: TaskActivityItem | null | undefined) {
  return String(task?.levelCode || '').trim() === 'custom'
}

function taskLevelText(task: TaskActivityItem | null | undefined) {
  const levelCode = String(task?.levelCode || '').trim()
  switch (levelCode) {
    case 'primary':
      return t('ppmx.team.taskLevelPrimary')
    case 'middle':
      return t('ppmx.team.taskLevelMiddle')
    case 'advanced':
      return t('ppmx.team.taskLevelAdvanced')
    case 'custom':
      return t('ppmx.team.taskLevelCustom')
    default:
      return t('ppmx.team.taskLevelDefault')
  }
}

function taskRewardText(task: TaskActivityItem | null | undefined) {
  if (!task) return formatAmount(0)

  const currency = String(task.rewardCurrency || '').trim()
  return currency ? `${formatAmount(task.rewardAmount)} ${currency}` : formatAmount(task.rewardAmount)
}

function formatTaskConfigAmount(value: number | string | undefined | null) {
  const normalized = numberValue(value)
  return normalized.toLocaleString('en-US', {
    minimumFractionDigits: Number.isInteger(normalized) ? 0 : 2,
    maximumFractionDigits: 2,
  })
}

function taskSubtitleText(task: TaskActivityItem | null | undefined) {
  const count = Math.max(0, Math.trunc(numberValue(task?.requiredInviteCount)))
  const amount = formatTaskConfigAmount(task?.requiredRechargeAmount)
  const currency = String(task?.rewardCurrency || APP_CURRENCY).trim() || APP_CURRENCY

  return t('ppmx.team.taskInviteRewardRequirement', {
    count: formatCount(count),
    amount,
    currency,
  })
}

function taskDeadlineMs(task: TaskActivityItem | null | undefined) {
  const deadline = String(task?.deadlineAt || '').trim()
  if (!deadline) return 0

  const value = new Date(deadline).getTime()
  return Number.isFinite(value) ? value : 0
}

function formatTaskCountdown(totalSeconds: number) {
  const safeSeconds = Math.max(0, Math.trunc(Number(totalSeconds) || 0))
  const hours = Math.floor(safeSeconds / 3600)
  const minutes = Math.floor((safeSeconds % 3600) / 60)
  const seconds = safeSeconds % 60

  return [hours, minutes, seconds].map(item => String(item).padStart(2, '0')).join(':')
}

function taskNeedText(task: TaskActivityItem | null | undefined) {
  const required = Math.max(0, Math.trunc(numberValue(task?.requiredInviteCount)))
  const progress = Math.max(0, Math.trunc(numberValue(task?.progressInviteCount)))
  const remaining = Math.max(0, required - progress)

  return t('ppmx.team.taskNeedFriends', { count: formatCount(remaining) })
}

function taskStatusText(task: TaskActivityItem | null | undefined) {
  if (!task) return t('ppmx.promo.claim')
  if (canClaimTask(task)) return t('ppmx.promo.claim')

  switch (Number(task.status ?? 0)) {
    case 1:
      return t('ppmx.team.taskCompleted')
    case 2:
      return t('ppmx.team.taskRewarded')
    case 3:
      return t('ppmx.team.taskExpired')
    case 4:
      return t('ppmx.team.taskCanceled')
    default:
      return `${formatCount(task.progressInviteCount)}/${formatCount(task.requiredInviteCount)}`
  }
}

function canClaimTask(task: TaskActivityItem | null | undefined) {
  if (!task) return false
  if (!task.claimed) return true

  const status = Number(task.status ?? 0)
  return status === 1 || status === 2 || status === 3 || status === 4
}

function customRewardPerUser(task: TaskActivityItem) {
  const count = Math.max(CUSTOM_TASK_MIN_INVITE_COUNT, Math.trunc(numberValue(task.requiredInviteCount)))
  const reward = numberValue(task.rewardAmount)
  return reward > 0 ? reward / count : 5
}

function adjustCustomInviteCount(delta: number) {
  const task = activeTask.value
  if (!task || !isCustomTask(task) || task.claimed) return

  const nextCount = Math.max(
    CUSTOM_TASK_MIN_INVITE_COUNT,
    Math.min(CUSTOM_TASK_MAX_INVITE_COUNT, Math.trunc(numberValue(task.requiredInviteCount)) + delta),
  )
  const rewardPerUser = customRewardPerUser(task)
  const updated: TaskActivityItem = {
    ...task,
    requiredInviteCount: nextCount,
    rewardAmount: Number((rewardPerUser * nextCount).toFixed(2)),
  }
  taskActivities.value = taskActivities.value.map((item, index) => index === taskTier.value ? updated : item)
}

function taskHistoryStatusText(task: TaskActivityItem) {
  switch (Number(task.status ?? 0)) {
    case 1:
      return t('ppmx.team.taskCompleted')
    case 2:
      return t('ppmx.team.taskRewarded')
    case 3:
      return t('ppmx.team.taskExpired')
    case 4:
      return t('ppmx.team.taskCanceled')
    default:
      return t('ppmx.team.taskInProgress')
  }
}

function formatDateText(value: string | null | undefined) {
  const raw = String(value || '').trim()
  if (!raw) return '--'

  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw

  return date.toLocaleString()
}

function requireRegisterAuth() {
  return pageChromeRef.value?.requireRegisterAuth() ?? false
}

async function claimTeamTask() {
  if (!requireRegisterAuth()) return

  const task = activeTask.value
  if (!task) {
    showToast(t('ppmx.team.taskNoActivity'))
    return
  }

  if (!canClaimTask(task)) {
    showToast(taskStatusText(task))
    return
  }

  taskClaiming.value = true

  try {
    const payload = isCustomTask(task)
      ? {
          levelCode: 'custom',
          inviteCount: Math.max(
            CUSTOM_TASK_MIN_INVITE_COUNT,
            Math.min(CUSTOM_TASK_MAX_INVITE_COUNT, Math.trunc(numberValue(task.requiredInviteCount))),
          ),
        }
      : {}
    const claimed = await claimTaskActivity(payload, currentLang.value)
    const normalizedClaimed = normalizeTaskDeadline(claimed)
    taskActivities.value = [normalizedClaimed]
    taskTier.value = 0
    taskDeadlineRefreshKey = ''
    taskHistoryLoaded.value = false
    showToast(t('ppmx.team.taskClaimSuccess'))
    void loadTeamData()
  } catch (error) {
    const code = error instanceof HttpError ? error.message : ''
    if (code === 'task_activity_has_active') {
      showToast(t('ppmx.team.taskActiveExists'))
      void loadTeamData()
      return
    }
    if (code === 'task_activity_already_claimed') {
      showToast(t('ppmx.team.taskAlreadyClaimed'))
      void loadTeamData()
      return
    }

    showToast(error instanceof Error && error.message ? error.message : t('ppmx.team.taskClaimFailed'))
  } finally {
    taskClaiming.value = false
  }
}

async function claimTeamTaskReward() {
  if (!requireRegisterAuth()) return

  const task = activeTask.value
  if (!task?.claimed || Number(task.status ?? 0) !== 0) {
    showToast(t('ppmx.team.taskNoActivity'))
    return
  }

  taskRewardClaiming.value = true

  try {
    const rewarded = await rewardTaskActivity(currentLang.value)
    const normalizedRewarded = normalizeTaskDeadline(rewarded)
    taskActivities.value = [normalizedRewarded]
    taskTier.value = 0
    taskDeadlineRefreshKey = ''
    taskHistoryLoaded.value = false
    showToast(t('ppmx.team.taskClaimSuccess'))
    void loadTeamData()
  } catch (error) {
    const code = error instanceof HttpError ? error.message : ''
    if (code === 'task_activity_not_reached') {
      showToast(t('ppmx.team.taskRewardNotReached'))
      return
    }
    if (code === 'task_activity_no_active') {
      void loadTeamData()
      showToast(t('ppmx.team.taskNoActivity'))
      return
    }

    showToast(error instanceof Error && error.message ? error.message : t('ppmx.team.taskClaimFailed'))
  } finally {
    taskRewardClaiming.value = false
  }
}

function openTaskHistory() {
  if (!requireRegisterAuth()) return

  taskHistoryOpen.value = true
  void loadTaskHistory()
}

function closeTaskHistory() {
  taskHistoryOpen.value = false
}

async function loadTaskHistory() {
  taskHistoryLoading.value = true
  taskHistoryError.value = ''

  try {
    const resp = await getTaskActivityRecords({ currentPage: 0, pageSize: 20 }, currentLang.value)
    taskRecords.value = Array.isArray(resp?.list) ? resp.list : []
    taskHistoryLoaded.value = true
  } catch (error) {
    taskHistoryError.value = error instanceof Error ? error.message : t('ppmx.team.taskHistoryLoadFailed')
  } finally {
    taskHistoryLoading.value = false
  }
}

function updateTaskCountdown() {
  taskCountdownNow.value = Date.now()

  const task = activeTask.value
  if (!task?.claimed || Number(task.status ?? 0) !== 0) return

  const deadline = taskDeadlineMs(task)
  if (!deadline || deadline > taskCountdownNow.value) return

  const refreshKey = `${task.recordId || task.configId || 'task'}:${deadline}`
  if (taskDeadlineRefreshKey === refreshKey) return

  taskDeadlineRefreshKey = refreshKey
  void loadTeamData()
}

function startTaskCountdownTimer() {
  if (taskCountdownTimer || typeof window === 'undefined') return

  updateTaskCountdown()
  taskCountdownTimer = window.setInterval(updateTaskCountdown, 1000)
}

function stopTaskCountdownTimer() {
  if (!taskCountdownTimer || typeof window === 'undefined') return

  window.clearInterval(taskCountdownTimer)
  taskCountdownTimer = undefined
}

function openTransferModal() {
  if (!requireRegisterAuth()) return

  transferModalOpen.value = true
}

function openCommissionWithdrawModal() {
  if (!requireRegisterAuth()) return
  if (rebateWithdrawDisabled.value) {
    showToast(t('ppmx.team.commissionWithdrawDisabledToast'))
    return
  }

  commissionWithdrawModalOpen.value = true
  if (!withdrawAccountsLoaded.value && !withdrawAccountsLoading.value) {
    loadCommissionWithdrawAccounts()
  }
}

function closeTransferModal() {
  transferModalOpen.value = false
}

function closeCommissionWithdrawModal() {
  commissionWithdrawModalOpen.value = false
}

function transferAllCommission() {
  const availableUnits = availableCommissionCents.value / 100
  transferAmount.value = availableUnits > 0 ? availableUnits.toFixed(2) : ''
}

function withdrawAllCommission() {
  const availableUnits = availableCommissionCents.value / 100
  commissionWithdrawAmount.value = availableUnits > 0 ? availableUnits.toFixed(2) : ''
}

function syncDefaultCommissionWithdrawAccount() {
  if (
    commissionWithdrawAccountId.value
    && commissionWithdrawAccountOptions.value.some(option => option.value === commissionWithdrawAccountId.value)
  ) {
    return
  }

  const defaultAccount = withdrawAccounts.value.find(account => account.countryCode === APP_COUNTRY_CODE && account.isDefault === 1)
    ?? withdrawAccounts.value.find(account => account.countryCode === APP_COUNTRY_CODE)

  commissionWithdrawAccountId.value = defaultAccount ? String(defaultAccount.id) : ''
}

async function loadCommissionWithdrawAccounts() {
  withdrawAccountsLoading.value = true
  withdrawAccountsError.value = false

  try {
    withdrawAccounts.value = await getWithdrawAccounts()
    withdrawAccountsLoaded.value = true
    syncDefaultCommissionWithdrawAccount()
  } catch {
    withdrawAccounts.value = []
    withdrawAccountsError.value = true
    withdrawAccountsLoaded.value = false
    commissionWithdrawAccountId.value = ''
  } finally {
    withdrawAccountsLoading.value = false
  }
}

function buildCommissionWithdrawFieldValues() {
  const account = selectedCommissionWithdrawAccount.value
  if (!account) return undefined

  const entries = Object.entries(parseWithdrawAccountData(account.accountData))
    .map(([key, value]) => [key, value.trim()] as const)
    .filter(([, value]) => value)

  return entries.length ? Object.fromEntries(entries) : undefined
}

function validateCommissionWithdrawSubmit() {
  if (rebateWithdrawDisabled.value) return t('ppmx.team.commissionWithdrawDisabledToast')
  if (commissionWithdrawAmountCents.value < MIN_COMMISSION_WITHDRAW_AMOUNT * 100) return t('ppmx.team.commissionWithdrawMinToast')
  if (commissionWithdrawAmountCents.value > availableCommissionCents.value) return t('ppmx.team.commissionWithdrawBalanceToast')
  if (withdrawAccountsError.value) return t('ppmx.team.commissionWithdrawAccountsError')
  if (!commissionWithdrawAccountId.value) return t('ppmx.team.commissionWithdrawAccountRequired')

  return ''
}

function getCommissionWithdrawErrorMessage(error: unknown) {
  const message = error instanceof Error ? error.message.trim() : ''

  if (!message) return t('ppmx.team.commissionWithdrawFailed')
  if (message === 'invalid_amount') return t('ppmx.team.commissionWithdrawAmountRequired')
  if (message === 'rebate_withdraw_disabled' || message === 'withdraw_disabled') return t('ppmx.team.commissionWithdrawDisabledToast')
  if (message === 'rebate_amount_insufficient' || message === 'user_balance_insufficient') return t('ppmx.team.commissionWithdrawBalanceToast')
  if (message === 'account_not_found' || message === 'account_country_mismatch') return t('ppmx.team.commissionWithdrawAccountsError')

  return message
}

async function submitTransferOrder() {
  if (availableCommissionCents.value <= 0) {
    showToast(t('ppmx.team.commissionTransferNone'))
    return
  }

  if (transferAmountValue.value <= 0) {
    showToast(t('ppmx.team.commissionTransferInvalid'))
    return
  }

  if (transferAmountValue.value > availableCommissionCents.value / 100 + 1e-9) {
    showToast(t('ppmx.team.commissionTransferExceed'))
    return
  }

  try {
    const resp = await transferRebateToBalance({ amount: Number(transferAmountValue.value.toFixed(2)) })

    if (stats.value) {
      stats.value = {
        ...stats.value,
        availableCommission: Math.max(0, Number(resp.rebateAmount ?? 0)),
      }
    }

    void userStore.loadUserInfo()
    closeTransferModal()
    showToast(t('ppmx.team.commissionTransferSuccess', { amount: formatAmount(resp.transferAmount) }))
  } catch (error) {
    const code = error instanceof HttpError ? error.message : ''
    if (code === 'invalid_amount') {
      showToast(t('ppmx.team.commissionTransferInvalid'))
      return
    }

    showToast(error instanceof Error && error.message ? error.message : t('ppmx.team.commissionTransferFailed'))
  }
}

function submitTransfer() {
  void transferSubmitLock.run().catch(() => {})
}

async function submitCommissionWithdrawOrder() {
  const message = validateCommissionWithdrawSubmit()
  if (message) {
    showToast(message)
    return
  }

  try {
    const order = await createRebateWithdrawOrder({
      amount: Number((commissionWithdrawAmountCents.value / 100).toFixed(2)),
      countryCode: APP_COUNTRY_CODE,
      accountId: Number(commissionWithdrawAccountId.value),
      fieldValues: buildCommissionWithdrawFieldValues(),
    })

    closeCommissionWithdrawModal()
    commissionWithdrawAmount.value = ''
    showToast(order.orderNo
      ? t('ppmx.team.commissionWithdrawSuccessWithOrder', { orderNo: order.orderNo })
      : t('ppmx.team.commissionWithdrawSuccess'),
    )

    await Promise.allSettled([
      loadTeamData(),
      userStore.loadUserInfo(),
    ])
  } catch (error) {
    showToast(getCommissionWithdrawErrorMessage(error))
    void loadTeamData()
  }
}

function submitCommissionWithdraw() {
  void commissionWithdrawSubmitLock.run().catch(() => {})
}

function openBank() {
  if (!requireRegisterAuth()) return

  closeCommissionWithdrawModal()
  void router.push('/bank')
}

async function copyInvite() {
  if (!requireRegisterAuth()) return
  if (!inviteLink.value) {
    showToast(t('ppmx.team.inviteUnavailable'))
    return
  }

  try {
    await navigator.clipboard.writeText(inviteLink.value)
    showToast(t('ppmx.team.copySuccess'))
  } catch {
    showToast(t('ppmx.team.copyFailure'))
  }
}

function shareInvite(type: 'telegram' | 'whatsapp') {
  if (!requireRegisterAuth()) return
  if (!inviteLink.value) {
    showToast(t('ppmx.team.inviteUnavailable'))
    return
  }

  const text = `${t('ppmx.team.shareText')} ${inviteLink.value}`
  const url = type === 'telegram'
    ? `https://t.me/share/url?url=${encodeURIComponent(inviteLink.value)}&text=${encodeURIComponent(t('ppmx.team.telegramText'))}`
    : `https://wa.me/?text=${encodeURIComponent(text)}`

  window.open(url, '_blank', 'noopener,noreferrer')
}

async function renderInviteQr(link: string) {
  const requestId = ++inviteQrRequestId
  inviteQrError.value = false

  if (!link) {
    inviteQrDataUrl.value = ''
    return
  }

  try {
    const dataUrl = await QRCode.toDataURL(link, {
      errorCorrectionLevel: 'H',
      margin: 1,
      width: 220,
      color: {
        dark: '#141416',
        light: '#ffffff',
      },
    })

    if (requestId === inviteQrRequestId) inviteQrDataUrl.value = dataUrl
  } catch {
    if (requestId === inviteQrRequestId) {
      inviteQrDataUrl.value = ''
      inviteQrError.value = true
    }
  }
}

async function loadTeamData() {
  if (!userStore.isLogin) {
    teamLoading.value = false
    teamLoaded.value = false
    teamError.value = ''
    return
  }

  teamLoading.value = true
  teamError.value = ''

  try {
    const [nextStats, nextRuleConfig, nextTaskActivities] = await Promise.all([
      getCurrentTgInviteStats(),
      getCurrentTgInviteRuleConfig(),
      getTaskActivityList(currentLang.value),
    ])

    stats.value = nextStats
    ruleConfig.value = nextRuleConfig
    taskActivities.value = normalizeTaskActivities(nextTaskActivities)
    if (taskTier.value >= taskActivities.value.length) {
      taskTier.value = Math.max(0, taskActivities.value.length - 1)
    }
    teamLoaded.value = true
  } catch (error) {
    teamLoaded.value = false
    teamError.value = error instanceof Error ? error.message : t('ppmx.team.loadFailed')
  } finally {
    teamLoading.value = false
  }
}

watch(inviteLink, link => {
  void renderInviteQr(link)
}, { immediate: true })

watch(
  () => userStore.isLogin,
  () => {
    void loadTeamData()
  },
)

watch(currentLang, () => {
  if (!userStore.isLogin) return

  void loadTeamData()
  if (taskHistoryOpen.value) void loadTaskHistory()
})

onMounted(() => {
  startTaskCountdownTimer()
  void loadTeamData()
})

onBeforeUnmount(() => {
  stopTaskCountdownTimer()
})
</script>

<template>
  <main id="page-team" class="ppmx-page ppmx-subpage ppmx-team-v2-page">
    <PpmxPageChrome ref="pageChromeRef">
      <section class="ppmx-subpage__narrow ppmx-team-v2-stack">
        <header class="ppmx-team-v2-hero reveal in">
          <span class="ppmx-eyebrow">{{ t('ppmx.team.eyebrow') }}</span>
          <h1 class="ppmx-display">{{ t('ppmx.team.title') }}</h1>
          <p>{{ t('ppmx.team.description') }}</p>
        </header>

        <PpmxGuestPanel
          v-if="!userStore.isLogin"
          id="teamV2Guest"
          icon="fa-user-group"
          :title="t('ppmx.team.authTitle')"
          :message="t('ppmx.team.authMessage')"
          :primary-text="t('ppmx.team.authAction')"
          @primary="requireRegisterAuth"
        />

        <AppState
          v-else-if="teamLoading"
          class="ppmx-team-state"
          status="loading"
          :title="t('ppmx.team.title')"
          skeleton-variant="team-page"
        />

        <AppState
          v-else-if="teamError"
          class="ppmx-team-state"
          status="error"
          :title="t('ppmx.team.loadFailed')"
          :message="teamError"
          :action-text="t('common.retry')"
          @retry="loadTeamData"
        />

        <template v-else-if="teamLoaded">
          <section class="ppmx-team-rebate ppmx-glass reveal in">
            <h2 class="ppmx-display">
              <i class="fa-solid fa-coins" />
              <span>{{ t('ppmx.team.rechargeRebateTitle') }}</span>
            </h2>
            <p>{{ t('ppmx.team.rechargeRebateDescription') }}</p>

            <div class="ppmx-team-rebate__table">
              <div class="ppmx-team-rebate__rows">
                <article
                  v-for="row in rechargeRebateRows"
                  :key="row.id"
                  class="ppmx-team-rebate__row"
                  :class="{ 'is-top': row.top }"
                >
                  <span class="ppmx-team-rebate__range">
                    <i class="fa-solid" :class="row.icon" />
                    <span>{{ t(`ppmx.team.${row.labelKey}`) }}</span>
                  </span>
                  <span class="ppmx-team-rebate__value">
                    <strong class="ppmx-team-rebate__amount">
                      {{ row.rate }}<em>%</em>
                    </strong>
                    <em class="ppmx-team-rebate__example">
                      {{ t('ppmx.team.rechargeRebateExample', { rate: row.rate }) }}
                    </em>
                  </span>
                </article>
              </div>
            </div>
          </section>

          <section id="teamV2Commission" class="ppmx-team-v2-commission reveal in">
            <div class="ppmx-team-v2-commission__glow" />
            <div class="ppmx-team-v2-commission__body">
              <p class="ppmx-eyebrow">{{ t('ppmx.team.commissionEyebrow') }}</p>
              <div class="ppmx-team-v2-commission__main">
                <div>
                  <span>{{ t('ppmx.team.commissionAvailable') }}</span>
                  <strong>{{ formatAmount(stats?.availableCommission) }} <em>{{ APP_CURRENCY }}</em></strong>
                </div>
                <div class="ppmx-team-v2-commission__actions">
                  <button type="button" @click="openTransferModal">
                    <i class="fa-solid fa-right-left" />
                    <span>{{ t('ppmx.team.commissionTransfer') }}</span>
                  </button>
                  <button class="is-gold" type="button" @click="openCommissionWithdrawModal">
                    <i class="fa-solid fa-money-bill-transfer" />
                    <span>{{ t('ppmx.team.commissionWithdraw') }}</span>
                  </button>
                </div>
              </div>

              <div class="ppmx-team-v2-commission__tiles">
                <article v-for="tile in commissionTiles" :key="tile.id">
                  <strong>{{ tile.value }}</strong>
                  <span>{{ tile.label }}</span>
                </article>
              </div>
            </div>
          </section>

          <section id="teamV2Kpis" class="ppmx-team-v2-kpis reveal in" :aria-label="t('ppmx.team.kpiAria')">
            <article
              v-for="kpi in teamV2Kpis"
              :key="kpi.id"
              :class="`is-${kpi.tone}`"
            >
              <i class="fa-solid" :class="kpi.icon" />
              <strong>{{ kpi.value }}</strong>
              <span>{{ kpi.label }}</span>
            </article>
          </section>

          <section id="teamV2Invite" class="ppmx-team-v2-invite reveal in">
            <div>
              <h2 class="ppmx-display">{{ t('ppmx.team.inviteTitle') }}</h2>
              <label for="teamV2InviteLink">{{ t('ppmx.team.inviteLinkLabel') }}</label>
              <div class="ppmx-team-v2-invite__link">
                <input
                  id="teamV2InviteLink"
                  :value="inviteLink"
                  :placeholder="t('ppmx.team.inviteUnavailable')"
                  readonly
                >
                <button type="button" :disabled="!inviteLink" @click="copyInvite">
                  <i class="fa-solid fa-copy" />
                  <span>{{ t('ppmx.team.copy') }}</span>
                </button>
              </div>
              <div class="ppmx-team-v2-share">
                <button type="button" @click="shareInvite('whatsapp')">
                  <i class="fa-brands fa-whatsapp" />
                  <span>WhatsApp</span>
                </button>
                <button type="button" @click="shareInvite('telegram')">
                  <i class="fa-brands fa-telegram" />
                  <span>Telegram</span>
                </button>
                <button type="button" @click="copyInvite">
                  <i class="fa-solid fa-share-nodes" />
                  <span>{{ t('ppmx.team.share') }}</span>
                </button>
              </div>
            </div>

            <aside class="ppmx-team-v2-qr" :aria-label="t('ppmx.team.qrAria')">
              <div :class="{ 'is-disabled': !inviteQrDataUrl }">
                <img v-if="inviteQrDataUrl" :src="inviteQrDataUrl" :alt="t('ppmx.team.qrAria')">
                <i v-else class="fa-solid fa-qrcode" />
              </div>
              <p>{{ inviteQrError ? t('ppmx.team.qrLoadFailed') : t('ppmx.team.qrText') }}</p>
            </aside>
          </section>

          <details id="teamV2Rules" class="ppmx-team-v2-rules reveal in">
            <summary>
              <span>
                <i class="fa-solid fa-circle-info" />
                {{ t('ppmx.team.rulesTitle') }}
              </span>
              <i class="fa-solid fa-chevron-down" />
            </summary>
            <ul>
              <li v-for="item in ruleItems" :key="item">{{ item }}</li>
            </ul>
          </details>
        </template>
      </section>
    </PpmxPageChrome>

    <PpmxWithdrawModal
      id="teamV2TaskHistoryModal"
      v-model="taskHistoryOpen"
      :eyebrow="t('ppmx.team.rebateTitle')"
      title-id="teamV2TaskHistoryModalTitle"
      :title="t('ppmx.team.taskHistoryTitle')"
      :close-aria-label="t('ppmx.team.taskHistoryClose')"
      @close="closeTaskHistory"
    >
      <AppState
        v-if="taskHistoryLoading"
        class="ppmx-team-task-history-state"
        status="loading"
        :title="t('ppmx.team.taskHistoryTitle')"
      />
      <AppState
        v-else-if="taskHistoryError"
        class="ppmx-team-task-history-state"
        status="error"
        :title="t('ppmx.team.taskHistoryLoadFailed')"
        :message="taskHistoryError"
        :action-text="t('common.retry')"
        @retry="loadTaskHistory"
      />
      <AppState
        v-else-if="taskHistoryLoaded && !taskRecords.length"
        class="ppmx-team-task-history-state"
        status="empty"
        :title="t('ppmx.team.taskHistoryEmpty')"
      />
      <div v-else class="ppmx-team-task-history-list">
        <article
          v-for="record in taskRecords"
          :key="`${record.configId}-${record.recordId || 0}`"
          class="ppmx-team-task-history-item"
        >
          <div>
            <strong>{{ record.title || taskLevelText(record) }}</strong>
            <span>{{ taskHistoryStatusText(record) }}</span>
          </div>
          <p v-if="record.subTitle">{{ record.subTitle }}</p>
          <dl>
            <div>
              <dt>{{ t('ppmx.team.taskProgressLabel') }}</dt>
              <dd>{{ formatCount(record.progressInviteCount) }}/{{ formatCount(record.requiredInviteCount) }}</dd>
            </div>
            <div>
              <dt>{{ t('ppmx.team.taskReward') }}</dt>
              <dd>{{ taskRewardText(record) }}</dd>
            </div>
            <div>
              <dt>{{ t('ppmx.team.taskDeadline') }}</dt>
              <dd>{{ formatDateText(record.deadlineAt) }}</dd>
            </div>
          </dl>
        </article>
      </div>
    </PpmxWithdrawModal>

    <PpmxWithdrawModal
      id="teamV2CommissionWithdrawModal"
      v-model="commissionWithdrawModalOpen"
      :eyebrow="t('ppmx.team.commissionEyebrow')"
      title-id="teamV2CommissionWithdrawModalTitle"
      :title="t('ppmx.team.commissionWithdrawTitle')"
      :close-aria-label="t('ppmx.team.commissionWithdrawClose')"
      @close="closeCommissionWithdrawModal"
    >
      <form class="ppmx-team-transfer-form" @submit.prevent="submitCommissionWithdraw">
        <p class="ppmx-team-transfer-sub">{{ t('ppmx.team.commissionWithdrawReviewDesc') }}</p>

        <label class="ppmx-withdraw-field">
          <span>{{ t('ppmx.team.commissionWithdrawAvailable') }}</span>
          <strong>{{ formatAmount(stats?.availableCommission) }} <em>{{ APP_CURRENCY }}</em></strong>
        </label>

        <label class="ppmx-withdraw-field">
          <span>{{ t('ppmx.team.commissionWithdrawAmountLabel') }}</span>
          <div class="ppmx-team-transfer-row">
            <input
              v-model="commissionWithdrawAmount"
              inputmode="decimal"
              min="50"
              step="0.01"
              type="number"
              :placeholder="t('ppmx.team.commissionWithdrawAmountPlaceholder')"
            >
            <button type="button" class="ppmx-team-transfer-all" @click="withdrawAllCommission">
              {{ t('ppmx.team.commissionTransferAll') }}
            </button>
          </div>
        </label>

        <div class="wd-fee-card">
          <div class="wd-fee-row">
            <span class="wd-fee-k">{{ t('ppmx.team.commissionWithdrawFee') }}</span>
            <span class="wd-fee-v">{{ formatMarketMoney(commissionWithdrawFeeCents) }}</span>
          </div>
          <div class="wd-fee-row">
            <span class="wd-fee-k">{{ t('ppmx.team.commissionWithdrawReceive') }}</span>
            <span class="wd-fee-v--gold">{{ formatMarketMoney(commissionWithdrawReceiveCents) }}</span>
          </div>
          <div class="wd-fee-row">
            <span class="wd-fee-k is-strong">{{ t('ppmx.team.commissionWithdrawDeduct') }}</span>
            <span class="wd-fee-v--gold">{{ formatMarketMoney(commissionWithdrawAmountCents) }}</span>
          </div>
        </div>

        <label class="ppmx-withdraw-field">
          <span>{{ t('ppmx.team.commissionWithdrawAccount') }}</span>
          <p v-if="withdrawAccountsLoading && !commissionWithdrawAccountOptions.length" class="ppmx-team-transfer-sub">
            {{ t('ppmx.team.commissionWithdrawAccountsLoading') }}
          </p>
          <PpmxSelectInput
            v-else
            v-model="commissionWithdrawAccountId"
            :options="commissionWithdrawAccountOptions"
            :placeholder="t('ppmx.team.commissionWithdrawAccountPlaceholder')"
            :title="t('ppmx.team.commissionWithdrawAccount')"
            :disabled="withdrawAccountsLoading || !commissionWithdrawAccountOptions.length"
          />
        </label>

        <div
          v-if="withdrawAccountsError || (!withdrawAccountsLoading && !commissionWithdrawAccountOptions.length)"
          class="ppmx-withdraw-account-note"
          :class="{ 'is-error': withdrawAccountsError }"
        >
          <i class="fa-solid" :class="withdrawAccountsError ? 'fa-triangle-exclamation' : 'fa-link'" />
          <span>
            {{ t(withdrawAccountsError ? 'ppmx.team.commissionWithdrawAccountsError' : 'ppmx.team.commissionWithdrawAccountNote') }}
          </span>
          <button type="button" @click="openBank">
            {{ t('ppmx.team.commissionWithdrawGoBank') }}
          </button>
        </div>

        <div class="wd-rules">
          <h3>{{ t('ppmx.team.commissionWithdrawRulesTitle') }}</h3>
          <ul>
            <li>{{ t('ppmx.team.commissionWithdrawRule1') }}</li>
            <li>{{ t('ppmx.team.commissionWithdrawRule2') }}</li>
            <li>{{ t('ppmx.team.commissionWithdrawRule3') }}</li>
          </ul>
        </div>

        <PpmxButton
          block
          size="lg"
          type="submit"
          variant="gold"
          :disabled="commissionWithdrawSubmitDisabled"
          :loading="isCommissionWithdrawSubmitting"
        >
          {{ isCommissionWithdrawSubmitting
            ? t('ppmx.team.commissionWithdrawSubmitting')
            : t('ppmx.team.commissionWithdrawConfirm') }}
        </PpmxButton>
      </form>
    </PpmxWithdrawModal>

    <PpmxWithdrawModal
      id="teamV2TransferModal"
      v-model="transferModalOpen"
      :eyebrow="t('ppmx.team.commissionEyebrow')"
      title-id="teamV2TransferModalTitle"
      :title="t('ppmx.team.commissionTransferTitle')"
      :close-aria-label="t('ppmx.team.commissionTransferClose')"
      @close="closeTransferModal"
    >
      <form class="ppmx-team-transfer-form" @submit.prevent="submitTransfer">
        <p class="ppmx-team-transfer-sub">{{ t('ppmx.team.commissionTransferSub') }}</p>
        <label class="ppmx-withdraw-field">
          <span>{{ t('ppmx.team.commissionTransferAvailable') }}</span>
          <strong>{{ formatAmount(stats?.availableCommission) }} <em>{{ APP_CURRENCY }}</em></strong>
        </label>
        <label class="ppmx-withdraw-field">
          <span>{{ t('ppmx.team.commissionTransferAmountLabel') }}</span>
          <div class="ppmx-team-transfer-row">
            <input
              v-model="transferAmount"
              inputmode="decimal"
              min="0"
              step="0.01"
              type="number"
              :placeholder="t('ppmx.team.commissionTransferPlaceholder')"
            >
            <button type="button" class="ppmx-team-transfer-all" @click="transferAllCommission">
              {{ t('ppmx.team.commissionTransferAll') }}
            </button>
          </div>
        </label>
        <p class="ppmx-team-transfer-sub">
          <i class="fa-solid fa-circle-info" />
          <span>{{ t('ppmx.team.commissionTransferWarn') }}</span>
        </p>
        <PpmxButton
          block
          size="lg"
          type="submit"
          variant="gold"
          :disabled="transferSubmitDisabled"
          :loading="isTransferSubmitting"
        >
          {{ isTransferSubmitting
            ? t('ppmx.team.commissionTransferSubmitting')
            : t('ppmx.team.commissionTransferConfirm') }}
        </PpmxButton>
      </form>
    </PpmxWithdrawModal>
  </main>
</template>
