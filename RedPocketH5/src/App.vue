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
  <van-config-provider>
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
}
</style>
