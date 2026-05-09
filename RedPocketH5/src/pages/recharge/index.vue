<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showConfirmDialog, showToast } from 'vant'
import type {
  AppCountryItem,
  AppPayMethodItem,
  AppRechargeChannelItem,
  RechargeField,
  RechargeFieldOption,
  RechargeFirstRecharge3DayPromotion,
  RechargeOrderAppReq,
  RechargeTodayFirstPromotion,
} from '@/api/user'
import {
  ackRechargeNotification,
  createRechargeOrder,
  getAppCountries,
  getCountryRechargeFields,
  getCountryRechargeInfo,
  getCurrentTgUserInfo,
  getPendingRechargeNotifications,
  getRechargePromotions,
} from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { useUserStore } from '@/stores'
import { formatCurrency, truncate2 } from '@/utils/currency'
import { trackFirstRechargePurchase } from '@/utils/facebook-pixel'
import { safeBack } from '@/utils/navigation'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()

const balance = ref(0)
const currentUserCountry = ref(String(userStore.userInfo?.country || '').trim())
const hideCountrySelector = ref(false)
const pageLoading = ref(true)

// 促销活动
// promoChoice: '' = 不参加, 'first' = 首充, 'today_first' = 今日首充
const firstRecharge3Day = ref<RechargeFirstRecharge3DayPromotion | null>(null)
const todayFirstRecharge = ref<RechargeTodayFirstPromotion | null>(null)
const promoChoice = ref<'' | 'first' | 'today_first'>('')

const hasFirst = computed(() => !!firstRecharge3Day.value?.visible)
const hasTodayFirst = computed(() => !!todayFirstRecharge.value?.visible)
const canSelectFirst = computed(() => !!firstRecharge3Day.value?.visible && !!firstRecharge3Day.value?.selectable)
const canSelectTodayFirst = computed(() => !!todayFirstRecharge.value?.visible && !!todayFirstRecharge.value?.selectable)
const showPromo = computed(() => hasFirst.value || hasTodayFirst.value)
const selectedActivityCode = computed(() => {
  if (promoChoice.value === 'first' && canSelectFirst.value)
    return firstRecharge3Day.value?.activityCode ?? ''
  if (promoChoice.value === 'today_first' && canSelectTodayFirst.value)
    return todayFirstRecharge.value?.activityCode ?? ''
  return ''
})

function formatPercent(value: number) {
  if (!Number.isFinite(value))
    return '0'
  return Number(truncate2(value).toFixed(2)).toString()
}

const firstRechargeGiftView = computed(() => {
  const promo = firstRecharge3Day.value
  const rates = Array.isArray(promo?.rates) ? promo.rates : []
  if (!promo || rates.length === 0) {
    return {
      badge: '',
      desc: t('rechargePage.firstRechargeDesc'),
      detail: '',
    }
  }

  const dayRate = (day: number) => Number(rates.find(item => Number(item.day) === day)?.rate || 0)
  return {
    badge: promo.todayRate > 0 ? `+${formatPercent(Number(promo.todayRate))}%` : '',
    desc: t('rechargePage.firstRechargeV2Desc', {
      day1Rate: formatPercent(dayRate(1)),
      day2Rate: formatPercent(dayRate(2)),
      day3Rate: formatPercent(dayRate(3)),
    }),
    detail: t('rechargePage.firstRechargeV2Detail'),
  }
})

async function loadFirstRechargeStatus() {
  try {
    const { data } = await getRechargePromotions()
    firstRecharge3Day.value = data?.firstRecharge3Day ?? null
    todayFirstRecharge.value = data?.todayFirstRecharge ?? null

    if (canSelectFirst.value)
      promoChoice.value = 'first'
    else if (canSelectTodayFirst.value)
      promoChoice.value = 'today_first'
    else
      promoChoice.value = ''
  }
  catch { /* 接口失败不影响主流程 */ }
}

function togglePromo(choice: 'first' | 'today_first') {
  if (choice === 'first' && !canSelectFirst.value)
    return
  if (choice === 'today_first' && !canSelectTodayFirst.value)
    return
  promoChoice.value = promoChoice.value === choice ? '' : choice
}

