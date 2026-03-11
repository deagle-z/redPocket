<script setup lang="ts">
import { getCurrentTgUserInfo } from '@/api/user'
import defaultAvatar from '@/assets/svg/avatar.svg'
import { rootRouteList } from '@/config/routes'
import { isLogin } from '@/utils/auth'
import { formatCurrency } from '@/utils/currency'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const profileLoading = ref(false)
const profile = reactive({
  avatar: defaultAvatar,
  balance: 0,
})

const show = computed(() => {
  if (route.path === '/wallet')
    return true
  if (route.name && rootRouteList.includes(String(route.name)))
    return true
  return false
})

const balanceText = computed(() => {
  const raw = Number(profile.balance ?? 0)
  return formatCurrency(raw)
})

async function loadCurrentUserInfo() {
  if (!isLogin()) {
    profile.balance = 0
    profile.avatar = defaultAvatar
    return
  }
  if (profileLoading.value)
    return
  try {
    profileLoading.value = true
    const { data } = await getCurrentTgUserInfo()
    profile.balance = Number(data?.balance ?? 0)
    profile.avatar = data?.avatar || defaultAvatar
  }
  catch {
    profile.balance = 0
    profile.avatar = defaultAvatar
  }
  finally {
    profileLoading.value = false
  }
}

onMounted(() => {
  loadCurrentUserInfo()
})

watch(() => route.fullPath, () => {
  if (show.value)
    loadCurrentUserInfo()
})

function goRecharge() {
  router.push('/recharge')
}
</script>

<template>
  <header v-if="show" class="app-header">
    <div class="brand-wrap">
      <span class="brand-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" class="icon-svg">
          <circle cx="12" cy="12" r="12" fill="currentColor" />
          <path d="M17.7 6.9L15.8 17c-.1.5-.4.6-.8.4l-2.8-2.1-1.4 1.3c-.2.2-.3.3-.7.3l.2-2.9 5.3-4.8c.2-.2 0-.3-.3-.2l-6.5 4.1-2.8-.9c-.6-.2-.6-.6.1-.9l11-4.2c.5-.2 1 .1.8.9z" fill="#7d2b00" />
        </svg>
      </span>
      <div class="brand-text">
        <span class="brand-title">{{ t('appTopHeader.brandTitle') }}</span>
        <span class="brand-subtitle">{{ t('appTopHeader.brandSubtitle') }}</span>
      </div>
    </div>

    <button type="button" class="balance-wrap" @click="goRecharge">
      <span class="balance-label">{{ t('appTopHeader.balance') }}</span>
      <span class="balance-value">{{ balanceText }}</span>
      <img class="user-avatar" :src="profile.avatar || defaultAvatar" alt="avatar">
    </button>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background:
    radial-gradient(circle at 14% 24%, rgba(255, 215, 0, 0.2), transparent 28%),
    linear-gradient(160deg, #970000 0%, #760000 56%, #5b0000 100%);
  color: #fff3da;
  min-height: 54px;
  box-sizing: border-box;
  position: sticky;
  top: 0;
  z-index: 100;
  border-bottom: 1px solid rgba(212, 175, 55, 0.55);
  box-shadow:
    0 10px 18px rgba(0, 0, 0, 0.28),
    inset 0 -1px 0 rgba(255, 248, 214, 0.12);
}

.brand-wrap,
.balance-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.brand-wrap {
  gap: 8px;
}

.brand-text {
  display: inline-flex;
  flex-direction: column;
  min-width: 0;
}

.brand-title {
  font-size: clamp(13px, 3.2vw, 17px);
  line-height: 1;
  letter-spacing: 0.3px;
  white-space: nowrap;
  color: #ffedc3;
  font-weight: 700;
}

.brand-subtitle {
  margin-top: 3px;
  font-size: clamp(9px, 2.2vw, 11px);
  line-height: 1;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: rgba(255, 232, 186, 0.75);
}

.brand-icon,
.balance-icon {
  width: clamp(24px, 5vw, 30px);
  height: clamp(24px, 5vw, 30px);
  border-radius: 50%;
  color: #f4c750;
  background: linear-gradient(180deg, #ffe9af 0%, #d4af37 100%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.6),
    0 4px 8px rgba(0, 0, 0, 0.24);
}

.balance-icon {
  color: #58a6e8;
  border: 1px solid rgba(255, 255, 255, 0.55);
}

.icon-svg {
  width: 100%;
  height: 100%;
}

.balance-wrap {
  border: none;
  padding: 6px 8px 6px 10px;
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(255, 251, 238, 0.95) 0%, rgba(246, 227, 184, 0.95) 100%);
  border: 1px solid rgba(212, 175, 55, 0.65);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.85),
    0 4px 10px rgba(0, 0, 0, 0.2);
  white-space: nowrap;
  gap: 5px;
  cursor: pointer;
}

.balance-wrap:active {
  transform: translateY(1px);
}

.balance-label {
  font-size: clamp(11px, 2.6vw, 13px);
  font-weight: 600;
  color: #7f2b00;
}

.balance-value {
  font-size: clamp(13px, 3vw, 16px);
  font-weight: 700;
  line-height: 1;
  background: linear-gradient(to bottom, #cfb53b, #8a6e14, #d4af37);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.user-avatar {
  width: clamp(24px, 4.8vw, 30px);
  height: clamp(24px, 4.8vw, 30px);
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid rgba(143, 45, 0, 0.4);
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
}

@media (max-width: 390px) {
  .app-header {
    padding: 9px 10px;
    gap: 8px;
  }

  .brand-subtitle {
    display: none;
  }

  .balance-wrap {
    padding: 5px 7px 5px 8px;
  }
}
</style>
