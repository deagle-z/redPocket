<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import type { AppCountryItem, CreateWithdrawOrderReq, RechargeField, RechargeFieldOption, WithdrawAccountItem } from '@/api/user'
import {
  createWithdrawOrder,
  getAppCountries,
  getCountryWithdrawFields,
  getCurrentTgUserInfo,
  getCurrentTgWithdrawSummary,
  getWithdrawAccounts,
} from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { useUserStore } from '@/stores'
import { formatCurrency, truncate2 } from '@/utils/currency'
import { safeBack } from '@/utils/navigation'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()

const MIN_WITHDRAW_AMOUNT = 5
const FREE_WITHDRAW_COUNT = 3
const WITHDRAW_FEE_RATE = 0.05

const balance = ref(0)
const nonWithdrawableAmount = ref(0)
const todayWithdrawCount = ref(0)
const currentUserCountry = ref('')
const hideCountrySelector = ref(false)
const pageLoading = ref(true)
const withdrawSummaryLoaded = ref(false)

// 国家列表
const countries = ref<AppCountryItem[]>([])
const selectedCountry = ref<AppCountryItem | null>(null)
const countriesLoading = ref(false)

// 提现字段
const withdrawFields = ref<RechargeField[]>([])
const fieldValues = ref<Record<string, string>>({})
const fieldsLoading = ref(false)

// 当前国家绑定的账户
const boundAccounts = ref<WithdrawAccountItem[]>([])
const selectedAccountId = ref<number | undefined>(undefined)

// 金额
const amountOptions = [5, 10, 100, 200, 500, 1000, 5000, 10000, 20000, 50000, 'custom']
const selectedAmount = ref<number | 'custom'>(amountOptions[0] as number)
const customAmount = ref('')
const submitLoading = ref(false)

const displayAmount = computed(() => {
  if (selectedAmount.value === 'custom')
    return customAmount.value ? Number(customAmount.value) : 0
  return selectedAmount.value
})

const localAmount = computed(() => {
  const coins = Number(displayAmount.value) || 0
  return formatLocalAmount(coins)
})

const localCurrencySymbol = computed(() => selectedCountry.value?.currencySymbol || '')
const withdrawableAmount = computed(() =>
  truncate2(Math.max(0, balance.value - nonWithdrawableAmount.value)),
)
const estimatedFee = computed(() => {
  if (todayWithdrawCount.value < FREE_WITHDRAW_COUNT)
    return 0
  return truncate2(Number(displayAmount.value || 0) * WITHDRAW_FEE_RATE)
})
const actualReceiveAmount = computed(() =>
  truncate2(Math.max(0, Number(displayAmount.value || 0) - estimatedFee.value)),
)

const canSubmit = computed(() =>
  truncate2(Number(displayAmount.value)) >= MIN_WITHDRAW_AMOUNT
  && truncate2(Number(displayAmount.value)) <= withdrawableAmount.value
  && withdrawSummaryLoaded.value
  && !submitLoading.value
  && !!selectedCountry.value
  && !fieldsLoading.value,
)

// select 选择器
const pickerVisible = ref(false)
const pickerField = ref<RechargeField | null>(null)
const pickerOptions = computed(() => parseFieldOptions(pickerField.value))

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

function onSelectOption(option: RechargeFieldOption) {
  if (pickerField.value)
    fieldValues.value[pickerField.value.fieldKey] = option.value
  pickerVisible.value = false
}

function chooseAmount(value: number | 'custom') {
  selectedAmount.value = value
  if (value !== 'custom')
    customAmount.value = ''
}

function goBack() {
  safeBack(router)
}

function showHelpTip() {
  showToast(t('withdrawPage.helpTip'))
}

function formatAmountText(value: number) {
  return truncate2(Number(value || 0)).toFixed(2)
}

function formatLocalAmount(coins: number) {
  const rate = selectedCountry.value?.rate || 1
  return truncate2((Number(coins) || 0) * rate).toFixed(2)
}

