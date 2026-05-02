<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { RouteMap } from 'vue-router'
import { useUserStore } from '@/stores'
import { getAuthCountry, setAuthCountry } from '@/utils/auth'
import { trackAttributionEvent } from '@/utils/attribution'
import { getSourceChannelCode } from '@/utils/source-channel'
import { showToast } from 'vant'
import { languageOptions, locale } from '@/utils/i18n'
import { safeBack } from '@/utils/navigation'
import AppPageHeader from '@/components/AppPageHeader.vue'
import emailIcon from '@/assets/svg/email.svg'
import lockIcon from '@/assets/svg/lock.svg'
import languageIcon from '@/assets/svg/language.svg'
import inviteIcon from '@/assets/svg/invite.svg'
import avatarIcon from '@/assets/svg/avatar.svg'
import imgRegisterHeader from '@/assets/images/register-header.jpg'
import imgTelegram from '@/assets/images/telegram.png'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const showLangPopup = ref(false)
const showCountryPopup = ref(false)
const tgBotId = Number(import.meta.env.VITE_TG_BOT_ID || 0)
const registerHeaderImage = imgRegisterHeader
const registerCountries = [
  { code: 'MX', nameKey: 'register.countryMexico', dialCode: '+52' },
  { code: 'ID', nameKey: 'register.countryIndonesia', dialCode: '+62' },
  { code: 'BR', nameKey: 'register.countryBrazil', dialCode: '+55' },
] as const
type RegisterCountryCode = typeof registerCountries[number]['code']
const registerCountryMap = Object.fromEntries(registerCountries.map(item => [item.code, item.nameKey])) as Record<RegisterCountryCode, string>
const registerCountryDialCodeMap = Object.fromEntries(registerCountries.map(item => [item.code, item.dialCode])) as Record<
  RegisterCountryCode,
  string
>
const registerPhoneRules: Record<RegisterCountryCode, RegExp> = {
  MX: /^[2-9]\d{9}$/,
  ID: /^0?8\d{8,11}$/,
  BR: /^[1-9]{2}9\d{8}$/,
}

const postData = reactive<{
  country: RegisterCountryCode
  phone: string
  firstName: string
  password: string
  confirmPassword: string
  inviteCode: string
}>({
  country: 'BR',
  phone: '',
  firstName: '',
  password: '',
  confirmPassword: '',
  inviteCode: '',
})

const currentCountryLabel = computed(() => t(registerCountryMap[postData.country] || 'register.countryBrazil'))
const currentCountryDialCode = computed(() => registerCountryDialCodeMap[postData.country] || '+55')

function detectRegisterCountry(): RegisterCountryCode {
  const browserLanguages = [
    ...(navigator.languages || []),
    navigator.language,
  ]
    .filter(Boolean)
    .map(lang => lang.toUpperCase())

  for (const lang of browserLanguages) {
    if (lang.includes('-BR') || lang.startsWith('PT'))
      return 'BR'
    if (lang.includes('-MX') || lang.startsWith('ES'))
      return 'MX'
    if (lang.includes('-ID') || lang.startsWith('ID'))
      return 'ID'
  }

  return 'BR'
}

function normalizePhoneInput(event: Event) {
  const input = event.target as HTMLInputElement
  postData.phone = input.value.replace(/\D+/g, '')
}

function isValidRegisterPhone(country: RegisterCountryCode, phone: string) {
  return registerPhoneRules[country].test(phone)
}

onMounted(() => {
  postData.country = (getAuthCountry() as RegisterCountryCode) || detectRegisterCountry()
  const queryCode = String(router.currentRoute.value.query.c || '').trim()
  const localCode = String(localStorage.getItem('invite_code') || '').trim()
  const inviteCode = queryCode || localCode
  if (inviteCode)
    postData.inviteCode = inviteCode
  if (queryCode)
    localStorage.setItem('invite_code', queryCode)
})

