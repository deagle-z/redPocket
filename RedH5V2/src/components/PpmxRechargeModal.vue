<script setup lang="ts">
import type {
  AppCountryItem,
  AppRechargeChannelItem,
  RechargeField,
  RechargeFieldOption,
  RechargeFirstStatusResp,
  RechargeOrderAppBack,
  RechargeOrderAppReq,
} from '@/api/user'
import {
  ackRechargeNotification,
  createRechargeOrderV2,
  getAppCountries,
  getCountryRechargeFields,
  getCountryRechargeInfo,
  getPendingRechargeNotifications,
  getRechargeIsFirstV2,
} from '@/api/user'
import {
  APP_COUNTRY_CODE,
  APP_COUNTRY_NAME_LOCAL,
  APP_CURRENCY,
  APP_CURRENCY_SYMBOL,
} from '@/config/market'
import { rechargeBonusAmountByNumber, rechargeBonusRateByNumber } from '@/utils/rechargeBonus'
import { formatMoney, toDisplayCents } from '@/utils/money'
import { getRechargeFieldInitialValues, setCachedRechargeFieldValues } from '@/utils/rechargeFieldCache'
import {
  DEFAULT_RECHARGE_WALLET_TYPE,
  type RechargeWalletType,
} from '@/utils/walletType'
import { trackEvent } from '@/utils/tracker'
import { showConfirmDialog, showFailToast, showSuccessToast, showToast } from 'vant'
import type { PageStateStatus } from '@/types/business'
import PpmxButton from '@/components/PpmxButton.vue'
import PpmxRechargePaymentModal from '@/components/PpmxRechargePaymentModal.vue'
import PpmxSelectInput from '@/components/PpmxSelectInput.vue'
import PpmxWithdrawModal from '@/components/PpmxWithdrawModal.vue'

defineOptions({ name: 'PpmxRechargeModal' })

const props = withDefaults(defineProps<{
  modelValue: boolean
}>(), {
  modelValue: false,
})

const emit = defineEmits<{
  close: []
  'update:modelValue': [value: boolean]
}>()

const CRYPTO_CHANNEL_CODE = 'USDT_TRC20'
const rechargeAmounts = [50, 100, 200, 500, 1000, 5000, 10000] as const
const MIN_RECHARGE_AMOUNT = rechargeAmounts[0]

const { t } = useI18n()
const userStore = useUserStore()
const router = useRouter()

const pageStatus = ref<PageStateStatus>('loading')
const pageError = ref('')
const countries = ref<AppCountryItem[]>([])
const selectedCountry = ref<AppCountryItem | null>(null)
const channels = ref<AppRechargeChannelItem[]>([])
const selectedChannelCode = ref('')
const selectedPayMethodCode = ref('')
const rechargeFields = ref<RechargeField[]>([])
const fieldValues = ref<Record<string, string>>({})
const selectedAmount = ref<number>(rechargeAmounts[0])
const selectedWalletType = ref<RechargeWalletType>(DEFAULT_RECHARGE_WALLET_TYPE)
const rechargeFirstStatus = ref<RechargeFirstStatusResp | null>(null)
const formError = ref('')
const isHydratingRechargeFields = ref(false)
const showPaymentModal = ref(false)
const paymentOrder = ref<RechargeOrderAppBack | null>(null)

const selectedChannel = computed(() => (
  channels.value.find(channel => channel.channelCode === selectedChannelCode.value) ?? null
))

const payMethods = computed(() => selectedChannel.value?.methods ?? [])

const selectedPayMethod = computed(() => (
  payMethods.value.find(method => method.methodCode === selectedPayMethodCode.value) ?? null
))
const isSelectedCryptoChannel = computed(() => isCryptoRechargeChannel(selectedChannel.value))
const visibleRechargeFields = computed(() => (
  isSelectedCryptoChannel.value ? [] : rechargeFields.value
))

const amount = computed(() => selectedAmount.value)

