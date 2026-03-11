<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { bindCurrentTgEmail, getCurrentTgUserInfo } from '@/api/user'
import { useUserStore } from '@/stores'
import { locale } from '@/utils/i18n'
import AppPageHeader from '@/components/AppPageHeader.vue'
import languageIcon from '@/assets/svg/language.svg'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()

const sending = ref(false)
const submitting = ref(false)
const countdown = ref(0)
const showLangPopup = ref(false)
const boundEmail = ref('')

const formData = reactive({
  email: '',
  code: '',
})

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

let timer: ReturnType<typeof setInterval> | null = null

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
    closeLanguagePopup()
    return
  }
  locale.value = lang
  showToast(t('login.language.changed'))
  closeLanguagePopup()
  setTimeout(() => {
    window.location.reload()
  }, 350)
}

function startCountdown() {
  countdown.value = 60
  timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      countdown.value = 0
      if (timer) {
        clearInterval(timer)
        timer = null
      }
    }
  }, 1000)
}

async function sendCode() {
  if (boundEmail.value) {
    showToast(t('bindEmailPage.boundToast'))
    return
  }
  const email = String(formData.email || '').trim()
  if (!email) {
    showToast(t('bindEmailPage.toastEnterEmail'))
    return
  }
  if (sending.value || countdown.value > 0)
    return
  try {
    sending.value = true
    await userStore.sendCode(email)
    showToast(t('bindEmailPage.toastSendSuccess'))
    startCountdown()
  }
  catch (error: any) {
    showToast(error?.message || t('bindEmailPage.toastSendFailed'))
  }
  finally {
    sending.value = false
  }
}

async function submitBindEmail() {
  if (boundEmail.value) {
    showToast(t('bindEmailPage.boundToast'))
    return
  }
  const email = String(formData.email || '').trim()
  const code = String(formData.code || '').trim()
  if (!email) {
    showToast(t('bindEmailPage.toastEnterEmail'))
    return
  }
  if (!code) {
    showToast(t('bindEmailPage.toastEnterCode'))
    return
  }
  if (submitting.value)
    return
  try {
    submitting.value = true
    await bindCurrentTgEmail({ email, code })
    showToast(t('bindEmailPage.toastBindSuccess'))
    router.back()
  }
  catch (error: any) {
    showToast(error?.message || t('bindEmailPage.toastBindFailed'))
  }
  finally {
    submitting.value = false
  }
}

async function loadCurrentEmail() {
  try {
    const { data } = await getCurrentTgUserInfo()
    const email = String(data?.email || '').trim()
    if (email) {
      boundEmail.value = email
      formData.email = email
    }
  }
  catch {}
}

onUnmounted(() => {
  if (timer)
    clearInterval(timer)
})

onMounted(() => {
  void loadCurrentEmail()
})
</script>

<template>
  <div class="bind-email-page">
    <div class="bind-email-shell">
      <AppPageHeader class="bind-email-header" :title="t('bindEmailPage.title')" @back="goBack" @right-click="openLanguagePopup">
        <template #right>
          <img :src="languageIcon" class="lang-icon" alt="language icon">
        </template>
      </AppPageHeader>

      <section class="hero-card">
        <p class="hero-kicker">
          {{ t('bindEmailPage.title') }}
        </p>
        <p class="hero-value">
          {{ boundEmail || t('bindEmailPage.emailPlaceholder') }}
        </p>
      </section>

      <section class="form-card">
        <div class="field-row">
          <label class="field-label">
            <span class="field-icon-wrap">
              <van-icon name="envelop-o" />
            </span>
            <span>{{ t('bindEmailPage.email') }}</span>
          </label>
          <input
            v-model="formData.email"
            type="text"
            class="field-input"
            :class="{ 'field-input--locked': !!boundEmail }"
            :placeholder="t('bindEmailPage.emailPlaceholder')"
            :readonly="!!boundEmail"
          >
        </div>

        <div class="field-row code-row">
          <label class="field-label">
            <span class="field-icon-wrap">
              <van-icon name="shield-o" />
            </span>
            <span>{{ t('bindEmailPage.code') }}</span>
          </label>
          <div class="code-group">
            <input
              v-model="formData.code"
              type="text"
              class="field-input"
              :class="{ 'field-input--locked': !!boundEmail }"
              :placeholder="t('bindEmailPage.codePlaceholder')"
              :readonly="!!boundEmail"
            >
            <button
              type="button"
              class="send-btn"
              :disabled="!!boundEmail || sending || countdown > 0"
              @click="sendCode"
            >
              <span v-if="countdown > 0">{{ countdown }}s</span>
              <span v-else-if="sending">{{ t('bindEmailPage.sending') }}</span>
              <span v-else>{{ t('bindEmailPage.send') }}</span>
            </button>
          </div>
        </div>
      </section>

      <button type="button" class="submit-btn" :disabled="!!boundEmail || submitting" @click="submitBindEmail">
        {{ boundEmail ? t('bindEmailPage.boundLabel') : (submitting ? t('bindEmailPage.submitting') : t('bindEmailPage.submit')) }}
      </button>
    </div>

    <van-popup v-model:show="showLangPopup" round position="bottom" class="lang-popup">
      <div class="lang-header">
        <span>{{ t('login.language.title') }}</span>
      </div>
      <button
        v-for="item in languageOptions"
        :key="item.code"
        type="button"
        class="lang-row"
        @click="selectLanguage(item.value)"
      >
        <span class="lang-left">
          <em class="lang-tag">{{ item.code }}</em>
          <span>{{ t(item.nativeTextKey) }}</span>
        </span>
        <span class="lang-right">{{ t(item.englishTextKey) }}</span>
      </button>
      <p class="lang-tip">
        {{ t('login.language.autoRefresh') }}
      </p>
    </van-popup>
  </div>
