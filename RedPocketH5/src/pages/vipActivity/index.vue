<script setup lang="ts">
import { useRouter } from 'vue-router'
import AppPageHeader from '@/components/AppPageHeader.vue'
import VipProgressPopup from '@/components/VipProgressPopup.vue'
import vipBrazilImage from '@/assets/images/long_image_brazil/VIP_short_brazil.png'
import vipEnglishImage from '@/assets/images/long_image_english/VIP_short_english.png'
import vipIndonesiaImage from '@/assets/images/long_image_indonesia/VIP_short_indonesia.png'
import vipMexicoImage from '@/assets/images/long_image_mexico/VIP_short_mexico.png'
import { locale } from '@/utils/i18n'
import { safeBack } from '@/utils/navigation'

const { t } = useI18n()
const router = useRouter()
const showVipPopup = ref(false)

const vipImageMap: Record<string, string> = {
  'pt-BR': vipBrazilImage,
  'en-US': vipEnglishImage,
  'id-ID': vipIndonesiaImage,
  'es-MX': vipMexicoImage,
}

const vipImage = computed(() => vipImageMap[locale.value] || vipEnglishImage)

function goBack() {
  safeBack(router)
}

function openVipPopup() {
  showVipPopup.value = true
}
</script>

<template>
  <main class="activity-image-page">
    <AppPageHeader :title="t('activityPage.vip')" @back="goBack" />

    <button type="button" class="activity-image-entry" @click="openVipPopup">
      <img
        class="activity-image"
        :src="vipImage"
        :alt="t('activityPage.vip')"
      >
    </button>

    <VipProgressPopup v-model:show="showVipPopup" />
  </main>
</template>

<style scoped>
.activity-image-page {
  min-height: 100vh;
  background-image:
    radial-gradient(circle at 18% 12%, rgba(212, 175, 55, 0.18), transparent 28%),
    radial-gradient(circle at 82% 84%, rgba(255, 215, 0, 0.12), transparent 24%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.04) 18px,
      rgba(212, 175, 55, 0.04) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #240000 60%, #150000 100%);
  padding: 0 12px calc(18px + env(safe-area-inset-bottom));
}

.activity-image-entry {
  display: block;
  width: 100%;
  padding: 0;
  border: none;
  background: transparent;
  text-align: left;
}

.activity-image {
  display: block;
  width: 100%;
  max-width: 100%;
  height: auto;
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.32),
    0 0 0 1px rgba(255, 248, 214, 0.08);
}
</style>

<route lang="json5">
{
  name: 'VipActivity',
}
</route>
