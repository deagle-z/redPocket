<script setup lang="ts">
import AppPageHeader from '@/components/AppPageHeader.vue'
import OrderRecordList from '@/components/OrderRecordList.vue'
import { getWithdrawOrderHistory } from '@/api/user'

const { t } = useI18n()
const router = useRouter()

function goBack() {
  router.back()
}
</script>

<template>
  <div class="order-record-page">
    <AppPageHeader :title="t('orderRecordPage.withdrawTitle')" @back="goBack" />
    <OrderRecordList kind="withdraw" :loader="getWithdrawOrderHistory" />
  </div>
</template>

<style scoped>
.order-record-page {
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
  padding: 0 12px calc(90px + env(safe-area-inset-bottom));
  color: #f8e8c6;
}
</style>

<route lang="json5">
{
  name: 'WithdrawRecord'
}
</route>
