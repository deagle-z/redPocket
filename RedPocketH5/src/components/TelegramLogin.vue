<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

type TelegramLoginMode = 'callback' | 'redirect'
type TelegramRequestAccess = 'read' | 'write'
type TelegramWidgetSize = 'small' | 'medium' | 'large'

interface TelegramAuthUser {
  id: number
  first_name?: string
  last_name?: string
  username?: string
  photo_url?: string
  auth_date?: number
  hash?: string
}

interface Props {
  mode: TelegramLoginMode
  telegramLogin: string
  redirectUrl?: string
  requestAccess?: TelegramRequestAccess
  size?: TelegramWidgetSize
  userpic?: boolean
  radius?: string
}

const props = withDefaults(defineProps<Props>(), {
  redirectUrl: '',
  requestAccess: 'write',
  size: 'large',
  userpic: true,
  radius: '',
})

const emit = defineEmits<{
  (e: 'callback', user: TelegramAuthUser): void
}>()

const telegramRootRef = ref<HTMLElement | null>(null)
const callbackKey = `__telegramAuth_${Math.random().toString(36).slice(2)}`

function onTelegramAuth(user: TelegramAuthUser) {
  emit('callback', user)
}

function appendWidgetScript() {
  if (!telegramRootRef.value)
    return

  telegramRootRef.value.innerHTML = ''

  const script = document.createElement('script')
  script.async = true
  script.src = 'https://telegram.org/js/telegram-widget.js?23'
  script.setAttribute('data-telegram-login', props.telegramLogin)
  script.setAttribute('data-size', props.size)
  script.setAttribute('data-userpic', String(props.userpic))
  script.setAttribute('data-request-access', props.requestAccess)

  if (props.radius)
    script.setAttribute('data-radius', props.radius)

  if (props.mode === 'callback') {
    ;(window as unknown as Record<string, unknown>)[callbackKey] = onTelegramAuth
    script.setAttribute('data-onauth', `window.${callbackKey}(user)`)
  }
  else if (props.redirectUrl) {
    script.setAttribute('data-auth-url', props.redirectUrl)
  }

  telegramRootRef.value.appendChild(script)
}

onMounted(() => {
  appendWidgetScript()
})

onBeforeUnmount(() => {
  const win = window as unknown as Record<string, unknown>
  if (win[callbackKey])
    delete win[callbackKey]
})
</script>

<template>
  <div ref="telegramRootRef" class="telegram-login-widget" />
</template>

<style scoped>
.telegram-login-widget {
  display: inline-flex;
  justify-content: center;
  width: 100%;
}
</style>
