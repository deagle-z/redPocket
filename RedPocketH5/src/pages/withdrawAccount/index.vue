<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showConfirmDialog, showToast } from 'vant'
import type { AppCountryItem, RechargeField, WithdrawAccountItem } from '@/api/user'
import {
  addWithdrawAccount,
  deleteWithdrawAccount,
  getAppCountries,
  getCountryWithdrawFields,
  getWithdrawAccounts,
  setDefaultWithdrawAccount,
  updateWithdrawAccount,
} from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'

const router = useRouter()

// 国家
const countries = ref<AppCountryItem[]>([])
const selectedCountry = ref<AppCountryItem | null>(null)
const countriesLoading = ref(false)

// 账户列表
const allAccounts = ref<WithdrawAccountItem[]>([])
const accountsLoading = ref(false)
const countryAccounts = computed(() =>
  allAccounts.value.filter(a => a.countryCode === selectedCountry.value?.countryCode),
)

// 字段配置
const withdrawFields = ref<RechargeField[]>([])
const fieldsLoading = ref(false)

// 表单状态
const showForm = ref(false)
const editingId = ref<number | null>(null) // null = 新增
const fieldValues = ref<Record<string, string>>({})
const submitLoading = ref(false)

const formTitle = computed(() => editingId.value ? '修改账户' : '绑定账户')

function showCenterToast(message: string) {
  showToast({ message, position: 'middle', teleport: '#app', wordBreak: 'break-word' })
}

function goBack() {
  router.back()
}

// 解析 accountData JSON
function parseAccountData(raw: string): Record<string, string> {
  try {
    return JSON.parse(raw) ?? {}
  }
  catch {
    return {}
  }
}

// 用字段定义中的 label 显示 accountData 内容
function getAccountDisplayLines(account: WithdrawAccountItem): Array<{ label: string, value: string }> {
  const data = parseAccountData(account.accountData)
  if (!withdrawFields.value.length) {
    return Object.entries(data).map(([k, v]) => ({ label: k, value: String(v) }))
  }
  return withdrawFields.value
    .filter(f => data[f.fieldKey] !== undefined && data[f.fieldKey] !== '')
    .map(f => ({ label: f.fieldLabel, value: String(data[f.fieldKey]) }))
}

async function loadWithdrawFields(code: string) {
  if (withdrawFields.value.length)
    return // 已加载
  fieldsLoading.value = true
  try {
    const { data } = await getCountryWithdrawFields(code)
    withdrawFields.value = data ?? []
  }
  catch {
    withdrawFields.value = []
  }
  finally {
    fieldsLoading.value = false
  }
}

async function loadAccounts() {
  accountsLoading.value = true
  try {
    const { data } = await getWithdrawAccounts()
    allAccounts.value = data ?? []
  }
  catch {
    allAccounts.value = []
  }
  finally {
    accountsLoading.value = false
  }
}

async function handleSelectCountry(country: AppCountryItem) {
  if (selectedCountry.value?.countryCode === country.countryCode)
    return
  selectedCountry.value = country
  withdrawFields.value = []
  closeForm()
  await loadWithdrawFields(country.countryCode)
}

async function loadCountries() {
  countriesLoading.value = true
  try {
    const { data } = await getAppCountries()
    countries.value = data ?? []
    if (countries.value.length) {
      selectedCountry.value = countries.value[0]
      await Promise.all([
        loadAccounts(),
        loadWithdrawFields(countries.value[0].countryCode),
      ])
    }
  }
  catch {
    countries.value = []
  }
  finally {
    countriesLoading.value = false
  }
}

function initFieldValues(preset?: Record<string, string>) {
  const init: Record<string, string> = {}
  for (const f of withdrawFields.value)
    init[f.fieldKey] = preset?.[f.fieldKey] ?? f.defaultValue ?? ''
  fieldValues.value = init
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
  fieldValues.value = {}
}

