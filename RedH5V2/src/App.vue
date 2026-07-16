<script setup lang="ts">
import type { RechargeSuccessNotification } from '@/api/user'
import { ackRechargeNotification, getPendingRechargeNotifications } from '@/api/user'
import { APP_CURRENCY, APP_CURRENCY_SYMBOL } from '@/config/market'
import { runtimeState, checkForceUpdate } from '@/utils/runtime'
import { canInstallPwa, promptInstallPwa } from '@/pwa'
import { resolveAppShell } from '@/utils/appShell'
import { createThirdPartyEventId, trackAttributionEvent } from '@/utils/attribution'
import { trackPurchase } from '@/utils/facebookPixel'
import { formatMoney, toDisplayCents } from '@/utils/money'
import { incrementLocalRechargeCount } from '@/utils/rechargeCount'
import { getTtlStorage, setTtlStorage } from '@/utils/storage'
import { trackEvent } from '@/utils/tracker'
import wsClient, { closeWebSocket, connectWebSocket } from '@/plugins/websocket'
import AppShell from '@/components/AppShell.vue'
import PpmxHomeShell from '@/components/PpmxHomeShell.vue'
import PpmxPaymentSuccessNotice from '@/components/PpmxPaymentSuccessNotice.vue'

const keepAliveNames = ['home', 'mine']
const userStore = useUserStore()
const prizePoolStore = usePrizePoolStore()
const { t } = useI18n()
const wsInitialized = ref(false)
const syncingRechargeNotifications = ref(false)
const processedRechargeOrders = new Set<string>()
const rechargeSuccessNotice = reactive({
  show: false,
  title: '',
  message: '',
  amountText: '',
})
const INSTALL_PWA_DISMISSED_KEY = 'ppmx_install_pwa_dismissed'
const installPwaDismissed = ref(false)
let rechargeNoticeTimer: number | null = null
const processedOrderTimers = new Map<string, number>()
let offRechargeSuccessListener: (() => void) | null = null
let offPrizePoolBalanceListener: (() => void) | null = null
let offOpenListener: (() => void) | null = null
let offSyncExpiredListener: (() => void) | null = null

const runtimeStatus = computed(() => {
  if (runtimeState.config.maintenance) return 'maintenance'
  if (checkForceUpdate(runtimeState.config)) return 'fallback'
  return 'ready'
})

const runtimeMessage = computed(() => {
  if (runtimeState.config.maintenance) {
    return runtimeState.config.maintenanceMessage || '系统正在维护，请稍后再试'
  }

  if (checkForceUpdate(runtimeState.config)) {
    return runtimeState.config.forceUpdateMessage || '当前版本过旧，请刷新或重新打开应用'
  }

  return ''
})

const showInstallPwa = computed(() => canInstallPwa.value && !installPwaDismissed.value)

function getMsUntilTomorrow() {
  const now = new Date()
  const tomorrow = new Date(now)
  tomorrow.setHours(24, 0, 0, 0)
  return Math.max(1, tomorrow.getTime() - now.getTime())
}

function reloadApp() {
  window.location.reload()
}

async function handleInstallPwa() {
  await promptInstallPwa()
}

function dismissInstallPwa() {
  installPwaDismissed.value = true

  setTtlStorage(INSTALL_PWA_DISMISSED_KEY, true, getMsUntilTomorrow())

  void trackEvent({
    eventName: 'pwa_install_prompt_dismissed',
    eventType: 'pwa',
  })
}

function normalizeRechargeSuccessMessage(message: unknown): RechargeSuccessNotification | null {
  const record = message && typeof message === 'object'
    ? message as Record<string, unknown>
    : {}
  const data = record.data && typeof record.data === 'object'
    ? record.data as Record<string, unknown>
    : record
  const orderNo = String(data.orderNo || '').trim()

  if (!orderNo) return null

  return {
    orderNo,
    channel: String(data.channel || ''),
    currency: String(data.currency || APP_CURRENCY),
    amount: Number(data.amount || 0),
    creditAmount: data.creditAmount == null ? null : Number(data.creditAmount),
    bonusAmount: data.bonusAmount == null ? null : Number(data.bonusAmount),
    status: Number(data.status || 0),
    isFirstRecharge: Boolean(data.isFirstRecharge),
    payTime: typeof data.payTime === 'string' ? data.payTime : null,
    frontendNotifyStatus: Number(data.frontendNotifyStatus || 0),
    frontendNotifyCount: Number(data.frontendNotifyCount || 0),
    frontendNotifyAt: typeof data.frontendNotifyAt === 'string' ? data.frontendNotifyAt : null,
    frontendNotifyAckAt: typeof data.frontendNotifyAckAt === 'string' ? data.frontendNotifyAckAt : null,
  }
}

