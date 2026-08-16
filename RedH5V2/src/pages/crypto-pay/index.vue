<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import type { CryptoRechargeOption, RechargeOrderAppBack } from '@/api/user'
import {
  cancelCryptoRechargeOrder,
  createCryptoRechargeOrder,
  getCryptoRechargeOptions,
  getCryptoRechargeOrderStatus,
} from '@/api/user'
import { APP_CURRENCY } from '@/config/market'
import QRCode from 'qrcode'
import { showConfirmDialog, showFailToast, showSuccessToast, showToast } from 'vant'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '../ppmx-home/types'

defineOptions({ name: 'CryptoPayPage' })

definePage({
  name: 'crypto-pay',
  meta: {
    titleKey: 'cryptoPay.title',
    shell: 'none',
    tabbar: false,
    requiresAuth: true,
  },
})

const CRYPTO_STATUS_PENDING = 0
const CRYPTO_STATUS_PAID = 1
const CRYPTO_STATUS_EXPIRED = 2
const CRYPTO_STATUS_CANCELED = 3
const STATUS_POLL_INTERVAL_MS = 3000

const { pickText } = usePpmxLocale()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const pageState = ref<'loading' | 'selecting' | 'creating' | 'waiting' | 'paid' | 'expired' | 'canceled' | 'error'>('loading')
const pageError = ref('')
const platformAmount = ref(0)
const options = ref<CryptoRechargeOption[]>([])
const selectedOption = ref<CryptoRechargeOption | null>(null)
const paymentOrder = ref<RechargeOrderAppBack | null>(null)
const orderStatus = ref<number | null>(null)
const countdownText = ref('--:--')
const qrDataUrl = ref('')
const qrError = ref(false)
const isRefreshing = ref(false)
const isCanceling = ref(false)

let statusTimer: ReturnType<typeof window.setTimeout> | undefined
let countdownTimer: ReturnType<typeof window.setInterval> | undefined
let expireRefreshTriggered = false

const text = {
  eyebrow: { es: 'PP Pay · Cripto', en: 'PP Pay · Crypto', zh: 'PP Pay · 加密货币' },
  title: { es: 'Pago USDT', en: 'USDT Payment', zh: 'USDT 支付' },
  sub: {
    es: 'Transfiere exactamente el monto indicado mediante USDT TRC20.',
    en: 'Transfer the exact amount below with USDT TRC20.',
    zh: '请使用 USDT TRC20 按页面金额转账。',
  },
  stepMethod: { es: 'Método', en: 'Method', zh: '方式' },
  stepPay: { es: 'Pagar', en: 'Pay', zh: '支付' },
  stepDone: { es: 'Confirmación', en: 'Confirm', zh: '确认' },
  platformAmount: { es: 'Monto de recarga', en: 'Recharge amount', zh: '充值金额' },
  amount: { es: 'Monto a transferir', en: 'Transfer amount', zh: '转账金额' },
  address: { es: 'Dirección receptora', en: 'Receiving address', zh: '收款地址' },
  copy: { es: 'Copiar', en: 'Copy', zh: '复制' },
  copied: { es: 'Copiado', en: 'Copied', zh: '已复制' },
  copyFailed: { es: 'Copia manualmente', en: 'Copy manually', zh: '请手动复制' },
  waiting: { es: 'Esperando pago', en: 'Waiting for payment', zh: '等待付款' },
  selecting: { es: 'Elige moneda', en: 'Choose token', zh: '选择币种' },
  creating: { es: 'Creando orden', en: 'Creating order', zh: '正在创建订单' },
  paid: { es: 'Pago confirmado', en: 'Payment confirmed', zh: '支付已确认' },
  expired: { es: 'Orden expirada', en: 'Order expired', zh: '订单已过期' },
  canceled: { es: 'Orden cancelada', en: 'Order canceled', zh: '订单已取消' },
  scanHint: { es: 'Escanea con tu wallet o copia la dirección.', en: 'Scan with your wallet or copy the address.', zh: '使用钱包扫码，或复制地址转账。' },
  selected: { es: 'Método seleccionado', en: 'Selected method', zh: '当前方式' },
  chooseTitle: { es: 'Selecciona la moneda', en: 'Select token', zh: '选择支付币种' },
  chooseSub: { es: 'Actualmente solo está disponible USDT en TRC20.', en: 'USDT on TRC20 is currently available.', zh: '当前仅开放 USDT · TRC20。' },
  createOrder: { es: 'Crear orden USDT', en: 'Create USDT order', zh: '生成 USDT 订单' },
  creatingOrder: { es: 'Creando...', en: 'Creating...', zh: '创建中...' },
  noticeTitle: { es: 'Antes de transferir', en: 'Before transfer', zh: '转账前确认' },
  noticeNet: { es: 'Solo envía USDT por Tron (TRC20). Otros activos o redes no se acreditarán.', en: 'Only send USDT on Tron (TRC20). Other assets or networks will not be credited.', zh: '仅支持 USDT Tron (TRC20)，其他币种或网络不会入账。' },
  noticeAmount: { es: 'El monto debe coincidir exactamente con {amount}.', en: 'The amount must match exactly {amount}.', zh: '转账金额需与 {amount} 完全一致。' },
  noticeOrder: { es: 'La acreditación se mostrará después de la confirmación de red.', en: 'Credit will appear after network confirmation.', zh: '链上确认后将自动更新入账状态。' },
  refresh: { es: 'Actualizar estado', en: 'Refresh status', zh: '刷新状态' },
  refreshing: { es: 'Pago pendiente de confirmación', en: 'Payment is still pending confirmation', zh: '订单仍在等待链上确认' },
  cancel: { es: 'Cancelar pago', en: 'Cancel payment', zh: '取消支付' },
  back: { es: 'Volver', en: 'Back', zh: '返回' },
  done: { es: 'Listo', en: 'Done', zh: '完成' },
  order: { es: 'Orden', en: 'Order', zh: '订单号' },
  network: { es: 'Red', en: 'Network', zh: '网络' },
  expires: { es: 'Expira en', en: 'Expires in', zh: '有效时间' },
  qrError: { es: 'No se pudo generar el QR', en: 'Could not generate QR', zh: '二维码生成失败' },
  loadFailed: { es: 'No se pudo cargar el pago cripto.', en: 'Could not load crypto payment.', zh: '加载虚拟货币支付失败。' },
  invalidAmount: { es: 'Monto inválido.', en: 'Invalid amount.', zh: '充值金额无效。' },
  noOption: { es: 'No hay método cripto disponible.', en: 'No crypto method is available.', zh: '暂无可用虚拟货币支付方式。' },
  confirmCycleTitle: { es: 'Continuar recarga', en: 'Continue recharge', zh: '继续充值' },
  confirmCycleMessage: { es: 'Hay un ciclo de actividad pendiente. Confirma para continuar.', en: 'An activity cycle is still unfinished. Confirm to continue.', zh: '当前存在未完成活动周期，确认后继续充值。' },
} satisfies Record<string, LocalizedText>