async function register() {
  const phone = postData.phone.replace(/\D+/g, '')
  postData.phone = phone

  if (!postData.country) {
    showToast(t('register.pleaseSelectCountry'))
    return
  }
  if (!phone) {
    showToast(t('register.pleaseEnterPhone'))
    return
  }
  if (!isValidRegisterPhone(postData.country, phone)) {
    showToast(t('register.invalidPhone'))
    return
  }
  const firstName = postData.firstName.trim()
  postData.firstName = firstName
  if (!firstName) {
    showToast(t('register.pleaseEnterNickname'))
    return
  }
  if ([...firstName].length > 128) {
    showToast(t('register.nicknameTooLong'))
    return
  }
  if (!postData.password) {
    showToast(t('register.pleaseEnterPassword'))
    return
  }
  if (!postData.confirmPassword) {
    showToast(t('register.pleaseEnterConfirmPassword'))
    return
  }
  if (postData.password !== postData.confirmPassword) {
    showToast(t('register.passwordsDoNotMatch'))
    return
  }
  try {
    loading.value = true
    trackAttributionEvent({
      eventName: 'register_submit',
      metadata: {
        country: postData.country,
        method: 'phone',
      },
    })
    await userStore.register({
      phone,
      country: postData.country,
      firstName,
      password: postData.password,
      inviteCode: postData.inviteCode.trim(),
      sourceChannelCode: getSourceChannelCode(),
    })
    trackAttributionEvent({
      eventName: 'register_success',
      metadata: {
        country: postData.country,
        method: 'phone',
      },
    })
    setAuthCountry(postData.country)
    showToast(t('register.registerSuccess'))
    router.push({ name: 'Login' as keyof RouteMap })
  }
  finally {
    loading.value = false
  }
}

interface TgAuthPayload {
  id: number
  first_name?: string
  last_name?: string
  username?: string
  photo_url?: string
  auth_date: number
  hash: string
}

async function handleTelegramLogin() {
  const tgLogin = (window as any)?.Telegram?.Login
  if (!tgLogin?.auth) {
    showToast(t('login.tgUnavailable'))
    return
  }
  if (!tgBotId) {
    showToast(t('login.missingBotConfig'))
    return
  }
  try {
    loading.value = true
    await new Promise<void>((resolve, reject) => {
      tgLogin.auth({ bot_id: tgBotId }, async (data: TgAuthPayload | false) => {
        if (!data) {
          reject(new Error(t('login.tgAuthCancelled')))
          return
        }
        try {
          trackAttributionEvent({
            eventName: 'telegram_auth_submit',
            metadata: {
              source: 'register',
            },
          })
          await userStore.loginByTelegram(data)
          trackAttributionEvent({
            eventName: 'telegram_auth_success',
            metadata: {
              source: 'register',
            },
          })
          resolve()
        }
        catch (error) {
          reject(error)
        }
      })
    })
    const { redirect, ...othersQuery } = router.currentRoute.value.query
    router.push({
      name: (redirect as keyof RouteMap) || 'Home',
      query: { ...othersQuery },
    })
  }
  catch (error: any) {
    showToast(error?.message || t('login.tgLoginFailed'))
  }
  finally {
    loading.value = false
  }
}

function goBack() {
  safeBack(router)
}

function goLogin() {
  router.push({ name: 'Login' as keyof RouteMap })
}

function openLanguagePopup() {
  showLangPopup.value = true
}

function closeLanguagePopup() {
  showLangPopup.value = false
}

function openCountryPopup() {
  showCountryPopup.value = true
}

function closeCountryPopup() {
  showCountryPopup.value = false
}

function selectCountry(country: RegisterCountryCode) {
  postData.country = country
  showCountryPopup.value = false
}

function selectLanguage(lang: string) {
  if (locale.value === lang) {
    showLangPopup.value = false
    return
  }
  locale.value = lang
  showToast(t('login.language.changed'))
  showLangPopup.value = false
  setTimeout(() => {
    window.location.reload()
  }, 350)
}
</script>

