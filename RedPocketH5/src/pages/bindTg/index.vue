<script setup lang="ts">
import { showToast } from 'vant'
import { bindCurrentTgChannelName } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { safeBack } from '@/utils/navigation'

const tgName = ref('')
const submitting = ref(false)
const { t } = useI18n()
const router = useRouter()
const CHANNEL_URL = 'https://t.me/LuckyCoinsMX'

function normalizeTgName(value: string) {
  const trimmed = value.trim()
  if (!trimmed)
    return ''
  return trimmed.startsWith('@') ? trimmed : `@${trimmed}`
}

function goBack() {
  safeBack(router)
}

async function handleBind() {
  const value = normalizeTgName(tgName.value)
  if (!value) {
    showToast(t('bindTgPage.toastTgNameRequired'))
    return
  }
  submitting.value = true
  try {
    const res = await bindCurrentTgChannelName(value)
    tgName.value = res.data.tgName
    showToast(t('bindTgPage.toastBindSuccess', { count: res.data.awardedCount || 0 }))
  }
  catch (err: any) {
    showToast(err?.message || t('bindTgPage.toastBindFailed'))
  }
  finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="bind-tg-page">
    <AppPageHeader :title="t('bindTgPage.title')" @back="goBack" />

    <section class="bind-hero">
      <div class="bind-hero__icon">
        <van-icon name="gift-o" />
      </div>
      <div class="bind-hero__copy">
        <p class="bind-hero__eyebrow">
          {{ t('bindTgPage.activityEyebrow') }}
        </p>
        <h2>{{ t('bindTgPage.activityTitle') }}</h2>
        <p class="bind-hero__desc">
          <span>{{ t('bindTgPage.activityDescBefore') }}</span>
          <a class="channel-link" :href="CHANNEL_URL" target="_blank" rel="noopener noreferrer">
            <van-icon name="guide-o" />
            <span>{{ CHANNEL_URL }}</span>
          </a>
          <span>{{ t('bindTgPage.activityDescAfter') }}</span>
        </p>
      </div>
    </section>

    <van-form class="bind-form" @submit="handleBind">
      <section class="bind-card">
        <label class="field-label" for="tgName">{{ t('bindTgPage.tgNameLabel') }}</label>
        <van-field
          id="tgName"
          v-model="tgName"
          name="tgName"
          clearable
          class="tg-field"
          :border="false"
          :placeholder="t('bindTgPage.tgNamePlaceholder')"
        >
          <template #left-icon>
            <van-icon name="contact-o" class="tg-field-icon" />
          </template>
        </van-field>
        <p class="field-hint">
          {{ t('bindTgPage.inputHint') }}
        </p>

        <button type="submit" class="bind-submit-btn" :disabled="submitting">
          <van-loading v-if="submitting" size="18" color="#7c2200" />
          <van-icon v-else name="success" />
          <span>{{ t('bindTgPage.bindButton') }}</span>
        </button>
      </section>
    </van-form>
  </div>
</template>

<style scoped>
.bind-tg-page {
  min-height: 100vh;
  padding: 8px 12px calc(90px + env(safe-area-inset-bottom));
  background-image:
    radial-gradient(circle at 20% 10%, rgba(212, 175, 55, 0.18), transparent 30%),
    radial-gradient(circle at 80% 88%, rgba(255, 215, 0, 0.12), transparent 28%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.04) 18px,
      rgba(212, 175, 55, 0.04) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #230000 62%, #160000 100%);
  color: #fff0c9;
}

.bind-hero {
  margin-top: 12px;
  padding: 18px 16px;
  border: 1px solid rgba(212, 175, 55, 0.46);
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(255, 244, 205, 0.12), rgba(255, 244, 205, 0) 42%),
    linear-gradient(160deg, rgba(140, 0, 0, 0.96) 0%, rgba(74, 0, 0, 0.96) 100%);
  box-shadow:
    0 14px 26px rgba(0, 0, 0, 0.32),
    inset 0 0 0 1px rgba(255, 248, 214, 0.1);
  display: flex;
  gap: 14px;
  align-items: flex-start;
}

