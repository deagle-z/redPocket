<script setup lang="ts">
import { useRouter } from 'vue-router'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { safeBack } from '@/utils/navigation'

const { t } = useI18n()
const router = useRouter()

function goBack() {
  safeBack(router)
}
</script>

<template>
  <main class="vip-activity-page">
    <AppPageHeader :title="t('profilePage.vipTitle')" @back="goBack" />

    <section class="vip-hero-card">
      <p class="vip-kicker">
        VIP Activity
      </p>
      <h2>{{ t('profilePage.vipTitle') }}</h2>
      <p>{{ t('profilePage.vipProgressHint', { level: t('profilePage.vipNextLevel'), amount: '--' }) }}</p>
    </section>

    <section class="vip-placeholder-card">
      <span class="vip-placeholder-icon">VIP</span>
      <h3>{{ t('profilePage.vipUpgradeReward') }}</h3>
      <p>VIP activity content is reserved here.</p>
    </section>
  </main>
</template>

<style scoped>
.vip-activity-page {
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
  color: #fff0c9;
}

.vip-hero-card,
.vip-placeholder-card {
  position: relative;
  overflow: hidden;
  border-radius: 18px;
  border: 1px solid rgba(212, 175, 55, 0.34);
  background: linear-gradient(165deg, rgba(118, 0, 0, 0.95), rgba(54, 0, 0, 0.96));
  box-shadow:
    0 12px 24px rgba(0, 0, 0, 0.3),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
}

.vip-hero-card {
  margin-top: 10px;
  padding: 20px 16px;
}

.vip-kicker {
  margin: 0;
  color: #ffd98b;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.vip-hero-card h2,
.vip-placeholder-card h3 {
  margin: 8px 0 0;
  color: #fff0c9;
  font-size: 24px;
  font-weight: 800;
}

.vip-hero-card p,
.vip-placeholder-card p {
  margin: 8px 0 0;
  color: rgba(255, 229, 186, 0.68);
  font-size: 13px;
  line-height: 1.6;
}

.vip-placeholder-card {
  margin-top: 12px;
  padding: 22px 16px;
  text-align: center;
}

.vip-placeholder-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  border: 1px solid rgba(255, 248, 214, 0.46);
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 17px;
  font-weight: 900;
  box-shadow: 0 10px 22px rgba(0, 0, 0, 0.24);
}
</style>

<route lang="json5">
{
  name: 'VipActivity',
}
</route>
