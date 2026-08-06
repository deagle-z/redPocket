<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores'
import { trackAttributionEvent } from '@/utils/attribution'
import { getAuthCountry, setAuthCountry } from '@/utils/auth'
import { showToast } from 'vant'
import { languageOptions, locale } from '@/utils/i18n'
import { safeBack } from '@/utils/navigation'
import { buildPhoneWithDialCode, normalizeNationalPhone, onlyPhoneDigits } from '@/utils/phone'
import AppPageHeader from '@/components/AppPageHeader.vue'
import emailIcon from '@/assets/svg/email.svg'
import lockIcon from '@/assets/svg/lock.svg'
import inviteIcon from '@/assets/svg/invite.svg'
import languageIcon from '@/assets/svg/language.svg'
import imgRegisterHeader from '@/assets/images/register-header.jpg'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const showLangPopup = ref(false)
const showCountryPopup = ref(false)
const loginCountries = [
  { code: 'PE', nameKey: 'login.countryPeru', dialCode: '+51' },
  { code: 'ID', nameKey: 'login.countryIndonesia', dialCode: '+62' },
  { code: 'BR', nameKey: 'login.countryBrazil', dialCode: '+55' },
] as const
type LoginCountryCode = typeof loginCountries[number]['code']
const loginCountryMap = Object.fromEntries(loginCountries.map(item => [item.code, item.nameKey])) as Record<LoginCountryCode, string>
const loginCountryDialCodeMap = Object.fromEntries(loginCountries.map(item => [item.code, item.dialCode])) as Record<
  LoginCountryCode,
  string
>
const postData = reactive<{
  country: LoginCountryCode
  phone: string
  password: string
}>({
  country: 'BR',
  phone: '',
  password: '',
})

const currentCountryDialCode = computed(() => loginCountryDialCodeMap[postData.country] || '+55')

function detectLoginCountry(): LoginCountryCode {
  const browserLanguages = [
    ...(navigator.languages || []),
    navigator.language,
  ]
    .filter(Boolean)
    .map(lang => lang.toUpperCase())

  for (const lang of browserLanguages) {
    if (lang.includes('-BR') || lang.startsWith('PT'))
      return 'BR'
    if (lang.includes('-PE') || lang.startsWith('ES'))
      return 'PE'
    if (lang.includes('-ID') || lang.startsWith('ID'))
      return 'ID'
  }

  return 'BR'
}

function normalizePhoneInput(event: Event) {
  const input = event.target as HTMLInputElement
  postData.phone = onlyPhoneDigits(input.value)
}

function buildLoginPhone(country: LoginCountryCode, rawPhone: string) {
  const nationalPhone = normalizeNationalPhone(country, rawPhone)
  return {
    nationalPhone,
    phoneWithDialCode: buildPhoneWithDialCode(country, rawPhone),
  }
}

onMounted(() => {
  const storedCountry = getAuthCountry() as LoginCountryCode
  postData.country = loginCountryMap[storedCountry] ? storedCountry : detectLoginCountry()
})

async function login() {
  const { nationalPhone, phoneWithDialCode } = buildLoginPhone(postData.country, postData.phone)
  postData.phone = nationalPhone

  if (!nationalPhone) {
    showToast(t('login.pleaseEnterPhone'))
    return
  }
  if (!postData.password) {
    showToast(t('login.pleaseEnterPassword'))
    return
  }
  try {
    loading.value = true
    trackAttributionEvent({
      eventName: 'login_submit',
      metadata: {
        method: 'phone',
      },
    })
    await userStore.login({ ...postData, phone: phoneWithDialCode })
    trackAttributionEvent({
      eventName: 'login_success',
      metadata: {
        method: 'phone',
      },
    })
    setAuthCountry(postData.country)
    const { redirect, ...othersQuery } = router.currentRoute.value.query
    const redirectPath = typeof redirect === 'string' && redirect ? redirect : '/'
    router.replace({
      path: redirectPath,
      query: {
        ...othersQuery,
      },
    })
  }
  finally {
    loading.value = false
  }
}

