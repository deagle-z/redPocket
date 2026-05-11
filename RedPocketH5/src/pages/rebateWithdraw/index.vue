<script setup lang="ts">
import { computed, ref } from 'vue'
import { showToast } from 'vant'
import { useRouter } from 'vue-router'
import type { AppCountryItem, CreateWithdrawOrderReq, RechargeField, RechargeFieldOption, WithdrawAccountItem } from '@/api/user'
import {
  createRebateWithdrawOrder,
  getAppCountries,
  getCountryWithdrawFields,
  getCurrentTgUserInfo,
  getWithdrawAccounts,
} from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { useUserStore } from '@/stores'
import { formatCurrency, truncate2 } from '@/utils/currency'
import { safeBack } from '@/utils/navigation'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()

const rebateBalance = ref(0)
const currentUserCountry = ref('')
const hideCountrySelector = ref(false)
const pageLoading = ref(true)
const countries = ref<AppCountryItem[]>([])
const selectedCountry = ref<AppCountryItem | null>(null)
const countriesLoading = ref(false)
const withdrawFields = ref<RechargeField[]>([])
const fieldValues = ref<Record<string, string>>({})
const fieldsLoading = ref(false)
const boundAccounts = ref<WithdrawAccountItem[]>([])
const selectedAccountId = ref<number | undefined>(undefined)
const amountOptions = [100, 200, 500, 1000, 5000, 10000, 20000, 50000, 'custom']
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

const canSubmit = computed(() =>
  Number(displayAmount.value) > 0
  && Number(displayAmount.value) <= rebateBalance.value
  && !submitLoading.value
  && !!selectedCountry.value
  && !fieldsLoading.value,
)

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
  showToast(t('transformPage.rebateWithdrawTip'))
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
    rebateBalance.value = Number(data?.rebate_amount ?? 0)
    currentUserCountry.value = String(data?.country || '').trim()
  }
  catch {
    rebateBalance.value = 0
    currentUserCountry.value = String(userStore.userInfo?.country || '').trim()
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
    if (accountsRes.status === 'fulfilled')
      applyBoundAccounts(accountsRes.value.data ?? [], code)
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
  if (!amount || amount <= 0) {
    showCenterToast(t('withdrawPage.invalidAmount'))
    return
  }
  if (amount > rebateBalance.value) {
    showCenterToast(t('transformPage.toastExceedBalance'))
    return
  }

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
    const { data } = await createRebateWithdrawOrder(req)
    rebateBalance.value = Number(data?.rebateAmount ?? truncate2(rebateBalance.value - amount))
    showCenterToast(t('withdrawPage.orderSuccess', { orderNo: data?.orderNo || '--' }))
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
  <div class="rebate-withdraw-page">
    <AppPageHeader :title="t('transformPage.rebateWithdraw')" @back="goBack" @right-click="showHelpTip">
      <template #right>
        <van-icon name="question-o" />
      </template>
    </AppPageHeader>

    <template v-if="pageLoading">
      <section class="card balance-card skeleton-card">
        <div class="skeleton-line skeleton-line-sm" />
        <div class="skeleton-line skeleton-line-xl" />
        <div class="skeleton-line skeleton-line-md" />
      </section>
      <section class="card skeleton-card">
        <div class="skeleton-line skeleton-line-title" />
        <div class="skeleton-input" />
        <div class="amount-grid">
          <div v-for="idx in 9" :key="idx" class="skeleton-amount-item" />
        </div>
      </section>
    </template>

    <template v-else>
      <section class="card balance-card">
        <p class="card-label">
          {{ t('transformPage.walletCommission') }}
        </p>
        <p class="card-value">
          <CoinAmount :text="formatCurrency(rebateBalance)" />
        </p>
        <p class="card-sub">
          {{ t('transformPage.rebateWithdrawNoFlow') }}
        </p>
      </section>

      <div v-if="!hideCountrySelector" class="country-bar">
        <div class="country-scroll">
          <button
            v-for="country in countries"
            :key="country.countryCode"
            type="button"
            class="country-pill"
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

      <section class="card">
        <h2 class="section-title">
          {{ t('withdrawPage.amountTitle') }}
        </h2>
        <van-field
          v-model="customAmount"
          type="number"
          :label="t('withdrawPage.amountLabel')"
          :placeholder="t('withdrawPage.amountPlaceholder')"
          class="custom-input"
          @focus="selectedAmount = 'custom'"
        />
        <div class="amount-grid">
          <button
            v-for="item in amountOptions"
            :key="item"
            type="button"
            class="amount-item"
            :class="{ active: selectedAmount === item }"
            @click="chooseAmount(item as number | 'custom')"
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
            <CoinAmount text="0.00" />
          </div>
          <div class="fee-row total">
            <span>{{ t('withdrawPage.actualDeduct') }}</span>
            <CoinAmount :text="formatAmountText(Number(displayAmount || 0))" />
          </div>
        </div>
      </section>

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

      <section class="card tips">
        <h2 class="tips-title">
          {{ t('withdrawPage.tipsTitle') }}
        </h2>
        <ol>
          <li>{{ t('transformPage.rebateWithdrawNoFlow') }}</li>
          <li>{{ t('withdrawPage.tips4') }}</li>
          <li>{{ t('withdrawPage.tips5') }}</li>
        </ol>
      </section>
    </template>

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
  name: 'RebateWithdraw',
}
</route>

