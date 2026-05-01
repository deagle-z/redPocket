<script setup lang="ts">
import { showToast } from 'vant'
import { appUpload, getCurrentTgUserInfo, setAudioOpen, updateCurrentTgAvatar, updateCurrentTgName } from '@/api/user'
import avatarPlaceholderIcon from '@/assets/my/question-circle.svg'

const { t } = useI18n()
const router = useRouter()

const PROFILE_AVATAR_KEY = 'profile_custom_avatar'
const AUDIO_OPEN_KEY = 'setting_audio_open'

const AVATAR_BASE = 'https://pub-bd25d6a357314ec1823d725e93570e3d.r2.dev/game/'
const avatarOptions = Array.from({ length: 9 }, (_, i) => `${AVATAR_BASE}avatar${i + 1}.png`)

// ── Name ──────────────────────────────────────────────────────────
const currentName = ref('')
const nameDraft = ref('')
const showNamePopup = ref(false)
const savingName = ref(false)

function openNamePopup() {
  nameDraft.value = currentName.value
  showNamePopup.value = true
}

function closeNamePopup() {
  showNamePopup.value = false
}

async function saveName() {
  if (savingName.value)
    return
  const nextName = nameDraft.value.trim()
  if (!nextName) {
    showToast(t('settingPage.nameRequired'))
    return
  }
  if ([...nextName].length > 64) {
    showToast(t('settingPage.nameTooLong'))
    return
  }
  savingName.value = true
  try {
    const { data } = await updateCurrentTgName(nextName)
    currentName.value = data?.username || nextName
    showToast(t('settingPage.nameUpdated'))
    closeNamePopup()
  }
  catch {
    showToast(t('common.requestFailed'))
  }
  finally {
    savingName.value = false
  }
}

// ── Avatar ────────────────────────────────────────────────────────
const currentAvatar = ref(localStorage.getItem(PROFILE_AVATAR_KEY) || '')
const showAvatarPopup = ref(false)
const uploadingAvatar = ref(false)
const avatarFileInput = ref<HTMLInputElement>()

function openAvatarPopup() {
  showAvatarPopup.value = true
}

function closeAvatarPopup() {
  showAvatarPopup.value = false
}

async function selectAvatar(url: string) {
  if (uploadingAvatar.value)
    return
  try {
    uploadingAvatar.value = true
    await updateCurrentTgAvatar(url)
    currentAvatar.value = url
    localStorage.setItem(PROFILE_AVATAR_KEY, url)
    showToast(t('profilePage.toastAvatarUpdated'))
    closeAvatarPopup()
  }
  finally {
    uploadingAvatar.value = false
  }
}

function triggerAvatarUpload() {
  if (uploadingAvatar.value)
    return
  avatarFileInput.value?.click()
}

async function handleAvatarFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target?.files?.[0]
  target.value = ''
  if (!file)
    return
  if (!file.type.startsWith('image/')) {
    showToast(t('profilePage.toastUploadImageOnly'))
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    showToast(t('profilePage.toastUploadTooLarge'))
    return
  }
  if (uploadingAvatar.value)
    return
  uploadingAvatar.value = true
  try {
    const uploadRes = await appUpload(file)
    const url = uploadRes?.data?.url || ''
    if (!url) {
      showToast(t('profilePage.toastUploadFailed'))
      return
    }
    await updateCurrentTgAvatar(url)
    currentAvatar.value = url
    localStorage.setItem(PROFILE_AVATAR_KEY, url)
    showToast(t('profilePage.toastAvatarUpdated'))
    closeAvatarPopup()
  }
  finally {
    uploadingAvatar.value = false
  }
}

// ── Audio ─────────────────────────────────────────────────────────
const audioOpen = ref(localStorage.getItem(AUDIO_OPEN_KEY) !== '0')
const savingAudio = ref(false)

async function loadSettingData() {
  try {
    const { data } = await getCurrentTgUserInfo()
    currentName.value = data?.username || ''
    if (data?.avatar) {
      currentAvatar.value = data.avatar
      localStorage.setItem(PROFILE_AVATAR_KEY, data.avatar)
    }
    const val = data?.audioOpen === 1
    audioOpen.value = val
    localStorage.setItem(AUDIO_OPEN_KEY, val ? '1' : '0')
  }
  catch { /* keep local values */ }
}

async function toggleAudio() {
  if (savingAudio.value)
    return
  const next = !audioOpen.value
  savingAudio.value = true
  try {
    await setAudioOpen(next ? 1 : 0)
    audioOpen.value = next
    localStorage.setItem(AUDIO_OPEN_KEY, next ? '1' : '0')
  }
  catch {
    showToast(t('common.requestFailed'))
  }
  finally {
    savingAudio.value = false
  }
}