const formattedAmount = computed(() => formatMarketAmount(amount.value))
const paymentOrderAmount = computed(() => formatMarketAmount(paymentOrder.value?.amount ?? amount.value))
const isFirstRecharge = computed(() => Boolean(rechargeFirstStatus.value?.isFirstRecharge))
const rechargeNumber = computed(
  () => Number(rechargeFirstStatus.value?.rechargeNumber) || (isFirstRecharge.value ? 1 : 2),
)
const giftRates = computed(() => rechargeFirstStatus.value?.giftRates)
const giftMinAmount = computed(() => rechargeFirstStatus.value?.giftMinAmount)
const isSportsWalletRecharge = computed(() => selectedWalletType.value === 'sport')
const effectiveBonusRate = computed(() => isSportsWalletRecharge.value ? 0 :
  rechargeBonusRateByNumber(amount.value, rechargeNumber.value, giftRates.value, giftMinAmount.value),
)
const effectiveBonusAmount = computed(() => isSportsWalletRecharge.value ? 0 :
  rechargeBonusAmountByNumber(amount.value, rechargeNumber.value, giftRates.value, giftMinAmount.value),
)
const formattedBonusAmount = computed(() => formatMarketAmount(effectiveBonusAmount.value))
const showPromo = computed(() => Boolean(rechargeFirstStatus.value) && !isSportsWalletRecharge.value)
const pageStateMessage = computed(() => {
  if (pageStatus.value === 'loading') return undefined
  if (pageStatus.value === 'empty') return t('recharge.emptyMessage')

  return t('recharge.errorMessage')
})

const rechargeWalletOptions = computed(() => [
  { label: t('recharge.walletBalance'), value: 'balance' },
  { label: t('recharge.walletSport'), value: 'sport' },
])

const selectedRechargeWalletBalance = computed(() => (
  isSportsWalletRecharge.value ? userStore.userInfo?.sportBalance : userStore.userInfo?.balance
))
const balanceText = computed(() => (
  formatMoney(toDisplayCents(selectedRechargeWalletBalance.value), { currency: APP_CURRENCY_SYMBOL })
))

const submitLock = useSubmitLock(submitRecharge, { minLockMs: 600 })
const isSubmitting = computed(() => submitLock.locked.value)

function formatMarketAmount(value: number) {
  if (!Number.isFinite(value) || value <= 0) return `${APP_CURRENCY_SYMBOL}0.00`

  return formatMoney(Math.round(value * 100), { currency: APP_CURRENCY_SYMBOL })
}

function formatRechargeTierDisplay(value: number) {
  return formatMarketAmount(value)
}

function amountBonus(item: number) {
  if (isSportsWalletRecharge.value) return 0

  return rechargeBonusAmountByNumber(item, rechargeNumber.value, giftRates.value, giftMinAmount.value)
}

function amountBonusText(item: number) {
  return `+${formatMarketAmount(amountBonus(item))}`
}

function formatPercent(value: number | undefined) {
  if (!Number.isFinite(value)) return '0%'

  return `${Math.round(Number(value) * 100)}%`
}

function isCryptoRechargeChannel(channel: AppRechargeChannelItem | null | undefined) {
  if (!channel) return false
  const channelCode = String(channel.channelCode || '').trim().toUpperCase()
  const providerType = String(channel.providerType || '').trim().toLowerCase()

  return channelCode === CRYPTO_CHANNEL_CODE || providerType === 'native'
}

function paymentOptionIconClass(channel: AppRechargeChannelItem) {
  if (isCryptoRechargeChannel(channel)) return 'fa-brands fa-bitcoin'

  return 'fa-solid fa-credit-card'
}

function paymentOptionToneClass(channel: AppRechargeChannelItem) {
  if (isCryptoRechargeChannel(channel)) return 'is-crypto'

  return 'is-cash'
}

function paymentOptionSubtext(channel: AppRechargeChannelItem) {
  if (isCryptoRechargeChannel(channel)) return 'USDT'

  return channel.methods?.[0]?.methodName || channel.providerType || ''
}

function normalizeOption(option: unknown): RechargeFieldOption | null {
  if (!option || typeof option !== 'object') return null
  const record = option as Record<string, unknown>
  const label = String(record.label ?? record.name ?? record.value ?? '').trim()
  const value = String(record.value ?? record.code ?? record.label ?? '').trim()

  return label && value ? { label, value } : null
}