async function handleSubmit() {
  if (!selectedCountry.value)
    return

  // 必填校验
  for (const f of withdrawFields.value) {
    if (f.isRequired === 1 && !fieldValues.value[f.fieldKey]?.trim()) {
      showCenterToast(f.errorTips || `${f.fieldLabel} 不能为空`)
      return
    }
  }

  submitLoading.value = true
  try {
    const accountData = JSON.stringify(fieldValues.value)
    if (editingId.value) {
      await updateWithdrawAccount(editingId.value, {
        countryCode: selectedCountry.value.countryCode,
        accountData,
      })
      showCenterToast('修改成功')
    }
    else {
      await addWithdrawAccount({
        countryCode: selectedCountry.value.countryCode,
        accountData,
      })
      showCenterToast('绑定成功')
    }
    closeForm()
    await loadAccounts()
  }
  catch {
    // request interceptor已提示错误
  }
  finally {
    submitLoading.value = false
  }
}

async function handleSetDefault(account: WithdrawAccountItem) {
  if (account.isDefault === 1)
    return
  try {
    await setDefaultWithdrawAccount(account.id)
    showCenterToast('已设为默认')
    await loadAccounts()
  }
  catch {}
}

async function handleDelete(account: WithdrawAccountItem) {
  try {
    await showConfirmDialog({
      title: '删除账户',
      message: '确认删除该提现账户？',
      teleport: '#app',
      confirmButtonColor: '#d4af37',
    })
    await deleteWithdrawAccount(account.id)
    showCenterToast('删除成功')
    await loadAccounts()
  }
  catch {}
}

onMounted(() => {
  loadCountries()
})
</script>

<template>
  <div class="wa-page">
    <AppPageHeader class="wa-header" title="提现账户" @back="goBack" />

    <!-- 国家选择 -->
    <div class="country-bar">
      <div class="country-scroll">
        <button
          v-for="c in countries"
          :key="c.countryCode"
          type="button"
          class="country-pill"
          :class="{ active: selectedCountry?.countryCode === c.countryCode }"
          @click="handleSelectCountry(c)"
        >
          <span class="country-code">{{ c.countryCode }}</span>
          <span class="country-name">{{ c.countryNameEn }}</span>
        </button>
        <div v-if="countriesLoading" class="country-pill-loading">
          <van-loading size="16" color="#d4af37" />
        </div>
      </div>
    </div>

    <!-- 账户列表 -->
    <section class="card">
      <div class="section-head">
        <h2>已绑定账户</h2>
        <button v-if="!showForm && withdrawFields.length" type="button" class="add-btn" @click="openAddForm">
          <van-icon name="plus" />
          绑定新账户
        </button>
      </div>

      <van-loading v-if="accountsLoading" size="20" color="#d4af37" class="section-loading" />

      <template v-else>
        <div v-if="countryAccounts.length" class="account-list">
          <div
            v-for="account in countryAccounts"
            :key="account.id"
            class="account-card"
            :class="{ 'account-card--default': account.isDefault === 1 }"
          >
            <div class="account-card__head">
              <span v-if="account.isDefault === 1" class="default-badge">默认</span>
              <span class="account-country">{{ account.countryCode }}</span>
              <div class="account-actions">
                <button type="button" class="action-btn" @click="openEditForm(account)">
                  <van-icon name="edit" />
                </button>
                <button
                  v-if="account.isDefault !== 1"
                  type="button"
                  class="action-btn"
                  @click="handleSetDefault(account)"
                >
                  <van-icon name="star-o" />
                </button>
                <button type="button" class="action-btn action-btn--del" @click="handleDelete(account)">
                  <van-icon name="delete-o" />
                </button>
              </div>
            </div>
            <div class="account-fields">
              <div
                v-for="line in getAccountDisplayLines(account)"
                :key="line.label"
                class="account-field-row"
              >
                <span class="field-label">{{ line.label }}</span>
                <span class="field-value">{{ line.value }}</span>
              </div>
            </div>
          </div>
        </div>

        <p v-else-if="!showForm" class="empty-tip">
          暂无绑定账户，请点击下方绑定
        </p>
      </template>

      <!-- 无账户时直接展示绑定按钮 -->
      <button
        v-if="!showForm && !accountsLoading && countryAccounts.length === 0 && withdrawFields.length"
        type="button"
        class="add-btn-block"
        @click="openAddForm"
      >
        <van-icon name="plus" />
        绑定账户
      </button>
    </section>

    <!-- 新增 / 编辑表单 -->
    <section v-if="showForm" class="card form-card">
      <div class="section-head">
        <h2>{{ formTitle }}</h2>
        <button type="button" class="close-btn" @click="closeForm">
          <van-icon name="cross" />
        </button>
      </div>

      <van-loading v-if="fieldsLoading" size="20" color="#d4af37" class="section-loading" />

      <template v-else-if="withdrawFields.length">
        <van-field
          v-for="field in withdrawFields"
          :key="field.fieldKey"
          v-model="fieldValues[field.fieldKey]"
          :type="field.fieldType === 'number' ? 'number' : field.fieldType === 'textarea' ? 'textarea' : 'text'"
          :label="field.fieldLabel"
          :placeholder="field.fieldPlaceholder || ''"
          :required="field.isRequired === 1"
          :maxlength="field.maxLength ?? undefined"
          class="custom-input wa-field"
          rows="3"
        />

        <van-button
          type="primary"
          round
          block
          class="submit-btn"
          :loading="submitLoading"
          @click="handleSubmit"
        >
          {{ editingId ? '保存修改' : '确认绑定' }}
        </van-button>
      </template>

      <p v-else class="empty-tip">
        该国家暂未配置提现字段
      </p>
    </section>
  </div>