<template>
  <div class="register-page">
    <div class="register-shell">
      <AppPageHeader
        class="register-header"
        :title="t('register.pageTitle')"
        @back="goBack"
        @right-click="openLanguagePopup"
      >
        <template #right>
          <img :src="languageIcon" class="lang-icon" alt="language icon">
        </template>
      </AppPageHeader>

      <section class="hero-card">
        <img
          class="hero-image"
          :src="registerHeaderImage"
          alt="register banner"
        >
        <div class="hero-content">
          <p class="hero-eyebrow">
            {{ t('appTopHeader.brandSubtitle') }}
          </p>
          <h2 class="hero-title">
            {{ t('register.pageTitle') }}
          </h2>
          <p class="hero-desc">
            {{ t('register.telegramLogin') }}
          </p>
        </div>
      </section>

      <section class="auth-card">
        <button type="button" class="tg-entry" @click="handleTelegramLogin">
          <span class="tg-entry__media">
            <img :src="imgTelegram" alt="telegram" class="tg-entry__icon">
          </span>
          <span class="tg-entry__content">
            <strong>{{ t('register.telegramLogin') }}</strong>
            <span>{{ t('login.telegramAuth') }}</span>
          </span>
          <span class="tg-entry__arrow">→</span>
        </button>

        <div class="form-card">
          <div class="form-row">
            <label class="form-label">
              <span class="icon-wrap">
                <img :src="languageIcon" alt="country" class="form-icon">
              </span>
              <span>{{ t('register.country') }}</span>
            </label>
            <button type="button" class="country-trigger" @click="openCountryPopup">
              <span>{{ currentCountryLabel }}</span>
              <span class="country-trigger-arrow" aria-hidden="true">▾</span>
            </button>
          </div>

          <div class="form-row">
            <label for="register-phone" class="form-label">
              <span class="icon-wrap">
                <img :src="emailIcon" alt="phone" class="form-icon">
              </span>
              <span>{{ t('register.phone') }}</span>
            </label>
            <div class="phone-input-wrap">
              <span class="phone-dial-code">{{ currentCountryDialCode }}</span>
              <input
                id="register-phone"
                v-model="postData.phone"
                type="tel"
                inputmode="tel"
                pattern="[0-9]*"
                autocomplete="tel-national"
                class="form-input phone-input"
                :placeholder="t('register.pleaseEnterPhone')"
                @input="normalizePhoneInput"
              >
            </div>
          </div>

          <div class="form-row">
            <label for="register-nickname" class="form-label">
              <span class="icon-wrap">
                <img :src="avatarIcon" alt="nickname" class="form-icon">
              </span>
              <span>{{ t('register.nickname') }}</span>
            </label>
            <input
              id="register-nickname"
              v-model="postData.firstName"
              type="text"
              autocomplete="nickname"
              class="form-input"
              maxlength="128"
              :placeholder="t('register.pleaseEnterNickname')"
            >
          </div>

          <div class="form-row">
            <label for="register-password" class="form-label">
              <span class="icon-wrap">
                <img :src="lockIcon" alt="password" class="form-icon">
              </span>
              <span>{{ t('register.password') }}</span>
            </label>
            <input
              id="register-password"
              v-model="postData.password"
              type="password"
              autocomplete="new-password"
              class="form-input"
              :placeholder="t('register.pleaseEnterPassword')"
            >
          </div>

          <div class="form-row">
            <label for="register-confirm-password" class="form-label">
              <span class="icon-wrap">
                <img :src="lockIcon" alt="confirm password" class="form-icon">
              </span>
              <span>{{ t('register.confirmPassword') }}</span>
            </label>
            <input
              id="register-confirm-password"
              v-model="postData.confirmPassword"
              type="password"
              autocomplete="new-password"
              class="form-input"
              :placeholder="t('register.pleaseEnterConfirmPassword')"
            >
          </div>

          <div class="form-row">
            <label for="register-invite-code" class="form-label">
              <span class="icon-wrap">
                <img :src="inviteIcon" alt="invite code" class="form-icon">
              </span>
              <span>{{ t('register.inviteCode') }}</span>
            </label>
            <input
              id="register-invite-code"
              v-model="postData.inviteCode"
              type="text"
              class="form-input"
              :placeholder="t('register.pleaseEnterInviteCode')"
            >
          </div>
        </div>

        <van-button
          :loading="loading"
          type="primary"
          round
          block
          class="register-btn"
          @click="register"
        >
          {{ t('register.confirm') }}
        </van-button>

        <p class="login-text">
          {{ t('register.alreadyHaveAccount') }}
          <button type="button" class="login-link" @click="goLogin">
            {{ t('register.loginNow') }}
          </button>
        </p>
      </section>

      <section class="feature-grid">
        <article class="feature-card">
          <div class="feature-icon feature-icon-game" aria-hidden="true">
            G
          </div>
          <div class="feature-title">
            {{ t('login.feature.game.title') }}
          </div>
          <div class="feature-desc">
            {{ t('login.feature.game.desc') }}
          </div>
        </article>
        <article class="feature-card">
          <div class="feature-icon feature-icon-coin" aria-hidden="true">
            C
          </div>
          <div class="feature-title">
            {{ t('login.feature.coin.title') }}
          </div>
          <div class="feature-desc">
            {{ t('login.feature.coin.desc') }}
          </div>
        </article>
        <article class="feature-card">
          <div class="feature-icon feature-icon-invite" aria-hidden="true">
            <img :src="inviteIcon" alt="" class="feature-icon-img">
          </div>
          <div class="feature-title">
            {{ t('login.feature.invite.title') }}
          </div>
          <div class="feature-desc">
            {{ t('login.feature.invite.desc') }}
          </div>
        </article>
      </section>
    </div>

    <van-popup v-model:show="showLangPopup" round position="bottom" class="language-popup">
      <div class="language-popup-header">
        <span class="language-popup-title">{{ t('login.language.title') }}</span>
        <button class="language-popup-close" @click="closeLanguagePopup">
          ×
        </button>
      </div>

      <div class="language-list">
        <button
          v-for="item in languageOptions"
          :key="item.value"
          class="language-item"
          :class="{ active: locale === item.value }"
          @click="selectLanguage(item.value)"
        >
          <span class="language-code">{{ item.code }}</span>
          <span class="language-text">
            <span class="native">{{ t(item.nativeTextKey) }}</span>
            <span class="english">{{ t(item.englishTextKey) }}</span>
          </span>
          <span v-if="locale === item.value" class="language-check">✓</span>
        </button>
      </div>

      <p class="language-tip">
        {{ t('login.language.autoRefresh') }}
      </p>
    </van-popup>

    <van-popup v-model:show="showCountryPopup" round position="bottom" class="language-popup">
      <div class="language-popup-header">
        <span class="language-popup-title">{{ t('register.country') }}</span>
        <button class="language-popup-close" @click="closeCountryPopup">
          ×
        </button>
      </div>

      <div class="language-list">
        <button
          v-for="item in registerCountries"
          :key="item.code"
          class="language-item"
          :class="{ active: postData.country === item.code }"
          @click="selectCountry(item.code)"
        >
          <span class="language-code">{{ item.code }}</span>
          <span class="language-text">
            <span class="native">{{ t(item.nameKey) }}</span>
          </span>
          <span v-if="postData.country === item.code" class="language-check">✓</span>
        </button>
      </div>
    </van-popup>
  </div>
