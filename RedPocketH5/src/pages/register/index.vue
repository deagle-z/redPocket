<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { RouteMap } from 'vue-router'
import { useUserStore } from '@/stores'
import { showToast } from 'vant'
import { locale } from '@/utils/i18n'
import AppPageHeader from '@/components/AppPageHeader.vue'
import emailIcon from '@/assets/svg/email.svg'
import lockIcon from '@/assets/svg/lock.svg'
import languageIcon from '@/assets/svg/language.svg'
import verifyIcon from '@/assets/svg/verify.svg'
import inviteIcon from '@/assets/svg/invite.svg'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const sendLoading = ref(false)
const countdown = ref(0)
const devCode = ref('')
const showLangPopup = ref(false)
const tgBotId = Number(import.meta.env.VITE_TG_BOT_ID || 0)

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
  code: '',
  password: '',
  confirmPassword: '',
  inviteCode: '',
})

let countdownTimer: ReturnType<typeof setInterval> | null = null

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

onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
  }
})

async function sendCode() {
  if (!postData.email) {
    showToast(t('register.pleaseEnterEmail'))
    return
  }
  try {
    sendLoading.value = true
    const res = await userStore.sendCode(postData.email)
    const code = res?.data?.code
    if (code) {
      devCode.value = String(code)
      postData.code = String(code)
      showToast(`dev code: ${code}`)
    }
    else {
      devCode.value = ''
    }
    startCountdown()
    showToast(t('register.sendCodeSuccess'))
  }
  catch (error: any) {
    showToast(error?.message || t('register.sendCodeSuccess'))
  }
  finally {
    sendLoading.value = false
  }
}

