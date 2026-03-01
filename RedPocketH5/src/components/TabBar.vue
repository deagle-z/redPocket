<script setup lang="ts">
import { rootRouteList } from '@/config/routes'

const route = useRoute()

const show = computed(() => {
  if (route.path === '/wallet')
    return true
  if (route.name && rootRouteList.includes(route.name)) {
    return true
  }
  return false
})

function isActive(name: string) {
  return route.name === name
}
</script>

<template>
  <van-tabbar v-if="show" route placeholder class="lp-tabbar">
    <van-tabbar-item replace :to="{ name: 'Home' }">
      <span class="tab-label" :class="{ 'tab-label--active': isActive('Home') }">
        {{ $t('tabbar.home') }}
      </span>
      <template #icon>
        <div class="icon-wrap" :class="{ 'icon-wrap--active': isActive('Home') }">
          <img src="@/assets/tabbar/home.svg" class="tab-icon">
        </div>
      </template>
    </van-tabbar-item>

    <van-tabbar-item replace :to="{ name: 'History' }">
      <span class="tab-label" :class="{ 'tab-label--active': isActive('History') }">
        {{ $t('tabbar.history') }}
      </span>
      <template #icon>
        <div class="icon-wrap" :class="{ 'icon-wrap--active': isActive('History') }">
          <img src="@/assets/tabbar/history.svg" class="tab-icon">
        </div>
      </template>
    </van-tabbar-item>

    <van-tabbar-item replace :to="{ name: 'SendPacket' }">
      <span class="tab-label tab-label--send">{{ $t('tabbar.sendPacket') }}</span>
      <template #icon>
        <div class="icon-wrap icon-wrap--special">
          <img src="@/assets/tabbar/dice.svg" class="tab-icon">
        </div>
      </template>
    </van-tabbar-item>

    <van-tabbar-item replace to="/wallet">
      <span class="tab-label" :class="{ 'tab-label--active': route.path === '/wallet' }">
        {{ $t('tabbar.wallet') }}
      </span>
      <template #icon>
        <div class="icon-wrap" :class="{ 'icon-wrap--active': route.path === '/wallet' }">
          <img src="@/assets/tabbar/wallet.svg" class="tab-icon">
        </div>
      </template>
    </van-tabbar-item>

    <van-tabbar-item replace :to="{ name: 'Profile' }">
      <span class="tab-label" :class="{ 'tab-label--active': isActive('Profile') }">
        {{ $t('tabbar.profile') }}
      </span>
      <template #icon>
        <div class="icon-wrap" :class="{ 'icon-wrap--active': isActive('Profile') }">
          <img src="@/assets/tabbar/profile.svg" class="tab-icon">
        </div>
      </template>
    </van-tabbar-item>
  </van-tabbar>
</template>

<style scoped>
:deep(.lp-tabbar.van-tabbar) {
  left: 10px;
  right: 10px;
  bottom: 10px;
  height: 68px;
  padding: 6px 8px calc(env(safe-area-inset-bottom) + 6px);
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(170, 196, 180, 0.46);
  border-radius: 18px;
  box-shadow:
    0 10px 28px rgba(15, 23, 42, 0.12),
    0 2px 10px rgba(60, 127, 91, 0.08);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

:deep(.lp-tabbar .van-tabbar-item--active) {
  color: inherit;
}

:deep(.lp-tabbar .van-tabbar-item) {
  color: inherit;
  min-height: 50px;
  border-radius: 12px;
  padding-bottom: 2px;
  transition:
    background-color 0.2s ease,
    transform 0.2s ease;
}

:deep(.lp-tabbar .van-tabbar-item:active) {
  transform: translateY(1px);
}

:deep(.lp-tabbar .van-tabbar-item__icon) {
  margin-bottom: 4px;
}

:deep(.lp-tabbar .van-tabbar-item:focus-visible) {
  outline: 2px solid #3dae6a;
  outline-offset: 1px;
}

.tab-label {
  font-size: 11px;
  font-weight: 500;
  color: #7e9c8c;
  line-height: 1.1;
  transition: color 0.2s ease;
}

.tab-label--active {
  color: #3dae6a;
  font-weight: 600;
}

.tab-label--send {
  color: #b07a2a;
}

.icon-wrap {
  width: 30px;
  height: 30px;
  border-radius: 10px;
  border: 1px solid #c7e0d0;
  background: linear-gradient(180deg, #f4fbf7 0%, #eaf5ee 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  transition:
    background 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.icon-wrap--active {
  background: #3dae6a;
  border-color: #2e9959;
  box-shadow: 0 6px 16px rgba(61, 174, 106, 0.32);
  transform: translateY(-1px);
}

.tab-icon {
  width: 17px;
  height: 17px;
  filter: brightness(0) saturate(100%) invert(57%) sepia(26%) saturate(700%) hue-rotate(103deg) brightness(98%);
  transition: filter 0.2s ease;
}

.icon-wrap--active .tab-icon {
  filter: brightness(0) invert(1);
}

.icon-wrap--special {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  background: linear-gradient(135deg, #ffa726 0%, #ff9800 50%, #fb8c00 100%);
  border: 1px solid #ffa726cc;
  box-shadow:
    0 0 12px rgba(255, 167, 38, 0.5),
    0 8px 14px rgba(255, 152, 0, 0.35),
    0 0 0 2px rgba(255, 167, 38, 0.22),
    inset 0 0 10px #ffffff40;
}

.icon-wrap--special .tab-icon {
  width: 20px;
  height: 20px;
  filter: brightness(0) invert(1);
}

@media (max-width: 390px) {
  :deep(.lp-tabbar.van-tabbar) {
    left: 6px;
    right: 6px;
    bottom: 6px;
    height: 64px;
    border-radius: 16px;
    padding: 5px 6px calc(env(safe-area-inset-bottom) + 4px);
  }

  .icon-wrap {
    width: 28px;
    height: 28px;
  }

  .icon-wrap--special {
    width: 34px;
    height: 34px;
  }

  .tab-label {
    font-size: 10px;
  }
}
</style>