</template>

<style scoped>
.register-page {
  min-height: 100vh;
  background-image:
    radial-gradient(circle at 16% 8%, rgba(212, 175, 55, 0.18), transparent 28%),
    radial-gradient(circle at 84% 90%, rgba(255, 215, 0, 0.12), transparent 24%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.04) 18px,
      rgba(212, 175, 55, 0.04) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #230000 60%, #160000 100%);
  color: #fff0c9;
  padding: 0 12px calc(28px + env(safe-area-inset-bottom));
}

.register-shell {
  width: 100%;
  max-width: 640px;
  margin: 0 auto;
  position: relative;
}

.register-header {
  margin-bottom: 12px;
}

.lang-icon {
  width: 24px;
  height: 24px;
  filter: brightness(0) saturate(100%) invert(84%) sepia(39%) saturate(612%) hue-rotate(338deg) brightness(105%)
    contrast(96%);
}

.hero-card,
.auth-card,
.feature-card {
  position: relative;
  overflow: hidden;
  border-radius: 18px;
  border: 1px solid rgba(212, 175, 55, 0.38);
  box-shadow:
    0 14px 28px rgba(0, 0, 0, 0.34),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
}

.hero-card::after,
.auth-card::after,
.feature-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}

.hero-card {
  min-height: 188px;
  margin-top: 8px;
  background: linear-gradient(155deg, rgba(122, 0, 0, 0.96) 0%, rgba(70, 0, 0, 0.97) 55%, rgba(38, 0, 0, 0.98) 100%);
}