// 国家列表
const countries = ref<AppCountryItem[]>([])
const selectedCountry = ref<AppCountryItem | null>(null)

// 充值通道与支付方式（按国家动态加载）
const channels = ref<AppRechargeChannelItem[]>([])
const selectedChannel = ref<AppRechargeChannelItem | null>(null)
const payMethods = computed<AppPayMethodItem[]>(() => selectedChannel.value?.methods ?? [])
const selectedPay = ref<AppPayMethodItem | null>(null)
const rechargeInfoLoading = ref(false)

// 充值字段
const rechargeFields = ref<RechargeField[]>([])
const fieldValues = ref<Record<string, string>>({})

// select 选择器
const pickerVisible = ref(false)
const pickerField = ref<RechargeField | null>(null)
const pickerColumns = computed(() =>
  parseFieldOptions(pickerField.value).map(o => ({ text: o.label, value: o.value })),
)

function parseFieldOptions(field: RechargeField | null): RechargeFieldOption[] {
  if (!field?.optionsJson)
    return []
  try {
    return JSON.parse(field.optionsJson)
  }
  catch {
    return []
  }
}

function getSelectLabel(field: RechargeField): string {
  const opts = parseFieldOptions(field)
  return opts.find(o => o.value === fieldValues.value[field.fieldKey])?.label
    ?? fieldValues.value[field.fieldKey]
    ?? ''
}

function openPicker(field: RechargeField) {
  pickerField.value = field
  pickerVisible.value = true
}

function onPickerConfirm({ selectedOptions }: { selectedOptions: Array<{ text: string, value: string }> }) {
  if (pickerField.value)
    fieldValues.value[pickerField.value.fieldKey] = selectedOptions[0]?.value ?? ''
  pickerVisible.value = false
}

const amountOptions = [20, 30, 50, 100, 200, 500, 1000, 5000, 10000, 'custom']
const selectedAmount = ref<number | 'custom'>(amountOptions[0] as number)
const customAmount = ref('')
const submitLoading = ref(false)

const displayAmount = computed(() => {
  if (selectedAmount.value === 'custom')
    return customAmount.value ? Number(customAmount.value) : 0
  return selectedAmount.value
})

const canSubmit = computed(() =>
  Number(displayAmount.value) > 0
  && !submitLoading.value
  && !!selectedChannel.value
  && !rechargeInfoLoading.value,
)

const localAmount = computed(() => {
  const coins = Number(displayAmount.value) || 0
  return formatLocalAmount(coins)
})

const localCurrencySymbol = computed(() => selectedCountry.value?.currencySymbol || '')

function chooseAmount(value: number | 'custom') {
  selectedAmount.value = value
  if (value !== 'custom')
    customAmount.value = ''
}

function selectChannel(ch: AppRechargeChannelItem) {
  selectedChannel.value = ch
  selectedPay.value = ch.methods[0] ?? null
}

function goBack() {
  safeBack(router)
}

function showCenterToast(message: string) {
  showToast({ message, position: 'middle', teleport: '#app', wordBreak: 'break-word' })
}

function showHelpTip() {
  showCenterToast(t('rechargePage.helpTip'))
}

function openPaymentUrl(url: string) {
  const opened = window.open(url, '_blank', 'noopener,noreferrer')
  if (opened) {
    opened.opener = null
    return
  }
  window.location.href = url
}

function formatLocalAmount(coins: number) {
  const rate = selectedCountry.value?.rate || 1
  return truncate2((Number(coins) || 0) * rate).toFixed(2)
}

async function loadBalance() {
  try {
    const { data } = await getCurrentTgUserInfo()
    balance.value = Number(data?.balance ?? 0)
  }
  catch {
    balance.value = 0
  }
}