function showCenterToast(message: string) {
  showToast({ message, position: 'middle', teleport: '#app', wordBreak: 'break-word' })
}

function parseAccountData(raw: string): Record<string, string> {
  try {
    return JSON.parse(raw) ?? {}
  }
  catch {
    return {}
  }
}

async function loadUserInfo() {
  try {
    const { data } = await getCurrentTgUserInfo()
    balance.value = Number(data?.balance ?? 0)
    currentUserCountry.value = String(data?.country || '').trim()
  }
  catch {
    balance.value = 0
    currentUserCountry.value = String(userStore.userInfo?.country || '').trim()
  }
}

async function loadWithdrawSummary() {
  withdrawSummaryLoaded.value = false
  try {
    const { data } = await getCurrentTgWithdrawSummary()
    balance.value = Number(data?.balance ?? balance.value)
    nonWithdrawableAmount.value = Number(data?.nonWithdrawableAmount ?? 0)
    todayWithdrawCount.value = Number(data?.todayWithdrawCount ?? 0)
    withdrawSummaryLoaded.value = true
  }
  catch {
    nonWithdrawableAmount.value = 0
    todayWithdrawCount.value = 0
  }
}

function applyWithdrawFields(fields: RechargeField[]) {
  withdrawFields.value = fields
  const init: Record<string, string> = {}
  for (const f of withdrawFields.value)
    init[f.fieldKey] = f.defaultValue ?? ''
  fieldValues.value = init
}

function applyBoundAccounts(accounts: WithdrawAccountItem[], code: string) {
  boundAccounts.value = accounts.filter(a => a.countryCode === code)
  const account = boundAccounts.value.find(a => a.isDefault === 1) ?? boundAccounts.value[0]
  if (account) {
    selectedAccountId.value = account.id
    const parsed = parseAccountData(account.accountData)
    for (const key of Object.keys(parsed)) {
      if (key in fieldValues.value)
        fieldValues.value[key] = parsed[key]
    }
  }
  else {
    selectedAccountId.value = undefined
  }
}

async function loadWithdrawDetails(code: string) {
  fieldsLoading.value = true
  withdrawFields.value = []
  fieldValues.value = {}
  boundAccounts.value = []
  selectedAccountId.value = undefined

  const [fieldsRes, accountsRes] = await Promise.allSettled([
    getCountryWithdrawFields(code),
    getWithdrawAccounts(),
  ])

  try {
    if (selectedCountry.value?.countryCode !== code)
      return

    const fields = fieldsRes.status === 'fulfilled' ? fieldsRes.value.data ?? [] : []
    applyWithdrawFields(fields)
    if (accountsRes.status === 'fulfilled') {
      applyBoundAccounts(accountsRes.value.data ?? [], code)
    }
  }
  finally {
    fieldsLoading.value = false
  }
}

async function handleSelectCountry(country: AppCountryItem) {
  if (selectedCountry.value?.countryCode === country.countryCode)
    return
  selectedCountry.value = country
  await loadWithdrawDetails(country.countryCode)
}

async function loadCountries() {
  countriesLoading.value = true
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
      void loadWithdrawDetails(selectedCountry.value.countryCode)
    }
  }
  catch {
    countries.value = []
  }
  finally {
    countriesLoading.value = false
  }
}

async function initPage() {
  pageLoading.value = true
  currentUserCountry.value = String(userStore.userInfo?.country || '').trim()
  try {
    await loadUserInfo()
    await loadWithdrawSummary()
    await loadCountries()
  }
  finally {
    pageLoading.value = false
  }
}