.hero-image {
  position: absolute;
  inset: 0;
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  opacity: 0.24;
}

.hero-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(105deg, rgba(62, 0, 0, 0.92) 10%, rgba(62, 0, 0, 0.6) 45%, rgba(62, 0, 0, 0.92) 100%),
    radial-gradient(circle at 82% 18%, rgba(212, 175, 55, 0.16), transparent 22%);
}

.hero-content {
  position: relative;
  z-index: 1;
  padding: 22px 20px 20px;
}

.hero-eyebrow {
  margin: 0 0 6px;
  color: #ffd98b;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.hero-title {
  margin: 0;
  color: #fff0c9;
  font-size: 28px;
  line-height: 1.12;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.hero-desc {
  margin: 10px 0 0;
  max-width: 240px;
  color: rgba(255, 229, 186, 0.78);
  font-size: 13px;
  line-height: 1.45;
}

.auth-card {
  margin-top: 14px;
  background:
    radial-gradient(rgba(212, 175, 55, 1) 1px, transparent 1px),
    linear-gradient(160deg, rgba(116, 0, 0, 0.96), rgba(52, 0, 0, 0.98));
  background-size:
    18px 18px,
    100% 100%;
  padding: 14px;
}

.tg-entry {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid rgba(212, 175, 55, 0.24);
  border-radius: 16px;
  background: rgba(255, 248, 214, 0.06);
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  box-shadow: inset 0 1px 0 rgba(255, 248, 214, 0.08);
  transition:
    transform 0.2s ease,
    background-color 0.2s ease;
}

.tg-entry:active {
  transform: translateY(1px);
}

.tg-entry__media {
  width: 46px;
  height: 46px;
  border-radius: 14px;
  background: linear-gradient(180deg, rgba(255, 248, 214, 0.16), rgba(212, 175, 55, 0.08));
  border: 1px solid rgba(212, 175, 55, 0.24);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
}

.tg-entry__icon {
  width: 28px;
  height: 28px;
}

.tg-entry__content {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 3px;
  text-align: left;
}

.tg-entry__content strong {
  color: #fff0c9;
  font-size: 14px;
  line-height: 1.35;
  font-weight: 800;
}

.tg-entry__content span {
  color: rgba(255, 229, 186, 0.66);
  font-size: 12px;
  line-height: 1.3;
}

.tg-entry__arrow {
  color: #ffd98b;
  font-size: 18px;
  line-height: 1;
}

.form-card {
  position: relative;
  overflow: hidden;
  margin-top: 14px;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  border-radius: 18px;
  border: 1px solid rgba(212, 175, 55, 0.24);
  box-shadow: inset 0 1px 0 rgba(255, 248, 214, 0.08);
  padding: 8px 14px;
}

.form-row {
  min-height: 86px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 12px;
  padding: 12px 0;
}

.form-row + .form-row {
  border-top: 1px solid rgba(212, 175, 55, 0.14);
}

.form-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  font-weight: 700;
  color: #ffe09a;
  cursor: text;
  letter-spacing: 0.02em;
}

