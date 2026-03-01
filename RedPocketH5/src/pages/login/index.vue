<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { RouteMap } from 'vue-router'
import { useUserStore } from '@/stores'
import { showToast } from 'vant'
import { locale } from '@/utils/i18n'
import AppPageHeader from '@/components/AppPageHeader.vue'
import emailIcon from '@/assets/svg/email.svg'
import lockIcon from '@/assets/svg/lock.svg'
import inviteIcon from '@/assets/svg/invite.svg'
import languageIcon from '@/assets/svg/language.svg'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const activeTab = ref<'telegram' | 'email'>('telegram')
const tgBotId = Number(import.meta.env.VITE_TG_BOT_ID || 0)
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
    router.push({
      name: (redirect as keyof RouteMap) || 'Home',
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

async function handleTelegramAuth() {
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
          await userStore.loginByTelegram(data)
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
  router.push({ name: 'ForgotPassword' as keyof RouteMap })
}

function goRegister() {
  router.push({ name: 'Register' as keyof RouteMap })
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

      <div class="banner">
        <img
          class="banner-image"
          src="https://game.luckypacket.me/images/register-header.jpg"
          alt="register banner"
        >
      </div>

      <div class="tabs">
        <button
          class="tab"
          :class="{ active: activeTab === 'telegram' }"
          @click="activeTab = 'telegram'"
        >
          Telegram
        </button>
        <button
          class="tab"
          :class="{ active: activeTab === 'email' }"
          @click="activeTab = 'email'"
        >
          {{ t('login.emailTab') }}
        </button>
      </div>

      <section v-if="activeTab === 'telegram'" class="telegram-panel">
        <div class="tg-logo">
          <img
            class="tg-logo-image"
            src="https://game.luckypacket.me/static/image/telegram.png"
            alt="telegram"
          >
        </div>
        <p class="tg-title">
          {{ t('login.telegramAuth') }}
        </p>
        <van-button class="tg-login-btn" :loading="loading" round block @click="handleTelegramAuth">
          {{ t('login.loginWithTelegram') }}
        </van-button>
      </section>

      <section v-else class="email-panel">
        <div class="email-form-card">
          <div class="email-form-row">
            <label for="login-email" class="email-form-label">
              <img :src="emailIcon" alt="email" class="email-form-icon">
              <span>{{ t('login.email') }}</span>
            </label>
            <input
              id="login-email"
              v-model="postData.email"
              type="text"
              class="email-form-input"
              :placeholder="t('login.emailOrUsername')"
            >
          </div>

          <div class="email-form-row">
            <label for="login-password" class="email-form-label">
              <img :src="lockIcon" alt="password" class="email-form-icon">
              <span>{{ t('login.password') }}</span>
            </label>
            <input
              id="login-password"
              v-model="postData.password"
              type="password"
              class="email-form-input"
              :placeholder="t('login.password')"
            >
          </div>
        </div>

        <div class="email-forgot-wrap">
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

      <section class="feature-card">
        <div class="feature-item">
          <div class="feature-icon feature-icon-game" aria-hidden="true">
            G
          </div>
          <div class="feature-title">
            {{ t('login.feature.game.title') }}
          </div>
          <div class="feature-desc">
            {{ t('login.feature.game.desc') }}
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon feature-icon-coin" aria-hidden="true">
            C
          </div>
          <div class="feature-title">
            {{ t('login.feature.coin.title') }}
          </div>
          <div class="feature-desc">
            {{ t('login.feature.coin.desc') }}
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon feature-icon-invite" aria-hidden="true">
            <img :src="inviteIcon" alt="" class="feature-icon-img">
          </div>
          <div class="feature-title">
            {{ t('login.feature.invite.title') }}
          </div>
          <div class="feature-desc">
            {{ t('login.feature.invite.desc') }}
          </div>
        </div>
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
  background:
    radial-gradient(1200px 500px at 50% -240px, rgba(84, 185, 105, 0.12), transparent 65%), var(--color-bg-page);
  padding: 0 var(--page-padding-x) 34px;
}

.login-shell {
  width: 100%;
  max-width: 640px;
  margin: 0 auto;
}

.login-header {
  margin: 0 calc(-1 * var(--page-padding-x));
  padding: 0 14px;
  background: var(--color-bg-card);
  border-bottom: 1px solid rgba(16, 24, 40, 0.06);
}

.lang-icon {
  width: 24px;
  height: 24px;
}

.banner {
  margin-top: 16px;
  border-radius: var(--radius-3xl);
  overflow: hidden;
  box-shadow: 0 10px 26px rgba(15, 23, 42, 0.08);
}

.banner-image {
  display: block;
  width: 100%;
  height: auto;
}

.tabs {
  margin-top: 18px;
  padding: 6px;
  background: rgba(255, 255, 255, 0.78);
  border-radius: 16px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  box-shadow: 0 4px 18px rgba(15, 23, 42, 0.05);
}

.tab {
  border: 0;
  background: transparent;
  border-radius: 12px;
  height: 52px;
  font-size: var(--font-md);
  font-weight: 500;
  color: var(--color-text-tab);
  position: relative;
  transition:
    color 0.2s ease,
    background-color 0.2s ease;
  cursor: pointer;
}

.tab.active {
  color: var(--color-primary);
  background: var(--color-bg-card);
}

.tab.active::after {
  content: '';
  position: absolute;
  left: 26%;
  right: 26%;
  bottom: 5px;
  height: 3px;
  border-radius: 3px;
  background: var(--color-primary);
}

.telegram-panel,
.email-panel {
  margin-top: 18px;
  text-align: center;
}

.tg-logo {
  width: 108px;
  border-radius: 0;
  margin: 0 auto;
}

.tg-logo-image {
  display: block;
  width: 100%;
  height: auto;
}

.tg-title {
  margin: 16px 0 20px;
  font-size: var(--font-lg);
  font-weight: 500;
  color: var(--color-text-muted);
}

.tg-login-btn {
  width: 100%;
  margin: 0 auto;
  background: var(--color-telegram);
  border: none;
  color: var(--color-bg-card);
  font-size: var(--font-base);
  font-weight: 500;
  height: 50px;
  box-shadow: 0 8px 20px rgba(50, 164, 235, 0.2);
}

.email-form-card {
  background: var(--color-bg-card);
  border-radius: var(--radius-2xl);
  border: 1px solid rgba(16, 24, 40, 0.08);
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.05);
  padding: 8px 16px;
}

.email-form-row {
  min-height: 86px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: stretch;
  gap: 10px;
  padding: 8px 0;
}

.email-form-row + .email-form-row {
  border-top: 1px solid var(--color-border);
}

.email-form-label {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  font-size: var(--font-md);
  font-weight: 500;
  color: var(--color-text-form);
  cursor: text;
}

.email-form-icon {
  width: 22px;
  height: 22px;
}

.email-form-input {
  width: 100%;
  border: 1px solid rgba(16, 24, 40, 0.14);
  background: rgba(255, 255, 255, 0.96);
  outline: none;
  border-radius: 10px;
  min-height: 46px;
  padding: 0 14px;
  font-size: var(--font-base);
  color: var(--color-text-input);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.email-form-input::placeholder {
  color: var(--color-text-muted);
}

.email-form-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 4px rgba(101, 177, 104, 0.16);
}

.email-forgot-wrap {
  margin-top: 14px;
  display: flex;
  justify-content: flex-end;
}

.email-forgot-btn {
  border: none;
  background: transparent;
  color: var(--color-primary-link);
  font-size: var(--font-md);
  padding: 6px 0;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.email-forgot-btn:active {
  opacity: 0.75;
}

.email-login-btn {
  margin-top: 10px;
  background: var(--color-primary-btn);
  border: none;
  height: 54px;
  font-size: var(--font-xl);
  font-weight: 500;
  letter-spacing: 0.2px;
  box-shadow: 0 10px 24px rgba(101, 177, 104, 0.34);
}

.email-signup-text {
  margin: 24px 0 0;
  font-size: var(--font-md);
  color: var(--color-text-body);
  line-height: 1.6;
}

.email-signup-link {
  border: none;
  background: transparent;
  color: var(--color-primary-link);
  font-size: var(--font-md);
  font-weight: 600;
  margin-left: 6px;
  cursor: pointer;
}

.feature-card {
  margin-top: 30px;
  background: linear-gradient(180deg, rgba(245, 250, 246, 0.95) 0%, rgba(240, 247, 241, 0.95) 100%);
  border: 1px solid rgba(91, 172, 106, 0.12);
  border-radius: 20px;
  padding: 18px 12px;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.feature-item {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.feature-icon {
  width: 56px;
  height: 56px;
  margin: 0 auto 12px;
  border-radius: 50%;
  background: linear-gradient(145deg, #61be73 0%, #4ca95f 100%);
  color: var(--color-bg-card);
  font-size: 20px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 18px rgba(76, 169, 95, 0.26);
}

.feature-icon-game {
  background: linear-gradient(145deg, #3c3f91 0%, #2a2e7e 100%);
}

.feature-icon-coin {
  background: linear-gradient(145deg, #f6be3f 0%, #e39b1f 100%);
}

.feature-icon-invite {
  background: linear-gradient(145deg, #61be73 0%, #4ca95f 100%);
}

.feature-icon-img {
  width: 28px;
  height: 28px;
  object-fit: contain;
  filter: brightness(0) invert(1);
}

.feature-title {
  font-size: var(--font-lg);
  font-weight: 600;
  color: var(--color-text-feature);
  line-height: 1.35;
}

.feature-desc {
  margin-top: 6px;
  font-size: var(--font-sm);
  color: var(--color-text-muted);
  line-height: 1.4;
  max-width: 10em;
}

.language-popup {
  min-height: 430px;
  padding: 10px 0 28px;
}

.language-popup-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  border-bottom: 1px solid var(--color-border);
}

.language-popup-title {
  font-size: var(--font-2xl);
  font-weight: 600;
  color: var(--color-text-primary);
}

.language-popup-close {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  border: none;
  background: transparent;
  font-size: 36px;
  color: var(--color-text-light);
  line-height: 1;
  cursor: pointer;
}

.language-list {
  padding: var(--page-padding-x);
}

.language-item {
  width: 100%;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  padding: 14px var(--page-padding-x);
  margin-bottom: 12px;
  background: var(--color-bg-card);
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
  border-color: var(--color-border-active);
  background: var(--color-primary-active);
}

.language-code {
  font-size: var(--font-md);
  color: var(--color-text-primary);
}

.language-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.language-text .native {
  font-size: var(--font-md);
  font-weight: 600;
  color: var(--color-text-primary);
}

.language-text .english {
  font-size: var(--font-xs);
  color: var(--color-text-en);
}

.language-check {
  font-size: var(--font-2xl);
  color: var(--color-primary-link);
  text-align: right;
}

.language-tip {
  margin: 10px var(--page-padding-x) 0;
  text-align: center;
  color: var(--color-text-muted);
  font-size: var(--font-sm);
}

@media (max-width: 390px) {
  .email-form-card {
    padding: 6px 12px;
  }

  .feature-card {
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
    padding: 14px 8px;
  }

  .feature-icon {
    width: 40px;
    height: 40px;
    margin-bottom: 8px;
    font-size: var(--font-sm);
  }

  .feature-icon-img {
    width: 20px;
    height: 20px;
  }

  .feature-title {
    font-size: var(--font-base);
    line-height: 1.2;
  }

  .feature-desc {
    margin-top: 4px;
    font-size: var(--font-xs);
    line-height: 1.25;
    max-width: 7em;
  }
}
</style>

<route lang="json5">
{
  name: 'Login'
}
</route>