async function loadRechargeInfo(code: string) {
  rechargeInfoLoading.value = true
  channels.value = []
  selectedChannel.value = null
  selectedPay.value = null
  rechargeFields.value = []
  fieldValues.value = {}
  try {
    const [infoRes, fieldsRes] = await Promise.all([
      getCountryRechargeInfo(code),
      getCountryRechargeFields(code),
    ])
    channels.value = infoRes.data?.channels ?? []
    if (channels.value.length)
      selectChannel(channels.value[0])
    const fields = fieldsRes.data ?? []
    rechargeFields.value = fields
    const init: Record<string, string> = {}
    for (const f of fields)
      init[f.fieldKey] = f.defaultValue ?? ''
    fieldValues.value = init
  }
  catch {
    // ignore
  }
  finally {
    rechargeInfoLoading.value = false
  }
}

async function handleSelectCountry(country: AppCountryItem) {
  if (selectedCountry.value?.countryCode === country.countryCode)
    return
  selectedCountry.value = country
  await loadRechargeInfo(country.countryCode)
}

async function loadCountries() {
  hideCountrySelector.value = false
  try {
    const { data } = await getAppCountries()
    countries.value = data ?? []
    if (countries.value.length) {
      const normalizedUserCountry = currentUserCountry.value.toUpperCase()
      const matchedCountry = normalizedUserCountry
        ? countries.value.find(item => item.countryCode.toUpperCase() === normalizedUserCountry) ?? null
        : null
      selectedCountry.value = matchedCountry ?? countries.value[0]
      hideCountrySelector.value = !!matchedCountry
      void loadRechargeInfo(selectedCountry.value.countryCode)
    }
  }
  catch {
    countries.value = []
    hideCountrySelector.value = false
  }
}

async function initPage() {
  pageLoading.value = true
  try {
    await loadCountries()
  }
  finally {
    pageLoading.value = false
    void loadBalance()
    void loadFirstRechargeStatus()
  }
}

async function handleSubmitRecharge() {
  if (!canSubmit.value)
    return
  const amount = truncate2(Number(displayAmount.value))
  if (!amount || amount <= 0) {
    showCenterToast(t('rechargePage.invalidAmount'))
    return
  }

  // 校验必填 + 正则
  for (const f of rechargeFields.value) {
    const val = fieldValues.value[f.fieldKey]?.trim() ?? ''
    const tip = f.errorTips || ''
    if (f.isRequired === 1 && !val) {
      showCenterToast(tip || t('rechargePage.requiredField', { field: f.fieldLabel }))
      return
    }
    if (val && f.regexRule) {
      const regex = new RegExp(f.regexRule)
      if (!regex.test(val)) {
        showCenterToast(tip || t('rechargePage.invalidField', { field: f.fieldLabel }))
        return
      }
    }
  }

  submitLoading.value = true
  try {
    const req: RechargeOrderAppReq = {
      amount,
      channel: selectedChannel.value?.channelCode ?? '',
      payMethod: selectedPay.value?.methodCode ?? '',
      currency: selectedCountry.value?.currencyCode,
      countryCode: selectedCountry.value?.countryCode ?? '',
      extraFields: rechargeFields.value.length ? { ...fieldValues.value } : undefined,
      activityCode: selectedActivityCode.value,
    }
    const { data } = await submitRechargeOrder(req)

    if (data?.payUrl) {
      showCenterToast(t('rechargePage.orderToPay'))
      openPaymentUrl(data.payUrl)
      return
    }

    if (data?.devCallback) {
      showCenterToast(t('rechargePage.orderRechargeSuccess', { orderNo: data.orderNo }))
      await syncPendingRechargeNotificationsForPixel()
      await loadBalance()
      return
    }

    showCenterToast(t('rechargePage.orderSuccess', { orderNo: data?.orderNo || '--' }))
  }
  catch (error) {
    if (error instanceof Error && error.message === 'recharge_confirm_cancelled')
      return
    showCenterToast(t('rechargePage.orderFailed'))
  }
  finally {
    submitLoading.value = false
  }
}

async function submitRechargeOrder(req: RechargeOrderAppReq) {
  try {
    const response = await createRechargeOrder(req)
    if (!response.data?.needConfirmUnfinishedActivityCycle)
      return response

    await confirmUnfinishedActivityCycle()
    return await createRechargeOrder({
      ...req,
      confirmUnfinishedActivityCycle: true,
    })
  }
  catch (error) {
    const message = error instanceof Error ? error.message : ''
    if (message !== 'unfinished_activity_cycle_confirm_required')
      throw error

    await confirmUnfinishedActivityCycle()
    return await createRechargeOrder({
      ...req,
      confirmUnfinishedActivityCycle: true,
    })
  }
}

