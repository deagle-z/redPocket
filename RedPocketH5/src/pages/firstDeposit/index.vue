<script setup lang="ts">
import { useRouter } from 'vue-router'
import AppPageHeader from '@/components/AppPageHeader.vue'
import firstDepositBrazilImage from '@/assets/images/first_deposit_bonus_short_brazil.png'
import firstDepositEnglishImage from '@/assets/images/first_deposit_bonus_short_english.png'
import firstDepositIndonesiaImage from '@/assets/images/first_deposit_bonus_short_indonesia.png'
import firstDepositMexicoImage from '@/assets/images/first_deposit_bonus_short_mexico.png'
import { locale } from '@/utils/i18n'
import { safeBack } from '@/utils/navigation'

const { t } = useI18n()
const router = useRouter()

const firstDepositImageMap: Record<string, string> = {
  'pt-BR': firstDepositBrazilImage,
  'en-US': firstDepositEnglishImage,
  'id-ID': firstDepositIndonesiaImage,
  'es-MX': firstDepositMexicoImage,
}

const firstDepositImage = computed(() => firstDepositImageMap[locale.value] || firstDepositEnglishImage)

function goBack() {
  safeBack(router)
}

function goRecharge() {
  router.push('/recharge')
}
</script>

<template>
  <main class="first-deposit-page">
    <AppPageHeader :title="t('rechargePage.firstRechargeTitle')" @back="goBack" />

    <button type="button" class="first-deposit-entry" @click="goRecharge">
      <img
        class="first-deposit-image"
        :src="firstDepositImage"
        :alt="t('rechargePage.firstRechargeTitle')"
      >
    </button>
  </main>
</template>

<style scoped>
.first-deposit-page {
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

.first-deposit-entry {
  display: block;
  width: 100%;
  padding: 0;
  border: 0;
  background: transparent;
}

.first-deposit-entry:active {
  transform: translateY(1px);
}

.first-deposit-image {
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
  name: 'FirstDeposit',
}
</route>