<style scoped>
.rebate-withdraw-page {
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
  margin-bottom: 12px;
  padding: 14px 16px;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  border: 1px solid rgba(212, 175, 55, 0.34);
  border-radius: 18px;
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.3),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
}

.card::after {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}

.balance-card {
  margin-top: 10px;
}

.card-label,
.card-sub {
  margin: 0;
  color: rgba(255, 229, 186, 0.62);
  font-size: 12px;
}

.card-value {
  margin: 8px 0;
  color: #ffd87f;
  font-size: 28px;
  font-weight: 800;
  line-height: 1;
}

.country-bar {
  margin: 10px 0 12px;
  overflow: hidden;
}

.country-scroll {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 2px;
}

.country-pill,
.country-pill-loading {
  min-width: 92px;
  height: 44px;
  flex: 0 0 auto;
  border: 1px solid rgba(212, 175, 55, 0.28);
  border-radius: 14px;
  color: rgba(255, 229, 186, 0.72);
  background: rgba(255, 248, 214, 0.06);
}

.country-pill.active {
  color: #5a1b00;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
}

.country-code,
.country-name {
  display: block;
  font-size: 11px;
  line-height: 16px;
}

.country-code {
  font-weight: 800;
}

.section-title,
.tips-title {
  margin: 0 0 12px;
  color: #ffd98b;
  font-size: 14px;
  font-weight: 800;
}

.custom-input {
  overflow: hidden;
  margin-bottom: 10px;
  background: rgba(255, 248, 214, 0.06);
  border: 1px solid rgba(212, 175, 55, 0.18);
  border-radius: 14px;
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

.amount-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.amount-item,
.skeleton-amount-item {
  min-height: 54px;
  border: 1px solid rgba(212, 175, 55, 0.22);
  border-radius: 14px;
  color: #fff0c9;
  background: rgba(255, 248, 214, 0.06);
}

.amount-item.active {
  color: #5a1b00;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
}

.amount-local {
  display: block;
  margin-top: 4px;
  font-size: 10px;
  color: currentColor;
  opacity: 0.72;
}

.local-amount-hint {
  margin: 10px 0;
  color: rgba(255, 229, 186, 0.66);
  font-size: 12px;
  text-align: center;
}

.submit-btn {
  margin-top: 12px;
  color: #5a1b00;
  font-weight: 800;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  border: none;
}

.fee-card {
  margin-top: 12px;
  padding: 10px 12px;
  background: rgba(0, 0, 0, 0.12);
  border-radius: 14px;
}

.fee-row,
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.fee-row {
  color: rgba(255, 229, 186, 0.68);
  font-size: 12px;
}

.fee-row.total {
  margin-top: 8px;
  color: #ffd87f;
  font-weight: 800;
}

.bind-btn,
.bind-btn-block {
  border: 1px solid rgba(255, 248, 214, 0.34);
  border-radius: 999px;
  color: #5a1b00;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  font-size: 12px;
  font-weight: 700;
}

.bind-btn {
  height: 30px;
  padding: 0 12px;
}

.bind-btn-block {
  display: block;
  height: 34px;
  margin: 12px auto 0;
  padding: 0 16px;
}

.section-loading {
  display: block;
  margin: 18px auto;
}

.empty-tip,
.tips li {
  color: rgba(255, 229, 186, 0.64);
  font-size: 12px;
}

.tips ol {
  margin: 0;
  padding-left: 18px;
}

.tips li + li {
  margin-top: 6px;
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

.skeleton-card {
  min-height: 120px;
}

.skeleton-line,
.skeleton-input,
.skeleton-amount-item {
  overflow: hidden;
  background: linear-gradient(90deg, rgba(255, 248, 214, 0.08), rgba(255, 248, 214, 0.16), rgba(255, 248, 214, 0.08));
  background-size: 200% 100%;
  animation: skeleton-shimmer 1.2s linear infinite;
}

.skeleton-line {
  height: 12px;
  margin-bottom: 12px;
  border-radius: 999px;
}

.skeleton-line-sm {
  width: 34%;
}

.skeleton-line-xl {
  width: 68%;
  height: 24px;
}

.skeleton-line-md {
  width: 48%;
}

.skeleton-line-title {
  width: 36%;
}

.skeleton-input {
  height: 44px;
  margin-bottom: 10px;
  border-radius: 14px;
}

@keyframes skeleton-shimmer {
  from {
    background-position: 200% 0;
  }

  to {
    background-position: -200% 0;
  }
}
</style>