async function confirmUnfinishedActivityCycle() {
  try {
    await showConfirmDialog({
      title: t('rechargePage.unfinishedActivityTitle'),
      message: t('rechargePage.unfinishedActivityMessage'),
      confirmButtonText: t('rechargePage.continueRecharge'),
      cancelButtonText: t('rechargePage.cancelRecharge'),
    })
  }
  catch {
    throw new Error('recharge_confirm_cancelled')
  }
}

async function syncPendingRechargeNotificationsForPixel() {
  try {
    const { data } = await getPendingRechargeNotifications()
    for (const item of data || []) {
      if (item.isFirstRecharge) {
        trackFirstRechargePurchase({
          orderNo: item.orderNo,
          amount: Number(item.amount || 0),
          currency: item.currency || 'BRL',
        })
      }
      await ackRechargeNotification(item.orderNo)
    }
  }
  catch (error) {
    console.warn('[recharge pixel] sync pending failed:', error)
  }
}

onMounted(() => {
  void initPage()
})
</script>

<template>
  <div class="recharge-page">
    <AppPageHeader class="recharge-header" :title="t('rechargePage.title')" @back="goBack" @right-click="showHelpTip">
      <template #right>
        <van-icon name="question-o" />
      </template>
    </AppPageHeader>

    <div v-if="pageLoading" class="recharge-skeleton" aria-hidden="true">
      <div class="skeleton-country-row">
        <span v-for="item in 4" :key="`country-${item}`" class="skeleton-pill" />
      </div>

      <section class="card balance-card skeleton-balance-card">
        <div>
          <span class="skeleton-line skeleton-line--label" />
          <span class="skeleton-line skeleton-line--value" />
        </div>
        <span class="skeleton-chip" />
      </section>

      <section class="card skeleton-card">
        <span class="skeleton-line skeleton-line--title" />
        <div class="skeleton-grid skeleton-grid--two">
          <span v-for="item in 2" :key="`channel-${item}`" class="skeleton-option" />
        </div>
      </section>

      <section class="card skeleton-card">
        <span class="skeleton-line skeleton-line--title" />
        <div class="skeleton-grid skeleton-grid--three">
          <span v-for="item in 6" :key="`amount-${item}`" class="skeleton-option skeleton-option--amount" />
        </div>
      </section>

      <section class="card skeleton-card">
        <span class="skeleton-line skeleton-line--title" />
        <div class="skeleton-list">
          <span v-for="item in 2" :key="`pay-${item}`" class="skeleton-row" />
        </div>
      </section>

      <section class="card skeleton-card">
        <span class="skeleton-line skeleton-line--title" />
        <div class="skeleton-list">
          <span v-for="item in 3" :key="`field-${item}`" class="skeleton-field" />
        </div>
      </section>

      <span class="skeleton-submit" />
    </div>

    <template v-else>
      <!-- 国家选择行 -->
      <div v-if="!hideCountrySelector" class="country-bar">
        <div class="country-scroll">
          <button
            v-for="country in countries" :key="country.countryCode" type="button" class="country-pill"
            :class="{ active: selectedCountry?.countryCode === country.countryCode }"
            @click="handleSelectCountry(country)"
          >
            <span class="country-code">{{ country.countryCode }}</span>
            <span class="country-name">{{ country.countryNameEn }}</span>
          </button>
        </div>
      </div>

      <section class="card balance-card">
        <div>
          <p class="card-label">
            {{ t('rechargePage.currentBalance') }}
          </p>
          <p class="card-value">
            <CoinAmount :text="formatCurrency(balance)" />
          </p>
        </div>
        <span class="card-chip"><img class="chip-coin" src="@/assets/svg/coin.svg" alt=""></span>
      </section>

      <!-- 充值通道 -->
      <section class="card">
        <h2>{{ t('rechargePage.channelTitle') }}</h2>
        <van-loading v-if="rechargeInfoLoading" size="20" color="#d4af37" class="section-loading" />
        <template v-else>
          <div v-if="channels.length" class="pill-group">
            <button
              v-for="ch in channels" :key="ch.channelCode" type="button" class="pill"
              :class="{ active: selectedChannel?.channelCode === ch.channelCode }" @click="selectChannel(ch)"
            >
              <img v-if="ch.icon" :src="ch.icon" class="channel-icon" alt="">
              {{ ch.channelName }}
            </button>
          </div>
          <p v-else class="empty-tip">
            {{ t('rechargePage.noChannel') }}
          </p>
        </template>
      </section>

      <section class="card">
        <h2>{{ t('rechargePage.amountTitle') }}</h2>
        <div class="amount-grid">
          <button
            v-for="item in amountOptions" :key="item" type="button" class="amount-item"
            :class="{ active: selectedAmount === item }" @click="chooseAmount(item as number | 'custom')"
          >
            <span v-if="item !== 'custom'">{{ item }}</span>
            <span v-else>{{ t('rechargePage.custom') }}</span>
            <span v-if="item !== 'custom' && selectedCountry?.rate" class="amount-local">
              {{ localCurrencySymbol }} {{ formatLocalAmount(Number(item)) }}
            </span>
            <span v-if="item === 'custom' && selectedCountry?.rate && Number(customAmount) > 0" class="amount-local">
              {{ localCurrencySymbol }} {{ formatLocalAmount(Number(customAmount)) }}
            </span>
          </button>
        </div>
        <van-field
          v-if="selectedAmount === 'custom'" v-model="customAmount" type="number"
          :label="t('rechargePage.customAmount')" :placeholder="t('rechargePage.customAmountPlaceholder')"
          class="custom-input"
        />
      </section>

      <!-- 支付方式 -->
      <section v-if="payMethods.length" class="card">
        <h2>{{ t('rechargePage.payMethodTitle') }}</h2>
        <div class="pay-list">
          <button
            v-for="method in payMethods" :key="method.methodCode" type="button" class="pay-item"
            :class="{ active: selectedPay?.methodCode === method.methodCode }" @click="selectedPay = method"
          >
            <div class="pay-left">
              <div class="pay-logo">
                <img v-if="method.icon" :src="method.icon" class="pay-logo-img" alt="">
                <span v-else>{{ method.methodName.charAt(0) }}</span>
              </div>
              <div>
                <p class="pay-name">
                  {{ method.methodName }}
                </p>
                <p class="pay-sub">
                  {{ method.methodCode }}
                </p>
              </div>
            </div>
            <span class="pay-check">
              <van-icon name="success" />
            </span>
          </button>
        </div>
      </section>

      <!-- 充值字段 -->
      <section v-if="rechargeFields.length" class="card">
        <h2>{{ t('rechargePage.fillInfo') }}</h2>
        <template v-for="field in rechargeFields" :key="field.fieldKey">
          <van-field
            v-if="field.fieldType === 'select'"
            :model-value="getSelectLabel(field)"
            :label="field.fieldLabel"
            :placeholder="field.fieldPlaceholder || ''"
            :required="field.isRequired === 1"
            is-link
            readonly
            class="custom-input recharge-field"
            @click="openPicker(field)"
          />
          <van-field
            v-else
            v-model="fieldValues[field.fieldKey]"
            :type="field.fieldType === 'number' ? 'number' : field.fieldType === 'textarea' ? 'textarea' : 'text'"
            :label="field.fieldLabel"
            :placeholder="field.fieldPlaceholder || ''"
            :required="field.isRequired === 1"
            :maxlength="field.maxLength ?? undefined"
            class="custom-input recharge-field"
            rows="3"
          />
        </template>
      </section>

      <!-- select 选择器弹窗 -->
      <van-popup v-model:show="pickerVisible" position="bottom" teleport="#app">
        <van-picker
          :columns="pickerColumns"
          @confirm="onPickerConfirm"
          @cancel="pickerVisible = false"
        />
      </van-popup>

      <!-- 促销活动选择 -->
      <section v-if="showPromo" class="card">
        <h2>{{ t('rechargePage.promoTitle') }}</h2>
        <div class="promo-list">
          <!-- 首充活动 -->
          <button
            v-if="hasFirst"
            type="button"
            class="promo-item"
            :class="{ active: promoChoice === 'first', disabled: !canSelectFirst }"
            :disabled="!canSelectFirst"
            @click="togglePromo('first')"
          >
            <span class="promo-radio">
              <span v-if="promoChoice === 'first'" class="promo-radio-dot" />
            </span>
            <div class="promo-body">
              <div class="promo-header-row">
                <span class="promo-name">{{ t('rechargePage.firstRechargeTitle') }}</span>
                <span v-if="firstRechargeGiftView.badge" class="promo-badge">{{ firstRechargeGiftView.badge }}</span>
              </div>
              <p class="promo-desc">
                {{ firstRechargeGiftView.desc }}
              </p>
              <p v-if="firstRechargeGiftView.detail" class="promo-desc promo-desc--sub">
                {{ firstRechargeGiftView.detail }}
              </p>
            </div>
          </button>

          <!-- 今日首充活动 -->
          <button
            v-if="hasTodayFirst"
            type="button"
            class="promo-item"
            :class="{ active: promoChoice === 'today_first', disabled: !canSelectTodayFirst }"
            :disabled="!canSelectTodayFirst"
            @click="togglePromo('today_first')"
          >
            <span class="promo-radio">
              <span v-if="promoChoice === 'today_first'" class="promo-radio-dot" />
            </span>
            <div class="promo-body">
              <div class="promo-header-row">
                <span class="promo-name">{{ t('rechargePage.todayFirstTitle') }}</span>
                <span v-if="todayFirstRecharge?.rate" class="promo-badge">+{{ formatPercent(Number(todayFirstRecharge.rate)) }}%</span>
              </div>
              <p class="promo-desc">
                {{ t('rechargePage.todayFirstDesc') }}
              </p>
            </div>
          </button>

          <!-- 不参加 -->
          <button
            type="button"
            class="promo-item"
            :class="{ active: promoChoice === '' }"
            @click="promoChoice = ''"
          >
            <span class="promo-radio">
              <span v-if="promoChoice === ''" class="promo-radio-dot" />
            </span>
            <div class="promo-body">
              <span class="promo-name promo-name--muted">{{ t('rechargePage.promoNone') }}</span>
            </div>
          </button>
        </div>
      </section>

      <div v-if="displayAmount && selectedCountry?.rate" class="local-amount-hint">
        ≈ {{ localCurrencySymbol }}{{ localAmount }} {{ selectedCountry?.currencyCode }}
      </div>

      <van-button
        type="primary" round block class="confirm-btn" :loading="submitLoading" :disabled="!canSubmit"
        @click="handleSubmitRecharge"
      >
        {{ t('rechargePage.submit') }}
      </van-button>
    </template>
  </div>
