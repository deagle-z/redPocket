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
  getWithdrawAccounts,
} from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { formatCurrency } from '@/utils/currency'

const { t } = useI18n()
const router = useRouter()

const balance = ref(0)
const frozen = ref(0)

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
  const rate = selectedCountry.value?.rate || 1
  return (coins * rate).toFixed(2)
})

const localCurrencySymbol = computed(() => selectedCountry.value?.currencySymbol || '')

const canSubmit = computed(() =>
  Number(displayAmount.value) > 0
  && !submitLoading.value
  && !!selectedCountry.value
  && !fieldsLoading.value,
)

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

function chooseAmount(value: number | 'custom') {
  selectedAmount.value = value
  if (value !== 'custom')
    customAmount.value = ''
}

function goBack() {
  router.back()
}

function showHelpTip() {
  showToast(t('withdrawPage.helpTip'))
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

async function loadBalance() {
  try {
    const { data } = await getCurrentTgUserInfo()
    balance.value = Number(data?.balance ?? 0)
    frozen.value = Number((data as any)?.frozen ?? 0)
  }
  catch {
    balance.value = 0
  }
}

async function loadWithdrawFields(code: string) {
  fieldsLoading.value = true
  withdrawFields.value = []
  fieldValues.value = {}
  try {
    const { data } = await getCountryWithdrawFields(code)
    withdrawFields.value = data ?? []
    // 初始化字段默认值
    const init: Record<string, string> = {}
    for (const f of withdrawFields.value)
      init[f.fieldKey] = f.defaultValue ?? ''
    fieldValues.value = init
  }
  catch {
    withdrawFields.value = []
  }
  finally {
    fieldsLoading.value = false
  }
}

async function loadBoundAccount(code: string) {
  try {
    const { data } = await getWithdrawAccounts()
    const all = data ?? []
    boundAccounts.value = all.filter(a => a.countryCode === code)
    // 优先默认账户，否则取第一个
    const account = boundAccounts.value.find(a => a.isDefault === 1) ?? boundAccounts.value[0]
    if (account) {
      selectedAccountId.value = account.id
      const parsed = parseAccountData(account.accountData)
      // 用账户数据覆盖字段值
      for (const key of Object.keys(parsed)) {
        if (key in fieldValues.value)
          fieldValues.value[key] = parsed[key]
      }
    }
    else {
      selectedAccountId.value = undefined
    }
  }
  catch {
    boundAccounts.value = []
    selectedAccountId.value = undefined
  }
}

async function handleSelectCountry(country: AppCountryItem) {
  if (selectedCountry.value?.countryCode === country.countryCode)
    return
  selectedCountry.value = country
  await loadWithdrawFields(country.countryCode)
  await loadBoundAccount(country.countryCode)
}

async function loadCountries() {
  countriesLoading.value = true
  try {
    const { data } = await getAppCountries()
    countries.value = data ?? []
    if (countries.value.length) {
      selectedCountry.value = countries.value[0]
      await loadWithdrawFields(countries.value[0].countryCode)
      await loadBoundAccount(countries.value[0].countryCode)
    }
  }
  catch {
    countries.value = []
  }
  finally {
    countriesLoading.value = false
  }
}

async function handleSubmitWithdraw() {
  if (!canSubmit.value)
    return
  const amount = Number(displayAmount.value)
  if (!amount || amount <= 0) {
    showCenterToast(t('withdrawPage.invalidAmount'))
    return
  }

  // 必填 + 正则校验
  for (const f of withdrawFields.value) {
    const val = fieldValues.value[f.fieldKey]?.trim() ?? ''
    if (f.isRequired === 1 && !val) {
      showCenterToast(f.errorTips || `${f.fieldLabel} 不能为空`)
      return
    }
    if (val && f.regexRule) {
      const regex = new RegExp(f.regexRule)
      if (!regex.test(val)) {
        showCenterToast(f.errorTips || `${f.fieldLabel} 格式不正确`)
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
    await loadBalance()
  }
  catch {
    showCenterToast(t('withdrawPage.orderFailed'))
  }
  finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  loadBalance()
  loadCountries()
})
</script>

<template>
  <div class="withdraw-page theme-withdraw">
    <AppPageHeader :title="t('withdrawPage.title')" @back="goBack" @right-click="showHelpTip">
      <template #right>
        <van-icon name="question-o" />
      </template>
    </AppPageHeader>

    <section class="card balance-card">
      <div>
        <p class="card-label">
          {{ t('withdrawPage.currentBalance') }}
        </p>
        <p class="card-value">
          <CoinAmount :text="formatCurrency(balance)" />
        </p>
        <p class="card-sub">
          {{ t('withdrawPage.frozenBalance', { amount: formatCurrency(frozen) }) }}
        </p>
      </div>
      <span class="card-chip"><img class="chip-coin" src="@/assets/svg/coin.svg" alt=""></span>
    </section>

    <!-- 国家选择行 -->
    <div class="country-bar">
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
          <span v-if="item !== 'custom'"><CoinAmount :text="`${item}`" /></span>
          <span v-else>{{ t('withdrawPage.custom') }}</span>
          <span v-if="item !== 'custom' && selectedCountry?.rate" class="amount-local">
            {{ localCurrencySymbol }} {{ (Number(item) * (selectedCountry?.rate ?? 1)).toFixed(2) }}
          </span>
        </button>
      </div>

      <div v-if="displayAmount && selectedCountry?.rate" class="local-amount-hint">
        ≈ {{ localCurrencySymbol }}{{ localAmount }} {{ selectedCountry?.currencyCode }}
      </div>

      <div class="balance-breakdown">
        <div class="balance-header">
          <div class="balance-title">
            <span class="balance-icon">◎</span>
            {{ t('withdrawPage.availableBalance') }}
          </div>
          <div class="balance-amount">
            <CoinAmount :text="formatCurrency(balance)" />
          </div>
        </div>
        <div class="balance-row">
          <div>
            <p class="row-title">
              {{ t('withdrawPage.freezing') }}
            </p>
          </div>
          <div class="row-right">
            <CoinAmount :text="formatCurrency(frozen)" />
          </div>
        </div>
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
          <CoinAmount :text="`${displayAmount || 0}`" />
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
        <li>{{ t('withdrawPage.tips1', { amount: '100' }) }}</li>
        <li>{{ t('withdrawPage.tips2') }}</li>
        <li>{{ t('withdrawPage.tips3') }}</li>
        <li>{{ t('withdrawPage.tips4') }}</li>
        <li>{{ t('withdrawPage.tips5') }}</li>
      </ol>
    </section>

    <!-- select 选择器弹窗 -->
    <van-popup v-model:show="pickerVisible" position="bottom" teleport="#app">
      <van-picker
        :columns="pickerColumns"
        @confirm="onPickerConfirm"
        @cancel="pickerVisible = false"
      />
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

.balance-breakdown,
.fee-card {
  margin-top: 14px;
  padding: 12px;
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.06);
}

.balance-header,
.balance-row,
.fee-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.balance-header {
  margin-bottom: 10px;
}

.balance-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #ffd98b;
  font-weight: 700;
}

.balance-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 999px;
  background: rgba(212, 175, 55, 0.16);
}

.balance-row + .balance-row,
.fee-row + .fee-row {
  border-top: 1px solid rgba(212, 175, 55, 0.12);
  margin-top: 10px;
  padding-top: 10px;
}

.row-right,
.fee-row {
  color: #fff0c9;
  gap: 8px;
}

.row-title {
  margin: 0;
  color: #fff0c9;
  font-weight: 700;
}

:deep(.custom-input),
:deep(.withdraw-field) {
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(212, 175, 55, 0.16);
  background: rgba(255, 248, 214, 0.05);
}

:deep(.withdraw-field) {
  margin-top: 10px;
}

:deep(.withdraw-field:first-of-type) {
  margin-top: 0;
}

:deep(.custom-input .van-field__label),
:deep(.custom-input .van-field__control),
:deep(.withdraw-field .van-field__label),
:deep(.withdraw-field .van-field__control) {
  color: #fff0c9;
}

:deep(.custom-input .van-field__control::placeholder),
:deep(.withdraw-field .van-field__control::placeholder) {
  color: rgba(255, 229, 186, 0.4);
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