async function handleSubmitWithdraw() {
  if (!canSubmit.value)
    return
  const amount = truncate2(Number(displayAmount.value))
  if (!amount || amount < MIN_WITHDRAW_AMOUNT) {
    showCenterToast(t('withdrawPage.invalidAmount'))
    return
  }
  if (amount > withdrawableAmount.value) {
    showCenterToast(t('withdrawPage.amountExceedsWithdrawable', { amount: formatCurrency(withdrawableAmount.value) }))
    return
  }

  // 必填 + 正则校验
  for (const f of withdrawFields.value) {
    const val = fieldValues.value[f.fieldKey]?.trim() ?? ''
    if (f.isRequired === 1 && !val) {
      showCenterToast(f.errorTips || t('withdrawPage.requiredField', { field: f.fieldLabel }))
      return
    }
    if (val && f.regexRule) {
      const regex = new RegExp(f.regexRule)
      if (!regex.test(val)) {
        showCenterToast(f.errorTips || t('withdrawPage.invalidField', { field: f.fieldLabel }))
        return
      }
    }
  }

  submitLoading.value = true
  try {
    const req: CreateWithdrawOrderReq = {
      amount,
      countryCode: selectedCountry.value?.countryCode ?? '',
      accountId: selectedAccountId.value,
      fieldValues: withdrawFields.value.length ? { ...fieldValues.value } : undefined,
    }
    const { data } = await createWithdrawOrder(req)
    showCenterToast(t('withdrawPage.orderSuccess', { orderNo: data?.orderNo || '--' }))
    await loadWithdrawSummary()
  }
  catch {
    showCenterToast(t('withdrawPage.orderFailed'))
  }
  finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  void initPage()
})
</script>