onMounted(() => {
  loadSettingData()
})
</script>

<template>
  <div class="setting-page">
    <AppPageHeader :title="t('settingPage.title')" @back="router.back()" />

    <div class="setting-body">
      <!-- 用户名 -->
      <p class="section-label">
        {{ t('settingPage.name') }}
      </p>
      <section class="menu-card">
        <button type="button" class="name-row" @click="openNamePopup">
          <span class="name-content">
            <span class="name-label">{{ t('settingPage.currentName') }}</span>
            <span class="name-value">{{ currentName || t('settingPage.nameUnset') }}</span>
          </span>
          <span class="name-hint">{{ t('settingPage.changeName') }}</span>
          <van-icon name="arrow" class="row-arrow" />
        </button>
      </section>

      <!-- 头像 -->
      <p class="section-label">
        {{ t('settingPage.avatar') }}
      </p>
      <section class="menu-card">
        <button type="button" class="avatar-row" @click="openAvatarPopup">
          <img
            :src="currentAvatar || avatarPlaceholderIcon"
            class="avatar-preview"
            alt="avatar"
          >
          <span class="avatar-hint">{{ t('settingPage.changeAvatar') }}</span>
          <van-icon name="arrow" class="row-arrow" />
        </button>
      </section>

      <!-- 音乐 -->
      <p class="section-label">
        {{ t('settingPage.music') }}
      </p>
      <section class="menu-card">
        <div class="toggle-row">
          <span class="toggle-label">{{ t('settingPage.bgMusic') }}</span>
          <van-switch
            :model-value="audioOpen"
            :loading="savingAudio"
            active-color="#d4af37"
            inactive-color="#444"
            @update:model-value="toggleAudio"
          />
        </div>
      </section>
    </div>

    <!-- 头像选择弹窗 -->
    <van-popup v-model:show="showAvatarPopup" round position="bottom" class="avatar-popup">
      <div class="avatar-popup-header">
        <span class="avatar-popup-title">{{ t('profilePage.avatarPopupTitle') }}</span>
        <button class="avatar-popup-close" @click="closeAvatarPopup">
          ×
        </button>
      </div>

      <div class="avatar-grid">
        <button
          v-for="item in avatarOptions"
          :key="item"
          type="button"
          class="avatar-option"
          :class="{ active: currentAvatar === item }"
          @click="selectAvatar(item)"
        >
          <img :src="item" alt="" class="avatar-option-img">
        </button>
      </div>

      <div class="avatar-divider">
        <span>{{ t('profilePage.avatarPopupOr') }}</span>
      </div>

      <div class="avatar-upload-wrap">
        <button type="button" class="avatar-upload-btn" :disabled="uploadingAvatar" @click="triggerAvatarUpload">
          <van-icon name="photograph" />
          <span>{{ uploadingAvatar ? t('profilePage.avatarUploading') : t('profilePage.avatarUpload') }}</span>
        </button>
        <input ref="avatarFileInput" type="file" accept="image/*" class="avatar-file-input" @change="handleAvatarFileChange">
      </div>
    </van-popup>

    <!-- 用户名修改弹窗 -->
    <van-popup v-model:show="showNamePopup" round position="bottom" class="name-popup">
      <div class="avatar-popup-header">
        <span class="avatar-popup-title">{{ t('settingPage.namePopupTitle') }}</span>
        <button class="avatar-popup-close" @click="closeNamePopup">
          ×
        </button>
      </div>

      <div class="name-form">
        <van-field
          v-model="nameDraft"
          class="name-field"
          :label="t('settingPage.nameInputLabel')"
          :placeholder="t('settingPage.namePlaceholder')"
          maxlength="64"
          show-word-limit
          clearable
        />
        <button type="button" class="name-save-btn" :disabled="savingName" @click="saveName">
          <van-icon name="success" />
          <span>{{ savingName ? t('settingPage.savingName') : t('settingPage.saveName') }}</span>
        </button>
      </div>
    </van-popup>
  </div>
</template>

<style scoped>
.setting-page {
  min-height: 100vh;
  background-image:
    radial-gradient(circle at 18% 10%, rgba(212, 175, 55, 0.18), transparent 28%),
    radial-gradient(circle at 84% 82%, rgba(255, 215, 0, 0.1), transparent 24%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.04) 18px,
      rgba(212, 175, 55, 0.04) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #240000 60%, #150000 100%);
  padding: 12px 12px calc(88px + env(safe-area-inset-bottom));
  color: #f8e8c6;
}

.setting-body {
  display: flex;
  flex-direction: column;
}

