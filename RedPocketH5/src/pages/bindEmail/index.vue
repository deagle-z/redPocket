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
    showToast('邮箱已绑定')
    return
  }
  const email = String(formData.email || '').trim()
  if (!email) {
    showToast('请输入邮箱')
    return
  }
  if (sending.value || countdown.value > 0)
    return
  try {
    sending.value = true
    await userStore.sendCode(email)
    showToast('发送成功')
    startCountdown()
  }
  catch (error: any) {
    showToast(error?.message || '发送失败')
  }
  finally {
    sending.value = false
  }
}

async function submitBindEmail() {
  if (boundEmail.value) {
    showToast('邮箱已绑定')
    return
  }
  const email = String(formData.email || '').trim()
  const code = String(formData.code || '').trim()
  if (!email) {
    showToast('请输入邮箱')
    return
  }
  if (!code) {
    showToast('请输入邮箱验证码')
    return
  }
  if (submitting.value)
    return
  try {
    submitting.value = true
    await bindCurrentTgEmail({ email, code })
    showToast('绑定成功')
    router.back()
  }
  catch (error: any) {
    showToast(error?.message || '绑定失败')
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
    <AppPageHeader title="邮箱" @back="goBack" @right-click="openLanguagePopup">
      <template #right>
        <img :src="languageIcon" class="lang-icon" alt="language icon">
      </template>
    </AppPageHeader>

    <section class="form-card">
      <div class="field-row">
        <label class="field-label">
          <van-icon name="friends-o" />
          <span>邮箱</span>
        </label>
        <input
          v-model="formData.email"
          type="text"
          class="field-input"
          placeholder="邮箱"
          :readonly="!!boundEmail"
        >
      </div>

      <div class="field-row code-row">
        <label class="field-label">
          <van-icon name="shield-o" />
          <span>邮箱验证码</span>
        </label>
        <div class="code-group">
          <input
            v-model="formData.code"
            type="text"
            class="field-input"
            placeholder="请输入邮箱验证码"
            :readonly="!!boundEmail"
          >
          <button
            type="button"
            class="send-btn"
            :disabled="!!boundEmail || sending || countdown > 0"
            @click="sendCode"
          >
            <span v-if="countdown > 0">{{ countdown }}s</span>
            <span v-else-if="sending">发送中</span>
            <span v-else>发送</span>
          </button>
        </div>
      </div>
    </section>

    <button type="button" class="submit-btn" :disabled="!!boundEmail || submitting" @click="submitBindEmail">
      {{ boundEmail ? '邮箱已绑定' : (submitting ? '提交中...' : '提交') }}
    </button>

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
  background: #f5f6fa;
  padding: 0 10px calc(16px + env(safe-area-inset-bottom));
}

.lang-icon {
  width: 18px;
  height: 18px;
  object-fit: contain;
}

.form-card {
  margin-top: 12px;
  border-radius: 12px;
  background: #fff;
  padding: 10px 12px;
}

.field-row {
  min-height: 54px;
  display: grid;
  grid-template-columns: 138px 1fr;
  align-items: center;
}

.field-row + .field-row {
  border-top: 1px solid #f0f0f5;
}

.field-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #1a1a2e;
  font-size: 13px;
}

.field-label .van-icon {
  color: #41b463;
  font-size: 16px;
}

.field-input {
  width: 100%;
  border: 0;
  outline: none;
  background: transparent;
  color: #1a1a2e;
  font-size: 13px;
  padding: 0;
}

.field-input::placeholder {
  color: #a8b0bf;
}

.code-row {
  align-items: center;
}

.code-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.send-btn {
  width: 62px;
  height: 34px;
  border: 0;
  border-radius: 6px;
  background: #eb5757;
  color: #fff;
  font-size: 13px;
}

.send-btn:disabled {
  opacity: 0.7;
}

.submit-btn {
  margin-top: 18px;
  width: 100%;
  height: 42px;
  border: 0;
  border-radius: 6px;
  background: #49ad62;
  color: #fff;
  font-size: 16px;
}

.submit-btn:disabled {
  opacity: 0.75;
}

.lang-popup {
  min-height: 220px;
  padding: 8px 0 12px;
}

.lang-header {
  text-align: center;
  color: #1a1a2e;
  font-size: 14px;
  font-weight: 600;
  padding: 8px 14px;
}

.lang-row {
  width: 100%;
  border: 0;
  background: #fff;
  padding: 10px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #1a1a2e;
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
  background: #edf0f5;
  color: #6b7280;
  font-size: 11px;
  font-style: normal;
}

.lang-right {
  color: #9ca3af;
  font-size: 11px;
}

.lang-tip {
  margin: 8px 0 0;
  text-align: center;
  color: #9ca3af;
  font-size: 11px;
}
</style>
