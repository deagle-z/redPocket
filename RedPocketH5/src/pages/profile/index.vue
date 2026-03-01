<script setup lang="ts">
import { getCurrentTgUserInfo, tgLogout } from '@/api/user'
import { showToast } from 'vant'
import { clearToken, isLogin } from '@/utils/auth'
import { locale } from '@/utils/i18n'
import avatarPlaceholderIcon from '@/assets/my/question-circle.svg'
import bankCardIcon from '@/assets/my/bank card.svg'
import chartBarAltIcon from '@/assets/my/chart-bar-alt.svg'
import customerServiceIcon from '@/assets/my/customer-service-fill.svg'
import downIcon from '@/assets/my/down2.svg'
import emailIcon from '@/assets/my/email.svg'
import gamesIcon from '@/assets/my/games-2.svg'
import passwordIcon from '@/assets/my/Password.svg'
import shareIcon from '@/assets/my/share.svg'
import teamIcon from '@/assets/my/team.svg'
import telegramIcon from '@/assets/my/telegram.svg'
import upIcon from '@/assets/my/up2.svg'
import walletIcon from '@/assets/my/wallet.svg'

type ExtraTone = 'success' | 'muted'

interface MenuItem {
  key: string
  label: string
  icon: string
  extra?: string
  tone?: ExtraTone
}

const accountMenus: MenuItem[] = [
  { key: 'wallet', label: '我的钱包', icon: walletIcon },
  { key: 'recharge', label: '充值', icon: downIcon },
  { key: 'withdraw', label: '提现', icon: upIcon },
  { key: 'withdraw-account', label: '提现账户', icon: bankCardIcon },
  { key: 'lucky-reward', label: '幸运转盘奖励', icon: gamesIcon },
]

const promoMenus: MenuItem[] = [
  { key: 'team', label: '我的团队', icon: teamIcon },
  { key: 'invite', label: '邀请好友', icon: shareIcon, extra: '奖励', tone: 'success' },
  { key: 'rebate', label: '佣金明细', icon: chartBarAltIcon },
]

const profileLoading = ref(false)
const profile = reactive({
  avatar: '',
  username: '',
  uid: '',
  tgId: 0,
  email: '',
  balance: 0,
  rebateAmount: 0,
})

const otherMenus = computed<MenuItem[]>(() => [
  { key: 'language', label: '语言（Language）', icon: shareIcon },
  { key: 'rules', label: '游戏规则', icon: gamesIcon },
  { key: 'bind-tg', label: '绑定Telegram', icon: telegramIcon, extra: formatMaskedNumber(profile.tgId), tone: 'muted' },
  { key: 'bind-email', label: '绑定邮箱', icon: emailIcon, extra: formatMaskedEmail(profile.email), tone: 'muted' },
  { key: 'change-password', label: '修改密码', icon: passwordIcon },
  { key: 'questions', label: '常见问题', icon: avatarPlaceholderIcon },
  { key: 'cs', label: '客服中心', icon: customerServiceIcon },
])

const displayName = computed(() => profile.username || '--')
const displayUid = computed(() => profile.uid || '--')
const displayBalance = computed(() => Number(profile.balance || 0).toFixed(2))
const displayRebateAmount = computed(() => Number(profile.rebateAmount || 0).toFixed(2))

function formatMaskedNumber(value: number) {
  const text = String(value || '').trim()
  if (!text)
    return '未绑定'
  if (text.length <= 4)
    return `${text.slice(0, 1)}***${text.slice(-1)}`
  return `${text.slice(0, 1)}***${text.slice(-3)}`
}

function formatMaskedEmail(email: string) {
  const text = String(email || '').trim()
  if (!text)
    return '未绑定'
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
  }
  catch {
    showToast('加载用户信息失败')
  }
  finally {
    profileLoading.value = false
  }
}

onMounted(() => {
  loadCurrentTgUserInfo()
})

const showLogoutDialog = ref(false)
const logoutLoading = ref(false)
const showLangPopup = ref(false)
const { t } = useI18n()
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
    showToast('已退出登录')
    router.replace({ name: 'Login' as keyof RouteMap })
  }
}
</script>