.section-label {
  padding: 12px 4px 6px;
  font-size: 12px;
  color: rgba(248, 232, 198, 0.5);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.menu-card {
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.28);
  background: linear-gradient(160deg, rgba(126, 0, 0, 0.72) 0%, rgba(43, 0, 0, 0.88) 100%);
  overflow: hidden;
}

/* Name row */
.name-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 14px 16px;
  background: transparent;
  text-align: left;
}

.name-content {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.name-label {
  font-size: 13px;
  color: rgba(248, 232, 198, 0.55);
}

.name-value {
  font-size: 16px;
  font-weight: 700;
  color: #fff0c9;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.name-hint {
  flex-shrink: 0;
  max-width: 108px;
  color: rgba(248, 232, 198, 0.62);
  font-size: 13px;
  text-align: right;
}

/* Avatar row */
.avatar-row {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  padding: 14px 16px;
  background: transparent;
  text-align: left;
}

.avatar-preview {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(212, 175, 55, 0.4);
  flex-shrink: 0;
}

.avatar-hint {
  flex: 1;
  font-size: 15px;
  color: #f8e8c6;
}

.row-arrow {
  color: rgba(248, 232, 198, 0.4);
  font-size: 14px;
}

/* Toggle row */
.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
}

.toggle-label {
  font-size: 15px;
  color: #f8e8c6;
}

/* Avatar popup — identical to profile page */
.avatar-popup {
  min-height: 420px;
  padding: 0 0 24px;
  background:
    radial-gradient(circle at top, rgba(212, 175, 55, 0.14), transparent 26%),
    linear-gradient(180deg, #540000 0%, #280000 100%);
  border: 1px solid rgba(212, 175, 55, 0.34);
}

.avatar-popup-header {
  height: 62px;
  border-bottom: 1px solid rgba(212, 175, 55, 0.16);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.avatar-popup-title {
  font-size: 18px;
  font-weight: 700;
  color: #fff0c9;
}

.avatar-popup-close {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  border: 0;
  background: transparent;
  color: rgba(255, 229, 186, 0.7);
  font-size: 18px;
  line-height: 1;
}

.avatar-grid {
  padding: 16px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px 14px;
}

.avatar-option {
  border: 0;
  background: transparent;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.avatar-option-img {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(255, 248, 214, 0.14);
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.24);
}

.avatar-option.active .avatar-option-img {
  border-color: #ffd87f;
  box-shadow: 0 0 0 4px rgba(212, 175, 55, 0.18);
}

.avatar-divider {
  margin: 8px 16px 0;
  display: flex;
  align-items: center;
  gap: 12px;
  color: rgba(255, 229, 186, 0.56);
  font-size: 14px;
}

.avatar-divider::before,
.avatar-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: rgba(212, 175, 55, 0.16);
}

.avatar-upload-wrap {
  padding: 14px 16px 0;
}

.avatar-upload-btn {
  width: 100%;
  height: 48px;
  border: 1px solid rgba(212, 175, 55, 0.34);
  border-radius: 14px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 700;
}

.avatar-upload-btn:disabled {
  opacity: 0.6;
}

.avatar-file-input {
  display: none;
}

.name-popup {
  min-height: 260px;
  padding: 0 0 calc(24px + env(safe-area-inset-bottom));
  background:
    radial-gradient(circle at top, rgba(212, 175, 55, 0.14), transparent 26%),
    linear-gradient(180deg, #540000 0%, #280000 100%);
  border: 1px solid rgba(212, 175, 55, 0.34);
}

.name-form {
  padding: 18px 16px 0;
}

.name-field {
  border: 1px solid rgba(212, 175, 55, 0.28);
  border-radius: 14px;
  overflow: hidden;
  --van-field-label-color: rgba(255, 240, 201, 0.74);
  --van-field-input-text-color: #fff0c9;
  --van-field-placeholder-text-color: rgba(255, 240, 201, 0.38);
  --van-cell-background: rgba(24, 0, 0, 0.38);
  --van-cell-group-background: transparent;
  --van-text-color-2: rgba(255, 240, 201, 0.5);
}

.name-save-btn {
  margin-top: 18px;
  width: 100%;
  height: 48px;
  border: 1px solid rgba(212, 175, 55, 0.34);
  border-radius: 14px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 700;
}

.name-save-btn:disabled {
  opacity: 0.6;
}

@media (max-width: 360px) {
  .name-row {
    gap: 8px;
    padding: 14px 12px;
  }

  .name-hint {
    max-width: 84px;
    font-size: 12px;
  }
}
</style>

<route lang="json5">
{
  name: 'Setting',
}
</route>