function fieldOptions(field: RechargeField) {
  if (!field.optionsJson) return []

  try {
    const parsed = JSON.parse(field.optionsJson) as unknown
    if (!Array.isArray(parsed)) return []

    return parsed.map(normalizeOption).filter(Boolean) as RechargeFieldOption[]
  } catch {
    return []
  }
}

function selectAmount(value: number) {
  selectedAmount.value = value
  formError.value = ''
}

function getFieldInputType(field: RechargeField) {
  return field.fieldType === 'number' || field.dataType === 'number' ? 'number' : 'text'
}

function validateFields() {
  for (const field of visibleRechargeFields.value) {
    const value = String(fieldValues.value[field.fieldKey] ?? '').trim()

    if (field.isRequired && !value) {
      return t('recharge.requiredField', { field: field.fieldLabel })
    }

    if (field.minLength && value && value.length < field.minLength) {
      return field.errorTips || t('recharge.invalidField', { field: field.fieldLabel })
    }

    if (field.maxLength && value && value.length > field.maxLength) {
      return field.errorTips || t('recharge.invalidField', { field: field.fieldLabel })
    }

    if (field.regexRule && value) {
      try {
        const regex = new RegExp(field.regexRule)
        if (!regex.test(value)) {
          return field.errorTips || t('recharge.invalidField', { field: field.fieldLabel })
        }
      } catch {
        return field.errorTips || t('recharge.invalidField', { field: field.fieldLabel })
      }
    }
  }

  return ''
}

function buildExtraFields() {
  const entries = visibleRechargeFields.value
    .map(field => [field.fieldKey, String(fieldValues.value[field.fieldKey] ?? '').trim()] as const)
    .filter(([, value]) => value)

  return entries.length ? Object.fromEntries(entries) : undefined
}

function hydrateRechargeFieldValues(fields: RechargeField[]) {
  isHydratingRechargeFields.value = true
  fieldValues.value = getRechargeFieldInitialValues(APP_COUNTRY_CODE, userStore.userInfo?.id, fields)

  void nextTick(() => {
    isHydratingRechargeFields.value = false
  })
}

function persistRechargeFieldValues() {
  if (isHydratingRechargeFields.value || !visibleRechargeFields.value.length) return

  setCachedRechargeFieldValues(
    APP_COUNTRY_CODE,
    userStore.userInfo?.id,
    visibleRechargeFields.value,
    fieldValues.value,
  )
}

function buildRechargePayload(confirmUnfinishedActivityCycle = false): RechargeOrderAppReq {
  return {
    amount: amount.value,
    channel: selectedChannelCode.value,
    payMethod: selectedPayMethodCode.value || undefined,
    walletType: selectedWalletType.value,
    currency: selectedCountry.value?.currencyCode || APP_CURRENCY,
    countryCode: APP_COUNTRY_CODE,
    merchantOrderNo: buildMerchantOrderNo(),
    extraFields: buildExtraFields(),
    confirmUnfinishedActivityCycle,
  }
}

function buildMerchantOrderNo() {
  const uid = userStore.userInfo?.id ?? 'guest'
  return `mch_${Date.now()}_${uid}`
}

function closeRechargeModal() {
  showPaymentModal.value = false
  paymentOrder.value = null
  emit('update:modelValue', false)
  emit('close')
}

async function goToCryptoPay() {
  closeRechargeModal()
  await router.push({
    path: '/crypto-pay',
    query: {
      amount: amount.value.toFixed(2),
      channelCode: selectedChannelCode.value,
      countryCode: APP_COUNTRY_CODE,
      walletType: selectedWalletType.value,
    },
  })
}

interface RechargeConfig {
  countries: AppCountryItem[]
  channels: AppRechargeChannelItem[]
  fields: RechargeField[]
}

// Static recharge config (countries / channels / fields) rarely changes, so we keep it
// for the session and render it instantly on reopen while dynamic data refreshes in the background.
let cachedRechargeConfig: RechargeConfig | null = null