const summaryItems = computed(() => [
  paymentOrder.value
    ? { label: pickText(text.order), value: paymentOrder.value.orderNo }
    : { label: pickText(text.platformAmount), value: platformAmountText.value },
  { label: pickText(text.network), value: networkName.value },
  paymentOrder.value
    ? { label: pickText(text.expires), value: countdownText.value }
    : { label: pickText(text.amount), value: estimatedAmountText.value },
])

const cryptoPayment = computed(() => paymentOrder.value?.cryptoPayment ?? null)
const orderNo = computed(() => paymentOrder.value?.orderNo ?? '')
const selectedToken = computed(() => cryptoPayment.value?.token || selectedOption.value?.token || 'USDT')
const networkName = computed(() => {
  const network = cryptoPayment.value?.network || selectedOption.value?.network || 'TRC20'

  return network.toUpperCase() === 'TRC20' ? 'Tron (TRC20)' : network
})
const platformAmountText = computed(() => `${formatAmount(platformAmount.value)} ${APP_CURRENCY}`)
const estimatedAmountText = computed(() => {
  const estimatedAmount = selectedOption.value?.estimatedAmount || '--'

  return `${estimatedAmount} ${selectedToken.value}`
})
const payAmount = computed(() => {
  const expectedAmount = cryptoPayment.value?.expectedAmount

  return `${expectedAmount || selectedOption.value?.estimatedAmount || '--'} ${selectedToken.value}`
})
const usdtAddress = computed(() => cryptoPayment.value?.receiveAddress || selectedOption.value?.receiveAddress || '')
const noticeAmountText = computed(() => pickText(text.noticeAmount).replace('{amount}', payAmount.value))
const statusText = computed(() => {
  switch (pageState.value) {
    case 'loading':
      return pickText(text.creating)
    case 'selecting':
      return pickText(text.selecting)
    case 'creating':
      return pickText(text.creating)
    case 'paid':
      return pickText(text.paid)
    case 'expired':
      return pickText(text.expired)
    case 'canceled':
      return pickText(text.canceled)
    default:
      return pickText(text.waiting)
  }
})
const isCreating = computed(() => pageState.value === 'creating')
const isPending = computed(() => Number(orderStatus.value) === CRYPTO_STATUS_PENDING && Boolean(orderNo.value))
const isTerminal = computed(() => [CRYPTO_STATUS_PAID, CRYPTO_STATUS_EXPIRED, CRYPTO_STATUS_CANCELED].includes(Number(orderStatus.value)))
const canCreateOrder = computed(() => pageState.value === 'selecting' && Boolean(selectedOption.value) && !isCreating.value)

function firstQueryValue(value: unknown) {
  if (Array.isArray(value)) return value[0]

  return value
}

function queryString(key: string) {
  const value = firstQueryValue(route.query[key])

  return typeof value === 'string' ? value.trim() : ''
}

function queryAmount() {
  const value = Number(queryString('amount'))

  return Number.isFinite(value) ? value : 0
}