</template>

<style scoped>
.bind-email-page {
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

.bind-email-shell {
  width: 100%;
  max-width: 640px;
  margin: 0 auto;
}

.bind-email-header {
  margin-bottom: 12px;
}

.lang-icon {
  width: 18px;
  height: 18px;
  object-fit: contain;
  filter: brightness(0) saturate(100%) invert(85%) sepia(39%) saturate(649%) hue-rotate(335deg) brightness(105%) contrast(97%);
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
  font-size: 16px;
  font-weight: 700;
  line-height: 1.45;
}

.form-card {
  margin-top: 0;
  padding: 10px 14px;
}

.field-row {
  min-height: 84px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 12px;
  align-items: center;
}

.field-row + .field-row {
  border-top: 1px solid rgba(212, 175, 55, 0.14);
}

.field-label {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  gap: 10px;
  color: #ffe09a;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.02em;
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
  box-shadow: inset 0 1px 0 rgba(255, 248, 214, 0.08);
}

.field-icon-wrap .van-icon {
  color: #ffe09a;
  font-size: 16px;
}

.field-input {
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

.field-input--locked {
  color: rgba(255, 229, 186, 0.64);
  background: rgba(255, 248, 214, 0.03);
}

.code-row {
  align-items: center;
}

.code-group {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.code-group .field-input {
  flex: 1;
  min-width: 0;
}

.send-btn {
  width: 100px;
  height: 44px;
  border: 1px solid rgba(255, 248, 214, 0.42);
  border-radius: 12px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 13px;
  font-weight: 800;
  box-shadow: 0 8px 18px rgba(90, 27, 0, 0.22);
}

.send-btn:disabled {
  opacity: 0.68;
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

.lang-popup {
  min-height: 220px;
  padding: 8px 0 12px;
  background: linear-gradient(180deg, #5c0202 0%, #290000 100%);
  color: #fff0c9;
  border-radius: 24px 24px 0 0;
}

.lang-header {
  text-align: center;
  color: #ffe09a;
  font-size: 14px;
  font-weight: 700;
  padding: 10px 14px;
}

.lang-row {
  width: 100%;
  border: 0;
  background: transparent;
  padding: 12px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #fff0c9;
  font-size: 13px;
}

.lang-left {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.lang-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 18px;
  border-radius: 9px;
  background: rgba(255, 223, 135, 0.14);
  color: #ffd98b;
  font-size: 11px;
  font-style: normal;
  border: 1px solid rgba(212, 175, 55, 0.22);
}

.lang-right {
  color: rgba(255, 229, 186, 0.54);
  font-size: 11px;
}

.lang-tip {
  margin: 8px 0 0;
  text-align: center;
  color: rgba(255, 229, 186, 0.48);
  font-size: 11px;
}

@media (max-width: 390px) {
  .bind-email-page {
    padding-left: 10px;
    padding-right: 10px;
  }

  .send-btn {
    width: 92px;
    font-size: 12px;
  }
}
</style>