</template>

<route lang="json5">
{
  name: "WithdrawAccount"
}
</route>

<style scoped>
.wa-page {
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
  padding: 0 12px calc(60px + env(safe-area-inset-bottom));
}

.wa-header {
  margin-bottom: 10px;
}

/* 国家条 */
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

/* 卡片 */
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

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.card h2 {
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #ffd98b;
  margin: 0;
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

/* 添加按钮（header右侧小按钮）*/
.add-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.5);
  background: rgba(255, 248, 214, 0.08);
  color: #ffd87f;
  font-size: 12px;
  font-weight: 700;
}

/* 无账户时大块绑定按钮 */
.add-btn-block {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  padding: 14px;
  margin-top: 8px;
  border-radius: 14px;
  border: 1.5px dashed rgba(212, 175, 55, 0.4);
  background: rgba(255, 248, 214, 0.04);
  color: rgba(255, 216, 127, 0.75);
  font-size: 14px;
  font-weight: 700;
}

/* 账户卡片 */
.account-list {
  display: grid;
  gap: 10px;
}

.account-card {
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.05);
  padding: 12px 14px;
}

.account-card--default {
  border-color: rgba(212, 175, 55, 0.55);
  background: linear-gradient(165deg, rgba(255, 223, 135, 0.1), rgba(116, 24, 0, 0.15));
}

.account-card__head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.default-badge {
  padding: 2px 8px;
  border-radius: 8px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 11px;
  font-weight: 800;
}

.account-country {
  font-size: 12px;
  font-weight: 700;
  color: rgba(255, 229, 186, 0.55);
  flex: 1;
}

.account-actions {
  display: flex;
  gap: 6px;
}

.action-btn {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  background: rgba(255, 248, 214, 0.07);
  color: #ffd87f;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
}

.action-btn--del {
  color: rgba(255, 100, 100, 0.85);
  border-color: rgba(255, 100, 100, 0.2);
}

.account-fields {
  display: grid;
  gap: 6px;
}

.account-field-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 13px;
}

.field-label {
  color: rgba(255, 229, 186, 0.55);
  flex-shrink: 0;
  min-width: 80px;
}

.field-value {
  color: #fff0c9;
  word-break: break-all;
  font-weight: 600;
}

/* 表单 */
.form-card {
  margin-top: 4px;
}

.close-btn {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 1px solid rgba(212, 175, 55, 0.24);
  background: rgba(255, 248, 214, 0.08);
  color: rgba(255, 229, 186, 0.7);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.wa-field {
  margin-bottom: 10px;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(212, 175, 55, 0.18);
  background: rgba(255, 248, 214, 0.05);
}

:deep(.wa-field .van-field__label),
:deep(.wa-field .van-field__control) {
  color: #fff0c9;
}

:deep(.wa-field .van-field__control::placeholder) {
  color: rgba(255, 229, 186, 0.4);
}

.custom-input {
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

.submit-btn {
  margin-top: 8px;
  border: none;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%) !important;
  color: #5a1b00 !important;
  font-weight: 800;
  letter-spacing: 0.08em;
  box-shadow: 0 12px 22px rgba(75, 25, 0, 0.28);
}
</style>