async function register() {
  if (!postData.email) {
    showToast(t('register.pleaseEnterEmail'))
    return
  }
  if (!postData.code) {
    showToast(t('register.pleaseEnterCode'))
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
    await userStore.register({ ...postData })
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
  router.back()
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

      <div class="banner">
        <img
          class="banner-image"
          src="https://game.luckypacket.me/images/register-header.jpg"
          alt="register banner"
        >
      </div>

      <div class="tg-link-wrap">
        <button type="button" class="tg-link" @click="handleTelegramLogin">
          {{ t('register.telegramLogin') }}
        </button>
      </div>

      <div class="form-card">
        <div class="form-row">
          <label for="register-email" class="form-label">
            <img :src="emailIcon" alt="email" class="form-icon">
            <span>{{ t('register.email') }}</span>
          </label>
          <input
            id="register-email"
            v-model="postData.email"
            type="text"
            class="form-input"
            :placeholder="t('register.pleaseEnterEmail')"
          >
        </div>

        <div class="form-row form-row--code">
          <label for="register-code" class="form-label">
            <img :src="verifyIcon" alt="verify code" class="form-icon">
            <span>{{ t('register.emailCode') }}</span>
          </label>
          <div class="code-input-group">
            <input
              id="register-code"
              v-model="postData.code"
              type="text"
              class="form-input"
              :placeholder="t('register.pleaseEnterCode')"
            >
            <button
              type="button"
              class="send-btn"
              :disabled="sendLoading || countdown > 0"
              @click="sendCode"
            >
              <span v-if="countdown > 0">{{ countdown }}s</span>
              <span v-else-if="sendLoading">...</span>
              <span v-else>{{ t('register.send') }}</span>
            </button>
          </div>
          <p v-if="devCode" class="dev-code-tip">
            dev code: {{ devCode }}
          </p>
        </div>

        <div class="form-row">
          <label for="register-password" class="form-label">
            <img :src="lockIcon" alt="password" class="form-icon">
            <span>{{ t('register.password') }}</span>
          </label>
          <input
            id="register-password"
            v-model="postData.password"
            type="password"
            class="form-input"
            :placeholder="t('register.pleaseEnterPassword')"
          >
        </div>

        <div class="form-row">
          <label for="register-confirm-password" class="form-label">
            <img :src="lockIcon" alt="confirm password" class="form-icon">
            <span>{{ t('register.confirmPassword') }}</span>
          </label>
          <input
            id="register-confirm-password"
            v-model="postData.confirmPassword"
            type="password"
            class="form-input"
            :placeholder="t('register.pleaseEnterConfirmPassword')"
          >
        </div>

        <div class="form-row">
          <label for="register-invite-code" class="form-label">
            <img :src="inviteIcon" alt="invite code" class="form-icon">
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
.register-page {
  min-height: 100vh;
  background:
    radial-gradient(1200px 500px at 50% -240px, rgba(84, 185, 105, 0.12), transparent 65%), var(--color-bg-page);
  padding: 0 var(--page-padding-x) 34px;
}

.register-shell {
  width: 100%;
  max-width: 640px;
  margin: 0 auto;
}

.register-header {
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

.tg-link-wrap {
  margin-top: 20px;
  text-align: center;
}

.tg-link {
  border: 1px solid rgba(101, 177, 104, 0.4);
  background: rgba(255, 255, 255, 0.7);
  color: var(--color-primary-link);
  font-size: var(--font-md);
  font-weight: 500;
  border-radius: 999px;
  min-height: 42px;
  padding: 0 18px;
  cursor: pointer;
  transition:
    background-color 0.2s ease,
    border-color 0.2s ease;
}

.tg-link:active {
  background: rgba(101, 177, 104, 0.14);
}

.form-card {
  margin-top: 18px;
  background: var(--color-bg-card);
  border-radius: var(--radius-2xl);
  border: 1px solid rgba(16, 24, 40, 0.08);
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.05);
  padding: 8px 16px;
}

.form-row {
  min-height: 86px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: stretch;
  gap: 10px;
  padding: 8px 0;
}

.form-row + .form-row {
  border-top: 1px solid var(--color-border);
}

.form-label {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  font-size: var(--font-md);
  font-weight: 500;
  color: var(--color-text-form);
  cursor: text;
}

.form-icon {
  width: 22px;
  height: 22px;
}

.form-input {
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

.form-input::placeholder {
  color: var(--color-text-muted);
}

.form-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 4px rgba(101, 177, 104, 0.16);
}

.form-row--code .code-input-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.form-row--code .form-input {
  flex: 1;
  min-width: 0;
}

.dev-code-tip {
  margin: 4px 0 0;
  font-size: var(--font-sm);
  color: var(--color-danger);
}

.send-btn {
  flex-shrink: 0;
  min-width: 94px;
  height: 42px;
  border: none;
  border-radius: 10px;
  background: var(--color-danger);
  color: var(--color-bg-card);
  font-size: var(--font-base);
  font-weight: 500;
  cursor: pointer;
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.send-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.send-btn:active {
  transform: translateY(1px);
}

.register-btn {
  margin-top: 12px;
  background: var(--color-primary-btn);
  border: none;
  height: 54px;
  font-size: var(--font-xl);
  font-weight: 500;
  letter-spacing: 0.2px;
  box-shadow: 0 10px 24px rgba(101, 177, 104, 0.34);
}

.login-text {
  margin: 22px 0 0;
  font-size: var(--font-md);
  color: var(--color-text-body);
  text-align: center;
  line-height: 1.6;
}

.login-link {
  border: 1px solid var(--color-primary-link);
  background: rgba(255, 255, 255, 0.68);
  color: var(--color-primary-link);
  font-size: var(--font-md);
  font-weight: 600;
  margin-left: 6px;
  border-radius: 12px;
  padding: 5px 14px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.login-link:active {
  background: rgba(101, 177, 104, 0.14);
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
  right: var(--page-padding-x);
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
  .form-card {
    padding: 6px 12px;
  }

  .form-row--code .code-input-group {
    gap: 8px;
  }

  .send-btn {
    min-width: 84px;
    height: 40px;
    font-size: var(--font-sm);
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
  name: 'Register'
}
</route>