function rememberProcessedOrder(orderNo: string) {
  processedRechargeOrders.add(orderNo)

  const existingTimer = processedOrderTimers.get(orderNo)
  if (existingTimer) clearTimeout(existingTimer)

  const timer = window.setTimeout(() => {
    processedRechargeOrders.delete(orderNo)
    processedOrderTimers.delete(orderNo)
  }, 10 * 60 * 1000)
  processedOrderTimers.set(orderNo, timer)
}

function showRechargeSuccessNotice(notification: RechargeSuccessNotification) {
  const amount = Number(notification.creditAmount || notification.amount || 0)

  rechargeSuccessNotice.title = t('recharge.paymentSuccessTitle')
  rechargeSuccessNotice.message = t('recharge.paymentSuccessMessage', {
    orderNo: notification.orderNo,
  })
  rechargeSuccessNotice.amountText = amount > 0
    ? t('recharge.paymentSuccessAmount', {
      amount: formatMoney(toDisplayCents(amount), { currency: APP_CURRENCY_SYMBOL }),
      currency: notification.currency || APP_CURRENCY,
    })
    : ''
  rechargeSuccessNotice.show = true

  if (rechargeNoticeTimer) clearTimeout(rechargeNoticeTimer)
  rechargeNoticeTimer = window.setTimeout(() => {
    rechargeSuccessNotice.show = false
  }, 3600)
}

async function ackRechargeSuccess(orderNo: string) {
  try {
    await ackRechargeNotification(orderNo)
  } catch (error) {
    console.warn('[recharge ws] ack failed:', orderNo, error)
    void trackEvent({
      eventName: 'recharge_notification_ack_failed',
      eventType: 'api_error',
      userId: userStore.userInfo?.id,
      payload: {
        orderNo,
        message: error instanceof Error ? error.message : String(error),
      },
    })
  }
}

async function trackRechargeSuccess(notification: RechargeSuccessNotification) {
  const orderNo = String(notification.orderNo || '').trim()
  const amount = Number(notification.amount || 0)

  if (!orderNo) return ''

  const eventId = createThirdPartyEventId('recharge_success')
  const metadata = {
    orderNo,
    amount,
    creditAmount: Number(notification.creditAmount || 0),
    bonusAmount: Number(notification.bonusAmount || 0),
    currency: notification.currency || APP_CURRENCY,
    channel: notification.channel || '',
    status: notification.status,
    isFirstRecharge: Boolean(notification.isFirstRecharge),
    payTime: notification.payTime || '',
  }

  void trackEvent({
    eventName: 'recharge_success',
    eventType: 'transaction',
    userId: userStore.userInfo?.id,
    payload: metadata,
  })
  void trackAttributionEvent({
    eventName: 'recharge_success',
    eventId,
    metadata,
  })

  if (amount > 0) {
    trackPurchase({
      orderNo,
      amount,
      currency: notification.currency || APP_CURRENCY,
      eventId,
    })
  }

  return eventId
}

async function handleRechargeSuccess(notification: RechargeSuccessNotification, showNotice = true) {
  const orderNo = String(notification.orderNo || '').trim()
  if (!orderNo) {
    await userStore.loadUserInfo()
    return
  }

  if (processedRechargeOrders.has(orderNo)) return
  rememberProcessedOrder(orderNo)
  incrementLocalRechargeCount()

  await trackRechargeSuccess(notification)
  if (showNotice) showRechargeSuccessNotice(notification)

  await Promise.allSettled([
    ackRechargeSuccess(orderNo),
    userStore.loadUserInfo(),
  ])
}

function handleRechargeSuccessMessage(message: unknown) {
  const notification = normalizeRechargeSuccessMessage(message)

  if (!notification) {
    void userStore.loadUserInfo()
    return
  }

  void handleRechargeSuccess(notification)
}

function handlePrizePoolBalanceMessage(message: unknown) {
  const record = message && typeof message === 'object'
    ? message as Record<string, unknown>
    : {}
  const data = record.data && typeof record.data === 'object'
    ? record.data as Record<string, unknown>
    : record
  const poolCode = String(data.poolCode || '').trim()
  const balance = Number(data.balance)

  if (!poolCode || !Number.isFinite(balance)) return

  prizePoolStore.updatePrizePoolBalance({
    poolCode,
    balance,
  })
}

async function syncPendingRechargeNotifications() {
  if (!userStore.token || syncingRechargeNotifications.value) return

  try {
    syncingRechargeNotifications.value = true
    const notifications = await getPendingRechargeNotifications()

    for (const notification of notifications) {
      await handleRechargeSuccess(notification)
    }
  } catch (error) {
    console.warn('[recharge ws] sync pending failed:', error)
    void trackEvent({
      eventName: 'recharge_notification_sync_failed',
      eventType: 'api_error',
      userId: userStore.userInfo?.id,
      payload: {
        message: error instanceof Error ? error.message : String(error),
      },
    })
  } finally {
    syncingRechargeNotifications.value = false
  }
}

