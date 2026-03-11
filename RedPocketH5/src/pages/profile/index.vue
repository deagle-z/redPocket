<script setup lang="ts">
import { appUpload, getCurrentTgUserInfo, tgLogout, updateCurrentTgAvatar } from '@/api/user'
import { showToast } from 'vant'
import { clearToken, isLogin } from '@/utils/auth'
import { locale } from '@/utils/i18n'
import avatarPlaceholderIcon from '@/assets/my/question-circle.svg'
import avatar1 from '@/assets/images/avatar1.png'
import avatar2 from '@/assets/images/avatar2.png'
import avatar3 from '@/assets/images/avatar3.png'
import avatar4 from '@/assets/images/avatar4.png'
import avatar5 from '@/assets/images/avatar5.png'
import avatar6 from '@/assets/images/avatar6.png'
import avatar7 from '@/assets/images/avatar7.png'
import avatar8 from '@/assets/images/avatar8.png'
import avatar9 from '@/assets/images/avatar9.png'
import bankCardIcon from '@/assets/my/bank card.svg?raw'
import chartBarAltIcon from '@/assets/my/chart-bar-alt.svg?raw'
import customerServiceIcon from '@/assets/my/customer-service-fill.svg?raw'
import downIcon from '@/assets/my/down2.svg?raw'
import emailIcon from '@/assets/my/email.svg?raw'
import gamesIcon from '@/assets/my/games-2.svg?raw'
import passwordIcon from '@/assets/my/Password.svg?raw'
import questionCircleIcon from '@/assets/my/question-circle.svg?raw'
import shareIcon from '@/assets/my/share.svg?raw'
import teamIcon from '@/assets/my/team.svg?raw'
import telegramIcon from '@/assets/my/telegram.svg?raw'
import upIcon from '@/assets/my/up2.svg?raw'
import walletIcon from '@/assets/my/wallet.svg?raw'
import { CURRENCY_SYMBOL } from '@/utils/currency'

const { t } = useI18n()

type ExtraTone = 'success' | 'muted'

interface MenuItem {
  key: string
  label: string
  icon: string
  extra?: string
  tone?: ExtraTone
}

function normalizeInlineSvg(svg: string) {
  return svg
    .replace(/<\?xml[\s\S]*?\?>/gi, '')
    .replace(/<!DOCTYPE[\s\S]*?>/gi, '')
    .replace(/fill="[^"]*"/gi, 'fill="currentColor"')
    .trim()
}

const profileLoading = ref(false)
const showAvatarPopup = ref(false)
const uploadingAvatar = ref(false)
const avatarFileInput = ref<HTMLInputElement>()
const PROFILE_AVATAR_KEY = 'profile_custom_avatar'
const avatarOptions = [avatar1, avatar2, avatar3, avatar4, avatar5, avatar6, avatar7, avatar8, avatar9]
const profile = reactive({
  avatar: '',
  username: '',
  uid: '',
  tgId: 0,
  email: '',
  balance: 0,
  rebateAmount: 0,
})

const accountMenus = computed<MenuItem[]>(() => [
  { key: 'wallet', label: t('profilePage.accountWallet'), icon: normalizeInlineSvg(walletIcon) },
  { key: 'recharge', label: t('profilePage.accountRecharge'), icon: normalizeInlineSvg(downIcon) },
  { key: 'withdraw', label: t('profilePage.accountWithdraw'), icon: normalizeInlineSvg(upIcon) },
  { key: 'withdraw-account', label: t('profilePage.accountWithdrawAccount'), icon: normalizeInlineSvg(bankCardIcon) },
  // { key: 'lucky-reward', label: t('profilePage.accountLuckyReward'), icon: gamesIcon },
])

const promoMenus = computed<MenuItem[]>(() => [
  { key: 'team', label: t('profilePage.promoTeam'), icon: normalizeInlineSvg(teamIcon) },
  { key: 'invite', label: t('profilePage.promoInvite'), icon: normalizeInlineSvg(shareIcon), extra: t('profilePage.promoInviteExtra'), tone: 'success' },
  { key: 'rebate', label: t('profilePage.promoRebate'), icon: normalizeInlineSvg(chartBarAltIcon) },
])

