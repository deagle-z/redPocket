<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import type { WithdrawAccountItem } from '@/api/user'
import {
  addWithdrawAccount,
  deleteWithdrawAccount,
  getWithdrawAccounts,
  setDefaultWithdrawAccount,
  updateWithdrawAccount,
} from '@/api/user'
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

const US_COUNTRY_CODE = 'US'
const US_COUNTRY_NAME = 'United States'
const BANK_CODE_VALUES = ['cashapp', 'paypal'] as const
type BankCode = typeof BANK_CODE_VALUES[number]
const BANK_CODE_OPTIONS: Array<{ label: string, value: BankCode }> = [
  { label: 'Cash App', value: 'cashapp' },
  { label: 'PayPal', value: 'paypal' },
]
type BankFieldKey = 'bankCode' | 'bankNumber' | 'bankAccountName'

const { t } = useI18n()
const pageStatus = ref<PageStateStatus>('loading')
const pageError = ref('')
const accountsLoading = ref(false)
const allAccounts = ref<WithdrawAccountItem[]>([])
const showForm = ref(false)
const editingId = ref<number | null>(null)
const fieldValues = ref<Record<BankFieldKey, string>>({
  bankCode: '',
  bankNumber: '',
  bankAccountName: '',
})

const countryAccounts = computed(() =>
  allAccounts.value.filter(account => account.countryCode === US_COUNTRY_CODE),
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
const selectedBankCode = computed(() => normalizeBankCode(fieldValues.value.bankCode))
const bankNumberPlaceholder = computed(() =>
  selectedBankCode.value === 'paypal' ? t('bank.paypalPlaceholder') : t('bank.cashappPlaceholder'),
)
const bankNumberInputMode = computed(() => selectedBankCode.value === 'paypal' ? 'email' : 'text')

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

function getAccountDisplayLines(account: WithdrawAccountItem) {
  const data = parseAccountData(account.accountData)
  const bankCode = normalizeBankCode(readAccountValue(data, 'bankCode', 'bank_code', 'bank'))
  const bankNumber = readAccountValue(data, 'bankNumber', 'bank_number', 'accNo', 'accountNumber', 'pixKey')
  const bankAccountName = readAccountValue(data, 'bankAccountName', 'bank_account_name', 'accName', 'accountName', 'receiverName', 'name', 'fullName')

  return [
    { label: t('bank.bankCode'), value: bankCodeLabel(bankCode) || bankCode },
    { label: t('bank.bankNumber'), value: bankNumber },
    { label: t('bank.bankAccountName'), value: bankAccountName },
  ].filter(line => line.value)
}

function initFieldValues(preset: Record<string, string> = {}) {
  fieldValues.value = {
    bankCode: normalizeBankCode(readAccountValue(preset, 'bankCode', 'bank_code', 'bank')),
    bankNumber: readAccountValue(preset, 'bankNumber', 'bank_number', 'accNo', 'accountNumber', 'pixKey'),
    bankAccountName: readAccountValue(preset, 'bankAccountName', 'bank_account_name', 'accName', 'accountName', 'receiverName', 'name', 'fullName'),
  }
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
    await loadAccounts()
    pageStatus.value = 'ready'
  } catch (error) {
    pageError.value = error instanceof Error ? error.message : t('bank.loadFailed')
    pageStatus.value = 'error'
  }
}

function openAddForm() {
  editingId.value = null
  initFieldValues()
  showForm.value = true
}

function openEditForm(account: WithdrawAccountItem) {
  editingId.value = account.id
  initFieldValues(parseAccountData(account.accountData))
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingId.value = null
  initFieldValues()
}

function readAccountValue(data: Record<string, string>, ...keys: string[]) {
  const normalized = new Map<string, string>()
  for (const [key, value] of Object.entries(data)) {
    const normalizedKey = normalizeAccountKey(key)
    if (normalizedKey && !normalized.has(normalizedKey)) {
      normalized.set(normalizedKey, String(value ?? '').trim())
    }
  }
  for (const key of keys) {
    const value = normalized.get(normalizeAccountKey(key))
    if (value) return value
  }
  return ''
}

function normalizeAccountKey(key: string) {
  return String(key || '').toLowerCase().replace(/[^a-z0-9]/g, '')
}

function normalizeBankCode(value: string): BankCode | '' {
  const normalized = String(value || '').trim().toLowerCase()
  return normalized === 'cashapp' || normalized === 'paypal' ? normalized : ''
}

function bankCodeLabel(value: string) {
  return BANK_CODE_OPTIONS.find(option => option.value === value)?.label ?? ''
}

function validateBankAccountFields() {
  const bankCode = normalizeBankCode(fieldValues.value.bankCode)
  const bankNumber = fieldValues.value.bankNumber.trim()
  const bankAccountName = fieldValues.value.bankAccountName.trim()

  if (!bankCode) {
    return t('bank.invalidBankCode')
  }

  if (bankCode === 'cashapp' && !/^\$[A-Za-z][A-Za-z0-9_]{0,19}$/.test(bankNumber)) {
    return t('bank.invalidCashapp')
  }

  if (bankCode === 'paypal' && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(bankNumber)) {
    return t('bank.invalidPaypal')
  }

  if (!/^[\p{L}][\p{L}\p{M}\s.'-]{1,79}$/u.test(bankAccountName)) {
    return t('bank.invalidBankAccountName')
  }

  return ''
}

async function submitBankAccount() {
  const message = validateBankAccountFields()
  if (message) {
    showToast(message)
    return
  }

  const accountData = JSON.stringify({
    bankCode: normalizeBankCode(fieldValues.value.bankCode),
    bankNumber: fieldValues.value.bankNumber.trim(),
    bankAccountName: fieldValues.value.bankAccountName.trim(),
  })

  if (editingId.value) {
    await updateWithdrawAccount(editingId.value, {
      countryCode: US_COUNTRY_CODE,
      accountData,
    })
    showToast(t('bank.toastUpdateSuccess'))
  } else {
    await addWithdrawAccount({
      countryCode: US_COUNTRY_CODE,
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
              <strong>{{ US_COUNTRY_CODE }}</strong>
              {{ US_COUNTRY_NAME }}
            </span>
            <em>USD</em>
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
                  <span>{{ account.isDefault === 1 ? t('bank.defaultBadge') : US_COUNTRY_CODE }}</span>
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

            <label class="ppmx-bank-field">
              <span>
                {{ t('bank.bankCode') }}
                <em>*</em>
              </span>
              <PpmxSelectInput
                v-model="fieldValues.bankCode"
                :options="BANK_CODE_OPTIONS"
                :placeholder="t('bank.bankCodePlaceholder')"
                :title="t('bank.bankCode')"
              />
            </label>

            <label class="ppmx-bank-field">
              <span>
                {{ t('bank.bankNumber') }}
                <em>*</em>
              </span>
              <input
                v-model="fieldValues.bankNumber"
                type="text"
                :inputmode="bankNumberInputMode"
                :maxlength="selectedBankCode === 'paypal' ? 254 : 21"
                :placeholder="bankNumberPlaceholder"
              >
            </label>

            <label class="ppmx-bank-field">
              <span>
                {{ t('bank.bankAccountName') }}
                <em>*</em>
              </span>
              <input
                v-model="fieldValues.bankAccountName"
                type="text"
                maxlength="80"
                :placeholder="t('bank.bankAccountNamePlaceholder')"
                autocomplete="name"
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
