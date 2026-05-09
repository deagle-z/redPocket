<script setup lang="ts">
import { showToast } from 'vant'
import AppPageHeader from '@/components/AppPageHeader.vue'
import { getTenantServiceLinks } from '@/api/user'
import { safeBack } from '@/utils/navigation'

const router = useRouter()
const { t } = useI18n()

const loading = ref(false)
const defaultServiceUrl = 'https://t.me/Osanvnei'
const serviceLinks = ref({
  tgServiceUrl: defaultServiceUrl,
  wsServiceUrl: defaultServiceUrl,
})

const serviceItems = computed(() => [
  {
    key: 'tg',
    title: 'Telegram',
    desc: serviceLinks.value.tgServiceUrl,
    url: serviceLinks.value.tgServiceUrl,
    icon: 'chat-o',
  },
  {
    key: 'ws',
    title: 'WhatsApp',
    desc: serviceLinks.value.wsServiceUrl,
    url: serviceLinks.value.wsServiceUrl,
    icon: 'service-o',
  },
])

function normalizeServiceLinks(data?: { tgServiceUrl?: string | null, wsServiceUrl?: string | null } | null) {
  return {
    tgServiceUrl: String(data?.tgServiceUrl || '').trim() || defaultServiceUrl,
    wsServiceUrl: String(data?.wsServiceUrl || '').trim() || defaultServiceUrl,
  }
}

function normalizeExternalUrl(url?: string | null) {
  const value = String(url || '').trim()
  if (!value)
    return ''
  if (/^[a-z][a-z\d+\-.]*:/i.test(value))
    return value
  return `https://${value}`
}

function openExternal(url?: string | null) {
  const target = normalizeExternalUrl(url)
  if (!target) {
    showToast('暂未配置客服链接')
    return
  }

  const opened = window.open(target, '_blank')
  if (opened) {
    opened.opener = null
  }
  else {
    window.location.href = target
  }
}

async function loadServiceLinks() {
  loading.value = true
  try {
    const { data } = await getTenantServiceLinks()
    serviceLinks.value = normalizeServiceLinks(data)
  }
  catch {
    serviceLinks.value = normalizeServiceLinks()
  }
  finally {
    loading.value = false
  }
}

function goBack() {
  safeBack(router)
}

onMounted(() => {
  loadServiceLinks()
})
</script>

<template>
  <div class="cs-page">
    <AppPageHeader :title="t('profilePage.serviceCs')" @back="goBack" />

    <section class="cs-panel">
      <van-loading v-if="loading" class="cs-loading" size="22" color="#d4af37" />

      <div v-else class="cs-list">
        <button
          v-for="item in serviceItems"
          :key="item.key"
          type="button"
          class="cs-item"
          :class="{ disabled: !item.url }"
          @click="openExternal(item.url)"
        >
          <span class="cs-icon">
            <van-icon :name="item.icon" />
          </span>
          <span class="cs-info">
            <strong>{{ item.title }}</strong>
            <em>{{ item.desc }}</em>
          </span>
          <van-icon name="arrow" class="cs-arrow" />
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.cs-page {
  min-height: 100vh;
  padding: 10px 12px calc(90px + env(safe-area-inset-bottom));
  background:
    radial-gradient(circle at 18% 10%, rgba(212, 175, 55, 0.16), transparent 28%),
    linear-gradient(180deg, #3e0000 0%, #240000 62%, #150000 100%);
}

.cs-panel {
  margin-top: 12px;
  padding: 18px 14px;
  border: 1px solid rgba(212, 175, 55, 0.18);
  border-radius: 18px;
  background: rgba(255, 248, 214, 0.06);
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.22);
}

.cs-loading {
  display: flex;
  justify-content: center;
  padding: 42px 0;
}

.cs-list {
  display: grid;
  gap: 12px;
}

.cs-item {
  min-height: 76px;
  width: 100%;
  display: grid;
  grid-template-columns: 46px minmax(0, 1fr) 18px;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid rgba(212, 175, 55, 0.2);
  border-radius: 14px;
  background: rgba(36, 0, 0, 0.42);
  color: #ffe6b3;
  text-align: left;
}

.cs-item:active {
  transform: translateY(1px);
}

.cs-item.disabled {
  opacity: 0.55;
}

.cs-icon {
  width: 46px;
  height: 46px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 14px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #540000;
  font-size: 23px;
}

.cs-info {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.cs-info strong {
  font-size: 16px;
  line-height: 20px;
}

.cs-info em {
  overflow: hidden;
  color: rgba(255, 229, 186, 0.72);
  font-size: 12px;
  font-style: normal;
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cs-arrow {
  color: rgba(255, 229, 186, 0.66);
  font-size: 16px;
}
</style>