async function fetchRechargeConfig(): Promise<RechargeConfig> {
  const [appCountries, info, fields] = await Promise.all([
    getAppCountries(),
    getCountryRechargeInfo(APP_COUNTRY_CODE),
    getCountryRechargeFields(APP_COUNTRY_CODE),
  ])

  return {
    countries: appCountries,
    channels: info?.channels ?? [],
    fields: fields ?? [],
  }
}

function applyRechargeConfig(config: RechargeConfig) {
  countries.value = config.countries
  selectedCountry.value = config.countries.find(country => country.countryCode === APP_COUNTRY_CODE) ?? null
  channels.value = config.channels
  rechargeFields.value = config.fields
  hydrateRechargeFieldValues(rechargeFields.value)
  selectedChannelCode.value = config.channels[0]?.channelCode ?? ''
  selectedPayMethodCode.value = ''

  if (!selectedCountry.value) {
    pageStatus.value = 'empty'
    pageError.value = t('recharge.noMarketCountry')
    return
  }

  pageStatus.value = config.channels.length ? 'ready' : 'empty'
  pageError.value = config.channels.length ? '' : t('recharge.noChannel')
}

async function loadRechargeModal() {
  pageError.value = ''
  formError.value = ''

  // Dynamic data (balance + first-recharge status) refreshes in the background and never
  // blocks the form from rendering. Balance shows the cached store value until it resolves.
  void userStore.loadUserInfo().catch(() => {})
  void refreshRechargeFirstStatus().catch(() => {})

  // Stale-while-revalidate: reopen renders instantly from the cached config.
  if (cachedRechargeConfig) {
    applyRechargeConfig(cachedRechargeConfig)
    return
  }

  pageStatus.value = 'loading'

  try {
    const config = await fetchRechargeConfig()
    cachedRechargeConfig = config
    applyRechargeConfig(config)
  } catch (error) {
    pageStatus.value = 'error'
    pageError.value = error instanceof Error ? error.message : t('recharge.loadFailed')
  }
}

async function refreshRechargeFirstStatus() {
  rechargeFirstStatus.value = await getRechargeIsFirstV2()
  return rechargeFirstStatus.value
}

async function syncRechargeNotifications() {
  try {
    const notifications = await getPendingRechargeNotifications()

    await Promise.all(
      notifications.map(notification => ackRechargeNotification(notification.orderNo)),
    )
  } catch {
    // Notification acknowledgement must not block the successful recharge flow.
  }
}

async function createOrderWithConfirm(payload: RechargeOrderAppReq) {
  try {
    const order = await createRechargeOrderV2(payload)
    if (!order.needConfirmUnfinishedActivityCycle) return order

    await confirmUnfinishedActivityCycle()
    return await createRechargeOrderV2({
      ...payload,
      confirmUnfinishedActivityCycle: true,
    })
  } catch (error) {
    const message = error instanceof Error ? error.message : ''
    if (message !== 'unfinished_activity_cycle_confirm_required') throw error

    await confirmUnfinishedActivityCycle()
    return await createRechargeOrderV2({
      ...payload,
      confirmUnfinishedActivityCycle: true,
    })
  }
}

async function confirmUnfinishedActivityCycle() {
  await showConfirmDialog({
    title: t('recharge.unfinishedActivityTitle'),
    message: t('recharge.unfinishedActivityMessage'),
    confirmButtonText: t('recharge.continueRecharge'),
    cancelButtonText: t('recharge.cancelRecharge'),
  })
}

async function confirmPaymentAmountMatch() {
  try {
    await showConfirmDialog({
      title: t('recharge.paymentAmountConfirmTitle'),
      message: t('recharge.paymentAmountConfirmMessage'),
      confirmButtonText: t('recharge.paymentAmountConfirmAction'),
      cancelButtonText: t('recharge.cancelRecharge'),
    })

    return true
  } catch {
    return false
  }
}