function goBack() {
  safeBack(router)
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

function selectCountry(country: LoginCountryCode) {
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

function goForgotPassword() {
  router.push('/resetpwd')
}

function goCustomerService() {
  router.push('/cs')
}
</script>

<template>
  <div class="login-page">
    <div class="login-shell">
      <AppPageHeader
        class="login-header"
        :title="t('login.pageTitle')"
        @back="goBack"
        @right-click="openLanguagePopup"
      >
        <template #right>
          <img :src="languageIcon" class="lang-icon" alt="language icon">
        </template>
      </AppPageHeader>

      <section class="brand-hero">
        <img
          class="hero-image"
          :src="imgRegisterHeader"
          alt="register banner"
        >
        <div class="hero-content">
          <p class="hero-eyebrow">
            {{ t('appTopHeader.brandSubtitle') }}
          </p>
          <h2 class="hero-title">
            {{ t('appTopHeader.brandTitle') }}
          </h2>
          <p class="hero-desc">
            {{ t('login.phoneTab') }}
          </p>
        </div>
      </section>

      <section class="auth-card">
        <section class="email-panel">
          <div class="form-heading">
            <p class="form-kicker">
              {{ t('login.phoneTab') }}
            </p>
            <h3 class="form-title">
              {{ t('login.login') }}
            </h3>
          </div>

          <div class="email-form-card">
            <div class="email-form-row">
              <label for="login-phone" class="email-form-label">
                <span class="icon-wrap">
                  <img :src="emailIcon" alt="phone" class="email-form-icon">
                </span>
                <span>{{ t('login.phone') }}</span>
              </label>
              <div class="phone-input-wrap">
                <button type="button" class="phone-country-trigger" @click="openCountryPopup">
                  <span>{{ currentCountryDialCode }}</span>
                  <span class="phone-country-arrow" aria-hidden="true">▾</span>
                </button>
                <input
                  id="login-phone"
                  v-model="postData.phone"
                  type="tel"
                  inputmode="tel"
                  pattern="[0-9]*"
                  autocomplete="tel-national"
                  class="email-form-input phone-input"
                  :placeholder="t('login.pleaseEnterPhone')"
                  @input="normalizePhoneInput"
                >
              </div>
            </div>

            <div class="email-form-row">
              <label for="login-password" class="email-form-label">
                <span class="icon-wrap">
                  <img :src="lockIcon" alt="password" class="email-form-icon">
                </span>
                <span>{{ t('login.password') }}</span>
              </label>
              <input
                id="login-password"
                v-model="postData.password"
                type="password"
                autocomplete="current-password"
                class="email-form-input"
                :placeholder="t('login.password')"
              >
            </div>
          </div>

          <div class="email-actions">
            <button type="button" class="email-forgot-btn" @click="goForgotPassword">
              {{ t('login.forgotPassword') }}
            </button>
          </div>

          <van-button
            :loading="loading"
            type="primary"
            round block
            class="email-login-btn"
            @click="login"
          >
            {{ t('login.login') }}
          </van-button>

        </section>
      </section>

      <section class="feature-grid">
        <button type="button" class="feature-card" @click="goCustomerService">
          <div class="feature-icon feature-icon-service" aria-hidden="true">
            CS
          </div>
          <div class="feature-title">
            {{ t('login.feature.game.title') }}
          </div>
          <div class="feature-desc">
            {{ t('login.feature.game.desc') }}
          </div>
        </button>
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
        <span class="language-popup-title">{{ t('login.country') }}</span>
        <button class="language-popup-close" @click="closeCountryPopup">
          ×
        </button>
      </div>

      <div class="language-list">
        <button
          v-for="item in loginCountries"
          :key="item.code"
          class="language-item"
          :class="{ active: postData.country === item.code }"
          @click="selectCountry(item.code)"
        >
          <span class="language-code">{{ item.dialCode }}</span>
          <span class="language-text">
            <span class="native">{{ t(item.nameKey) }}</span>
            <span class="english">{{ item.code }}</span>
          </span>
          <span v-if="postData.country === item.code" class="language-check">✓</span>
        </button>
      </div>
    </van-popup>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  --login-gold-line: linear-gradient(
    90deg,
    transparent 0%,
    var(--color-gold-dark) 22%,
    var(--color-gold-light) 50%,
    var(--color-gold-dark) 78%,
    transparent 100%
  );
  --login-panel-bg: linear-gradient(
    160deg,
    color-mix(in srgb, var(--color-primary-dark) 72%, var(--color-game-bg-start)),
    var(--color-game-bg-mid) 78%
  );
  --login-subtle-border: color-mix(in srgb, var(--color-gold) 28%, transparent);
  --login-strong-border: color-mix(in srgb, var(--color-gold) 48%, transparent);
  --login-muted-text: var(--color-game-text-soft);
  background-image:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-game-bg-start) 84%, var(--color-primary-dark)) 0%,
      var(--color-game-bg-mid) 56%,
      var(--color-game-bg-end) 100%
    ),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 22px,
      color-mix(in srgb, var(--color-gold) 5%, transparent) 22px,
      color-mix(in srgb, var(--color-gold) 5%, transparent) 23px
    );
  color: var(--color-game-text);
  width: 100%;
  overflow-x: hidden;
  padding: 0 0 calc(24px + env(safe-area-inset-bottom));
}