<template>
  <div class="withdraw-page theme-withdraw">
    <AppPageHeader :title="t('withdrawPage.title')" @back="goBack" @right-click="showHelpTip">
      <template #right>
        <van-icon name="question-o" />
      </template>
    </AppPageHeader>

    <template v-if="pageLoading">
      <section class="card balance-card skeleton-card">
        <div class="skeleton-balance-copy">
          <div class="skeleton-line skeleton-line-sm" />
          <div class="skeleton-line skeleton-line-xl" />
          <div class="skeleton-line skeleton-line-md" />
        </div>
        <div class="skeleton-coin" />
      </section>

      <div class="country-bar skeleton-country-bar">
        <div class="country-scroll">
          <div v-for="idx in 4" :key="`country-skeleton-${idx}`" class="skeleton-pill" />
        </div>
      </div>

      <section class="card skeleton-card skeleton-amount-card">
        <div class="skeleton-line skeleton-line-title" />
        <div class="skeleton-input" />
        <div class="amount-grid skeleton-amount-grid">
          <div v-for="idx in amountOptions.length" :key="`amount-skeleton-${idx}`" class="skeleton-amount-item" />
        </div>
        <div class="skeleton-line skeleton-line-center" />
        <div class="skeleton-submit" />
        <div class="skeleton-fee-card">
          <div class="skeleton-fee-row">
            <div class="skeleton-line skeleton-line-sm" />
            <div class="skeleton-line skeleton-line-xs" />
          </div>
          <div class="skeleton-fee-row">
            <div class="skeleton-line skeleton-line-md" />
            <div class="skeleton-line skeleton-line-xs" />
          </div>
        </div>
      </section>

      <section class="card skeleton-card">
        <div class="section-head">
          <div class="skeleton-line skeleton-line-title" />
          <div class="skeleton-bind-btn" />
        </div>
        <div v-for="idx in 4" :key="`field-skeleton-${idx}`" class="skeleton-field" />
      </section>

      <section class="card tips skeleton-card">
        <div class="skeleton-line skeleton-line-title" />
        <div v-for="idx in 5" :key="`tip-skeleton-${idx}`" class="skeleton-tip-line" />
      </section>
    </template>

    <template v-else>
      <section class="card balance-card">
        <div>
          <p class="card-label">
            {{ t('withdrawPage.currentBalance') }}
          </p>
          <p class="card-value">
            <CoinAmount :text="formatCurrency(balance)" />
          </p>
          <div class="card-sub-list">
            <p class="card-sub">
              <CurrencyText :text="`${t('withdrawPage.withdrawableAmount')}: ${formatCurrency(withdrawableAmount)}`" />
            </p>
            <p class="card-sub">
              <CurrencyText :text="`${t('withdrawPage.nonWithdrawableAmount')}: ${formatCurrency(nonWithdrawableAmount)}`" />
            </p>
          </div>
        </div>
        <span class="card-chip"><img class="chip-coin" src="@/assets/svg/coin.svg" alt=""></span>
      </section>

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
          <div v-if="countriesLoading" class="country-pill-loading">
            <van-loading size="16" color="#d4af37" />
          </div>
        </div>
      </div>

      <!-- 提现金额 -->
      <section class="card">
        <h2 class="section-title">
          {{ t('withdrawPage.amountTitle') }}
        </h2>
        <van-field
          v-model="customAmount" type="number" :label="t('withdrawPage.amountLabel')" :placeholder="t('withdrawPage.amountPlaceholder')" class="custom-input"
          @focus="selectedAmount = 'custom'"
        />
        <div class="amount-grid">
          <button
            v-for="item in amountOptions" :key="item" type="button" class="amount-item"
            :class="{ active: selectedAmount === item }" @click="chooseAmount(item as number | 'custom')"
          >
            <span v-if="item !== 'custom'"><CoinAmount :text="formatAmountText(Number(item))" /></span>
            <span v-else>{{ t('withdrawPage.custom') }}</span>
            <span v-if="item !== 'custom' && selectedCountry?.rate" class="amount-local">
              {{ localCurrencySymbol }} {{ formatLocalAmount(Number(item)) }}
            </span>
          </button>
        </div>

        <div v-if="displayAmount && selectedCountry?.rate" class="local-amount-hint">
          ≈ {{ localCurrencySymbol }}{{ localAmount }} {{ selectedCountry?.currencyCode }}
        </div>

        <van-button type="primary" round block class="submit-btn" :loading="submitLoading" :disabled="!canSubmit" @click="handleSubmitWithdraw">
          {{ t('withdrawPage.submit') }}
        </van-button>

        <div class="fee-card">
          <div class="fee-row">
            <span>{{ t('withdrawPage.serviceFee') }}</span>
            <CoinAmount :text="formatAmountText(estimatedFee)" />
          </div>
          <div class="fee-row">
            <span>{{ t('withdrawPage.actualReceive') }}</span>
            <CoinAmount :text="formatAmountText(actualReceiveAmount)" />
          </div>
          <div class="fee-row total">
            <span>{{ t('withdrawPage.actualDeduct') }}</span>
            <CoinAmount :text="formatAmountText(Number(displayAmount || 0))" />
          </div>
        </div>
      </section>

      <!-- 收款信息 -->
      <section class="card">
        <div class="section-head">
          <h2 class="section-title">
            {{ t('withdrawPage.receiverTitle') }}
          </h2>
          <button type="button" class="bind-btn" @click="router.push('/withdrawAccount')">
            {{ t('withdrawPage.bindAccount') }}
          </button>
        </div>

        <van-loading v-if="fieldsLoading" size="20" color="#d4af37" class="section-loading" />

        <template v-else-if="withdrawFields.length">
          <template v-for="field in withdrawFields" :key="field.fieldKey">
            <van-field
              v-if="field.fieldType === 'select'"
              :model-value="getSelectLabel(field)"
              :label="field.fieldLabel"
              :placeholder="field.fieldPlaceholder || ''"
              :required="field.isRequired === 1"
              is-link
              readonly
              class="custom-input withdraw-field"
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
              class="custom-input withdraw-field"
              rows="3"
            />
          </template>
        </template>

        <p v-else-if="!fieldsLoading && selectedCountry" class="empty-tip">
          {{ t('withdrawPage.noFields') }}
          <button type="button" class="bind-btn-block" @click="router.push('/withdrawAccount')">
            {{ t('withdrawPage.bindAccount') }}
          </button>
        </p>
      </section>

      <!-- 提示 -->
      <section class="card tips">
        <h2 class="tips-title">
          {{ t('withdrawPage.tipsTitle') }}
        </h2>
        <ol>
          <li>{{ t('withdrawPage.tips1', { amount: MIN_WITHDRAW_AMOUNT }) }}</li>
          <li>{{ t('withdrawPage.tips2') }}</li>
          <li>{{ t('withdrawPage.tips3') }}</li>
          <li>{{ t('withdrawPage.tips4') }}</li>
          <li>{{ t('withdrawPage.tips5') }}</li>
        </ol>
      </section>
    </template>

    <!-- select 选择器弹窗 -->
    <van-popup v-if="!pageLoading" v-model:show="pickerVisible" round position="bottom" teleport="#app" class="withdraw-select-popup">
      <div class="withdraw-select-popup-header">
        <span class="withdraw-select-popup-title">{{ pickerField?.fieldLabel }}</span>
        <button type="button" class="withdraw-select-popup-close" @click="pickerVisible = false">
          ×
        </button>
      </div>

      <div v-if="pickerOptions.length" class="withdraw-select-list">
        <button
          v-for="option in pickerOptions"
          :key="option.value"
          type="button"
          class="withdraw-select-item"
          :class="{ active: pickerField && fieldValues[pickerField.fieldKey] === option.value }"
          @click="onSelectOption(option)"
        >
          <span class="withdraw-select-code">{{ option.value }}</span>
          <span class="withdraw-select-text">{{ option.label }}</span>
          <span v-if="pickerField && fieldValues[pickerField.fieldKey] === option.value" class="withdraw-select-check">✓</span>
        </button>
      </div>
      <p v-else class="withdraw-select-empty">
        {{ t('withdrawPage.noFields') }}
      </p>
    </van-popup>
  </div>