async function refreshCurrentUserInfo() {
  if (!userStore.token) return

  try {
    await userStore.loadUserInfo()
  } catch (error) {
    console.warn('[auth] load current user failed:', error)
  }
}

function initWebSocket() {
  if (wsInitialized.value || !userStore.token) return

  connectWebSocket()
  wsInitialized.value = true
}

function shutdownWebSocket() {
  if (!wsInitialized.value) return

  closeWebSocket()
  wsInitialized.value = false
}

onMounted(() => {
  installPwaDismissed.value = getTtlStorage<boolean>(INSTALL_PWA_DISMISSED_KEY) === true

  offRechargeSuccessListener = wsClient.on('recharge_success', handleRechargeSuccessMessage)
  offPrizePoolBalanceListener = wsClient.on('prize_pool_balance', handlePrizePoolBalanceMessage)
  offOpenListener = wsClient.onOpen(() => {
    void syncPendingRechargeNotifications()
  })
  offSyncExpiredListener = wsClient.onSyncExpired(() => {
    void syncPendingRechargeNotifications()
  })

  if (userStore.token) {
    void refreshCurrentUserInfo()
    initWebSocket()
    void syncPendingRechargeNotifications()
  }
})

watch(() => userStore.token, (token, oldToken) => {
  if (token && !oldToken) {
    initWebSocket()
    void syncPendingRechargeNotifications()
    return
  }

  if (!token && oldToken) {
    shutdownWebSocket()
  }
})

onUnmounted(() => {
  offRechargeSuccessListener?.()
  offPrizePoolBalanceListener?.()
  offOpenListener?.()
  offSyncExpiredListener?.()
  if (rechargeNoticeTimer) clearTimeout(rechargeNoticeTimer)
  processedOrderTimers.forEach(timer => clearTimeout(timer))
  processedOrderTimers.clear()
  shutdownWebSocket()
})
</script>

<template>
  <AppState
    v-if="runtimeStatus !== 'ready'"
    :status="runtimeStatus"
    fullscreen
    :message="runtimeMessage"
    action-text="刷新"
    @retry="reloadApp"
  />

  <RouterView v-else v-slot="{ Component, route }">
    <AppShell v-if="resolveAppShell(route.meta) === 'default'">
      <KeepAlive :include="keepAliveNames">
        <component
          :is="Component"
          v-if="route.meta.keepAlive"
          :key="String(route.name)"
        />
      </KeepAlive>

      <component
        :is="Component"
        v-if="!route.meta.keepAlive"
        :key="route.fullPath"
      />
    </AppShell>

    <PpmxHomeShell v-else-if="resolveAppShell(route.meta) === 'ppmx'">
      <KeepAlive :include="keepAliveNames">
        <component
          :is="Component"
          v-if="route.meta.keepAlive"
          :key="String(route.name)"
        />
      </KeepAlive>

      <component
        :is="Component"
        v-if="!route.meta.keepAlive"
        :key="route.fullPath"
      />
    </PpmxHomeShell>

    <template v-else>
      <KeepAlive :include="keepAliveNames">
        <component
          :is="Component"
          v-if="route.meta.keepAlive"
          :key="String(route.name)"
        />
      </KeepAlive>

      <component
        :is="Component"
        v-if="!route.meta.keepAlive"
        :key="route.fullPath"
      />
    </template>
  </RouterView>

  <aside
    v-if="showInstallPwa"
    class="install-pwa-button"
    role="status"
    :aria-label="t('pwa.installTitle')"
  >
    <div
      class="install-pwa-button__main"
    >
      <button
        class="install-pwa-button__install"
        type="button"
        :aria-label="t('pwa.installTitle')"
        @click="handleInstallPwa"
      >
        <span class="install-pwa-button__icon" aria-hidden="true">
          <i class="fa-solid fa-mobile-screen-button" />
        </span>
        <span>{{ t('pwa.installApp') }}</span>
      </button>
      <button
        class="install-pwa-button__close"
        type="button"
        :aria-label="t('pwa.closeInstallPrompt')"
        @click="dismissInstallPwa"
      >
        <i class="fa-solid fa-xmark" aria-hidden="true" />
      </button>
    </div>
  </aside>

  <PpmxPaymentSuccessNotice
    :show="rechargeSuccessNotice.show"
    :title="rechargeSuccessNotice.title"
    :message="rechargeSuccessNotice.message"
    :amount-text="rechargeSuccessNotice.amountText"
  />
</template>