.login-shell {
  width: calc(100% - 24px);
  max-width: 640px;
  margin: 0 auto;
  position: relative;
  overflow: hidden;
}

.login-header {
  margin-bottom: 12px;
}

.lang-icon {
  width: 24px;
  height: 24px;
  filter: brightness(0) saturate(100%) invert(84%) sepia(39%) saturate(612%) hue-rotate(338deg) brightness(105%)
    contrast(96%);
}

.auth-card,
.feature-card {
  position: relative;
  overflow: hidden;
  border-radius: var(--radius-3xl);
  border: 1px solid var(--login-subtle-border);
  box-shadow:
    var(--shadow-game-float),
    inset 0 1px 0 color-mix(in srgb, var(--color-gold-soft) 8%, transparent);
}

.auth-card::after,
.feature-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--login-gold-line);
}

.brand-hero {
  position: relative;
  min-height: 132px;
  margin-top: 6px;
  overflow: hidden;
  border-radius: var(--radius-3xl);
  border: 1px solid var(--login-subtle-border);
  background: var(--login-panel-bg);
  box-shadow: 0 12px 24px color-mix(in srgb, var(--color-game-bg-end) 72%, transparent);
}

.hero-image {
  position: absolute;
  inset: 0;
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  opacity: 0.18;
}

.brand-hero::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(
      100deg,
      color-mix(in srgb, var(--color-game-bg-start) 90%, transparent) 0%,
      color-mix(in srgb, var(--color-game-bg-start) 62%, transparent) 54%,
      color-mix(in srgb, var(--color-game-bg-end) 86%, transparent) 100%
    ),
    linear-gradient(180deg, color-mix(in srgb, var(--color-gold) 12%, transparent), transparent 56%);
}

.brand-hero::after {
  content: '';
  position: absolute;
  inset: auto 18px 0;
  height: 1px;
  background: var(--login-gold-line);
}

.hero-content {
  position: relative;
  z-index: 1;
  padding: 18px 18px 16px;
}

.hero-eyebrow {
  margin: 0 0 6px;
  color: var(--color-gold-light);
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.hero-title {
  margin: 0;
  color: var(--color-game-text);
  font-size: 26px;
  line-height: 1.12;
  font-weight: 800;
}

.hero-desc {
  margin: 8px 0 0;
  max-width: 220px;
  color: var(--login-muted-text);
  font-size: 13px;
  line-height: 1.45;
}

.auth-card {
  margin-top: 12px;
  background:
    radial-gradient(color-mix(in srgb, var(--color-gold) 22%, transparent) 1px, transparent 1px), var(--login-panel-bg);
  background-size:
    20px 20px,
    100% 100%;
  padding: 14px;
}

.email-panel {
  margin: 0;
  min-width: 0;
}

.form-heading {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}

.form-kicker {
  margin: 0;
  color: var(--color-gold-light);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.form-title {
  margin: 0;
  color: var(--color-game-text);
  font-size: var(--font-xl);
  font-weight: 800;
  line-height: 1.2;
  text-wrap: pretty;
}

.email-form-card {
  position: relative;
  overflow: hidden;
  background: color-mix(in srgb, var(--color-game-input) 86%, transparent);
  border-radius: var(--radius-2xl);
  border: 1px solid var(--login-subtle-border);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--color-gold-soft) 8%, transparent);
  padding: 4px 14px;
}

.email-form-row {
  min-height: 78px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 9px;
  padding: 10px 0;
}

