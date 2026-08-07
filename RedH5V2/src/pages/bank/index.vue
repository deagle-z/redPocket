<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import type {
  AppCountryItem,
  RechargeField,
  RechargeFieldOption,
  WithdrawAccountItem,
} from '@/api/user'
import {
  addWithdrawAccount,
  deleteWithdrawAccount,
  getAppCountries,
  getWithdrawAccounts,
  setDefaultWithdrawAccount,
  updateWithdrawAccount,
} from '@/api/user'
import {
  APP_COUNTRY_CODE,
  APP_COUNTRY_NAME,
  APP_COUNTRY_NAME_LOCAL,
  APP_CURRENCY,
} from '@/config/market'
import { buildWithdrawFields } from '@/config/payFields'
import type { PageStateStatus } from '@/types/business'
import { showConfirmDialog, showToast } from 'vant'
import PpmxSkeleton from '@/components/PpmxSkeleton.vue'
import PpmxSelectInput from '@/components/PpmxSelectInput.vue'
import PpmxButton from '@/components/PpmxButton.vue'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'

definePage({
  name: 'bank',
  meta: {
    titleKey: 'bank.title',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: true,
  },
})

const { t } = useI18n()
const pageStatus = ref<PageStateStatus>('loading')
const pageError = ref('')
const accountsLoading = ref(false)
const countries = ref<AppCountryItem[]>([])
const selectedCountry = ref<AppCountryItem | null>(null)
const allAccounts = ref<WithdrawAccountItem[]>([])
const showForm = ref(false)
const editingId = ref<number | null>(null)
const fieldValues = ref<Record<string, string>>({})

// 秘鲁 VCPAYPEN 通道代付字段前端写死，见 @/config/payFields。
// accNo 的长度/校验规则依赖已选 bankCode（电子钱包填手机号，银行填 20 位 CCI）。
const withdrawFields = computed<RechargeField[]>(
  () => buildWithdrawFields(t, fieldValues.value.bankCode),
)

const activeCountryCode = computed(() => selectedCountry.value?.countryCode || APP_COUNTRY_CODE)
const activeCountryName = computed(() => selectedCountry.value?.countryNameEn || APP_COUNTRY_NAME)
const activeCurrencyCode = computed(() => selectedCountry.value?.currencyCode || APP_CURRENCY)

const countryAccounts = computed(() =>
  allAccounts.value.filter(account => account.countryCode === activeCountryCode.value),
)

const formTitle = computed(() =>
  editingId.value ? t('bank.formTitleEdit') : t('bank.formTitleBind'),
)

const pageStateMessage = computed(() => {
  if (pageStatus.value === 'loading') return undefined

  return t('bank.errorMessage')
})

const submitLock = useSubmitLock(submitBankAccount, { minLockMs: 600 })
const isSubmitting = computed(() => submitLock.locked.value)

