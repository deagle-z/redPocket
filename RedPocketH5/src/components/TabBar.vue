<script setup lang="ts">
import { rootRouteList } from '@/config/routes'

const route = useRoute()
const tabbarStyle = {
  '--van-tabbar-z-index': '99',
  '--van-tabbar-background': '#650400',
  '--van-tabbar-item-active-background': 'transparent',
  background: 'linear-gradient(170deg, rgba(125, 0, 0, 0.98) 0%, rgba(78, 0, 0, 0.98) 58%, rgba(46, 0, 0, 0.98) 100%)',
}

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
  <van-tabbar v-if="show" route placeholder class="lp-tabbar" :style="tabbarStyle">
    <van-tabbar-item replace class="tab-item" :to="{ name: 'Home' }">
      <span class="tab-label" :class="{ 'tab-label--active': isActive('Home') }">
        {{ $t('tabbar.home') }}
      </span>
      <template #icon>
        <div class="icon-wrap" :class="{ 'icon-wrap--active': isActive('Home') }">
          <img src="@/assets/tabbar/home.svg" class="tab-icon">
        </div>
      </template>
    </van-tabbar-item>

    <van-tabbar-item replace class="tab-item" :to="{ name: 'History' }">
      <span class="tab-label" :class="{ 'tab-label--active': isActive('History') }">
        {{ $t('tabbar.history') }}
      </span>
      <template #icon>
        <div class="icon-wrap" :class="{ 'icon-wrap--active': isActive('History') }">
          <img src="@/assets/tabbar/history.svg" class="tab-icon">
        </div>
      </template>
    </van-tabbar-item>

    <van-tabbar-item replace class="tab-item" :to="{ name: 'Team' }">
      <span class="tab-label" :class="{ 'tab-label--active': isActive('Team') }">
        {{ $t('tabbar.team') }}
      </span>
      <template #icon>
        <div class="icon-wrap" :class="{ 'icon-wrap--active': isActive('Team') }">
          <img src="@/assets/tabbar/team.svg" class="tab-icon">
        </div>
      </template>
    </van-tabbar-item>

    <van-tabbar-item replace class="tab-item" to="/wallet">
      <span class="tab-label" :class="{ 'tab-label--active': route.path === '/wallet' }">
        {{ $t('tabbar.wallet') }}
      </span>
      <template #icon>
        <div class="icon-wrap" :class="{ 'icon-wrap--active': route.path === '/wallet' }">
          <img src="@/assets/tabbar/wallet.svg" class="tab-icon">
        </div>
      </template>
    </van-tabbar-item>

    <van-tabbar-item replace class="tab-item" :to="{ name: 'Profile' }">
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
  --van-tabbar-z-index: 99 !important;
  --van-tabbar-background: #650400 !important;
  --van-border-color: transparent !important;
  z-index: 99 !important;
  left: 10px;
  right: 10px;
  bottom: 10px;
  height: 70px;
  padding: 7px 10px calc(env(safe-area-inset-bottom) + 7px);
  overflow: visible;
  isolation: isolate;
  background:
    radial-gradient(rgba(212, 175, 55, 1) 1px, transparent 1px),
    linear-gradient(170deg, rgba(125, 0, 0, 0.98) 0%, rgba(78, 0, 0, 0.98) 58%, rgba(46, 0, 0, 0.98) 100%) !important;
  background-size: 18px 18px, 100% 100%;
  border: 1px solid rgba(212, 175, 55, 0.44);
  border-radius: 22px;
  box-shadow:
    0 14px 32px rgba(0, 0, 0, 0.42),
    inset 0 0 0 1px rgba(255, 248, 214, 0.12),
    0 0 0 1px rgba(212, 175, 55, 0.24);
}

:deep(.lp-tabbar.van-tabbar),
:deep(.lp-tabbar.van-tabbar--fixed),
:deep(.lp-tabbar.van-safe-area-bottom) {
  z-index: 99 !important;
  background-color: #650400 !important;
}

:deep(.lp-tabbar.van-tabbar)::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  border-radius: 22px 22px 0 0;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
  z-index: 1;
  pointer-events: none;
}

:deep(.lp-tabbar .van-tabbar-item--active) {
  color: inherit;
  background-color: transparent !important;
}

:deep(.lp-tabbar .van-tabbar-item) {
  position: relative;
  z-index: 2;
  color: inherit;
  min-height: 52px;
  border-radius: 14px;
  padding: 4px 0 2px;
  transition:
    background-color 0.2s ease,
    transform 0.2s ease;
}

:deep(.lp-tabbar .van-tabbar-item--active) {
  background: linear-gradient(180deg, rgba(255, 248, 214, 0.08) 0%, rgba(255, 248, 214, 0) 100%);
}

:deep(.lp-tabbar .van-tabbar-item:active) {
  transform: translateY(1px);
}

:deep(.lp-tabbar .van-tabbar-item__icon) {
  margin-bottom: 4px;
}

:deep(.lp-tabbar .van-tabbar-item__text) {
  overflow: visible;
}

:deep(.lp-tabbar .van-tabbar-item:focus-visible) {
  outline: 2px solid #d4af37;
  outline-offset: 1px;
}

.tab-label {
  font-size: 11px;
  font-weight: 600;
  color: rgba(255, 229, 186, 0.56);
  letter-spacing: 0.04em;
  line-height: 1.1;
  transition:
    color 0.2s ease,
    transform 0.2s ease;
}

.tab-label--active {
  color: #ffe09a;
  font-weight: 700;
  text-shadow: 0 0 8px rgba(212, 175, 55, 0.35);
}

.icon-wrap {
  position: relative;
  width: 34px;
  height: 34px;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid rgba(255, 248, 214, 0.16);
  background: linear-gradient(180deg, rgba(118, 22, 10, 0.96) 0%, rgba(88, 0, 0, 0.96) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    0 6px 12px rgba(0, 0, 0, 0.26);
  transition:
    background 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.icon-wrap::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.18) 0%, rgba(255, 255, 255, 0) 55%);
  pointer-events: none;
}

.icon-wrap--active {
  background: linear-gradient(160deg, #b91b1b 0%, #7b0000 100%);
  border-color: rgba(212, 175, 55, 0.55);
  box-shadow:
    0 8px 18px rgba(122, 0, 0, 0.42),
    0 0 12px rgba(212, 175, 55, 0.18),
    inset 0 1px 0 rgba(255, 240, 201, 0.28);
}

.tab-icon {
  width: 18px;
  height: 18px;
  filter: brightness(0) saturate(100%) invert(85%) sepia(20%) saturate(300%) hue-rotate(340deg) brightness(80%);
  transition: filter 0.2s ease;
}

.icon-wrap--active .tab-icon {
  filter: brightness(0) invert(1);
}

@media (max-width: 390px) {
  :deep(.lp-tabbar.van-tabbar) {
    left: 6px;
    right: 6px;
    bottom: 6px;
    height: 66px;
    border-radius: 18px;
    padding: 8px 6px calc(env(safe-area-inset-bottom) + 6px);
  }

  .icon-wrap {
    width: 32px;
    height: 32px;
  }

  .tab-label {
    font-size: 10px;
  }
}
</style>
