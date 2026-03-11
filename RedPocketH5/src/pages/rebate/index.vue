<script setup lang="ts">
import { useRouter } from 'vue-router'
import AppPageHeader from '@/components/AppPageHeader.vue'

const { t } = useI18n()
const router = useRouter()

function goBack() {
  router.back()
}
</script>

<template>
  <div class="rebate-page">
    <AppPageHeader :title="t('profilePage.promoRebate')" @back="goBack" />

    <section class="rebate-hero">
      <p class="rebate-eyebrow">
        Commission Center
      </p>
      <h2 class="rebate-title">
        {{ t('profilePage.promoRebate') }}
      </h2>
      <p class="rebate-text">
        Rebate records and commission settlement UI are being aligned to the home page style.
      </p>
    </section>

    <section class="rebate-card">
      <div class="rebate-card__line" />
      <p class="rebate-card__label">
        Coming Soon
      </p>
      <p class="rebate-card__copy">
        This page now uses the same red and gold visual system as the home page, and the detailed rebate modules can continue to be filled in here.
      </p>
    </section>
  </div>
</template>

<style scoped>
.rebate-page {
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
}

.rebate-hero,
.rebate-card {
  position: relative;
  overflow: hidden;
  border-radius: 18px;
  border: 1px solid rgba(212, 175, 55, 0.34);
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.3),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
}

.rebate-hero {
  margin-top: 10px;
  padding: 20px 16px;
}

.rebate-eyebrow,
.rebate-card__label {
  margin: 0;
  color: #ffd98b;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.rebate-title {
  margin: 8px 0 0;
  color: #fff0c9;
  font-size: 24px;
  font-weight: 800;
}

.rebate-text,
.rebate-card__copy {
  margin: 8px 0 0;
  color: rgba(255, 229, 186, 0.68);
  font-size: 13px;
  line-height: 1.6;
}

.rebate-card {
  margin-top: 12px;
  padding: 18px 16px;
}

.rebate-card__line {
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 18%, #ffd700 50%, #b8860b 82%, transparent 100%);
}
</style>

<route lang="json5">
{
  name: 'Rebate',
}
</route>