</template>

<route lang="json5">
{
  name: "Withdraw"
}
</route>

<style scoped>
.withdraw-page {
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
  padding: 0 12px calc(90px + env(safe-area-inset-bottom));
  color: #fff0c9;
}

.card {
  position: relative;
  overflow: hidden;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  border-radius: 18px;
  border: 1px solid rgba(212, 175, 55, 0.34);
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.3),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
  padding: 14px 16px;
  margin-bottom: 12px;
}

.card::after {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}

.skeleton-card {
  pointer-events: none;
}

.skeleton-line,
.skeleton-coin,
.skeleton-pill,
.skeleton-input,
.skeleton-amount-item,
.skeleton-submit,
.skeleton-bind-btn,
.skeleton-field,
.skeleton-tip-line {
  position: relative;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(255, 248, 214, 0.1);
}

.skeleton-line::after,
.skeleton-coin::after,
.skeleton-pill::after,
.skeleton-input::after,
.skeleton-amount-item::after,
.skeleton-submit::after,
.skeleton-bind-btn::after,
.skeleton-field::after,
.skeleton-tip-line::after {
  content: '';
  position: absolute;
  inset: 0;
  transform: translateX(-100%);
  background: linear-gradient(90deg, transparent, rgba(255, 231, 166, 0.2), transparent);
  animation: skeleton-shimmer 1.25s ease-in-out infinite;
}

@keyframes skeleton-shimmer {
  100% {
    transform: translateX(100%);
  }
}

.skeleton-balance-copy {
  width: min(68%, 240px);
}

.skeleton-line {
  height: 12px;
}

.skeleton-line-xs {
  width: 54px;
}

.skeleton-line-sm {
  width: 84px;
}

.skeleton-line-md {
  width: 138px;
}

.skeleton-line-xl {
  width: 188px;
  height: 28px;
  margin: 12px 0 10px;
}

.skeleton-line-title {
  width: 132px;
  height: 16px;
  margin-bottom: 14px;
}

.skeleton-line-center {
  width: 178px;
  height: 14px;
  margin: 14px auto 0;
}

.skeleton-coin {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  background: rgba(255, 223, 135, 0.16);
}