.bind-hero__icon {
  flex: 0 0 auto;
  width: 48px;
  height: 48px;
  border-radius: 16px;
  background: linear-gradient(180deg, #fff4ba 0%, #ffd14b 42%, #b86b08 100%);
  color: #7a2100;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  box-shadow:
    0 10px 20px rgba(0, 0, 0, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.42);
}

.bind-hero__copy {
  min-width: 0;
}

.bind-hero__eyebrow {
  margin: 0 0 6px;
  color: #ffd98b;
  font-size: 11px;
  line-height: 1;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.bind-hero h2 {
  margin: 0;
  color: #fff0c9;
  font-size: 20px;
  line-height: 1.2;
  font-weight: 900;
}

.bind-hero__desc {
  margin: 8px 0 0;
  color: rgba(255, 229, 186, 0.76);
  font-size: 13px;
  line-height: 1.5;
}

.channel-link {
  position: relative;
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  margin: 4px 4px 2px 0;
  padding: 4px 9px;
  border: 1px solid rgba(255, 236, 157, 0.58);
  border-radius: 999px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.28), transparent 48%),
    linear-gradient(135deg, #ffec9c 0%, #ffbb00 48%, #d27900 100%);
  box-shadow:
    0 6px 14px rgba(0, 0, 0, 0.26),
    0 0 14px rgba(255, 187, 0, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.5);
  color: #4a0c00;
  font-size: 12px;
  line-height: 1.2;
  font-weight: 900;
  text-decoration: none;
  vertical-align: middle;
  transition:
    transform 160ms ease,
    box-shadow 160ms ease,
    filter 160ms ease;
}

.channel-link::after {
  content: '';
  position: absolute;
  inset: -3px;
  border-radius: inherit;
  border: 1px solid rgba(255, 236, 157, 0.2);
  opacity: 0;
  transform: scale(0.96);
  transition:
    opacity 160ms ease,
    transform 160ms ease;
  pointer-events: none;
}

.channel-link:active {
  transform: translateY(1px) scale(0.98);
  filter: saturate(1.08);
  box-shadow:
    0 3px 8px rgba(0, 0, 0, 0.24),
    0 0 10px rgba(255, 187, 0, 0.14),
    inset 0 1px 0 rgba(255, 255, 255, 0.42);
}

.channel-link:hover::after,
.channel-link:focus-visible::after {
  opacity: 1;
  transform: scale(1);
}

.channel-link :deep(.van-icon) {
  flex: 0 0 auto;
  margin-right: 5px;
  font-size: 13px;
}

.channel-link span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bind-form {
  margin-top: 14px;
}

.bind-card {
  padding: 16px;
  border: 1px solid rgba(212, 175, 55, 0.4);
  border-radius: 18px;
  background: linear-gradient(170deg, rgba(72, 0, 0, 0.94), rgba(34, 0, 0, 0.96));
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.26),
    inset 0 0 0 1px rgba(255, 248, 214, 0.07);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.field-label {
  color: #ffd98b;
  font-size: 13px;
  font-weight: 800;
}

:deep(.tg-field.van-cell) {
  min-height: 50px;
  padding: 0 14px;
  border: 1px solid rgba(255, 229, 186, 0.28);
  border-radius: 14px;
  background: rgba(255, 248, 214, 0.08);
  align-items: center;
}

:deep(.tg-field .van-field__control) {
  color: #fff0c9;
  font-size: 15px;
  font-weight: 700;
}

:deep(.tg-field .van-field__control::placeholder) {
  color: rgba(255, 229, 186, 0.42);
}

.tg-field-icon {
  color: #ffd98b;
  font-size: 16px;
}

.field-hint {
  margin: 0;
  color: rgba(255, 229, 186, 0.62);
  font-size: 12px;
  line-height: 1.45;
}

.bind-submit-btn {
  width: 100%;
  min-height: 48px;
  margin-top: 8px;
  border: 1px solid rgba(255, 236, 157, 0.54);
  border-radius: 999px;
  background: linear-gradient(180deg, #fff9d0 0%, #ffd358 48%, #bf780c 100%);
  box-shadow:
    0 10px 20px rgba(0, 0, 0, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.52);
  color: #7c2200;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 900;
}

.bind-submit-btn:disabled {
  opacity: 0.72;
}
</style>
