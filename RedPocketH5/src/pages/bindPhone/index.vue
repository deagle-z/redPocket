<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { bindCurrentTgPhone, getCurrentTgUserInfo, sendRegisterSMSCode } from '@/api/user'
import { safeBack } from '@/utils/navigation'
import { getAuthCountry } from '@/utils/auth'
import { normalizeNationalPhone, onlyPhoneDigits } from '@/utils/phone'
import AppPageHeader from '@/components/AppPageHeader.vue'
import languageIcon from '@/assets/svg/language.svg'
import verifyIcon from '@/assets/svg/verify.svg'

const { t } = useI18n()
const router = useRouter()

const submitting = ref(false)
const sendLoading = ref(false)
const countdown = ref(0)
const boundPhone = ref('')
const boundCountry = ref('')

const bindPhoneCountries = [
  { code: 'BR', nameKey: 'bindPhonePage.countryBrazil', dialCode: '+55' },
  { code: 'MX', nameKey: 'bindPhonePage.countryMexico', dialCode: '+52' },
  { code: 'ID', nameKey: 'bindPhonePage.countryIndonesia', dialCode: '+62' },
] as const
type BindPhoneCountryCode = typeof bindPhoneCountries[number]['code']
const countryMap = Object.fromEntries(bindPhoneCountries.map(item => [item.code, item])) as Record<BindPhoneCountryCode, typeof bindPhoneCountries[number]>

const formData = reactive<{
  country: BindPhoneCountryCode
  phone: string
  code: string
}>({
  country: 'BR',
  phone: '',
  code: '',
})

let countdownTimer: ReturnType<typeof setInterval> | null = null

function goBack() {
  safeBack(router)
}

function normalizePhoneInput(event: Event) {
  const input = event.target as HTMLInputElement
  formData.phone = onlyPhoneDigits(input.value)
}

function startCountdown() {
  countdown.value = 60
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(countdownTimer!)
      countdownTimer = null
      countdown.value = 0
    }
  }, 1000)
}

function formatBoundPhone() {
  const country = boundCountry.value as BindPhoneCountryCode
  const dialCode = countryMap[country]?.dialCode || ''
  if (!boundPhone.value)
    return t('bindPhonePage.phonePlaceholder')
  return `${dialCode} ${boundPhone.value}`
}

async function sendCode() {
  const phone = normalizeNationalPhone(formData.country, formData.phone)
  formData.phone = phone
  if (!phone) {
    showToast(t('bindPhonePage.toastEnterPhone'))
    return
  }
  if (sendLoading.value || countdown.value > 0)
    return
  try {
    sendLoading.value = true
    await sendRegisterSMSCode(phone, formData.country)
    startCountdown()
    showToast(t('bindPhonePage.toastSendSuccess'))
  }
  catch (error: any) {
    showToast(error?.message || t('bindPhonePage.toastSendFailed'))
  }
  finally {
    sendLoading.value = false
  }
}

async function submitBindPhone() {
  const phone = normalizeNationalPhone(formData.country, formData.phone)
  formData.phone = phone
  if (!phone) {
    showToast(t('bindPhonePage.toastEnterPhone'))
    return
  }
  if (!formData.code) {
    showToast(t('bindPhonePage.toastEnterCode'))
    return
  }
  if (submitting.value)
    return
  try {
    submitting.value = true
    await bindCurrentTgPhone({
      phone,
      country: formData.country,
      code: formData.code.trim(),
    })
    showToast(t('bindPhonePage.toastBindSuccess'))
    safeBack(router)
  }
  catch (error: any) {
    showToast(error?.message || t('bindPhonePage.toastBindFailed'))
  }
  finally {
    submitting.value = false
  }
}