async function submitRecharge() {
  formError.value = ''

  if (!selectedCountry.value) throw new Error(t('recharge.noMarketCountry'))
  if (!selectedChannel.value) throw new Error(t('recharge.selectChannel'))
  if (!isSelectedCryptoChannel.value && payMethods.value.length && !selectedPayMethod.value) throw new Error(t('recharge.selectPayMethod'))
  if (!Number.isFinite(amount.value) || amount.value <= 0) throw new Error(t('recharge.invalidAmount'))
  if (amount.value < MIN_RECHARGE_AMOUNT) throw new Error(t('recharge.minAmount', { amount: formatMarketAmount(MIN_RECHARGE_AMOUNT) }))

  if (isSelectedCryptoChannel.value) {
    await trackEvent({
      eventName: 'recharge_submit',
      eventType: 'transaction',
      userId: userStore.userInfo?.id,
      payload: {
        amount: amount.value,
        channel: selectedChannelCode.value,
        payMethod: 'TRC20',
        walletType: selectedWalletType.value,
        bonusRate: effectiveBonusRate.value,
        bonusAmount: effectiveBonusAmount.value,
        isFirstRecharge: isFirstRecharge.value,
        crypto: true,
      },
    })
    await goToCryptoPay()
    return
  }

  const fieldError = validateFields()
  if (fieldError) throw new Error(fieldError)
  persistRechargeFieldValues()

  if (!await confirmPaymentAmountMatch()) return

  await refreshRechargeFirstStatus()
  const payload = buildRechargePayload()

  await trackEvent({
    eventName: 'recharge_submit',
    eventType: 'transaction',
    userId: userStore.userInfo?.id,
    payload: {
      amount: payload.amount,
      channel: payload.channel,
      payMethod: payload.payMethod,
      walletType: payload.walletType,
      bonusRate: effectiveBonusRate.value,
      bonusAmount: effectiveBonusAmount.value,
      isFirstRecharge: isFirstRecharge.value,
    },
  })

  const order = await createOrderWithConfirm(payload)

  await trackEvent({
    eventName: 'recharge_order_created',
    eventType: 'transaction',
    userId: userStore.userInfo?.id,
    payload: {
      orderNo: order.orderNo,
      amount: order.amount,
      channel: order.channel,
      payMethod: order.payMethod,
      walletType: payload.walletType,
      status: order.status,
      hasPayUrl: Boolean(order.payUrl),
      bonusAmount: isSportsWalletRecharge.value ? 0 : order.bonusAmount ?? effectiveBonusAmount.value,
      isFirstRecharge: isFirstRecharge.value,
    },
  })

  if (order.payUrl) {
    paymentOrder.value = order
    showPaymentModal.value = true
    showToast(t('recharge.orderToPay'))
    return
  }

  closeRechargeModal()

  if (order.devCallback) {
    showSuccessToast(t('recharge.orderSuccess'))
    await Promise.allSettled([
      userStore.loadUserInfo(),
      syncRechargeNotifications(),
    ])
    return
  }

  showSuccessToast(t('recharge.orderCreated', { orderNo: order.orderNo }))
}

async function runSubmit() {
  try {
    await submitLock.run()
  } catch (error) {
    const message = error instanceof Error ? error.message : t('recharge.orderFailed')

    formError.value = message
    showFailToast(message)
  }
}

watch(selectedChannelCode, () => {
  selectedPayMethodCode.value = selectedChannel.value?.methods?.[0]?.methodCode ?? ''
})

watch(fieldValues, persistRechargeFieldValues, { deep: true })

watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen) {
      void loadRechargeModal()
      return
    }

    formError.value = ''
    selectedWalletType.value = DEFAULT_RECHARGE_WALLET_TYPE
    showPaymentModal.value = false
    paymentOrder.value = null
  },
)
</script>

