<script setup lang="ts">
import { showToast } from 'vant'

interface TelegramAuthUser {
  id: number
  first_name?: string
  last_name?: string
  username?: string
  photo_url?: string
  auth_date?: number
  hash?: string
}

const telegramUser = ref<TelegramAuthUser | null>(null)
const { t } = useI18n()

function handleTelegramCallback(user: TelegramAuthUser) {
  telegramUser.value = user
  showToast(t('bindTgPage.toastAuthSuccess', { user: user.first_name || user.username || user.id }))
}
</script>

<template>
  <div class="bind-tg-page">
    <h2>{{ t('bindTgPage.title') }}</h2>

    <TelegramLogin
      mode="callback"
      telegram-login="luckRedBoomPacket66Bot"
      size="large"
      request-access="write"
      @callback="handleTelegramCallback"
    />

    <pre v-if="telegramUser" class="user-json">{{ telegramUser }}</pre>
  </div>
</template>

<style scoped>
.bind-tg-page {
  padding: 16px;
}

.user-json {
  margin-top: 16px;
  padding: 12px;
  border-radius: 8px;
  background: #f5f7fb;
  color: #334155;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