.skeleton-country-bar {
  min-height: 42px;
}

.skeleton-pill {
  flex: 0 0 auto;
  width: 96px;
  height: 34px;
  border-radius: 20px;
}

.skeleton-input {
  width: 100%;
  height: 46px;
  border-radius: 14px;
}

.skeleton-amount-grid {
  pointer-events: none;
}

.skeleton-amount-item {
  height: 64px;
  border-radius: 14px;
}

.skeleton-submit {
  height: 44px;
  margin-top: 14px;
  border-radius: 999px;
  background: rgba(255, 223, 135, 0.18);
}

.skeleton-fee-card {
  margin-top: 14px;
  padding: 12px;
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.14);
  background: rgba(255, 248, 214, 0.04);
}

.skeleton-fee-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.skeleton-fee-row + .skeleton-fee-row {
  border-top: 1px solid rgba(212, 175, 55, 0.1);
  margin-top: 10px;
  padding-top: 10px;
}

.skeleton-bind-btn {
  width: 76px;
  height: 28px;
  border-radius: 14px;
}

.skeleton-field {
  height: 54px;
  border-radius: 14px;
}

.skeleton-field + .skeleton-field {
  margin-top: 10px;
}

.skeleton-tip-line {
  height: 12px;
  border-radius: 8px;
}

.skeleton-tip-line + .skeleton-tip-line {
  margin-top: 10px;
}

.skeleton-tip-line:nth-child(3) {
  width: 86%;
}

.skeleton-tip-line:nth-child(4) {
  width: 92%;
}

.skeleton-tip-line:nth-child(5) {
  width: 78%;
}

.skeleton-tip-line:nth-child(6) {
  width: 84%;
}

.balance-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-label,
.card-sub {
  color: rgba(255, 229, 186, 0.6);
}

.card-value,
.balance-amount {
  margin: 0;
  color: #ffd87f;
  font-size: 24px;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.card-sub {
  margin: 6px 0 0;
  font-size: 12px;
}

.card-sub-list {
  margin-top: 6px;
}

.card-sub-list .card-sub {
  margin-top: 3px;
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
}

.chip-coin {
  width: 100%;
  height: 100%;
  display: block;
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
  align-items: center;
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

.country-pill-loading {
  display: flex;
  align-items: center;
  padding: 4px 8px;
}

.card h2,
.section-title,
.tips-title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #ffd98b;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.section-head .section-title {
  margin: 0;
}

.bind-btn {
  padding: 5px 12px;
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.5);
  background: rgba(255, 248, 214, 0.08);
  color: #ffd87f;
  font-size: 12px;
  font-weight: 700;
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
  margin: 4px 0 8px;
}

.bind-btn-block {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  padding: 14px;
  margin-top: 10px;
  border-radius: 14px;
  border: 1.5px dashed rgba(212, 175, 55, 0.4);
  background: rgba(255, 248, 214, 0.04);
  color: rgba(255, 216, 127, 0.75);
  font-size: 14px;
  font-weight: 700;
}

.amount-grid {
  margin-top: 12px;
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

.amount-item.active {
  border-color: rgba(255, 248, 214, 0.34);
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.18), rgba(116, 24, 0, 0.28));
  color: #ffd87f;
}

.amount-item.active .amount-local {
  color: rgba(255, 229, 186, 0.7);
}

.local-amount-hint {
  text-align: center;
  margin-top: 14px;
  font-size: 14px;
  color: rgba(255, 229, 186, 0.75);
  letter-spacing: 0.04em;
}

.fee-card {
  margin-top: 14px;
  padding: 12px;
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.06);
}

.fee-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.fee-row + .fee-row {
  border-top: 1px solid rgba(212, 175, 55, 0.12);
  margin-top: 10px;
  padding-top: 10px;
}

.fee-row {
  color: #fff0c9;
  gap: 8px;
}

:deep(.custom-input),
:deep(.withdraw-field) {
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(212, 175, 55, 0.16);
  background: rgba(255, 248, 214, 0.05);
}