.email-form-row + .email-form-row {
  border-top: 1px solid color-mix(in srgb, var(--color-gold) 14%, transparent);
}

.email-form-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  font-weight: 700;
  color: color-mix(in srgb, var(--color-gold-light) 74%, var(--color-game-text));
  cursor: text;
}

.icon-wrap {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-gold) 13%, transparent);
  border: 1px solid var(--login-subtle-border);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--color-gold-soft) 8%, transparent);
}

.email-form-icon {
  width: 16px;
  height: 16px;
  filter: brightness(0) saturate(100%) invert(85%) sepia(39%) saturate(649%) hue-rotate(335deg) brightness(105%)
    contrast(97%);
}

.email-form-input {
  width: 100%;
  min-width: 0;
  min-height: 46px;
  padding: 0 14px;
  border: 1px solid var(--login-subtle-border);
  border-radius: var(--radius-xl);
  background: color-mix(in srgb, var(--color-gold-soft) 5%, transparent);
  outline: none;
  color: var(--color-game-text);
  font-size: 14px;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.email-form-input::placeholder {
  color: color-mix(in srgb, var(--color-game-text) 44%, transparent);
}

.email-form-input:focus {
  border-color: color-mix(in srgb, var(--color-gold-light) 72%, transparent);
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--color-gold) 14%, transparent);
  background: color-mix(in srgb, var(--color-gold-soft) 8%, transparent);
}

.phone-input-wrap {
  display: flex;
  align-items: center;
  min-height: 46px;
  border: 1px solid var(--login-subtle-border);
  border-radius: var(--radius-xl);
  background: color-mix(in srgb, var(--color-gold-soft) 5%, transparent);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.phone-input-wrap:focus-within {
  border-color: color-mix(in srgb, var(--color-gold-light) 72%, transparent);
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--color-gold) 14%, transparent);
  background: color-mix(in srgb, var(--color-gold-soft) 8%, transparent);
}

.phone-country-trigger {
  flex: 0 0 auto;
  min-width: 72px;
  min-height: 46px;
  padding: 0 10px 0 14px;
  border: 0;
  border-right: 1px solid var(--login-subtle-border);
  background: transparent;
  color: var(--color-gold-light);
  font-size: 14px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  cursor: pointer;
}

.phone-country-arrow {
  color: var(--login-muted-text);
  font-size: 11px;
  line-height: 1;
}

.phone-input {
  flex: 1 1 0;
  width: 0;
  min-width: 0;
  border: 0;
  background: transparent;
}

.phone-input:focus {
  box-shadow: none;
  background: transparent;
}

.email-actions {
  margin-top: 10px;
  display: flex;
  justify-content: flex-end;
}

.email-forgot-btn {
  border: none;
  background: transparent;
  color: var(--color-gold-light);
  font-size: 13px;
  font-weight: 700;
  min-height: 44px;
  padding: 0;
  cursor: pointer;
}

.email-forgot-btn:active {
  opacity: 0.75;
}

:deep(.email-login-btn.van-button) {
  margin-top: 8px;
  width: 100%;
  height: 54px;
  border: 1px solid color-mix(in srgb, var(--color-gold-soft) 34%, transparent);
  background: var(--color-game-gold-gradient);
  color: var(--color-text-title);
  font-size: 16px;
  font-weight: 800;
  box-shadow:
    0 14px 24px color-mix(in srgb, var(--color-game-bg-end) 42%, transparent),
    0 8px 16px color-mix(in srgb, var(--color-gold-dark) 26%, transparent);
  transition:
    transform 0.18s ease,
    filter 0.18s ease,
    opacity 0.18s ease;
}

:deep(.email-login-btn.van-button:active) {
  transform: translateY(1px);
  filter: brightness(0.96);
}

:deep(.email-login-btn.van-button--disabled) {
  opacity: 0.7;
  filter: saturate(0.82);
}

.email-signup-text {
  margin: 14px 0 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 8px;
  text-align: center;
  font-size: 13px;
  color: var(--login-muted-text);
  line-height: 1.6;
}

.email-signup-link {
  min-height: 44px;
  padding: 0 14px;
  border: none;
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-gold) 16%, transparent);
  color: var(--color-gold-light);
  font-size: 13px;
  font-weight: 800;
  cursor: pointer;
  box-shadow: inset 0 0 0 1px var(--login-subtle-border);
}

