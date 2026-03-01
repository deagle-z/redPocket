<script setup lang="ts">
import { getCurrentTgUserInfo } from '@/api/user'
import defaultAvatar from '@/assets/svg/avatar.svg'
import { rootRouteList } from '@/config/routes'
import { isLogin } from '@/utils/auth'

const route = useRoute()
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
  if (Number.isNaN(raw))
    return '₱0.00'
  return `₱${raw.toFixed(2)}`
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
</script>

<template>
  <header v-if="show" class="app-header">
    <div class="brand-wrap">
      <span class="brand-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" class="icon-svg">
          <circle cx="12" cy="12" r="12" fill="currentColor" />
          <path d="M17.7 6.9L15.8 17c-.1.5-.4.6-.8.4l-2.8-2.1-1.4 1.3c-.2.2-.3.3-.7.3l.2-2.9 5.3-4.8c.2-.2 0-.3-.3-.2l-6.5 4.1-2.8-.9c-.6-.2-.6-.6.1-.9l11-4.2c.5-.2 1 .1.8.9z" fill="#4caf65" />
        </svg>
      </span>
      <span class="brand-title">红包雷游戏</span>
    </div>

    <div class="balance-wrap">
      <span class="balance-label">余额：</span>
      <span class="balance-value">{{ balanceText }}</span>
      <img class="user-avatar" :src="profile.avatar || defaultAvatar" alt="avatar">
    </div>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  padding: 2vw 3vw;
  background-color: var(--color-primary-btn);
  color: #fff;
  min-height: 8vw;
  box-sizing: border-box;
  position: sticky;
  top: 0;
  z-index: 100;
  box-shadow: 0 0.266667vw 1.066667vw #0000001a;
}

.brand-wrap,
.balance-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.brand-title {
  font-size: clamp(16px, 3.4vw, 22px);
  font-weight: 700;
  line-height: 1;
  letter-spacing: 0.2px;
  white-space: nowrap;
}

.brand-icon,
.balance-icon {
  width: clamp(20px, 4.2vw, 26px);
  height: clamp(20px, 4.2vw, 26px);
  border-radius: 50%;
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
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
  white-space: nowrap;
}

.balance-label {
  font-size: clamp(12px, 2.9vw, 17px);
  font-weight: 600;
}

.balance-value {
  font-size: clamp(14px, 3.3vw, 20px);
  font-weight: 700;
}

.user-avatar {
  width: clamp(22px, 4.4vw, 28px);
  height: clamp(22px, 4.4vw, 28px);
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid rgba(255, 255, 255, 0.55);
  background: rgba(255, 255, 255, 0.85);
}

@media (max-width: 390px) {
  .app-header {
    padding: 10px 12px;
  }
}
</style>