async function loadCurrentPhone() {
  try {
    const { data } = await getCurrentTgUserInfo()
    boundPhone.value = String(data?.phone || '').trim()
    const country = String(data?.country || getAuthCountry() || 'BR').toUpperCase()
    formData.country = countryMap[country as BindPhoneCountryCode] ? country as BindPhoneCountryCode : 'BR'
    boundCountry.value = formData.country
    if (boundPhone.value)
      formData.phone = boundPhone.value
  }
  catch {}
}

onMounted(() => {
  void loadCurrentPhone()
})

onUnmounted(() => {
  if (countdownTimer)
    clearInterval(countdownTimer)
})
</script>

<template>
  <div class="bind-phone-page">
    <div class="bind-phone-shell">
      <AppPageHeader class="bind-phone-header" :title="t('bindPhonePage.title')" @back="goBack" />

      <section class="hero-card">
        <p class="hero-kicker">
          {{ boundPhone ? t('bindPhonePage.currentPhone') : t('bindPhonePage.title') }}
        </p>
        <p class="hero-value">
          {{ formatBoundPhone() }}
        </p>
        <p class="hero-copy">
          {{ boundPhone ? t('bindPhonePage.changeHint') : t('bindPhonePage.bindHint') }}
        </p>
      </section>

      <section class="form-card">
        <div class="field-row">
          <label class="field-label">
            <span class="field-icon-wrap">
              <img :src="languageIcon" alt="" class="field-icon">
            </span>
            <span>{{ t('bindPhonePage.country') }}</span>
          </label>
          <div class="country-options">
            <button
              v-for="item in bindPhoneCountries"
              :key="item.code"
              type="button"
              class="country-btn"
              :class="{ active: formData.country === item.code }"
              @click="formData.country = item.code"
            >
              <span>{{ t(item.nameKey) }}</span>
              <em>{{ item.dialCode }}</em>
            </button>
          </div>
        </div>

        <div class="field-row">
          <label for="bind-phone" class="field-label">
            <span class="field-icon-wrap">
              <van-icon name="phone-o" />
            </span>
            <span>{{ t('bindPhonePage.phone') }}</span>
          </label>
          <div class="phone-input-wrap">
            <span class="phone-dial-code">{{ countryMap[formData.country].dialCode }}</span>
            <input
              id="bind-phone"
              v-model="formData.phone"
              type="tel"
              inputmode="tel"
              autocomplete="tel"
              class="field-input phone-input"
              :placeholder="t('bindPhonePage.phonePlaceholder')"
              @input="normalizePhoneInput"
            >
          </div>
        </div>

        <div class="field-row">
          <label for="bind-phone-code" class="field-label">
            <span class="field-icon-wrap">
              <img :src="verifyIcon" alt="" class="field-icon">
            </span>
            <span>{{ t('bindPhonePage.code') }}</span>
          </label>
          <div class="code-input-wrap">
            <input
              id="bind-phone-code"
              v-model="formData.code"
              type="text"
              inputmode="numeric"
              maxlength="6"
              class="field-input"
              :placeholder="t('bindPhonePage.codePlaceholder')"
            >
            <button type="button" class="send-btn" :disabled="sendLoading || countdown > 0" @click="sendCode">
              <span v-if="countdown > 0">{{ countdown }}s</span>
              <span v-else-if="sendLoading">{{ t('bindPhonePage.sending') }}</span>
              <span v-else>{{ t('bindPhonePage.send') }}</span>
            </button>
          </div>
        </div>
      </section>

      <button type="button" class="submit-btn" :disabled="submitting" @click="submitBindPhone">
        {{ submitting ? t('bindPhonePage.submitting') : (boundPhone ? t('bindPhonePage.changeSubmit') : t('bindPhonePage.submit')) }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.bind-phone-page {
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

.bind-phone-shell {
  width: 100%;
  max-width: 640px;
  margin: 0 auto;
}

.bind-phone-header {
  margin-bottom: 12px;
}

.hero-card,
.form-card {
  position: relative;
  overflow: hidden;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  border-radius: 18px;
  border: 1px solid rgba(212, 175, 55, 0.22);
  box-shadow:
    0 16px 34px rgba(0, 0, 0, 0.26),
    inset 0 1px 0 rgba(255, 248, 214, 0.08);
}

.hero-card::before,
.form-card::before {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 2px;
  background: linear-gradient(90deg, transparent, rgba(255, 215, 0, 0.82), transparent);
}

.hero-card {
  margin-bottom: 12px;
  padding: 16px;
}

.hero-kicker {
  margin: 0 0 8px;
  color: rgba(255, 229, 186, 0.58);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.hero-value {
  margin: 0;
  color: #fff0c9;
  font-size: 18px;
  font-weight: 800;
  line-height: 1.35;
}

.hero-copy {
  margin: 8px 0 0;
  color: rgba(255, 229, 186, 0.6);
  font-size: 12px;
  line-height: 1.5;
}

.form-card {
  padding: 10px 14px;
}

.field-row {
  min-height: 84px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 12px;
}

.field-row + .field-row {
  border-top: 1px solid rgba(212, 175, 55, 0.14);
}

.field-label {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #ffe09a;
  font-size: 13px;
  font-weight: 700;
}

.field-icon-wrap {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 10px;
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.18), rgba(212, 175, 55, 0.08));
  border: 1px solid rgba(212, 175, 55, 0.26);
}

.field-icon,
.field-icon-wrap .van-icon {
  width: 16px;
  height: 16px;
  color: #ffe09a;
}

.country-options {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.country-btn {
  min-height: 44px;
  border: 1px solid rgba(212, 175, 55, 0.18);
  border-radius: 13px;
  background: rgba(255, 248, 214, 0.04);
  color: rgba(255, 240, 201, 0.72);
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 3px;
  font-size: 12px;
  font-weight: 700;
}

.country-btn em {
  color: rgba(255, 229, 186, 0.46);
  font-size: 10px;
  font-style: normal;
}

.country-btn.active {
  border-color: rgba(255, 223, 135, 0.72);
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.2), rgba(212, 175, 55, 0.08));
  color: #ffe09a;
}

