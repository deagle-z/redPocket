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
            <img :src="emailIcon" alt="email" class="reset-icon">
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
            <img :src="verifyIcon" alt="code" class="reset-icon">
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
            <img :src="lockIcon" alt="new password" class="reset-icon">
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
  background:
    radial-gradient(1200px 500px at 50% -240px, rgba(84, 185, 105, 0.12), transparent 65%), var(--color-bg-page);
  padding: 0 var(--page-padding-x) 32px;
}

.reset-shell {
  width: 100%;
  max-width: 640px;
  margin: 0 auto;
}

.reset-header {
  margin: 0 calc(-1 * var(--page-padding-x));
  padding: 0 14px;
  background: var(--color-bg-card);
  border-bottom: 1px solid rgba(16, 24, 40, 0.06);
}

.reset-form-card {
  margin-top: 20px;
  background: var(--color-bg-card);
  border-radius: var(--radius-2xl);
  border: 1px solid rgba(16, 24, 40, 0.08);
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.05);
  padding: 8px 16px;
}

.reset-row {
  min-height: 86px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 10px;
  padding: 8px 0;
}

.reset-row + .reset-row {
  border-top: 1px solid var(--color-border);
}

.reset-label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-text-form);
  font-size: var(--font-md);
  font-weight: 500;
  cursor: text;
}

.reset-icon {
  width: 22px;
  height: 22px;
}

.reset-input {
  width: 100%;
  border: 1px solid rgba(16, 24, 40, 0.14);
  border-radius: 10px;
  min-height: 46px;
  padding: 0 14px;
  background: rgba(255, 255, 255, 0.96);
  outline: none;
  color: var(--color-text-input);
  font-size: var(--font-base);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.reset-input::placeholder {
  color: var(--color-text-muted);
}

.reset-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 4px rgba(101, 177, 104, 0.16);
}

.reset-code-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.send-btn {
  min-width: 96px;
  height: 42px;
  border: none;
  border-radius: 10px;
  background: var(--color-danger);
  color: var(--color-bg-card);
  font-size: var(--font-sm);
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.send-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.reset-submit-btn {
  margin-top: 18px;
  background: var(--color-primary-btn);
  border: none;
  height: 54px;
  font-size: var(--font-xl);
  font-weight: 500;
  box-shadow: 0 10px 24px rgba(101, 177, 104, 0.34);
}
</style>

<route lang="json5">
{
  name: 'ForgotPassword'
}
</route>