<template>
  <div class="profile-page">
    <section class="profile-card">
      <div class="profile-top">
        <div class="avatar-box">
          <img :src="profile.avatar || avatarPlaceholderIcon" alt="" class="avatar-icon">
        </div>
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
        <div class="balance-item">
          <strong class="balance-value">{{ displayBalance }}</strong>
          <span class="balance-label">余额(₱)</span>
        </div>
        <div class="balance-divider" />
        <div class="balance-item">
          <strong class="balance-value">{{ displayRebateAmount }}</strong>
          <span class="balance-label">佣金(₱)</span>
        </div>
      </div>
    </section>

    <p class="section-label">
      账户管理
    </p>
    <section class="menu-card">
      <button v-for="item in accountMenus" :key="item.key" type="button" class="menu-row" @click="onMenuClick(item)">
        <div class="menu-left">
          <img :src="item.icon" alt="" class="menu-icon">
          <span class="menu-text">{{ item.label }}</span>
        </div>
        <div class="menu-right">
          <span v-if="item.extra" class="menu-extra" :class="item.tone">{{ item.extra }}</span>
          <van-icon name="arrow" class="menu-arrow" />
        </div>
      </button>
    </section>

    <p class="section-label">
      推广中心
    </p>
    <section class="menu-card">
      <button v-for="item in promoMenus" :key="item.key" type="button" class="menu-row" @click="onMenuClick(item)">
        <div class="menu-left">
          <img :src="item.icon" alt="" class="menu-icon">
          <span class="menu-text">{{ item.label }}</span>
        </div>
        <div class="menu-right">
          <span v-if="item.extra" class="menu-extra" :class="item.tone">{{ item.extra }}</span>
          <van-icon name="arrow" class="menu-arrow" />
        </div>
      </button>
    </section>

    <p class="section-label">
      其他服务
    </p>
    <section class="menu-card">
      <button v-for="item in otherMenus" :key="item.key" type="button" class="menu-row" @click="onMenuClick(item)">
        <div class="menu-left">
          <img :src="item.icon" alt="" class="menu-icon">
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
        退出登录
      </button>
    </div>

    <AppConfirmDialog
      v-model:show="showLogoutDialog"
      title="确认退出"
      cancel-text="取消"
      confirm-text="确认"
      @confirm="handleConfirmLogout"
    >
      确定要退出登录吗？
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
  </div>
</template>

<style scoped>
.profile-page {
  min-height: 100vh;
  background: #f2f3f7;
  padding-bottom: calc(16px + env(safe-area-inset-bottom));
}

.profile-card {
  border-radius: 0 0 16px 16px;
  background: #fff;
  padding: 16px 16px 20px;
}

.profile-top {
  display: flex;
  align-items: center;
  gap: 12px;
}

.avatar-box {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: #e8f0fe;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-icon {
  width: 26px;
  height: 26px;
  opacity: 0.5;
}

.profile-meta {
  flex: 1;
  min-width: 0;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-name {
  margin: 0;
  color: #1a1a2e;
  font-size: 30px;
  line-height: 1.1;
  font-weight: 700;
}

.vip-tag {
  border: 1px solid #f4bf2c;
  border-radius: 10px;
  color: #c08a00;
  font-size: 11px;
  line-height: 1;
  padding: 4px 6px;
  font-weight: 500;
}

.user-id {
  margin: 4px 0 0;
  color: #9ca3af;
  font-size: 13px;
  line-height: 1;
}

.balance-row {
  margin-top: 16px;
  border-top: 1px solid #f0f0f5;
  display: flex;
  align-items: center;
  height: 60px;
}

.balance-item {
  flex: 1;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.balance-value {
  color: #2dc84d;
  font-size: 30px;
  line-height: 1;
  font-weight: 700;
}

.balance-label {
  color: #9ca3af;
  font-size: 12px;
  line-height: 1;
}

.balance-divider {
  width: 1px;
  height: 30px;
  background: #f0f0f5;
}

.section-label {
  margin: 14px 0 6px;
  padding: 0 16px;
  color: #9ca3af;
  font-size: 13px;
  line-height: 1;
  font-weight: 500;
}

.menu-card {
  margin: 0 8px;
  border-radius: 12px;
  background: #fff;
  overflow: hidden;
}

.menu-row {
  width: 100%;
  height: 48px;
  border: none;
  background: #fff;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.menu-row + .menu-row {
  border-top: 1px solid #f8f8fb;
}

.menu-left {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 12px;
}

.menu-icon {
  width: 20px;
  height: 20px;
  flex: 0 0 auto;
}

.menu-text {
  color: #1a1a2e;
  font-size: 15px;
  line-height: 1;
  font-weight: 400;
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
  color: #2dc84d;
}

.menu-extra.muted {
  color: #9ca3af;
}

.menu-arrow {
  color: #d1d5db;
  font-size: 14px;
}

.logout-wrap {
  padding: 20px 16px 24px;
}

.logout-btn {
  width: 100%;
  height: 48px;
  border-radius: 24px;
  border: 1px solid #ff4d4f;
  background: #fff;
  color: #ff4d4f;
  font-size: 15px;
  line-height: 1;
  font-weight: 500;
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
  .user-name {
    font-size: 26px;
  }

  .balance-value {
    font-size: 26px;
  }
}
</style>

<route lang="json5">
{
  name: 'Profile'
}
</route>