function formatAmount(value: number) {
  if (!Number.isFinite(value)) return '0.00'

  return value.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

function normalizeStatus(status: unknown) {
  const numeric = Number(status)

  return Number.isFinite(numeric) ? numeric : CRYPTO_STATUS_PENDING
}

function merchantOrderNo() {
  const uid = userStore.userInfo?.id ?? 'user'

  return `h5_crypto_${Date.now()}_${uid}`
}

async function loadOptions() {
  cleanupTimers()
  pageState.value = 'loading'
  pageError.value = ''
  paymentOrder.value = null
  orderStatus.value = null
  qrDataUrl.value = ''
  qrError.value = false
  expireRefreshTriggered = false

  const amount = queryAmount()
  if (!Number.isFinite(amount) || amount <= 0) {
    pageState.value = 'error'
    pageError.value = pickText(text.invalidAmount)
    return
  }

  platformAmount.value = amount

  try {
    const result = await getCryptoRechargeOptions(amount)
    options.value = result.options ?? []
    selectedOption.value = options.value[0] ?? null

    if (!selectedOption.value) {
      pageState.value = 'error'
      pageError.value = pickText(text.noOption)
      return
    }

    pageState.value = 'selecting'
  } catch (error) {
    pageState.value = 'error'
    pageError.value = error instanceof Error ? error.message : pickText(text.loadFailed)
  }
}

async function createOrder(confirmUnfinishedActivityCycle = false): Promise<RechargeOrderAppBack> {
  const option = selectedOption.value
  if (!option) throw new Error(pickText(text.noOption))

  const order = await createCryptoRechargeOrder({
    amount: platformAmount.value,
    network: option.network,
    token: option.token,
    activityCode: queryString('activityCode') || undefined,
    merchantOrderNo: merchantOrderNo(),
    confirmUnfinishedActivityCycle,
  })

  if (order.needConfirmUnfinishedActivityCycle && !confirmUnfinishedActivityCycle) {
    await confirmUnfinishedActivityCycleDialog()

    return createOrder(true)
  }

  return order
}

async function confirmUnfinishedActivityCycleDialog() {
  await showConfirmDialog({
    title: pickText(text.confirmCycleTitle),
    message: pickText(text.confirmCycleMessage),
    confirmButtonText: pickText(text.createOrder),
    cancelButtonText: pickText(text.cancel),
  })
}

async function confirmCreateOrder() {
  if (!canCreateOrder.value) return
  pageState.value = 'creating'

  try {
    const order = await createOrder(false)
    if (!order.cryptoPayment) throw new Error(pickText(text.loadFailed))

    paymentOrder.value = order
    orderStatus.value = normalizeStatus(order.status)
    pageState.value = 'waiting'
    expireRefreshTriggered = false
    await renderQr()
    startCountdown()
    startStatusPolling()
    showToast(pickText(text.waiting))
  } catch (error) {
    pageState.value = 'selecting'
    showFailToast(error instanceof Error ? error.message : pickText(text.loadFailed))
  }
}

async function renderQr() {
  qrError.value = false
  const address = cryptoPayment.value?.receiveAddress || ''
  if (!address) {
    qrDataUrl.value = ''
    qrError.value = true
    return
  }

  try {
    qrDataUrl.value = await QRCode.toDataURL(address, {
      errorCorrectionLevel: 'H',
      margin: 1,
      scale: 8,
      color: {
        dark: '#141416',
        light: '#ffffff',
      },
    })
  } catch {
    qrDataUrl.value = ''
    qrError.value = true
  }
}

async function copyValue(value: string) {
  try {
    await navigator.clipboard.writeText(value)
    showToast(pickText(text.copied))
  } catch {
    showToast(pickText(text.copyFailed))
  }
}

function applyStatus(status: Awaited<ReturnType<typeof getCryptoRechargeOrderStatus>>) {
  const previousAddress = cryptoPayment.value?.receiveAddress || ''
  orderStatus.value = normalizeStatus(status.status)

  if (paymentOrder.value) {
    paymentOrder.value = {
      ...paymentOrder.value,
      orderNo: status.orderNo || paymentOrder.value.orderNo,
      status: status.rechargeStatus,
      cryptoPayment: status.cryptoPayment ?? paymentOrder.value.cryptoPayment,
    }
  }

  if (status.cryptoPayment && previousAddress !== status.cryptoPayment.receiveAddress) {
    void renderQr()
  }

  if (orderStatus.value === CRYPTO_STATUS_PAID) {
    pageState.value = 'paid'
    stopStatusPolling()
    stopCountdown()
    showSuccessToast(pickText(text.paid))
    void userStore.loadUserInfo()
    return
  }

  if (orderStatus.value === CRYPTO_STATUS_EXPIRED) {
    pageState.value = 'expired'
    stopStatusPolling()
    stopCountdown()
    return
  }

  if (orderStatus.value === CRYPTO_STATUS_CANCELED) {
    pageState.value = 'canceled'
    stopStatusPolling()
    stopCountdown()
    return
  }

  pageState.value = 'waiting'
}

async function refreshStatus(showPendingToast = true) {
  if (!orderNo.value || isRefreshing.value) return
  isRefreshing.value = true

  try {
    const status = await getCryptoRechargeOrderStatus(orderNo.value)
    applyStatus(status)

    if (showPendingToast && isPending.value) {
      showToast(pickText(text.refreshing))
    }
  } catch (error) {
    if (showPendingToast) {
      showFailToast(error instanceof Error ? error.message : pickText(text.loadFailed))
    }
    throw error
  } finally {
    isRefreshing.value = false
  }
}

function startStatusPolling() {
  stopStatusPolling()
  if (!isPending.value || isTerminal.value) return

  statusTimer = window.setTimeout(async () => {
    try {
      await refreshStatus(false)
    } catch {
      // The next poll keeps the page recoverable when a transient request fails.
    }

    if (isPending.value && !isTerminal.value) {
      startStatusPolling()
    }
  }, STATUS_POLL_INTERVAL_MS)
}

function stopStatusPolling() {
  if (statusTimer) {
    window.clearTimeout(statusTimer)
    statusTimer = undefined
  }
}

function startCountdown() {
  stopCountdown()
  updateCountdown()
  countdownTimer = window.setInterval(updateCountdown, 1000)
}

function stopCountdown() {
  if (countdownTimer) {
    window.clearInterval(countdownTimer)
    countdownTimer = undefined
  }
}

function updateCountdown() {
  const expireTime = cryptoPayment.value?.expireTime || ''
  const expireMs = Date.parse(expireTime)
  if (!Number.isFinite(expireMs)) {
    countdownText.value = '--:--'
    return
  }

  const remainingSeconds = Math.max(0, Math.ceil((expireMs - Date.now()) / 1000))
  const minutes = Math.floor(remainingSeconds / 60)
  const seconds = remainingSeconds % 60
  countdownText.value = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`

  if (remainingSeconds <= 0 && isPending.value && !expireRefreshTriggered) {
    expireRefreshTriggered = true
    void refreshStatus(false).catch(() => {})
  }
}

async function cancelPayment() {
  if (!orderNo.value || !isPending.value || isCanceling.value) return
  isCanceling.value = true

  try {
    const status = await cancelCryptoRechargeOrder(orderNo.value)
    applyStatus(status)
    showToast(pickText(text.canceled))
  } catch (error) {
    showFailToast(error instanceof Error ? error.message : pickText(text.loadFailed))
  } finally {
    isCanceling.value = false
  }
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
    return
  }

  void router.push('/profile')
}

function cleanupTimers() {
  stopStatusPolling()
  stopCountdown()
}

function handlePageHide() {
  cleanupTimers()
}

onMounted(() => {
  window.addEventListener('pagehide', handlePageHide)
  void userStore.loadUserInfo().catch(() => {})
  void loadOptions()
})

onBeforeUnmount(() => {
  window.removeEventListener('pagehide', handlePageHide)
  cleanupTimers()
})
</script>

<template>
  <main id="cryptoPayFlow" class="ppmx-page ppmx-crypto-pay-page">
    <section class="ppmx-crypto-pay-stack">
      <header class="ppmx-crypto-pay-head reveal">
        <div class="ppmx-eyebrow">{{ pickText(text.eyebrow) }}</div>
        <h1 class="ppmx-display">{{ pickText(text.title) }}</h1>
        <p>{{ pickText(text.sub) }}</p>
      </header>

      <section class="ppmx-crypto-pay-card reveal" :aria-label="pickText(text.title)">
        <header class="ppmx-crypto-pay-card__head">
          <div class="ppmx-crypto-pay-brand">
            <span class="ppmx-crypto-pay-token" aria-hidden="true">
              <svg viewBox="0 0 32 32" role="img">
                <circle cx="16" cy="16" r="16" fill="#22a079" />
                <path fill="#fff" d="M8 8.5h16v3.25h-6.15v2.12c3.45.18 6.05.88 6.05 1.72s-2.6 1.54-6.05 1.72v6.19h-3.7v-6.19c-3.45-.18-6.05-.88-6.05-1.72s2.6-1.54 6.05-1.72v-2.12H8V8.5Zm6.15 6.82c-2.03.13-3.43.38-3.43.67s1.4.54 3.43.67v-1.34Zm3.7 1.34c2.03-.13 3.43-.38 3.43-.67s-1.4-.54-3.43-.67v1.34Z" />
              </svg>
            </span>
            <div>
              <span>{{ pickText(text.selected) }}</span>
              <strong>{{ selectedToken }} · {{ networkName }}</strong>
            </div>
          </div>
          <span class="ppmx-crypto-pay-status" :class="{ 'is-success': pageState === 'paid', 'is-ended': pageState === 'expired' || pageState === 'canceled' }">
            <i aria-hidden="true" />
            {{ statusText }}
          </span>
        </header>

        <nav class="ppmx-crypto-pay-steps" aria-label="Payment steps">
          <div class="ppmx-crypto-pay-step" :class="{ 'is-done': paymentOrder, 'is-active': !paymentOrder }">
            <span><i v-if="paymentOrder" class="fa-solid fa-check" /><template v-else>1</template></span>
            <small>{{ pickText(text.stepMethod) }}</small>
          </div>
          <i class="ppmx-crypto-pay-step-line" :class="{ 'is-done': paymentOrder }" aria-hidden="true" />
          <div class="ppmx-crypto-pay-step" :class="{ 'is-active': paymentOrder && !isTerminal, 'is-done': pageState === 'paid' }">
            <span><i v-if="pageState === 'paid'" class="fa-solid fa-check" /><template v-else>2</template></span>
            <small>{{ pickText(text.stepPay) }}</small>
          </div>
          <i class="ppmx-crypto-pay-step-line" :class="{ 'is-done': pageState === 'paid' }" aria-hidden="true" />
          <div class="ppmx-crypto-pay-step" :class="{ 'is-active': isTerminal }">
            <span>3</span>
            <small>{{ pickText(text.stepDone) }}</small>
          </div>
        </nav>

        <section v-if="pageState === 'loading'" class="ppmx-crypto-pay-state">
          <i class="fa-solid fa-circle-notch fa-spin" />
          <p>{{ pickText(text.creating) }}</p>
        </section>

        <section v-else-if="pageState === 'error'" class="ppmx-crypto-pay-state">
          <i class="fa-solid fa-triangle-exclamation" />
          <p>{{ pageError || pickText(text.loadFailed) }}</p>
          <button class="ppmx-crypto-pay-refresh" type="button" @click="loadOptions">
            <i class="fa-solid fa-arrows-rotate" />
            {{ pickText(text.refresh) }}
          </button>
        </section>

        <template v-else>
          <section v-if="!paymentOrder" class="ppmx-crypto-pay-select">
            <div>
              <h2>{{ pickText(text.chooseTitle) }}</h2>
              <p>{{ pickText(text.chooseSub) }}</p>
            </div>
            <button
              v-for="option in options"
              :key="`${option.token}-${option.network}`"
              class="ppmx-crypto-pay-option"
              :class="{ 'is-active': selectedOption === option }"
              type="button"
              @click="selectedOption = option"
            >
              <span class="ppmx-crypto-pay-option__icon">
                <i class="fa-solid fa-coins" />
              </span>
              <span>
                <strong>{{ option.token }} · {{ option.network }}</strong>
                <small>{{ option.estimatedAmount }} {{ option.token }}</small>
              </span>
              <i class="fa-solid fa-check" />
            </button>
          </section>

          <section v-if="paymentOrder" class="ppmx-crypto-pay-qr" :aria-label="pickText(text.scanHint)">
            <div class="ppmx-crypto-pay-qr__plate">
              <span class="ppmx-crypto-pay-corner ppmx-crypto-pay-corner--tl" />
              <span class="ppmx-crypto-pay-corner ppmx-crypto-pay-corner--tr" />
              <span class="ppmx-crypto-pay-corner ppmx-crypto-pay-corner--bl" />
              <span class="ppmx-crypto-pay-corner ppmx-crypto-pay-corner--br" />
              <img v-if="qrDataUrl" :src="qrDataUrl" :alt="pickText(text.scanHint)">
              <i v-else class="fa-solid" :class="qrError ? 'fa-triangle-exclamation' : 'fa-qrcode'" aria-hidden="true" />
            </div>
            <p>{{ qrError ? pickText(text.qrError) : pickText(text.scanHint) }}</p>
          </section>

          <section v-if="paymentOrder" class="ppmx-crypto-pay-fields">
            <label>
              <span>{{ pickText(text.amount) }}</span>
              <div>
                <input :value="payAmount" readonly>
                <button type="button" @click="copyValue(payAmount)">
                  <i class="fa-regular fa-copy" />
                  {{ pickText(text.copy) }}
                </button>
              </div>
            </label>
            <label>
              <span>{{ pickText(text.address) }}</span>
              <div>
                <input :value="usdtAddress" readonly>
                <button type="button" @click="copyValue(usdtAddress)">
                  <i class="fa-regular fa-copy" />
                  {{ pickText(text.copy) }}
                </button>
              </div>
            </label>
          </section>

          <dl class="ppmx-crypto-pay-summary">
            <div v-for="item in summaryItems" :key="item.label">
              <dt>{{ item.label }}</dt>
              <dd>{{ item.value }}</dd>
            </div>
          </dl>

          <section class="ppmx-crypto-pay-notice">
            <h2>
              <i class="fa-solid fa-circle-info" />
              {{ pickText(text.noticeTitle) }}
            </h2>
            <p>
              <i class="fa-solid fa-triangle-exclamation" />
              {{ pickText(text.noticeNet) }}
            </p>
            <p>
              <i class="fa-solid fa-check" />
              {{ noticeAmountText }}
            </p>
            <p>
              <i class="fa-solid fa-check" />
              {{ pickText(text.noticeOrder) }}
            </p>
          </section>

          <div class="ppmx-crypto-pay-actions">
            <button
              v-if="!paymentOrder"
              class="ppmx-crypto-pay-refresh"
              type="button"
              :disabled="!canCreateOrder"
              @click="confirmCreateOrder"
            >
              <i class="fa-solid" :class="isCreating ? 'fa-circle-notch fa-spin' : 'fa-arrow-right'" />
              {{ isCreating ? pickText(text.creatingOrder) : pickText(text.createOrder) }}
            </button>
            <template v-else-if="isPending">
              <button class="ppmx-crypto-pay-refresh" type="button" :disabled="isRefreshing" @click="refreshStatus()">
                <i class="fa-solid" :class="isRefreshing ? 'fa-circle-notch fa-spin' : 'fa-arrows-rotate'" />
                {{ pickText(text.refresh) }}
              </button>
              <button class="ppmx-crypto-pay-refresh is-secondary" type="button" :disabled="isCanceling" @click="cancelPayment">
                <i class="fa-solid" :class="isCanceling ? 'fa-circle-notch fa-spin' : 'fa-xmark'" />
                {{ pickText(text.cancel) }}
              </button>
            </template>
            <button v-else class="ppmx-crypto-pay-refresh" type="button" @click="goBack">
              <i class="fa-solid fa-check" />
              {{ pickText(text.done) }}
            </button>
          </div>
        </template>
      </section>
    </section>
  </main>
</template>

<style scoped>
.ppmx-crypto-pay-stack {
  display: grid;
  gap: 18px;
  width: calc(100% - 32px);
  max-width: 358px;
  padding: clamp(28px, 5vw, 56px) 0;
  margin: 0 auto;
}

.ppmx-crypto-pay-page {
  min-height: 100vh;
  min-height: 100dvh;
  color: var(--ppmx-text, #f5f5f7);
  background:
    radial-gradient(760px 420px at 18% -10%, rgba(200, 16, 46, 0.18), transparent 66%),
    radial-gradient(680px 360px at 100% 6%, rgba(255, 215, 0, 0.08), transparent 64%),
    linear-gradient(180deg, #060608 0%, #08080b 56%, #050507 100%);
}

:global(body:has(#cryptoPayFlow) .install-pwa-button) {
  display: none;
}

.ppmx-crypto-pay-head {
  display: grid;
  gap: 10px;
}

.ppmx-crypto-pay-head h1,
.ppmx-crypto-pay-head p {
  margin: 0;
}

.ppmx-crypto-pay-head h1 {
  color: #fff;
  font-size: clamp(2rem, 9vw, 3.2rem);
  line-height: 0.98;
}

.ppmx-crypto-pay-head p {
  max-width: 100%;
  color: rgba(245, 245, 247, 0.48);
  font-size: 0.92rem;
  font-weight: 700;
  line-height: 1.55;
  overflow-wrap: anywhere;
}

.ppmx-crypto-pay-card {
  display: grid;
  gap: 20px;
  width: calc(100vw - 32px);
  max-width: 358px;
  margin: 0 auto;
  padding: clamp(20px, 4vw, 28px);
  margin: 0 auto;
  color: #1a1a1a;
  overflow: hidden;
  background:
    radial-gradient(520px 240px at 24% -6%, rgba(200, 16, 46, 0.13), transparent 62%),
    linear-gradient(180deg, #fff8f8 0%, #ffffff 38%, #f8f8fa 100%);
  border: 1px solid rgba(26, 26, 32, 0.09);
  border-radius: 28px;
  box-shadow:
    0 34px 80px -44px rgba(0, 0, 0, 0.72),
    inset 0 1px 0 rgba(255, 255, 255, 0.94);
}

.ppmx-crypto-pay-brand,
.ppmx-crypto-pay-status {
  display: flex;
  align-items: center;
}

.ppmx-crypto-pay-card__head {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
  align-items: flex-start;
}

.ppmx-crypto-pay-brand {
  min-width: 0;
  gap: 12px;
}

.ppmx-crypto-pay-token {
  display: grid;
  width: 46px;
  height: 46px;
  flex: 0 0 auto;
  place-items: center;
  background: #f0fffa;
  border: 1px solid rgba(34, 160, 121, 0.16);
  border-radius: 14px;
  box-shadow: 0 14px 30px -22px rgba(34, 160, 121, 0.7);
}

.ppmx-crypto-pay-token svg {
  display: block;
  width: 32px;
  height: 32px;
}

.ppmx-crypto-pay-brand span {
  display: block;
  color: #71717a;
  font-size: 0.76rem;
  font-weight: 800;
  line-height: 1.25;
}

.ppmx-crypto-pay-brand strong {
  display: block;
  margin-top: 3px;
  overflow: hidden;
  color: #18181b;
  font-size: 0.98rem;
  font-weight: 900;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: normal;
}

.ppmx-crypto-pay-status {
  flex: 0 0 auto;
  justify-self: start;
  gap: 7px;
  min-height: 30px;
  padding: 0 10px;
  color: #b8860b;
  font-size: 0.76rem;
  font-weight: 900;
  background: rgba(255, 215, 0, 0.16);
  border: 1px solid rgba(184, 134, 11, 0.14);
  border-radius: 999px;
}

.ppmx-crypto-pay-status i {
  width: 7px;
  height: 7px;
  background: #f0bc3d;
  border-radius: 999px;
  box-shadow: 0 0 0 5px rgba(240, 188, 61, 0.16);
}

.ppmx-crypto-pay-status.is-success {
  color: #13834d;
  background: rgba(30, 158, 90, 0.14);
  border-color: rgba(30, 158, 90, 0.18);
}

.ppmx-crypto-pay-status.is-success i {
  background: #1e9e5a;
  box-shadow: 0 0 0 5px rgba(30, 158, 90, 0.16);
}

.ppmx-crypto-pay-status.is-ended {
  color: #71717a;
  background: rgba(113, 113, 122, 0.12);
  border-color: rgba(113, 113, 122, 0.16);
}

.ppmx-crypto-pay-status.is-ended i {
  background: #8a8a92;
  box-shadow: 0 0 0 5px rgba(138, 138, 146, 0.14);
}

.ppmx-crypto-pay-steps {
  display: grid;
  grid-template-columns: auto minmax(20px, 1fr) auto minmax(20px, 1fr) auto;
  gap: 8px;
  align-items: start;
}

.ppmx-crypto-pay-step {
  display: grid;
  gap: 6px;
  justify-items: center;
  min-width: 54px;
}

.ppmx-crypto-pay-step span {
  display: grid;
  width: 26px;
  height: 26px;
  place-items: center;
  color: #8a8a92;
  font-size: 0.72rem;
  font-weight: 900;
  background: rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 999px;
}

.ppmx-crypto-pay-step small {
  color: #8a8a92;
  font-size: 0.68rem;
  font-weight: 800;
  line-height: 1.2;
  text-align: center;
}

.ppmx-crypto-pay-step.is-done span {
  color: #fff;
  background: #1e9e5a;
  border-color: #1e9e5a;
}

.ppmx-crypto-pay-step.is-active span {
  color: #fff;
  background: #c8102e;
  border-color: #c8102e;
}

.ppmx-crypto-pay-step-line {
  height: 2px;
  margin-top: 12px;
  background: rgba(0, 0, 0, 0.08);
  border-radius: 999px;
}

.ppmx-crypto-pay-step-line.is-done {
  background: #1e9e5a;
}

.ppmx-crypto-pay-state,
.ppmx-crypto-pay-select {
  display: grid;
  gap: 14px;
}

.ppmx-crypto-pay-state {
  justify-items: center;
  padding: 28px 12px;
  text-align: center;
}

.ppmx-crypto-pay-state > i {
  color: #c8102e;
  font-size: 2rem;
}

.ppmx-crypto-pay-state p,
.ppmx-crypto-pay-select h2,
.ppmx-crypto-pay-select p {
  margin: 0;
}

.ppmx-crypto-pay-state p {
  color: #4a4a52;
  font-size: 0.86rem;
  font-weight: 800;
  line-height: 1.45;
}

.ppmx-crypto-pay-select h2 {
  color: #18181b;
  font-size: 1rem;
  font-weight: 900;
}

.ppmx-crypto-pay-select p {
  margin-top: 4px;
  color: #71717a;
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.4;
}

.ppmx-crypto-pay-option {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
  width: 100%;
  padding: 13px;
  color: #18181b;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid rgba(26, 26, 32, 0.1);
  border-radius: 16px;
  transition:
    border-color 0.18s ease,
    box-shadow 0.22s ease,
    transform 0.18s var(--ppmx-ease);
}

.ppmx-crypto-pay-option.is-active {
  border-color: rgba(34, 160, 121, 0.45);
  box-shadow: 0 18px 34px -28px rgba(34, 160, 121, 0.78);
}

.ppmx-crypto-pay-option:active {
  transform: scale(0.99);
}

.ppmx-crypto-pay-option__icon {
  display: grid;
  width: 40px;
  height: 40px;
  place-items: center;
  color: #13834d;
  background: #f0fffa;
  border: 1px solid rgba(34, 160, 121, 0.16);
  border-radius: 13px;
}

.ppmx-crypto-pay-option strong,
.ppmx-crypto-pay-option small {
  display: block;
}

.ppmx-crypto-pay-option strong {
  color: #18181b;
  font-size: 0.9rem;
  font-weight: 900;
}

.ppmx-crypto-pay-option small {
  margin-top: 3px;
  color: #71717a;
  font-size: 0.76rem;
  font-weight: 800;
}

.ppmx-crypto-pay-option > .fa-check {
  color: #1e9e5a;
  opacity: 0;
}

.ppmx-crypto-pay-option.is-active > .fa-check {
  opacity: 1;
}

.ppmx-crypto-pay-qr {
  display: grid;
  gap: 10px;
  justify-items: center;
  padding: 18px 0 2px;
}

.ppmx-crypto-pay-qr__plate {
  position: relative;
  display: grid;
  width: 188px;
  height: 188px;
  padding: 14px;
  place-items: center;
  background: #fff;
  border: 1px solid rgba(26, 26, 32, 0.08);
  border-radius: 22px;
  box-shadow: 0 18px 42px -30px rgba(26, 26, 32, 0.45);
}

.ppmx-crypto-pay-qr__plate img {
  display: block;
  width: 154px;
  height: 154px;
  border-radius: 6px;
}

.ppmx-crypto-pay-qr__plate > i {
  color: #c8102e;
  font-size: 3rem;
}

.ppmx-crypto-pay-corner {
  position: absolute;
  z-index: 2;
  width: 22px;
  height: 22px;
  border: 3px solid #ffd700;
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.16));
}

.ppmx-crypto-pay-corner--tl {
  top: 9px;
  left: 9px;
  border-right: 0;
  border-bottom: 0;
  border-top-left-radius: 7px;
}

.ppmx-crypto-pay-corner--tr {
  top: 9px;
  right: 9px;
  border-bottom: 0;
  border-left: 0;
  border-top-right-radius: 7px;
}

.ppmx-crypto-pay-corner--bl {
  bottom: 9px;
  left: 9px;
  border-top: 0;
  border-right: 0;
  border-bottom-left-radius: 7px;
}

.ppmx-crypto-pay-corner--br {
  right: 9px;
  bottom: 9px;
  border-top: 0;
  border-left: 0;
  border-bottom-right-radius: 7px;
}

.ppmx-crypto-pay-qr p {
  margin: 0;
  color: #8a8a92;
  font-size: 0.76rem;
  font-weight: 700;
  text-align: center;
}

.ppmx-crypto-pay-fields {
  display: grid;
  gap: 14px;
}

.ppmx-crypto-pay-fields label {
  display: grid;
  gap: 7px;
  min-width: 0;
}

.ppmx-crypto-pay-fields label > span {
  color: #71717a;
  font-size: 0.76rem;
  font-weight: 800;
}

.ppmx-crypto-pay-fields label > div {
  display: grid;
  grid-template-columns: 1fr;
  gap: 9px;
}

.ppmx-crypto-pay-fields input {
  width: 100%;
  min-width: 0;
  height: 46px;
  padding: 0 13px;
  color: #18181b;
  font-size: 0.85rem;
  font-weight: 800;
  background: #f7f7f8;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 12px;
}

.ppmx-crypto-pay-fields label:last-child input {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.74rem;
  font-weight: 700;
}

.ppmx-crypto-pay-fields button,
.ppmx-crypto-pay-refresh {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  min-width: 86px;
  height: 46px;
  color: #18181b;
  font-size: 0.82rem;
  font-weight: 900;
  cursor: pointer;
  background: #fff;
  border: 1px solid rgba(26, 26, 32, 0.12);
  border-radius: 12px;
  transition:
    background 0.18s ease,
    transform 0.18s var(--ppmx-ease),
    box-shadow 0.22s ease;
}

.ppmx-crypto-pay-fields button:hover,
.ppmx-crypto-pay-refresh:hover {
  background: #fff7f7;
  box-shadow: 0 14px 28px -22px rgba(200, 16, 46, 0.6);
  transform: translateY(-1px);
}

.ppmx-crypto-pay-fields button:active,
.ppmx-crypto-pay-refresh:active {
  transform: scale(0.98);
}

.ppmx-crypto-pay-fields button {
  width: 100%;
}

.ppmx-crypto-pay-summary,
.ppmx-crypto-pay-notice {
  display: grid;
  gap: 10px;
  padding: 14px;
  background: #fff;
  border: 1px solid rgba(26, 26, 32, 0.08);
  border-radius: 16px;
  box-shadow: 0 6px 18px -14px rgba(20, 20, 22, 0.18);
}

.ppmx-crypto-pay-summary {
  margin: 0;
}

.ppmx-crypto-pay-summary div {
  display: grid;
  gap: 12px;
  min-width: 0;
}

.ppmx-crypto-pay-summary dt,
.ppmx-crypto-pay-summary dd {
  margin: 0;
  min-width: 0;
}

.ppmx-crypto-pay-summary dt {
  color: #71717a;
  font-size: 0.76rem;
  font-weight: 800;
}

.ppmx-crypto-pay-summary dd {
  overflow: hidden;
  color: #18181b;
  font-size: 0.82rem;
  font-weight: 900;
  text-align: left;
  text-overflow: ellipsis;
  white-space: normal;
  word-break: break-all;
}

.ppmx-crypto-pay-notice h2 {
  display: flex;
  gap: 8px;
  align-items: center;
  margin: 0 0 2px;
  color: #18181b;
  font-size: 0.86rem;
  font-weight: 900;
}

.ppmx-crypto-pay-notice p {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  margin: 0;
  color: #4a4a52;
  font-size: 0.76rem;
  font-weight: 700;
  line-height: 1.45;
}

.ppmx-crypto-pay-notice p:first-of-type i {
  color: #b8860b;
}

.ppmx-crypto-pay-notice p:not(:first-of-type) i {
  color: #1e9e5a;
}

.ppmx-crypto-pay-refresh {
  width: 100%;
  color: #fff;
  background: linear-gradient(180deg, #e21f33, #c8102e);
  border-color: #c8102e;
  box-shadow: 0 16px 40px -18px rgba(200, 16, 46, 0.72);
}

.ppmx-crypto-pay-refresh:hover {
  background: linear-gradient(180deg, #ef3447, #c8102e);
}

.ppmx-crypto-pay-refresh.is-secondary {
  color: #18181b;
  background: #fff;
  border-color: rgba(26, 26, 32, 0.14);
  box-shadow: none;
}

.ppmx-crypto-pay-refresh.is-secondary:hover {
  background: #fff7f7;
}

.ppmx-crypto-pay-refresh:disabled {
  cursor: not-allowed;
  opacity: 0.62;
  transform: none;
}

.ppmx-crypto-pay-actions {
  display: grid;
  gap: 10px;
}

@media (min-width: 540px) {
  .ppmx-crypto-pay-stack,
  .ppmx-crypto-pay-card {
    max-width: 560px;
  }

  .ppmx-crypto-pay-card__head {
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
  }

  .ppmx-crypto-pay-status {
    justify-self: end;
  }

  .ppmx-crypto-pay-brand strong {
    white-space: nowrap;
  }

  .ppmx-crypto-pay-fields label > div {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .ppmx-crypto-pay-fields button {
    width: auto;
  }

  .ppmx-crypto-pay-summary div {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .ppmx-crypto-pay-summary dd {
    text-align: right;
    white-space: nowrap;
    word-break: normal;
  }
}

@media (max-width: 430px) {
  .ppmx-crypto-pay-card {
    border-radius: 22px;
  }

  .ppmx-crypto-pay-status {
    min-height: 28px;
    padding: 0 8px;
    font-size: 0.68rem;
  }
}
</style>