const otherMenus = computed<MenuItem[]>(() => [
  { key: 'language', label: t('profilePage.serviceLanguage'), icon: normalizeInlineSvg(shareIcon) },
  { key: 'rules', label: t('profilePage.serviceRules'), icon: normalizeInlineSvg(gamesIcon) },
  { key: 'bind-tg', label: t('profilePage.serviceBindTg'), icon: normalizeInlineSvg(telegramIcon), extra: formatMaskedNumber(profile.tgId), tone: 'muted' },
  { key: 'bind-email', label: t('profilePage.serviceBindEmail'), icon: normalizeInlineSvg(emailIcon), extra: formatMaskedEmail(profile.email), tone: 'muted' },
  { key: 'change-password', label: t('profilePage.serviceChangePassword'), icon: normalizeInlineSvg(passwordIcon) },
  { key: 'questions', label: t('profilePage.serviceQuestions'), icon: normalizeInlineSvg(questionCircleIcon) },
  { key: 'cs', label: t('profilePage.serviceCs'), icon: normalizeInlineSvg(customerServiceIcon) },
])

const displayName = computed(() => profile.username || '--')
const displayUid = computed(() => profile.uid || '--')
const displayBalance = computed(() => Number(profile.balance || 0).toFixed(2))
const displayRebateAmount = computed(() => Number(profile.rebateAmount || 0).toFixed(2))

function formatMaskedNumber(value: number) {
  const text = String(value || '').trim()
  if (!text)
    return t('profilePage.notBound')
  if (text.length <= 4)
    return `${text.slice(0, 1)}***${text.slice(-1)}`
  return `${text.slice(0, 1)}***${text.slice(-3)}`
}

function formatMaskedEmail(email: string) {
  const text = String(email || '').trim()
  if (!text)
    return t('profilePage.notBound')
  const atIndex = text.indexOf('@')
  if (atIndex <= 1)
    return text
  const prefix = text.slice(0, atIndex)
  return `${prefix.slice(0, 1)}***${prefix.slice(-1)}${text.slice(atIndex)}`
}

async function loadCurrentTgUserInfo() {
  if (!isLogin()) {
    return
  }
  if (profileLoading.value)
    return
  try {
    profileLoading.value = true
    const { data } = await getCurrentTgUserInfo()
    profile.avatar = data?.avatar || ''
    profile.username = data?.username || ''
    profile.uid = data?.uid || ''
    profile.tgId = Number(data?.tg_id || 0)
    profile.email = data?.email || ''
    profile.balance = Number(data?.balance || 0)
    profile.rebateAmount = Number(data?.rebate_amount || 0)
    const customAvatar = localStorage.getItem(PROFILE_AVATAR_KEY) || ''
    if (customAvatar)
      profile.avatar = customAvatar
  }
  catch {
    showToast(t('profilePage.toastLoadFailed'))
  }
  finally {
    profileLoading.value = false
  }
}

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
    profile.avatar = url
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
    profile.avatar = url
    localStorage.setItem(PROFILE_AVATAR_KEY, url)
    showToast(t('profilePage.toastAvatarUpdated'))
    closeAvatarPopup()
  }
  finally {
    uploadingAvatar.value = false
  }
}

onMounted(() => {
  loadCurrentTgUserInfo()
})

const showLogoutDialog = ref(false)
const logoutLoading = ref(false)
const showLangPopup = ref(false)
const router = useRouter()
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

