<script setup lang="ts">
import { useLocalStorage } from '@vueuse/core'
import { useRouteCacheStore } from '@/stores'
import { STORAGE_TOKEN_KEY } from '@/stores/mutation-type'
import wsClient, { connectWebSocket } from '@/plugins/websocket'
import AppTopHeader from '@/components/AppTopHeader.vue'
import TabBar from '@/components/TabBar.vue'

const routeCacheStore = useRouteCacheStore()
const accessToken = useLocalStorage<string | null>(STORAGE_TOKEN_KEY, '')
const wsInitialized = ref(false)

const keepAliveRouteNames = computed(() => {
  return routeCacheStore.routeCaches
})

// Thai Red & Gold — Vant component theme overrides
const vantThemeVars = {
  colorPrimary: '#c0392b',
  colorSuccess: '#c0392b',
  buttonPrimaryBackground: 'linear-gradient(135deg, #a93226, #c0392b)',
  buttonPrimaryBorderColor: '#a93226',
  buttonPrimaryColor: '#ffffff',
  tabbarItemActiveColor: '#c0392b',
  tabbarItemTextColor: '#4a3030',
  navBarIconColor: '#c0392b',
  navBarTitleTextColor: '#2c0a07',
  switchOnBackground: '#c0392b',
  sliderActiveBackground: '#c0392b',
  checkboxCheckedIconColor: '#c0392b',
  radioCheckedIconColor: '#c0392b',
  fieldLabelColor: '#5c2d1e',
}

function initWebSocket() {
  if (wsInitialized.value)
    return
  connectWebSocket()
  wsInitialized.value = true
}

function closeWebSocket() {
  if (!wsInitialized.value)
    return
  wsClient.close()
  wsInitialized.value = false
}

onMounted(() => {
  if (accessToken.value)
    initWebSocket()
})

watch(accessToken, (token, oldToken) => {
  const hasToken = !!token
  const hadToken = !!oldToken
  if (hasToken && !hadToken) {
    initWebSocket()
    return
  }
  if (!hasToken && hadToken)
    closeWebSocket()
})
</script>

<template>
  <van-config-provider :theme-vars="vantThemeVars">
    <router-view v-slot="{ Component }">
      <section class="app-wrapper">
        <AppTopHeader />
        <keep-alive :include="keepAliveRouteNames">
          <component :is="Component" />
        </keep-alive>
        <TabBar />
      </section>
    </router-view>
  </van-config-provider>
</template>

<style scoped>
.app-wrapper {
  width: 100%;
  position: relative;
  min-height: 100vh;
  background: linear-gradient(160deg, #fff8f5 0%, #faf0eb 100%);
  /* Subtle Thai diagonal gold stripe watermark */
  background-image:
    linear-gradient(160deg, #fff8f5 0%, #faf0eb 100%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.04) 18px,
      rgba(212, 175, 55, 0.04) 20px
    );
  background-blend-mode: normal;
}
</style>
