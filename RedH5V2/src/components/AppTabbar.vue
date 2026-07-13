<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const active = computed({
  get() {
    return route.path
  },
  set(path: string) {
    router.push(path)
  },
})
</script>

<template>
  <van-tabbar v-model="active" route safe-area-inset-bottom>
    <van-tabbar-item to="/" icon="wap-home-o">{{ t('nav.home') }}</van-tabbar-item>
    <van-tabbar-item to="/profile" icon="user-o">{{ t('nav.account') }}</van-tabbar-item>
  </van-tabbar>
</template>