.withdraw-field {
  display: block;
  margin-top: 10px;
}

.withdraw-field:first-of-type {
  margin-top: 0;
}

:deep(.custom-input .van-field__label),
:deep(.custom-input .van-field__control) {
  color: #fff0c9;
}

:deep(.custom-input .van-field__control::placeholder) {
  color: rgba(255, 229, 186, 0.4);
}

:deep(.withdraw-field .van-field__label) {
  width: 100%;
  margin: 0 0 6px;
  color: #fff0c9;
  font-size: 12px;
  line-height: 1.2;
}

:deep(.withdraw-field .van-field__label span) {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.withdraw-field .van-field__value) {
  width: 100%;
  min-width: 0;
}

:deep(.withdraw-field .van-field__body),
:deep(.withdraw-field .van-field__control) {
  width: 100%;
}

:global(.withdraw-select-popup.van-popup) {
  max-height: 72vh;
  min-height: 320px;
  overflow: hidden;
  padding: 10px 0 calc(22px + env(safe-area-inset-bottom));
  border: 1px solid rgba(212, 175, 55, 0.34);
  border-radius: 24px 24px 0 0;
  background:
    radial-gradient(circle at 12% 10%, rgba(212, 175, 55, 0.18), transparent 22%),
    linear-gradient(180deg, #540000 0%, #280000 100%);
  box-shadow: 0 -12px 32px rgba(0, 0, 0, 0.48);
}

:global(.withdraw-select-popup-header) {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 56px;
  border-bottom: 1px solid rgba(212, 175, 55, 0.15);
}

:global(.withdraw-select-popup-title) {
  max-width: calc(100% - 96px);
  overflow: hidden;
  color: #fff0c9;
  font-size: 20px;
  font-weight: 800;
  line-height: 1.2;
  letter-spacing: 0.04em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:global(.withdraw-select-popup-close) {
  position: absolute;
  top: 50%;
  right: 16px;
  border: none;
  background: transparent;
  color: #ffd98b;
  font-size: 18px;
  line-height: 1;
  transform: translateY(-50%);
}

:global(.withdraw-select-list) {
  max-height: calc(72vh - 96px);
  overflow-y: auto;
  padding: 16px 14px 0;
}

:global(.withdraw-select-item) {
  display: grid;
  grid-template-columns: minmax(42px, auto) 1fr 24px;
  align-items: center;
  width: 100%;
  min-height: 52px;
  margin-bottom: 12px;
  padding: 12px 16px;
  border: 1px solid rgba(212, 175, 55, 0.14);
  border-radius: 16px;
  background: rgba(255, 248, 214, 0.05);
  text-align: left;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

:global(.withdraw-select-item.active) {
  border-color: rgba(212, 175, 55, 0.52);
  background: rgba(212, 175, 55, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 248, 214, 0.08);
}

:global(.withdraw-select-code) {
  min-width: 34px;
  padding-right: 10px;
  color: #ffd98b;
  font-size: 13px;
  font-weight: 800;
}

:global(.withdraw-select-text) {
  min-width: 0;
  overflow: hidden;
  color: #fff0c9;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:global(.withdraw-select-check) {
  color: #ffd98b;
  font-size: 20px;
  text-align: right;
}

:global(.withdraw-select-empty) {
  margin: 48px 18px 0;
  color: rgba(255, 229, 186, 0.58);
  font-size: 13px;
  line-height: 1.5;
  text-align: center;
}

.submit-btn {
  margin-top: 14px;
  border: none;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%) !important;
  color: #5a1b00 !important;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.fee-row.total {
  font-weight: 800;
  color: #ffd87f;
}

.tips ol {
  margin: 0;
  padding-left: 18px;
}

.tips li + li {
  margin-top: 6px;
}

.tips li,
.card-sub {
  color: rgba(255, 229, 186, 0.6);
  font-size: 13px;
}
</style>