</template>

<route lang="json5">
{
  name: "Recharge"
}
</route>

<style scoped>
.recharge-page {
  min-height: 100vh;
  background-image:
    radial-gradient(circle at 18% 12%, rgba(212, 175, 55, 0.18), transparent 28%),
    radial-gradient(circle at 82% 84%, rgba(255, 215, 0, 0.12), transparent 24%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.04) 18px,
      rgba(212, 175, 55, 0.04) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #240000 60%, #150000 100%);
  color: #fff0c9;
  padding: 0 12px calc(90px + env(safe-area-inset-bottom));
}

.recharge-header {
  margin-bottom: 10px;
}

.recharge-skeleton {
  display: flex;
  flex-direction: column;
}

.skeleton-country-row {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  overflow: hidden;
}

.skeleton-pill,
.skeleton-line,
.skeleton-chip,
.skeleton-option,
.skeleton-row,
.skeleton-field,
.skeleton-submit {
  position: relative;
  overflow: hidden;
  background: rgba(255, 248, 214, 0.08);
}

.skeleton-pill::after,
.skeleton-line::after,
.skeleton-chip::after,
.skeleton-option::after,
.skeleton-row::after,
.skeleton-field::after,
.skeleton-submit::after {
  content: '';
  position: absolute;
  inset: 0;
  transform: translateX(-100%);
  background: linear-gradient(90deg, transparent, rgba(255, 248, 214, 0.12), transparent);
  animation: skeleton-shimmer 1.4s infinite;
}