.icon-wrap {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 10px;
  background: linear-gradient(180deg, rgba(255, 223, 135, 0.18), rgba(212, 175, 55, 0.08));
  border: 1px solid rgba(212, 175, 55, 0.26);
  box-shadow: inset 0 1px 0 rgba(255, 248, 214, 0.08);
}

.form-icon {
  width: 16px;
  height: 16px;
  filter: brightness(0) saturate(100%) invert(85%) sepia(39%) saturate(649%) hue-rotate(335deg) brightness(105%)
    contrast(97%);
}

.form-input {
  width: 100%;
  min-height: 48px;
  padding: 0 14px;
  border: 1px solid rgba(212, 175, 55, 0.22);
  border-radius: 14px;
  background: rgba(255, 248, 214, 0.05);
  outline: none;
  color: #fff4d1;
  font-size: 14px;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.form-input::placeholder {
  color: rgba(255, 229, 186, 0.42);
}

.form-input:focus {
  border-color: rgba(255, 223, 135, 0.72);
  box-shadow: 0 0 0 4px rgba(212, 175, 55, 0.14);
  background: rgba(255, 248, 214, 0.08);
}

.phone-input-wrap {
  display: flex;
  align-items: center;
  min-height: 48px;
  border: 1px solid rgba(212, 175, 55, 0.22);
  border-radius: 14px;
  background: rgba(255, 248, 214, 0.05);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.phone-input-wrap:focus-within {
  border-color: rgba(255, 223, 135, 0.72);
  box-shadow: 0 0 0 4px rgba(212, 175, 55, 0.14);
  background: rgba(255, 248, 214, 0.08);
}

.phone-dial-code {
  flex: 0 0 auto;
  min-width: 58px;
  padding: 0 10px 0 14px;
  color: #ffd77a;
  font-size: 14px;
  font-weight: 700;
  line-height: 1;
  border-right: 1px solid rgba(212, 175, 55, 0.24);
}

.phone-input {
  min-width: 0;
  border: 0;
  background: transparent;
}

.phone-input:focus {
  box-shadow: none;
  background: transparent;
}

.country-trigger {
  width: 100%;
  min-height: 48px;
  padding: 0 14px;
  border: 1px solid rgba(212, 175, 55, 0.22);
  border-radius: 14px;
  background: rgba(255, 248, 214, 0.05);
  color: #fff4d1;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.country-trigger:active {
  transform: translateY(1px);
}

.country-trigger-arrow {
  color: rgba(255, 233, 188, 0.72);
  font-size: 14px;
}

:deep(.register-btn.van-button) {
  margin-top: 14px;
  height: 54px;
  border: 1px solid rgba(255, 248, 214, 0.34);
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 16px;
  font-weight: 800;
  box-shadow:
    0 14px 26px rgba(0, 0, 0, 0.18),
    0 8px 18px rgba(90, 27, 0, 0.24);
}

:deep(.register-btn.van-button--disabled) {
  opacity: 0.72;
}

.login-text {
  margin: 16px 0 0;
  font-size: 13px;
  color: rgba(255, 229, 186, 0.74);
  text-align: center;
  line-height: 1.6;
}

.login-link {
  margin-left: 8px;
  padding: 6px 14px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 13px;
  font-weight: 800;
  cursor: pointer;
  box-shadow: 0 8px 16px rgba(75, 25, 0, 0.24);
}

.login-link:active {
  transform: translateY(1px);
}

.feature-grid {
  margin-top: 14px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.feature-card {
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  padding: 16px 10px 14px;
}

.feature-icon {
  width: 48px;
  height: 48px;
  margin: 0 auto 12px;
  border-radius: 50%;
  color: #fff7df;
  font-size: 15px;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(212, 175, 55, 0.45);
  box-shadow:
    inset 0 1px 0 rgba(255, 248, 214, 0.18),
    0 8px 18px rgba(0, 0, 0, 0.22);
}

.feature-icon-game {
  background: linear-gradient(145deg, #9a1212 0%, #6a0000 100%);
}

.feature-icon-coin {
  background: linear-gradient(145deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
}

.feature-icon-invite {
  background: linear-gradient(145deg, #c21a1a 0%, #8a0505 100%);
}

.feature-icon-img {
  width: 24px;
  height: 24px;
  object-fit: contain;
  filter: brightness(0) saturate(100%) invert(97%) sepia(44%) saturate(534%) hue-rotate(320deg) brightness(104%)
    contrast(96%);
}

.feature-title {
  text-align: center;
  font-size: 13px;
  font-weight: 700;
  color: #fff0c9;
  line-height: 1.35;
}

.feature-desc {
  margin-top: 6px;
  text-align: center;
  font-size: 11px;
  color: rgba(255, 229, 186, 0.66);
  line-height: 1.45;
}

:deep(.language-popup.van-popup) {
  min-height: 430px;
  padding: 10px 0 28px;
  background:
    radial-gradient(circle at 12% 10%, rgba(212, 175, 55, 0.18), transparent 22%),
    linear-gradient(180deg, #540000 0%, #280000 100%);
  border-radius: 24px 24px 0 0;
  border: 1px solid rgba(212, 175, 55, 0.34);
  box-shadow: 0 -12px 32px rgba(0, 0, 0, 0.48);
}

.language-popup-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  border-bottom: 1px solid rgba(212, 175, 55, 0.15);
}

.language-popup-title {
  font-size: 20px;
  font-weight: 800;
  color: #fff0c9;
  letter-spacing: 0.04em;
}

.language-popup-close {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  border: none;
  background: transparent;
  font-size: 18px;
  color: #ffd98b;
  line-height: 1;
  cursor: pointer;
}

.language-list {
  padding: 16px 14px 0;
}

.language-item {
  width: 100%;
  margin-bottom: 12px;
  padding: 14px 16px;
  border: 1px solid rgba(212, 175, 55, 0.14);
  border-radius: 16px;
  background: rgba(255, 248, 214, 0.05);
  display: grid;
  grid-template-columns: 34px 1fr 24px;
  align-items: center;
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.language-item.active {
  border-color: rgba(212, 175, 55, 0.52);
  background: rgba(212, 175, 55, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 248, 214, 0.08);
}

.language-code {
  font-size: 14px;
  color: #fff0c9;
  font-weight: 700;
}

.language-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.language-text .native {
  font-size: 14px;
  font-weight: 700;
  color: #fff0c9;
}

.language-text .english {
  font-size: 11px;
  color: rgba(255, 229, 186, 0.66);
}

.language-check {
  font-size: 20px;
  color: #ffd98b;
  text-align: right;
}

.language-tip {
  margin: 10px 14px 0;
  text-align: center;
  color: rgba(255, 229, 186, 0.6);
  font-size: 12px;
}

@media (max-width: 390px) {
  .hero-content {
    padding: 20px 16px 18px;
  }

  .hero-title {
    font-size: 24px;
  }

  .auth-card {
    padding: 12px;
  }

  .form-card {
    padding: 8px 12px;
  }

  .feature-grid {
    gap: 8px;
  }

  .feature-icon {
    width: 40px;
    height: 40px;
    margin-bottom: 8px;
  }

  .feature-icon-img {
    width: 20px;
    height: 20px;
  }

  .feature-desc {
    margin-top: 4px;
    font-size: 10px;
  }
}
</style>

<route lang="json5">
{
  name: 'Register'
}
</route>
