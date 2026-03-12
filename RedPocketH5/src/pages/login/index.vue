<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores'
import { showToast } from 'vant'
import { locale } from '@/utils/i18n'
import AppPageHeader from '@/components/AppPageHeader.vue'
import emailIcon from '@/assets/svg/email.svg'
import lockIcon from '@/assets/svg/lock.svg'
import inviteIcon from '@/assets/svg/invite.svg'
import languageIcon from '@/assets/svg/language.svg'
import imgRegisterHeader from '@/assets/images/register-header.jpg'
import imgTelegram from '@/assets/images/telegram.png'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const activeTab = ref<'telegram' | 'email'>('telegram')
const tgBotUsername = import.meta.env.VITE_TG_BOT_USERNAME || 'luckRedBoomPacket66Bot'
const showLangPopup = ref(false)
const languageOptions = [
  {
    code: 'CN',
    value: 'zh-CN',
    nativeTextKey: 'login.language.zhNative',
    englishTextKey: 'login.language.zhEn',
  },
  {
    code: 'US',
    value: 'en-US',
    nativeTextKey: 'login.language.enNative',
    englishTextKey: 'login.language.enEn',
  },
]

const postData = reactive({
  email: '',
  password: '',
})

async function login() {
  if (!postData.email) {
    showToast(t('login.pleaseEnterEmail'))
    return
  }
  if (!postData.password) {
    showToast(t('login.pleaseEnterPassword'))
    return
  }
  try {
    loading.value = true
    await userStore.login({ ...postData })
    const { redirect, ...othersQuery } = router.currentRoute.value.query
    const redirectPath = typeof redirect === 'string' && redirect ? redirect : '/'
    router.push({
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

interface TgAuthPayload {
  id: number
  first_name?: string
  last_name?: string
  username?: string
  photo_url?: string
  auth_date: number
  hash: string
}

async function handleTelegramAuth(user: TgAuthPayload) {
  try {
    loading.value = true
    await userStore.loginByTelegram(user)
    const { redirect, ...othersQuery } = router.currentRoute.value.query
    const redirectPath = typeof redirect === 'string' && redirect ? redirect : '/'
    router.push({
      path: redirectPath,
      query: {
        ...othersQuery,
      },
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
  router.back()
}

function openLanguagePopup() {
  showLangPopup.value = true
}

function closeLanguagePopup() {
  showLangPopup.value = false
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

function goRegister() {
  router.push('/register')
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

      <section class="hero-card">
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
            {{ t('login.loginWithTelegram') }}
          </p>
        </div>
      </section>

      <section class="auth-card">
        <div class="tabs">
          <button
            class="tab"
            :class="{ active: activeTab === 'telegram' }"
            type="button"
            @click="activeTab = 'telegram'"
          >
            Telegram
          </button>
          <button
            class="tab"
            :class="{ active: activeTab === 'email' }"
            type="button"
            @click="activeTab = 'email'"
          >
            {{ t('login.emailTab') }}
          </button>
        </div>

        <section v-if="activeTab === 'telegram'" class="telegram-panel">
          <div class="telegram-badge">
            <img
              class="tg-logo-image"
              :src="imgTelegram"
              alt="telegram"
            >
          </div>
          <p class="panel-title">
            {{ t('login.telegramAuth') }}
          </p>
          <p class="panel-subtitle">
            @{{ tgBotUsername }}
          </p>
          <div class="telegram-widget-wrap">
            <TelegramLogin
              mode="callback"
              :telegram-login="tgBotUsername"
              size="large"
              radius="18"
              request-access="write"
              @callback="handleTelegramAuth"
            />
          </div>
          <van-loading v-if="loading" class="tg-loading" size="20px" />
        </section>

        <section v-else class="email-panel">
          <div class="email-form-card">
            <div class="email-form-row">
              <label for="login-email" class="email-form-label">
                <span class="icon-wrap">
                  <img :src="emailIcon" alt="email" class="email-form-icon">
                </span>
                <span>{{ t('login.email') }}</span>
              </label>
              <input
                id="login-email"
                v-model="postData.email"
                type="text"
                autocomplete="username"
                class="email-form-input"
                :placeholder="t('login.emailOrUsername')"
              >
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

          <p class="email-signup-text">
            {{ t('login.noAccountYet') }}
            <button type="button" class="email-signup-link" @click="goRegister">
              {{ t('login.signUpNow') }}
            </button>
          </p>
        </section>
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
  </div>
</template>

<style scoped>
.login-page {
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

.login-shell {
  width: 100%;
  max-width: 640px;
  margin: 0 auto;
  position: relative;
}

.login-header {
  margin-bottom: 12px;
}

.lang-icon {
  width: 24px;
  height: 24px;
  filter: brightness(0) saturate(100%) invert(84%) sepia(39%) saturate(612%) hue-rotate(338deg) brightness(105%) contrast(96%);
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
  max-width: 220px;
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

.tabs {
  padding: 4px;
  background: rgba(255, 248, 214, 0.08);
  border: 1px solid rgba(212, 175, 55, 0.2);
  border-radius: 999px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}

.tab {
  border: 0;
  background: transparent;
  border-radius: 999px;
  height: 42px;
  font-size: 14px;
  font-weight: 700;
  color: rgba(255, 229, 186, 0.74);
  transition:
    color 0.2s ease,
    background-color 0.2s ease,
    transform 0.2s ease;
  cursor: pointer;
}

.tab.active {
  color: #5a1b00;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  box-shadow:
    0 10px 18px rgba(75, 25, 0, 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.telegram-panel,
.email-panel {
  margin-top: 16px;
}

.telegram-panel {
  text-align: center;
}

.telegram-badge {
  width: 78px;
  height: 78px;
  margin: 0 auto;
  border-radius: 24px;
  border: 1px solid rgba(212, 175, 55, 0.42);
  background: linear-gradient(180deg, rgba(255, 248, 214, 0.16), rgba(212, 175, 55, 0.08));
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 1px 0 rgba(255, 248, 214, 0.08);
}

.tg-logo-image {
  display: block;
  width: 50px;
  height: 50px;
}

.panel-title {
  margin: 16px 0 6px;
  color: #fff0c9;
  font-size: 20px;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.panel-subtitle {
  margin: 0;
  color: rgba(255, 229, 186, 0.68);
  font-size: 12px;
  line-height: 1.4;
}

.telegram-widget-wrap {
  margin-top: 16px;
  border-radius: 16px;
  padding: 16px 12px;
  background: rgba(255, 248, 214, 0.05);
  border: 1px solid rgba(212, 175, 55, 0.18);
}

.tg-loading {
  margin-top: 14px;
}

.email-form-card {
  position: relative;
  overflow: hidden;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  border-radius: 18px;
  border: 1px solid rgba(212, 175, 55, 0.24);
  box-shadow: inset 0 1px 0 rgba(255, 248, 214, 0.08);
  padding: 8px 14px;
}

.email-form-row {
  min-height: 86px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 12px;
  padding: 12px 0;
}

.email-form-row + .email-form-row {
  border-top: 1px solid rgba(212, 175, 55, 0.14);
}

.email-form-label {
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

.email-form-icon {
  width: 16px;
  height: 16px;
  filter: brightness(0) saturate(100%) invert(85%) sepia(39%) saturate(649%) hue-rotate(335deg) brightness(105%) contrast(97%);
}

.email-form-input {
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

.email-form-input::placeholder {
  color: rgba(255, 229, 186, 0.42);
}

.email-form-input:focus {
  border-color: rgba(255, 223, 135, 0.72);
  box-shadow: 0 0 0 4px rgba(212, 175, 55, 0.14);
  background: rgba(255, 248, 214, 0.08);
}

.email-actions {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.email-forgot-btn {
  border: none;
  background: transparent;
  color: #ffd98b;
  font-size: 13px;
  font-weight: 700;
  padding: 4px 0;
  cursor: pointer;
}

.email-forgot-btn:active {
  opacity: 0.75;
}

:deep(.email-login-btn.van-button) {
  margin-top: 12px;
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

:deep(.email-login-btn.van-button--disabled) {
  opacity: 0.72;
}

.email-signup-text {
  margin: 16px 0 0;
  text-align: center;
  font-size: 13px;
  color: rgba(255, 229, 186, 0.74);
  line-height: 1.6;
}

.email-signup-link {
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
  filter: brightness(0) saturate(100%) invert(97%) sepia(44%) saturate(534%) hue-rotate(320deg) brightness(104%) contrast(96%);
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

  .email-form-card {
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
  name: 'Login'
}
</route>
