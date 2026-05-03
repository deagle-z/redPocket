<script setup lang="ts">
import { useLocalStorage } from '@vueuse/core'
import { useRouteCacheStore } from '@/stores'
import { STORAGE_TOKEN_KEY } from '@/stores/mutation-type'
import type { RechargeSuccessNotification } from '@/api/user'
import { ackRechargeNotification, getPendingRechargeNotifications } from '@/api/user'
import wsClient, { connectWebSocket } from '@/plugins/websocket'
import { trackFirstRechargePurchase } from '@/utils/facebook-pixel'
import { showToast } from 'vant'
import AppTopHeader from '@/components/AppTopHeader.vue'
import TabBar from '@/components/TabBar.vue'
import AppConfirmDialog from '@/components/AppConfirmDialog.vue'
import { usePwaInstall } from '@/composables/usePwaInstall'

const { t } = useI18n()
const { showDialog: showPwaDialog, triggerInstall, dismiss: dismissPwa } = usePwaInstall()

const routeCacheStore = useRouteCacheStore()
const accessToken = useLocalStorage<string | null>(STORAGE_TOKEN_KEY, '')
const wsInitialized = ref(false)
const syncingRechargeNotify = ref(false)

const keepAliveRouteNames = computed(() => {
  return routeCacheStore.routeCaches
})

// Thai Red & Gold — Vant component theme overrides
const vantThemeVars = {
  colorPrimary: '#c0392b',
  colorSuccess: '#c0392b',
  buttonPrimaryBackground: 'linear-gradient(135deg, #a93226, #c0392b)',
  buttonPrimaryBorderColor: '#a93226',
  buttonPrimaryColor: '#ffffff',
  tabbarItemActiveColor: '#c0392b',
  tabbarItemTextColor: '#4a3030',
  navBarIconColor: '#c0392b',
  navBarTitleTextColor: '#2c0a07',
  switchOnBackground: '#c0392b',
  sliderActiveBackground: '#c0392b',
  checkboxCheckedIconColor: '#c0392b',
  radioCheckedIconColor: '#c0392b',
  fieldLabelColor: '#5c2d1e',
}

function initWebSocket() {
  if (wsInitialized.value)
    return
  connectWebSocket()
  wsInitialized.value = true
}

function closeWebSocket() {
  if (!wsInitialized.value)
    return
  wsClient.close()
  wsInitialized.value = false
}

async function ackRechargeSuccess(orderNo: string, showSuccessToast = true) {
  const normalizedOrderNo = String(orderNo || '').trim()
  if (!normalizedOrderNo)
    return
  try {
    await ackRechargeNotification(normalizedOrderNo)
    if (showSuccessToast)
      showToast(t('rechargePage.orderRechargeSuccess', { orderNo: normalizedOrderNo }))
  }
  catch (error) {
    console.warn('[recharge ws] ack failed:', normalizedOrderNo, error)
  }
}

function trackRechargeSuccessPixel(item: RechargeSuccessNotification) {
  if (!item?.isFirstRecharge)
    return
  trackFirstRechargePurchase({
    orderNo: item.orderNo,
    amount: Number(item.amount || 0),
    currency: item.currency || 'BRL',
  })
}

function handleRechargeSuccessMessage(message: any) {
  const data = message?.data || message || {}
  trackRechargeSuccessPixel(data)
  void ackRechargeSuccess(data.orderNo)
}

async function syncPendingRechargeNotifications() {
  if (!accessToken.value || syncingRechargeNotify.value)
    return
  try {
    syncingRechargeNotify.value = true
    const { data } = await getPendingRechargeNotifications()
    for (const item of data || []) {
      trackRechargeSuccessPixel(item)
      await ackRechargeSuccess(item.orderNo, false)
    }
  }
  catch (error) {
    console.warn('[recharge ws] sync pending failed:', error)
  }
  finally {
    syncingRechargeNotify.value = false
  }
}

onMounted(() => {
  wsClient.on('recharge_success', handleRechargeSuccessMessage)
  wsClient.onOpen(() => {
    void syncPendingRechargeNotifications()
  })
  if (accessToken.value)
    initWebSocket()
  void syncPendingRechargeNotifications()
})

watch(accessToken, (token, oldToken) => {
  const hasToken = !!token
  const hadToken = !!oldToken
  if (hasToken && !hadToken) {
    initWebSocket()
    void syncPendingRechargeNotifications()
    return
  }
  if (!hasToken && hadToken)
    closeWebSocket()
})
</script>

<template>
  <van-config-provider :theme-vars="vantThemeVars">
    <router-view v-slot="{ Component }">
      <section class="app-wrapper">
        <AppTopHeader />
        <keep-alive :include="keepAliveRouteNames">
          <component :is="Component" />
        </keep-alive>
        <TabBar />
      </section>
    </router-view>
  </van-config-provider>

  <AppConfirmDialog
    v-model:show="showPwaDialog"
    :title="t('pwaInstall.title')"
    :confirm-text="t('pwaInstall.confirm')"
    :cancel-text="t('pwaInstall.cancel')"
    :close-on-click-overlay="false"
    @confirm="triggerInstall"
    @cancel="dismissPwa"
  >
    {{ t('pwaInstall.message') }}
  </AppConfirmDialog>
</template>

<style scoped>
.app-wrapper {
  width: 100%;
  position: relative;
  min-height: 100vh;
  background: linear-gradient(160deg, #fff8f5 0%, #faf0eb 100%);
  /* Subtle Thai diagonal gold stripe watermark */
  background-image:
    linear-gradient(160deg, #fff8f5 0%, #faf0eb 100%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.04) 18px,
      rgba(212, 175, 55, 0.04) 20px
    );
  background-blend-mode: normal;
}
</style>