.skeleton-pill {
  flex: 0 0 86px;
  height: 30px;
  border-radius: 20px;
}

.skeleton-balance-card {
  min-height: 86px;
}

.skeleton-line {
  display: block;
  border-radius: 999px;
}

.skeleton-line--label {
  width: 96px;
  height: 13px;
  margin-bottom: 12px;
}

.skeleton-line--value {
  width: 132px;
  height: 28px;
}

.skeleton-line--title {
  width: 120px;
  height: 14px;
  margin-bottom: 14px;
}

.skeleton-chip {
  width: 44px;
  height: 44px;
  border-radius: 14px;
}

.skeleton-grid {
  display: grid;
  gap: 10px;
}

.skeleton-grid--two {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.skeleton-grid--three {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.skeleton-option {
  height: 46px;
  border-radius: 14px;
}

.skeleton-option--amount {
  height: 54px;
}

.skeleton-list {
  display: grid;
  gap: 10px;
}

.skeleton-row {
  height: 64px;
  border-radius: 16px;
}

.skeleton-field {
  height: 70px;
  border-radius: 14px;
}

.skeleton-submit {
  display: block;
  height: 44px;
  margin-top: 10px;
  border-radius: 999px;
  background: rgba(212, 175, 55, 0.24);
}

/* 国家选择条 */
.country-bar {
  margin-bottom: 12px;
  overflow: hidden;
}

.country-scroll {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
  overscroll-behavior-x: contain;
  touch-action: pan-x;
  scroll-snap-type: x proximity;
}

.country-scroll::-webkit-scrollbar {
  display: none;
}

.country-pill {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid rgba(212, 175, 55, 0.28);
  background: rgba(255, 248, 214, 0.06);
  color: rgba(255, 240, 201, 0.7);
  font-size: 13px;
  white-space: nowrap;
  transition: all 0.15s;
  scroll-snap-align: start;
}

.country-pill.active {
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  border-color: transparent;
  color: #5a1b00;
  font-weight: 700;
  box-shadow: 0 4px 12px rgba(75, 25, 0, 0.3);
}

.country-code {
  font-weight: 700;
  font-size: 12px;
}

.country-name {
  font-size: 12px;
}

.card {
  position: relative;
  overflow: hidden;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  border-radius: 18px;
  padding: 14px 16px;
  border: 1px solid rgba(212, 175, 55, 0.34);
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.3),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
  margin-bottom: 12px;
}

.card::after {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}

.balance-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-label {
  font-size: 13px;
  color: rgba(255, 229, 186, 0.62);
  margin: 0 0 6px;
}

.card-value {
  font-size: 26px;
  font-weight: 800;
  margin: 0;
  color: #ffd87f;
  letter-spacing: 0.04em;
}

.card-chip {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px;
  box-shadow: 0 8px 18px rgba(75, 25, 0, 0.24);
}

.chip-coin {
  width: 100%;
  height: 100%;
  display: block;
}

.card h2 {
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #ffd98b;
  margin: 0 0 12px;
}

.section-loading {
  display: flex;
  justify-content: center;
  padding: 8px 0;
}

.empty-tip {
  font-size: 13px;
  color: rgba(255, 229, 186, 0.45);
  text-align: center;
  margin: 4px 0 0;
}

.pill-group {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.pill {
  flex: 1;
  min-width: 80px;
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.06);
  padding: 12px 8px;
  font-weight: 700;
  color: #fff0c9;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.pill.active {
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  border-color: rgba(255, 248, 214, 0.34);
  box-shadow: 0 8px 18px rgba(75, 25, 0, 0.24);
}

.channel-icon {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  object-fit: contain;
}

.amount-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.amount-item {
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.06);
  padding: 12px 0;
  font-weight: 700;
  color: #fff0c9;
  font-size: 14px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
}