<template>
  <PpmxWithdrawModal
    id="rechargeModal"
    :model-value="props.modelValue"
    :eyebrow="APP_CURRENCY"
    title-id="rechargeModalTitle"
    :title="t('recharge.title')"
    :close-aria-label="t('common.close')"
    @update:model-value="emit('update:modelValue', $event)"
    @close="emit('close')"
  >
    <div id="depositModal" class="ppmx-recharge-modal ppmx-recharge-modal--profile">
      <AppState
        v-if="pageStatus !== 'ready'"
        class="ppmx-recharge-state ppmx-recharge-state--modal"
        :status="pageStatus"
        :title="pageError || undefined"
        :message="pageStateMessage"
        :action-text="t('common.retry')"
        skeleton-variant="withdraw-modal"
        @retry="loadRechargeModal"
      />

      <template v-else>
        <section class="ppmx-recharge-balance">
          <span class="ppmx-recharge-balance__icon">
            <i class="fa-solid fa-wallet" />
          </span>
          <p>{{ t('recharge.availableBalance') }}</p>
          <strong>{{ balanceText }} <em>{{ selectedCountry?.currencyCode || APP_CURRENCY }}</em></strong>
          <small>
            <i class="fa-solid fa-location-dot" />
            {{ t('recharge.marketCountry', { country: selectedCountry?.countryNameCn || APP_COUNTRY_NAME_LOCAL }) }} · {{ selectedCountry?.currencyCode || APP_CURRENCY }}
          </small>
        </section>

        <form class="ppmx-recharge-panel ppmx-recharge-form" @submit.prevent="runSubmit">
          <section class="ppmx-recharge-section ppmx-recharge-section--wallet">
            <label class="ppmx-recharge-field">
              <span>{{ t('recharge.walletTypeTitle') }}</span>
              <PpmxSelectInput
                v-model="selectedWalletType"
                :options="rechargeWalletOptions"
                :title="t('recharge.walletTypeTitle')"
              />
            </label>
          </section>

          <section class="ppmx-recharge-section ppmx-recharge-section--amount">
            <div class="ppmx-recharge-panel__head">
              <span class="ppmx-eyebrow">{{ t('recharge.stepAmount') }}</span>
              <h2>{{ t('recharge.amountTitle') }}</h2>
            </div>
            <div class="ppmx-recharge-amount-grid">
              <button
                v-for="item in rechargeAmounts"
                :key="item"
                class="ppmx-recharge-amount"
                :class="{ 'is-active': selectedAmount === item }"
                type="button"
                @click="selectAmount(item)"
              >
                <span v-if="amountBonus(item) > 0" class="ppmx-recharge-amount__bonus">{{ amountBonusText(item) }}</span>
                {{ formatRechargeTierDisplay(item) }}
              </button>
            </div>
          </section>

          <section class="ppmx-recharge-section ppmx-recharge-section--channel">
            <div class="ppmx-recharge-panel__head">
              <span class="ppmx-eyebrow">{{ t('recharge.stepChannel') }}</span>
              <h2>{{ t('recharge.channelTitle') }}</h2>
            </div>
            <div class="ppmx-recharge-choice-grid">
              <button
                v-for="channel in channels"
                :key="channel.channelCode"
                class="ppmx-recharge-choice"
                :class="{ 'is-active': selectedChannelCode === channel.channelCode }"
                :data-testid="`pay-channel-${channel.channelCode === CRYPTO_CHANNEL_CODE ? 'crypto' : channel.channelCode}`"
                type="button"
                @click="selectedChannelCode = channel.channelCode"
              >
                <span class="ppmx-recharge-option-icon" :class="paymentOptionToneClass(channel)">
                  <i :class="paymentOptionIconClass(channel)" />
                </span>
                <span class="ppmx-recharge-option-copy">
                  <strong>{{ channel.channelName }}</strong>
                  <small>{{ paymentOptionSubtext(channel) }}</small>
                </span>
                <span class="ppmx-recharge-option-check" aria-hidden="true">
                  <i class="fa-solid fa-check" />
                </span>
              </button>
            </div>
          </section>

          <section v-if="visibleRechargeFields.length" class="ppmx-recharge-section ppmx-recharge-section--fields">
            <div class="ppmx-recharge-panel__head">
              <span class="ppmx-eyebrow">{{ t('recharge.stepInfo') }}</span>
              <h2>{{ t('recharge.fillInfo') }}</h2>
            </div>
            <div class="ppmx-recharge-fields">
              <label
                v-for="field in visibleRechargeFields"
                :key="field.fieldKey"
                class="ppmx-recharge-field"
              >
                <span>
                  {{ field.fieldLabel }}
                  <em v-if="field.isRequired">*</em>
                </span>
                <textarea
                  v-if="field.fieldType === 'textarea'"
                  v-model="fieldValues[field.fieldKey]"
                  :maxlength="field.maxLength || undefined"
                  :placeholder="field.fieldPlaceholder || ''"
                  rows="3"
                />
                <PpmxSelectInput
                  v-else-if="field.fieldType === 'select'"
                  v-model="fieldValues[field.fieldKey]"
                  :options="fieldOptions(field)"
                  :placeholder="field.fieldPlaceholder || t('recharge.selectPlaceholder')"
                  :title="field.fieldLabel"
                />
                <input
                  v-else
                  v-model="fieldValues[field.fieldKey]"
                  :type="getFieldInputType(field)"
                  :inputmode="field.fieldType === 'number' ? 'decimal' : 'text'"
                  :maxlength="field.maxLength || undefined"
                  :placeholder="field.fieldPlaceholder || ''"
                >
              </label>
            </div>
          </section>

          <section v-if="showPromo" class="ppmx-recharge-section ppmx-recharge-promo">
            <div class="ppmx-recharge-panel__head">
              <span class="ppmx-eyebrow">
                <i class="fa-solid fa-bolt" />
                {{ t('recharge.promoEyebrow') }}
              </span>
              <h2>
                {{ isFirstRecharge ? t('recharge.firstRechargeBonusTitle') : t('recharge.afterFirstRechargeBonusTitle') }}
                <em>{{ formatPercent(effectiveBonusRate) }}</em>
              </h2>
            </div>
            <div class="ppmx-recharge-promo__list">
              <article class="ppmx-recharge-promo-card is-active">
                <i class="fa-solid fa-crown" />
                <span>
                  <strong>{{ t('recharge.bonusRateTitle', { rate: formatPercent(effectiveBonusRate) }) }}</strong>
                  <small>{{ t('recharge.bonusAmountDesc', { amount: formattedBonusAmount }) }}</small>
                </span>
                <b>{{ formattedBonusAmount }}</b>
              </article>
            </div>
          </section>

          <p v-if="formError" class="ppmx-recharge-error">
            <i class="fa-solid fa-circle-exclamation" />
            {{ formError }}
          </p>

          <footer class="ppmx-recharge-submit">
            <div>
              <span>{{ t('recharge.payAmount') }}</span>
              <strong>{{ formattedAmount }} <em>{{ selectedCountry?.currencyCode || APP_CURRENCY }}</em></strong>
              <small v-if="showPromo">
                {{ t('recharge.bonusAmountDesc', { amount: formattedBonusAmount }) }}
              </small>
            </div>
            <PpmxButton
              block
              data-testid="recharge-submit"
              size="lg"
              type="submit"
              :loading="isSubmitting"
              :disabled="isSubmitting"
            >
              {{ t('recharge.submit') }}
              <i class="fa-solid fa-arrow-right" />
            </PpmxButton>
          </footer>
        </form>

        <details class="wd-rules ppmx-recharge-rules">
          <summary>
            <span>
              <i class="fa-solid fa-circle-info" />
              {{ t('recharge.rulesTitle') }}
            </span>
            <i class="fa-solid fa-chevron-down" />
          </summary>
          <ul>
            <li>{{ t('recharge.rule1') }}</li>
            <li>{{ t('recharge.rule2') }}</li>
            <li>{{ t('recharge.rule3') }}</li>
          </ul>
        </details>
      </template>
    </div>
  </PpmxWithdrawModal>

  <!-- Legacy PpmxBaseModal payment popup retired; the depositModal-style popup below is the active payment flow. -->
  <PpmxRechargePaymentModal
    v-model="showPaymentModal"
    :order="paymentOrder"
    :amount-text="paymentOrderAmount"
    :method-text="paymentOrder?.payMethod || selectedPayMethod?.methodName || selectedPayMethodCode || '-'"
    :currency-fallback="APP_CURRENCY"
  />
</template>
