<script setup lang="ts">
import { rootRouteList } from '@/config/routes'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

/**
 * Get page title
 * Located in src/locales/json
 */
const title = computed(() => {
  if (route.name) {
    return t(`navbar.${route.name}`)
  }

  return t('navbar.Undefined')
})

/**
 * Show the left arrow
 * If route name is in rootRouteList, hide left arrow
 */
const showLeftArrow = computed(() => {
  if (route.name && rootRouteList.includes(route.name)) {
    return false
  }

  return true
})

function onBack() {
  if (window.history.state.back) {
    history.back()
  }
  else {
    router.replace('/')
  }
}
</script>

<template>
  <VanNavBar
    :title="title"
    :fixed="true"
    :left-arrow="showLeftArrow"
    placeholder clickable
    class="thai-navbar"
    @click-left="onBack"
  />
</template>

<style scoped>
:deep(.thai-navbar.van-nav-bar) {
  background:
    radial-gradient(circle at 14% 50%, rgba(255, 215, 0, 0.14), transparent 32%),
    linear-gradient(160deg, #8c0000 0%, #6a0000 55%, #520000 100%);
  border-bottom: 1px solid rgba(212, 175, 55, 0.45);
  box-shadow:
    0 6px 16px rgba(0, 0, 0, 0.28),
    inset 0 -1px 0 rgba(255, 248, 214, 0.08);
}

/* Gold top accent stripe */
:deep(.thai-navbar.van-nav-bar)::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 20%, #ffd700 50%, #b8860b 80%, transparent 100%);
}

:deep(.thai-navbar .van-nav-bar__title) {
  color: #ffd98b;
  font-weight: 700;
  font-size: 16px;
  letter-spacing: 0.04em;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.35);
}

:deep(.thai-navbar .van-nav-bar__arrow),
:deep(.thai-navbar .van-nav-bar__left),
:deep(.thai-navbar .van-nav-bar__right) {
  color: #d4af37;
}

:deep(.thai-navbar .van-nav-bar__text) {
  color: #d4af37;
  font-weight: 600;
}
</style>