.phone-input-wrap,
.code-input-wrap {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
}

.phone-dial-code {
  min-width: 56px;
  height: 48px;
  border: 1px solid rgba(212, 175, 55, 0.22);
  border-radius: 14px;
  background: rgba(255, 248, 214, 0.05);
  color: #ffd98b;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 800;
}

.field-input {
  min-width: 0;
  width: 100%;
  border: 1px solid rgba(212, 175, 55, 0.22);
  border-radius: 14px;
  outline: none;
  background: rgba(255, 248, 214, 0.05);
  color: #fff4d1;
  font-size: 14px;
  min-height: 48px;
  padding: 0 14px;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.field-input::placeholder {
  color: rgba(255, 229, 186, 0.42);
}

.field-input:focus {
  border-color: rgba(255, 223, 135, 0.72);
  box-shadow: 0 0 0 4px rgba(212, 175, 55, 0.14);
  background: rgba(255, 248, 214, 0.08);
}

.send-btn {
  flex: 0 0 96px;
  height: 48px;
  border: 1px solid rgba(255, 248, 214, 0.28);
  border-radius: 14px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 13px;
  font-weight: 800;
}

.send-btn:disabled {
  opacity: 0.66;
}

.submit-btn {
  margin-top: 18px;
  width: 100%;
  height: 52px;
  border: 1px solid rgba(255, 248, 214, 0.34);
  border-radius: 999px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 16px;
  font-weight: 800;
  box-shadow:
    0 14px 26px rgba(0, 0, 0, 0.18),
    0 8px 18px rgba(90, 27, 0, 0.24);
}

.submit-btn:disabled {
  opacity: 0.72;
}

@media (max-width: 390px) {
  .bind-phone-page {
    padding-left: 10px;
    padding-right: 10px;
  }

  .send-btn {
    flex-basis: 84px;
  }
}
</style>
