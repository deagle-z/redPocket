<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { RouteMap } from 'vue-router'
import { showToast } from 'vant'
import { useUserStore } from '@/stores'
import AppPageHeader from '@/components/AppPageHeader.vue'
import emailIcon from '@/assets/svg/email.svg'
import lockIcon from '@/assets/svg/lock.svg'
import verifyIcon from '@/assets/svg/verify.svg'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const sendLoading = ref(false)
const countdown = ref(0)

const postData = reactive({
  email: '',
  code: '',
  newPassword: '',
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
  if (countdownTimer)
    clearInterval(countdownTimer)
})

async function sendCode() {
  if (!postData.email) {
    showToast(t('forgotPassword.pleaseEnterEmail'))
    return
  }
  try {
    sendLoading.value = true
    await userStore.sendCode(postData.email)
    startCountdown()
    showToast(t('forgotPassword.sendCodeSuccess'))
  }
  catch (error: any) {
    showToast(error?.message || t('forgotPassword.getCode'))
  }
  finally {
    sendLoading.value = false
  }
}

async function submitReset() {
  if (!postData.email) {
    showToast(t('forgotPassword.pleaseEnterEmail'))
    return
  }
  if (!postData.code) {
    showToast(t('forgotPassword.pleaseEnterCode'))
    return
  }
  if (!postData.newPassword) {
    showToast(t('forgotPassword.pleaseEnterPassword'))
    return
  }

  try {
    loading.value = true
    await userStore.reset({ ...postData })
    showToast(t('forgotPassword.passwordResetSuccess'))
    router.push({ name: 'Login' as keyof RouteMap })
  }
  catch (error: any) {
    showToast(error?.message || t('forgotPassword.confirm'))
  }
  finally {
    loading.value = false
  }
}

function goBack() {
  router.back()
}
</script>

<template>
  <div class="reset-page">
    <div class="reset-shell">
      <AppPageHeader
        class="reset-header"
        :title="t('forgotPassword.pageTitle')"
        @back="goBack"
      />

      <div class="reset-form-card">
        <div class="reset-row">
          <label for="reset-email" class="reset-label">
            <span class="reset-icon-wrap">
              <img :src="emailIcon" alt="email" class="reset-icon">
            </span>
            <span>{{ t('forgotPassword.email') }}</span>
          </label>
          <input
            id="reset-email"
            v-model="postData.email"
            type="text"
            class="reset-input"
            :placeholder="t('forgotPassword.pleaseEnterEmail')"
          >
        </div>

        <div class="reset-row">
          <label for="reset-code" class="reset-label">
            <span class="reset-icon-wrap">
              <img :src="verifyIcon" alt="code" class="reset-icon">
            </span>
            <span>{{ t('forgotPassword.code') }}</span>
          </label>
          <div class="reset-code-wrap">
            <input
              id="reset-code"
              v-model="postData.code"
              type="text"
              class="reset-input"
              :placeholder="t('forgotPassword.pleaseEnterCode')"
            >
            <button
              type="button"
              class="send-btn"
              :disabled="sendLoading || countdown > 0"
              @click="sendCode"
            >
              <span v-if="countdown > 0">{{ countdown }}s</span>
              <span v-else-if="sendLoading">{{ t('forgotPassword.gettingCode') }}</span>
              <span v-else>{{ t('forgotPassword.send') }}</span>
            </button>
          </div>
        </div>

        <div class="reset-row">
          <label for="reset-password" class="reset-label">
            <span class="reset-icon-wrap">
              <img :src="lockIcon" alt="new password" class="reset-icon">
            </span>
            <span>{{ t('forgotPassword.newPassword') }}</span>
          </label>
          <input
            id="reset-password"
            v-model="postData.newPassword"
            type="password"
            class="reset-input"
            :placeholder="t('forgotPassword.pleaseEnterPassword')"
          >
        </div>
      </div>

      <van-button
        :loading="loading"
        type="primary"
        round
        block
        class="reset-submit-btn"
        @click="submitReset"
      >
        {{ t('forgotPassword.confirm') }}
      </van-button>
    </div>
  </div>
</template>

<style scoped>
.reset-page {
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

.reset-shell {
  width: 100%;
  max-width: 640px;
  margin: 0 auto;
}

.reset-header {
  margin-bottom: 12px;
}

.reset-form-card {
  position: relative;
  overflow: hidden;
  margin-top: 8px;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  border-radius: 18px;
  border: 1px solid rgba(212, 175, 55, 0.24);
  box-shadow:
    0 16px 34px rgba(0, 0, 0, 0.26),
    inset 0 1px 0 rgba(255, 248, 214, 0.08);
  padding: 8px 14px;
}

.reset-form-card::before {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 2px;
  background: linear-gradient(90deg, transparent, rgba(255, 215, 0, 0.85), transparent);
}

.reset-row {
  min-height: 88px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 12px;
  padding: 12px 0;
}

.reset-row + .reset-row {
  border-top: 1px solid rgba(212, 175, 55, 0.14);
}

.reset-label {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #ffe09a;
  font-size: 13px;
  font-weight: 700;
  cursor: text;
  letter-spacing: 0.02em;
}

.reset-icon-wrap {
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

.reset-icon {
  width: 16px;
  height: 16px;
  filter: brightness(0) saturate(100%) invert(85%) sepia(39%) saturate(649%) hue-rotate(335deg) brightness(105%) contrast(97%);
}

.reset-input {
  width: 100%;
  border: 1px solid rgba(212, 175, 55, 0.22);
  border-radius: 14px;
  min-height: 48px;
  padding: 0 14px;
  background: rgba(255, 248, 214, 0.05);
  outline: none;
  color: #fff4d1;
  font-size: 14px;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.reset-input::placeholder {
  color: rgba(255, 229, 186, 0.42);
}

.reset-input:focus {
  border-color: rgba(255, 223, 135, 0.72);
  box-shadow: 0 0 0 4px rgba(212, 175, 55, 0.14);
  background: rgba(255, 248, 214, 0.08);
}

.reset-code-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
}

.reset-code-wrap .reset-input {
  flex: 1;
  min-width: 0;
}

.send-btn {
  min-width: 102px;
  height: 44px;
  border: 1px solid rgba(255, 248, 214, 0.42);
  border-radius: 12px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 13px;
  font-weight: 800;
  cursor: pointer;
  box-shadow: 0 8px 18px rgba(90, 27, 0, 0.22);
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.send-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

:deep(.reset-submit-btn.van-button) {
  margin-top: 18px;
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

:deep(.reset-submit-btn.van-button--disabled) {
  opacity: 0.72;
}

@media (max-width: 390px) {
  .reset-page {
    padding-left: 10px;
    padding-right: 10px;
  }

  .send-btn {
    min-width: 92px;
    font-size: 12px;
  }
}
</style>

<route lang="json5">
{
  name: 'ForgotPassword'
}
</route>