.amount-local {
  font-size: 11px;
  font-weight: 400;
  color: rgba(255, 229, 186, 0.55);
  line-height: 1;
}

.amount-item.active .amount-local {
  color: rgba(255, 229, 186, 0.7);
}

.amount-item.active {
  border-color: rgba(255, 248, 214, 0.34);
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.18), rgba(116, 24, 0, 0.28));
  color: #ffd87f;
}

.custom-input {
  margin-top: 12px;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(212, 175, 55, 0.18);
  background: rgba(255, 248, 214, 0.05);
}

:deep(.custom-input .van-field__label),
:deep(.custom-input .van-field__control) {
  color: #fff0c9;
}

:deep(.custom-input .van-field__control::placeholder) {
  color: rgba(255, 229, 186, 0.4);
}

.pay-list {
  display: grid;
  gap: 10px;
}

.pay-item {
  width: 100%;
  border-radius: 16px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.05);
  padding: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  text-align: left;
}

.pay-item.active {
  border-color: rgba(255, 248, 214, 0.34);
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.16), rgba(116, 24, 0, 0.24));
  box-shadow: 0 8px 18px rgba(75, 25, 0, 0.24);
}

.pay-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pay-logo {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  overflow: hidden;
  flex-shrink: 0;
}

.pay-logo-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.pay-name {
  font-weight: 700;
  margin: 0;
  color: #fff0c9;
}