.feature-grid {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.feature-card {
  background: color-mix(in srgb, var(--color-game-card) 72%, transparent);
  color: inherit;
  border: 1px solid color-mix(in srgb, var(--color-gold) 22%, transparent);
  padding: 12px 8px 11px;
  cursor: pointer;
  font: inherit;
  box-shadow:
    0 10px 18px color-mix(in srgb, var(--color-game-bg-end) 38%, transparent),
    inset 0 1px 0 color-mix(in srgb, var(--color-gold-soft) 5%, transparent);
}

.feature-icon {
  width: 38px;
  height: 38px;
  margin: 0 auto 8px;
  border-radius: 50%;
  color: var(--color-gold-soft);
  font-size: 15px;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--login-strong-border);
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-gold-soft) 18%, transparent),
    0 8px 14px color-mix(in srgb, var(--color-game-bg-end) 42%, transparent);
}

.feature-icon-service {
  background: linear-gradient(145deg, var(--color-primary) 0%, var(--color-primary-link) 100%);
}

.feature-icon-coin {
  background: var(--color-game-gold-gradient);
  color: var(--color-text-title);
}

.feature-icon-invite {
  background: linear-gradient(145deg, var(--color-primary-medium) 0%, var(--color-primary-dark) 100%);
}

.feature-icon-img {
  width: 20px;
  height: 20px;
  object-fit: contain;
  filter: brightness(0) saturate(100%) invert(97%) sepia(44%) saturate(534%) hue-rotate(320deg) brightness(104%)
    contrast(96%);
}

.feature-title {
  text-align: center;
  font-size: 12px;
  font-weight: 700;
  color: var(--color-game-text);
  line-height: 1.35;
}

.feature-desc {
  margin-top: 4px;
  text-align: center;
  font-size: 10px;
  color: var(--login-muted-text);
  line-height: 1.45;
}

:deep(.language-popup.van-popup) {
  min-height: 420px;
  padding: 10px 0 28px;
  background:
    radial-gradient(circle at 12% 10%, color-mix(in srgb, var(--color-gold) 18%, transparent), transparent 22%),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-primary-link) 68%, var(--color-game-bg-start)) 0%,
      var(--color-game-bg-mid) 100%
    );
  border-radius: 24px 24px 0 0;
  border: 1px solid var(--login-strong-border);
  box-shadow: 0 -12px 32px color-mix(in srgb, var(--color-game-bg-end) 62%, transparent);
}

.language-popup-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  border-bottom: 1px solid color-mix(in srgb, var(--color-gold) 15%, transparent);
}

.language-popup-title {
  font-size: var(--font-xl);
  font-weight: 800;
  color: var(--color-game-text);
}

.language-popup-close {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  border: none;
  background: transparent;
  font-size: 18px;
  color: var(--color-gold-light);
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
  border: 1px solid color-mix(in srgb, var(--color-gold) 14%, transparent);
  border-radius: var(--radius-2xl);
  background: color-mix(in srgb, var(--color-gold-soft) 5%, transparent);
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
  border-color: var(--login-strong-border);
  background: color-mix(in srgb, var(--color-gold) 12%, transparent);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--color-gold-soft) 8%, transparent);
}

.language-code {
  font-size: 14px;
  color: var(--color-game-text);
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
  color: var(--color-game-text);
}

.language-text .english {
  font-size: 11px;
  color: var(--login-muted-text);
}

.language-check {
  font-size: 20px;
  color: var(--color-gold-light);
  text-align: right;
}

.language-tip {
  margin: 10px 14px 0;
  text-align: center;
  color: var(--login-muted-text);
  font-size: 12px;
}

@media (max-width: 390px) {
  .login-shell {
    width: calc(100% - 20px);
  }

  .hero-content {
    padding: 16px 16px 14px;
  }

  .hero-title {
    font-size: 24px;
  }

  .auth-card {
    padding: 12px;
  }

  .email-form-card {
    padding: 4px 12px;
  }

  .feature-grid {
    gap: 8px;
  }

  .feature-icon {
    width: 34px;
    height: 34px;
  }

  .feature-icon-img {
    width: 18px;
    height: 18px;
  }

  .feature-desc {
    font-size: 10px;
  }
}
</style>

<route lang="json5">
{
  name: 'Login'
}
</route>