function parseAccountData(raw: string): Record<string, string> {
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

function getFieldInputType(field: RechargeField) {
  return field.fieldType === 'number' || field.dataType === 'number' ? 'number' : 'text'
}

function getAccountDisplayLines(account: WithdrawAccountItem) {
  const data = parseAccountData(account.accountData)

  return withdrawFields.value
    .map(field => ({
      label: field.fieldLabel,
      value: String(data[field.fieldKey] ?? '').trim(),
    }))
    .filter(line => line.value)
}

function initFieldValues(preset: Record<string, string> = {}) {
  fieldValues.value = Object.fromEntries(
    withdrawFields.value.map(field => [
      field.fieldKey,
      String(preset[field.fieldKey] ?? field.defaultValue ?? '').trim(),
    ]),
  )
}

async function loadAccounts() {
  accountsLoading.value = true
  try {
    allAccounts.value = await getWithdrawAccounts()
  } finally {
    accountsLoading.value = false
  }
}

async function loadBankPage() {
  pageStatus.value = 'loading'
  pageError.value = ''

  try {
    countries.value = await getAppCountries()
    selectedCountry.value = countries.value.find(country => country.countryCode === APP_COUNTRY_CODE) ?? null

    if (!selectedCountry.value) {
      pageStatus.value = 'empty'
      pageError.value = t('bank.noMarketCountry', { country: APP_COUNTRY_NAME_LOCAL })
      return
    }

    await loadAccounts()
    initFieldValues()
    pageStatus.value = 'ready'
  } catch (error) {
    pageError.value = error instanceof Error ? error.message : t('bank.loadFailed')
    pageStatus.value = 'error'
  }
}

function openAddForm() {
  if (!withdrawFields.value.length) {
    showToast(t('bank.noFields'))
    return
  }

  editingId.value = null
  initFieldValues()
  showForm.value = true
}

function openEditForm(account: WithdrawAccountItem) {
  if (!withdrawFields.value.length) {
    showToast(t('bank.noFields'))
    return
  }

  editingId.value = account.id
  initFieldValues(parseAccountData(account.accountData))
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingId.value = null
  initFieldValues()
}

function validateBankAccountFields() {
  for (const field of withdrawFields.value) {
    const value = String(fieldValues.value[field.fieldKey] ?? '').trim()

    if (field.isRequired && !value) {
      return t('bank.requiredField', { field: field.fieldLabel })
    }

    if (field.minLength && value && value.length < field.minLength) {
      return field.errorTips || t('bank.invalidField', { field: field.fieldLabel })
    }

    if (field.maxLength && value && value.length > field.maxLength) {
      return field.errorTips || t('bank.invalidField', { field: field.fieldLabel })
    }

    if (field.regexRule && value) {
      try {
        const regex = new RegExp(field.regexRule)
        if (!regex.test(value)) {
          return field.errorTips || t('bank.invalidField', { field: field.fieldLabel })
        }
      } catch {
        return field.errorTips || t('bank.invalidField', { field: field.fieldLabel })
      }
    }
  }

  return ''
}

function buildAccountData() {
  return JSON.stringify(Object.fromEntries(
    withdrawFields.value.map(field => [
      field.fieldKey,
      String(fieldValues.value[field.fieldKey] ?? '').trim(),
    ]),
  ))
}

async function submitBankAccount() {
  if (!selectedCountry.value) {
    showToast(t('bank.noMarketCountry', { country: APP_COUNTRY_NAME_LOCAL }))
    return
  }

  const message = validateBankAccountFields()
  if (message) {
    showToast(message)
    return
  }

  const accountData = buildAccountData()

  if (editingId.value) {
    await updateWithdrawAccount(editingId.value, {
      countryCode: activeCountryCode.value,
      accountData,
    })
    showToast(t('bank.toastUpdateSuccess'))
  } else {
    await addWithdrawAccount({
      countryCode: activeCountryCode.value,
      accountData,
    })
    showToast(t('bank.toastBindSuccess'))
  }

  closeForm()
  await loadAccounts()
}

function handleSubmit() {
  void submitLock.run()
}

async function handleSetDefault(account: WithdrawAccountItem) {
  if (account.isDefault === 1) return

  await setDefaultWithdrawAccount(account.id)
  showToast(t('bank.toastSetDefaultSuccess'))
  await loadAccounts()
}

async function handleDelete(account: WithdrawAccountItem) {
  try {
    await showConfirmDialog({
      title: t('bank.deleteTitle'),
      message: t('bank.deleteConfirm'),
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      confirmButtonColor: '#f0bc3d',
    })
  } catch {
    return
  }

  await deleteWithdrawAccount(account.id)
  showToast(t('bank.toastDeleteSuccess'))
  if (editingId.value === account.id) closeForm()
  await loadAccounts()
}

onMounted(() => {
  void loadBankPage()
})
</script>

<template>
  <main class="ppmx-page ppmx-subpage ppmx-bank-page">
    <PpmxPageChrome>
      <section class="ppmx-subpage__compact ppmx-bank-stack">
        <header class="ppmx-subpage-head ppmx-bank-head reveal">
          <div class="ppmx-eyebrow">{{ t('bank.eyebrow') }}</div>
          <h1 class="ppmx-display">{{ t('bank.title') }}</h1>
          <p>{{ t('bank.sub') }}</p>
        </header>

        <AppState
          v-if="pageStatus !== 'ready'"
          class="ppmx-bank-state"
          :status="pageStatus"
          :title="pageError || undefined"
          :message="pageStateMessage"
          :action-text="t('common.retry')"
          skeleton-variant="bank-page"
          @retry="loadBankPage"
        />

        <template v-else>
          <section class="ppmx-bank-country reveal">
            <span>
              <strong>{{ activeCountryCode }}</strong>
              {{ activeCountryName }}
            </span>
            <em>{{ activeCurrencyCode }}</em>
          </section>

          <section id="bankList" class="ppmx-bank-list" :aria-label="t('bank.boundAccounts')">
            <div class="ppmx-bank-section-head">
              <h2>{{ t('bank.boundAccounts') }}</h2>
              <button
                v-if="!showForm"
                class="ppmx-bank-add"
                type="button"
                @click="openAddForm"
              >
                <i class="fa-solid fa-plus" />
                <span>{{ t('bank.bindNewAccount') }}</span>
              </button>
            </div>

            <PpmxSkeleton
              v-if="accountsLoading"
              class="reveal"
              variant="bank-account-list"
              :items="2"
              :aria-label="t('common.loading')"
            />

            <template v-else>
              <div v-if="!countryAccounts.length && !showForm" class="ppmx-bank-empty reveal">
                <i class="fa-solid fa-building-columns" />
                <p>{{ t('bank.emptyTip') }}</p>
              </div>

              <article
                v-for="account in countryAccounts"
                :key="account.id"
                class="ppmx-bank-card reveal"
                :class="{ 'is-default': account.isDefault === 1 }"
              >
                <div class="ppmx-bank-card__glow" />
                <div class="ppmx-bank-card__body">
                  <div>
                    <h2>
                      <i class="fa-solid fa-building-columns" />
                      {{ t('bank.title') }}
                    </h2>
                    <div class="ppmx-bank-fields">
                      <p
                        v-for="line in getAccountDisplayLines(account)"
                        :key="line.label"
                        class="ppmx-bank-field-row"
                      >
                        <span>{{ line.label }}</span>
                        <strong>{{ line.value }}</strong>
                      </p>
                    </div>
                  </div>
                  <span>{{ account.isDefault === 1 ? t('bank.defaultBadge') : activeCountryCode }}</span>
                </div>
                <footer>
                  <button type="button" @click="openEditForm(account)">
                    <i class="fa-solid fa-pen" />
                    <span>{{ t('bank.edit') }}</span>
                  </button>
                  <button
                    v-if="account.isDefault !== 1"
                    type="button"
                    @click="handleSetDefault(account)"
                  >
                    <i class="fa-solid fa-star" />
                    <span>{{ t('bank.setDefault') }}</span>
                  </button>
                  <button type="button" @click="handleDelete(account)">
                    <i class="fa-solid fa-trash-can" />
                    <span>{{ t('bank.remove') }}</span>
                  </button>
                </footer>
              </article>
            </template>
          </section>

          <section v-if="showForm" class="ppmx-bank-form reveal">
            <div class="ppmx-bank-form__head">
              <span class="ppmx-eyebrow">
                <i class="fa-solid fa-link" />
                {{ t('bank.add') }}
              </span>
              <button class="ppmx-bank-form__close" type="button" @click="closeForm">
                <i class="fa-solid fa-xmark" />
              </button>
              <h2>{{ formTitle }}</h2>
            </div>

            <label
              v-for="field in withdrawFields"
              :key="field.fieldKey"
              class="ppmx-bank-field"
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
                :placeholder="field.fieldPlaceholder || t('bank.selectPlaceholder')"
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

            <PpmxButton
              class="ppmx-bank-submit"
              block
              size="lg"
              type="button"
              :loading="isSubmitting"
              :disabled="isSubmitting"
              @click="handleSubmit"
            >
              <i class="fa-solid fa-link" />
              {{ editingId ? t('bank.saveChanges') : t('bank.confirmBind') }}
            </PpmxButton>
          </section>

          <button
            v-if="!showForm && !accountsLoading && !countryAccounts.length"
            class="ppmx-bank-add-block reveal"
            type="button"
            @click="openAddForm"
          >
            <i class="fa-solid fa-plus" />
            <span>{{ t('bank.bindNewAccount') }}</span>
          </button>

        </template>
      </section>
    </PpmxPageChrome>
  </main>
</template>