function openLogoutDialog() {
  showLogoutDialog.value = true
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

function goByPath(path: string) {
  router.push(path)
}

function goWallet() {
  goByPath('/wallet')
}

function goTransform() {
  goByPath('/transform')
}

function onMenuClick(item: MenuItem) {
  switch (item.key) {
    case 'wallet':
      goByPath('/wallet')
      break
    case 'recharge':
      goByPath('/recharge')
      break
    case 'withdraw':
      goByPath('/withdraw')
      break
    case 'withdraw-account':
      goByPath('/bindTg')
      break
    case 'lucky-reward':
      goByPath('/history')
      break
    case 'team':
      goByPath('/team')
      break
    case 'invite':
      goByPath('/invite')
      break
    case 'rebate':
      goByPath('/rebate')
      break
    case 'language':
      openLanguagePopup()
      break
    case 'rules':
      goByPath('/questions')
      break
    case 'bind-tg':
      goByPath('/bindTg')
      break
    case 'bind-email':
      goByPath('/bindEmail')
      break
    case 'change-password':
      goByPath('/resetpwd')
      break
    case 'questions':
      goByPath('/questions')
      break
    case 'cs':
      goByPath('/cs')
      break
    default:
      break
  }
}

async function handleConfirmLogout() {
  if (logoutLoading.value)
    return

  logoutLoading.value = true
  try {
    if (isLogin())
      await tgLogout()
  }
  catch {
    // Even if the backend token is already invalid, clear local auth and continue.
  }
  finally {
    clearToken()
    showLogoutDialog.value = false
    logoutLoading.value = false
    showToast(t('profilePage.toastLogoutSuccess'))
    router.replace('/login')
  }
}
</script>

<template>
  <div class="profile-page">
    <section class="profile-card">
      <div class="profile-top">
        <button class="avatar-box" type="button" @click="openAvatarPopup">
          <img :src="profile.avatar || avatarPlaceholderIcon" alt="" class="avatar-icon">
        </button>
        <div class="profile-meta">
          <div class="name-row">
            <h3 class="user-name">
              {{ displayName }}
            </h3>
            <span class="vip-tag">VIP 0</span>
          </div>
          <p class="user-id">
            ID: {{ displayUid }}
          </p>
        </div>
      </div>

      <div class="balance-row">
        <button class="balance-item" type="button" @click="goWallet">
          <strong class="balance-value">{{ displayBalance }}</strong>
          <span class="balance-label">{{ t('profilePage.balanceLabel', { symbol: CURRENCY_SYMBOL }) }}</span>
        </button>
        <div class="balance-divider" />
        <button class="balance-item" type="button" @click="goTransform">
          <strong class="balance-value">{{ displayRebateAmount }}</strong>
          <span class="balance-label">{{ t('profilePage.commissionLabel', { symbol: CURRENCY_SYMBOL }) }}</span>
        </button>
      </div>
    </section>

    <p class="section-label">
      {{ t('profilePage.sectionAccount') }}
    </p>
    <section class="menu-card">
      <button v-for="item in accountMenus" :key="item.key" type="button" class="menu-row" @click="onMenuClick(item)">
        <div class="menu-left">
          <span class="menu-icon" v-html="item.icon" />
          <span class="menu-text">{{ item.label }}</span>
        </div>
        <div class="menu-right">
          <span v-if="item.extra" class="menu-extra" :class="item.tone">{{ item.extra }}</span>
          <van-icon name="arrow" class="menu-arrow" />
        </div>
      </button>
    </section>

    <p class="section-label">
      {{ t('profilePage.sectionPromo') }}
    </p>
    <section class="menu-card">
      <button v-for="item in promoMenus" :key="item.key" type="button" class="menu-row" @click="onMenuClick(item)">
        <div class="menu-left">
          <span class="menu-icon" v-html="item.icon" />
          <span class="menu-text">{{ item.label }}</span>
        </div>
        <div class="menu-right">
          <span v-if="item.extra" class="menu-extra" :class="item.tone">{{ item.extra }}</span>
          <van-icon name="arrow" class="menu-arrow" />
        </div>
      </button>
    </section>

    <p class="section-label">
      {{ t('profilePage.sectionService') }}
    </p>
    <section class="menu-card">
      <button v-for="item in otherMenus" :key="item.key" type="button" class="menu-row" @click="onMenuClick(item)">
        <div class="menu-left">
          <span class="menu-icon" v-html="item.icon" />
          <span class="menu-text">{{ item.label }}</span>
        </div>
        <div class="menu-right">
          <span v-if="item.extra" class="menu-extra" :class="item.tone">{{ item.extra }}</span>
          <van-icon name="arrow" class="menu-arrow" />
        </div>
      </button>
    </section>

    <div class="logout-wrap">
      <button type="button" class="logout-btn" @click="openLogoutDialog">
        {{ t('profilePage.logout') }}
      </button>
    </div>

    <AppConfirmDialog
      v-model:show="showLogoutDialog"
      :title="t('profilePage.logoutTitle')"
      :cancel-text="t('profilePage.logoutCancel')"
      :confirm-text="t('profilePage.logoutConfirm')"
      @confirm="handleConfirmLogout"
    >
      {{ t('profilePage.logoutContent') }}
    </AppConfirmDialog>

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
          :class="{ active: profile.avatar === item }"
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
  </div>
</template>

<style scoped>
.profile-page {
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

.profile-card {
  position: relative;
  overflow: hidden;
  border-radius: 18px;
  border: 1px solid rgba(212, 175, 55, 0.42);
  background: linear-gradient(160deg, rgba(126, 0, 0, 0.96) 0%, rgba(82, 0, 0, 0.97) 60%, rgba(43, 0, 0, 0.98) 100%);
  padding: 18px 16px 18px;
  box-shadow:
    0 14px 28px rgba(0, 0, 0, 0.34),
    inset 0 0 0 1px rgba(255, 248, 214, 0.12);
}

.profile-card::after {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}

.profile-card::before {
  content: '';
  position: absolute;
  inset: 3px 0 0;
  background-image: radial-gradient(rgba(212, 175, 55, 1) 1px, transparent 1px);
  background-size: 18px 18px;
  opacity: 0.05;
  pointer-events: none;
}

.profile-top {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}

.avatar-box {
  width: 62px;
  height: 62px;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 222, 138, 0.18), rgba(92, 18, 0, 0.62));
  border: 1px solid rgba(212, 175, 55, 0.44);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  box-shadow:
    0 10px 18px rgba(0, 0, 0, 0.26),
    inset 0 1px 0 rgba(255, 255, 255, 0.18);
}