.pay-sub {
  margin: 2px 0 0;
  font-size: 12px;
  color: rgba(255, 229, 186, 0.58);
}

.pay-check {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 1px solid rgba(212, 175, 55, 0.24);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #ffd87f;
  background: rgba(255, 248, 214, 0.08);
}

.pay-item:not(.active) .pay-check {
  color: transparent;
}

.recharge-field {
  display: block;
  margin-top: 10px;
}

.recharge-field:first-of-type {
  margin-top: 0;
}

:deep(.recharge-field .van-field__label) {
  width: 100%;
  margin: 0 0 6px;
  font-size: 12px;
  line-height: 1.2;
}

:deep(.recharge-field .van-field__label span) {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.recharge-field .van-field__value) {
  width: 100%;
  min-width: 0;
}

:deep(.recharge-field .van-field__body),
:deep(.recharge-field .van-field__control) {
  width: 100%;
}

.promo-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.promo-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  width: 100%;
  padding: 14px 12px;
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.05);
  text-align: left;
  transition: all 0.15s;
}

.promo-item.active {
  border-color: rgba(212, 175, 55, 0.6);
  background: linear-gradient(160deg, rgba(255, 223, 135, 0.1), rgba(116, 24, 0, 0.2));
}

.promo-item.disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.promo-radio {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid rgba(212, 175, 55, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 1px;
}

.promo-item.active .promo-radio {
  border-color: #d4af37;
}

.promo-radio-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #d4af37;
}

.promo-body {
  flex: 1;
  min-width: 0;
}

.promo-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}

.promo-name {
  font-size: 14px;
  font-weight: 700;
  color: #fff0c9;
}

.promo-name--muted {
  color: rgba(255, 229, 186, 0.6);
  font-weight: 400;
}

.promo-badge {
  font-size: 13px;
  font-weight: 700;
  color: #ffd87f;
  white-space: nowrap;
}

.promo-desc {
  margin: 0;
  font-size: 12px;
  color: rgba(255, 229, 186, 0.55);
  line-height: 1.4;
}

.promo-desc--sub {
  margin-top: 4px;
  color: rgba(255, 229, 186, 0.72);
}

.local-amount-hint {
  text-align: center;
  margin-top: 14px;
  font-size: 14px;
  color: rgba(255, 229, 186, 0.75);
  letter-spacing: 0.04em;
}

.confirm-btn {
  margin-top: 10px;
  border: none;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%) !important;
  color: #5a1b00 !important;
  font-weight: 800;
  letter-spacing: 0.08em;
  box-shadow: 0 12px 22px rgba(75, 25, 0, 0.28);
}

@keyframes skeleton-shimmer {
  100% {
    transform: translateX(100%);
  }
}
</style>