.avatar-icon {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
  opacity: 1;
}

.profile-meta {
  flex: 1;
  min-width: 0;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.user-name {
  margin: 0;
  color: #fff0c9;
  font-size: 19px;
  line-height: 1.1;
  font-weight: 800;
  letter-spacing: 0.03em;
}

.vip-tag {
  border: 1px solid rgba(255, 248, 214, 0.46);
  border-radius: 999px;
  color: #5a1b00;
  font-size: 11px;
  line-height: 1;
  padding: 4px 8px;
  font-weight: 700;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  box-shadow: 0 4px 10px rgba(75, 25, 0, 0.25);
}

.user-id {
  margin: 4px 0 0;
  color: rgba(255, 229, 186, 0.68);
  font-size: 13px;
  line-height: 1;
}

.balance-row {
  position: relative;
  z-index: 1;
  margin-top: 16px;
  border-top: 1px solid rgba(212, 175, 55, 0.2);
  display: flex;
  align-items: center;
  height: 60px;
}

.balance-item {
  border: none;
  background: transparent;
  padding: 0;
  flex: 1;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 4px;
  cursor: pointer;
}

.balance-value {
  color: #ffd87f;
  font-size: 19px;
  line-height: 1;
  font-weight: 800;
}

.balance-label {
  color: rgba(255, 229, 186, 0.68);
  font-size: 12px;
  line-height: 1;
}

.balance-divider {
  width: 1px;
  height: 30px;
  background: rgba(212, 175, 55, 0.2);
}

.section-label {
  margin: 16px 0 8px;
  padding: 0 4px;
  color: #ffd98b;
  font-size: 11px;
  line-height: 1;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.menu-card {
  border-radius: 16px;
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  border: 1px solid rgba(212, 175, 55, 0.34);
  overflow: hidden;
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.28),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
}

.menu-row {
  width: 100%;
  min-height: 54px;
  border: none;
  background: transparent;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  transition: background-color 0.2s ease;
}

.menu-row:active {
  background: rgba(255, 248, 214, 0.05);
}

.menu-row + .menu-row {
  border-top: 1px solid rgba(212, 175, 55, 0.14);
}

.menu-left {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 10px;
}

.menu-icon {
  width: 18px;
  height: 18px;
  flex: 0 0 auto;
  color: #ffe7bf;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.menu-icon :deep(svg) {
  display: block;
  width: 100%;
  height: 100%;
}

.menu-icon :deep(path) {
  fill: currentColor;
}

.menu-text {
  color: #fff0c9;
  font-size: 15px;
  line-height: 1;
  font-weight: 600;
}

.menu-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.menu-extra {
  font-size: 13px;
  line-height: 1;
}

.menu-extra.success {
  color: #ffd87f;
}

.menu-extra.muted {
  color: rgba(255, 229, 186, 0.66);
}

.menu-arrow {
  color: rgba(255, 229, 186, 0.4);
  font-size: 14px;
}

.logout-wrap {
  padding: 22px 4px 0;
}

.logout-btn {
  width: 100%;
  height: 46px;
  border-radius: 24px;
  border: 1px solid rgba(212, 175, 55, 0.42);
  background: linear-gradient(180deg, #a51515 0%, #650000 100%);
  color: #fff0c9;
  font-size: 15px;
  line-height: 1;
  font-weight: 700;
  box-shadow:
    0 12px 22px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 248, 214, 0.18);
}

.language-popup {
  min-height: 430px;
  padding: 10px 0 28px;
  background:
    radial-gradient(circle at top, rgba(212, 175, 55, 0.14), transparent 26%),
    linear-gradient(180deg, #540000 0%, #280000 100%);
  border: 1px solid rgba(212, 175, 55, 0.34);
}

.language-popup-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  border-bottom: 1px solid rgba(212, 175, 55, 0.18);
}

.language-popup-title {
  font-size: 18px;
  font-weight: 700;
  color: #fff0c9;
}

.language-popup-close {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  border: none;
  background: transparent;
  font-size: 18px;
  color: rgba(255, 229, 186, 0.7);
  line-height: 1;
  cursor: pointer;
}

.language-list {
  padding: 14px;
}

.language-item {
  width: 100%;
  border: 1px solid rgba(212, 175, 55, 0.18);
  border-radius: 16px;
  padding: 14px;
  margin-bottom: 12px;
  background: linear-gradient(165deg, rgba(120, 0, 0, 0.84), rgba(58, 0, 0, 0.9));
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
  border-color: rgba(255, 248, 214, 0.4);
  background: linear-gradient(165deg, rgba(142, 38, 0, 0.94), rgba(88, 0, 0, 0.95));
  box-shadow: 0 10px 18px rgba(0, 0, 0, 0.24);
}

.language-code {
  font-size: 15px;
  color: #ffd98b;
}

.language-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.language-text .native {
  font-size: 14px;
  font-weight: 600;
  color: #fff0c9;
}

.language-text .english {
  font-size: 11px;
  color: rgba(255, 229, 186, 0.6);
}

.language-check {
  font-size: 20px;
  color: #ffd87f;
  text-align: right;
}

.language-tip {
  margin: 10px 16px 0;
  text-align: center;
  color: rgba(255, 229, 186, 0.56);
  font-size: 12px;
}

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

@media (max-width: 390px) {
  .user-name {
    font-size: 16px;
  }

  .balance-value {
    font-size: 16px;
  }
}
</style>

<route lang="json5">
{
  name: 'Profile'
}
</route>
