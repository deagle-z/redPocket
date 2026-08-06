<script setup lang="ts">
import { showToast } from 'vant'
import { getTenantServiceLinks } from '@/api/user'
import type { LocalizedText } from '@/pages/ppmx-home/types'
import { usePpmxLocale } from '@/pages/ppmx-home/composables/usePpmxLocale'
import AppState from './AppState.vue'

defineOptions({ name: 'PpmxSupportModal' })

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  close: []
  'update:modelValue': [value: boolean]
}>()

const { pickText } = usePpmxLocale()
const loading = ref(false)
const loaded = ref(false)
const failed = ref(false)
const serviceLinks = ref({
  tgServiceUrl: '',
  wsServiceUrl: '',
})

const supportText = {
  title: { es: 'Atención al cliente', en: 'Customer support', zh: '在线客服' },
  sub: {
    es: 'Soporte 24/7. Te respondemos en minutos.',
    en: '24/7 support. We reply in minutes.',
    zh: '全天候客服，几分钟内回复。',
  },
  telegram: { es: 'Telegram', en: 'Telegram', zh: 'Telegram' },
  whatsapp: { es: 'WhatsApp', en: 'WhatsApp', zh: 'WhatsApp' },
  close: { es: 'Cerrar soporte', en: 'Close support', zh: '关闭客服弹窗' },
  open: { es: 'Abrir canal', en: 'Open channel', zh: '打开客服渠道' },
  safe: {
    es: 'Usa solo canales oficiales PP.PE. Nunca compartas tu contraseña.',
    en: 'Use PP.PE official channels only. Never share your password.',
    zh: '请仅使用 PP.PE 官方渠道，切勿泄露密码。',
  },
  loadingTitle: { es: 'Cargando soporte', en: 'Loading support', zh: '正在加载客服' },
  loadingMessage: {
    es: 'Estamos obteniendo los canales oficiales.',
    en: 'We are loading the official channels.',
    zh: '正在获取官方客服渠道。',
  },
  errorTitle: { es: 'No se pudo cargar soporte', en: 'Could not load support', zh: '客服加载失败' },
  errorMessage: {
    es: 'Revisa tu conexión e inténtalo nuevamente.',
    en: 'Check your connection and try again.',
    zh: '请检查网络后重试。',
  },
  emptyTitle: { es: 'Sin canales disponibles', en: 'No channels available', zh: '暂无客服渠道' },
  emptyMessage: {
    es: 'Aún no hay enlaces de atención configurados.',
    en: 'Support links are not configured yet.',
    zh: '当前暂未配置客服链接。',
  },
  retry: { es: 'Reintentar', en: 'Retry', zh: '重试' },
  notConfigured: {
    es: 'Canal de soporte no configurado',
    en: 'Support channel is not configured',
    zh: '客服渠道暂未配置',
  },
} satisfies Record<string, LocalizedText>

const supportChannelMeta = [
  {
    id: 'telegram',
    icon: 'fa-telegram',
    label: supportText.telegram,
    linkKey: 'tgServiceUrl',
    tone: 'telegram',
  },
  {
    id: 'whatsapp',
    icon: 'fa-whatsapp',
    label: supportText.whatsapp,
    linkKey: 'wsServiceUrl',
    tone: 'whatsapp',
  },
] as const

const supportChannels = computed(() =>
  supportChannelMeta
    .map((channel) => {
      const url = normalizeExternalUrl(serviceLinks.value[channel.linkKey])

      return {
        ...channel,
        handle: serviceLinks.value[channel.linkKey],
        href: url,
      }
    })
    .filter(channel => Boolean(channel.href)),
)

function normalizeExternalUrl(url?: string | null) {
  const value = String(url || '').trim()
  if (!value) return ''
  if (/^[a-z][a-z\d+\-.]*:/i.test(value)) return value

  return `https://${value}`
}

function normalizeServiceLinks(data?: { tgServiceUrl?: string | null; wsServiceUrl?: string | null } | null) {
  return {
    tgServiceUrl: String(data?.tgServiceUrl || '').trim(),
    wsServiceUrl: String(data?.wsServiceUrl || '').trim(),
  }
}

async function loadServiceLinks() {
  loading.value = true
  failed.value = false

  try {
    serviceLinks.value = normalizeServiceLinks(await getTenantServiceLinks())
    loaded.value = true
  } catch {
    serviceLinks.value = normalizeServiceLinks()
    loaded.value = false
    failed.value = true
  } finally {
    loading.value = false
  }
}

function openExternal(url?: string | null) {
  const target = normalizeExternalUrl(url)
  if (!target) {
    showToast(pickText(supportText.notConfigured))
    return
  }

  const opened = window.open(target, '_blank', 'noopener,noreferrer')
  if (opened) {
    opened.opener = null
    return
  }

  window.location.href = target
}

function closeModal() {
  emit('update:modelValue', false)
  emit('close')
}

watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen) void loadServiceLinks()
  },
)
</script>

<template>
  <Teleport to="body">
    <Transition name="ppmx-support-fade">
      <div
        v-if="modelValue"
        id="supportModal"
        class="ppmx-support-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="supportModalTitle"
      >
        <button
          class="ppmx-support-modal__backdrop"
          type="button"
          :aria-label="pickText(supportText.close)"
          @click="closeModal"
        />

        <section class="ppmx-support-card" role="document">
          <button
            class="ppmx-support-close"
            type="button"
            :aria-label="pickText(supportText.close)"
            @click="closeModal"
          >
            <i class="fa-solid fa-xmark" />
          </button>

          <div class="ppmx-support-icon" aria-hidden="true">
            <i class="fa-solid fa-headset" />
          </div>

          <h2 id="supportModalTitle">{{ pickText(supportText.title) }}</h2>
          <p>{{ pickText(supportText.sub) }}</p>

          <AppState
            v-if="loading"
            class="ppmx-support-state"
            status="loading"
            :title="pickText(supportText.loadingTitle)"
            :message="pickText(supportText.loadingMessage)"
            skeleton-variant="support-modal"
          />

          <AppState
            v-else-if="failed"
            class="ppmx-support-state"
            status="error"
            :title="pickText(supportText.errorTitle)"
            :message="pickText(supportText.errorMessage)"
            :action-text="pickText(supportText.retry)"
            @retry="loadServiceLinks"
          />

          <AppState
            v-else-if="loaded && supportChannels.length === 0"
            class="ppmx-support-state"
            status="empty"
            :title="pickText(supportText.emptyTitle)"
            :message="pickText(supportText.emptyMessage)"
            :action-text="pickText(supportText.retry)"
            @retry="loadServiceLinks"
          />

          <div v-else class="ppmx-support-channels">
            <button
              v-for="channel in supportChannels"
              :key="channel.id"
              class="ppmx-support-channel"
              :class="`is-${channel.tone}`"
              type="button"
              :aria-label="`${pickText(supportText.open)} ${pickText(channel.label)}`"
              @click="openExternal(channel.href)"
            >
              <span class="ppmx-support-channel__icon" aria-hidden="true">
                <i class="fa-brands" :class="channel.icon" />
              </span>
              <span class="ppmx-support-channel__body">
                <strong>{{ pickText(channel.label) }}</strong>
                <small>{{ channel.handle }}</small>
              </span>
              <i class="fa-solid fa-arrow-up-right-from-square ppmx-support-channel__arrow" />
            </button>
          </div>

          <div class="ppmx-support-safe">
            <i class="fa-solid fa-shield-halved" />
            <span>{{ pickText(supportText.safe) }}</span>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.ppmx-support-modal {
  position: fixed;
  inset: 0;
  z-index: 10020;
  display: grid;
  place-items: center;
  padding: 18px;
  color: var(--app-text, #f5f5f7);
  font-family:
    "Montserrat",
    system-ui,
    -apple-system,
    "Segoe UI",
    sans-serif;
}

.ppmx-support-modal__backdrop {
  position: absolute;
  inset: 0;
  cursor: pointer;
  background:
    radial-gradient(circle at 50% 18%, rgba(20, 184, 166, 0.18), transparent 42%),
    radial-gradient(circle at 76% 72%, rgba(200, 16, 46, 0.2), transparent 42%),
    rgba(0, 0, 0, 0.72);
  border: 0;
  backdrop-filter: blur(10px) saturate(135%);
  -webkit-backdrop-filter: blur(10px) saturate(135%);
}

.ppmx-support-card {
  position: relative;
  display: grid;
  gap: 16px;
  width: min(100%, 448px);
  padding: 28px;
  overflow: hidden;
  background:
    radial-gradient(circle at 14% 0%, rgba(20, 184, 166, 0.14), transparent 38%),
    radial-gradient(circle at 100% 12%, rgba(255, 215, 0, 0.08), transparent 40%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.028)),
    rgba(9, 9, 12, 0.96);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 28px;
  box-shadow:
    0 34px 100px rgba(0, 0, 0, 0.82),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.ppmx-support-card::before {
  position: absolute;
  top: 0;
  right: 34px;
  left: 34px;
  height: 1px;
  content: "";
  background: linear-gradient(90deg, transparent, rgba(45, 212, 191, 0.72), transparent);
}

.ppmx-support-close {
  position: absolute;
  top: 16px;
  right: 16px;
  display: grid;
  width: 36px;
  height: 36px;
  place-items: center;
  color: rgba(245, 245, 247, 0.58);
  cursor: pointer;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 999px;
  transition:
    color 0.2s ease,
    background 0.2s ease,
    transform 0.18s ease;
}

.ppmx-support-close:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.ppmx-support-close:active {
  transform: scale(0.95);
}

.ppmx-support-icon {
  display: grid;
  width: 52px;
  height: 52px;
  place-items: center;
  color: #2dd4bf;
  font-size: 20px;
  background: rgba(20, 184, 166, 0.12);
  border: 1px solid rgba(45, 212, 191, 0.18);
  border-radius: 18px;
  box-shadow: 0 16px 34px -24px rgba(45, 212, 191, 0.75);
}

.ppmx-support-card h2 {
  margin: 0;
  padding-right: 42px;
  color: #fff;
  font-size: clamp(1.35rem, 4vw, 1.72rem);
  font-weight: 900;
  line-height: 1.12;
}

.ppmx-support-card p {
  margin: -8px 0 0;
  color: rgba(245, 245, 247, 0.58);
  font-size: 0.92rem;
  font-weight: 700;
  line-height: 1.55;
}

.ppmx-support-state {
  min-height: 0;
  padding: 4px 0 0;
}

.ppmx-support-channels {
  display: grid;
  gap: 12px;
  margin-top: 6px;
}

.ppmx-support-channel {
  display: flex;
  gap: 14px;
  align-items: center;
  width: 100%;
  min-height: 78px;
  padding: 16px;
  color: #fff;
  font: inherit;
  text-align: left;
  text-decoration: none;
  cursor: pointer;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.075), rgba(255, 255, 255, 0.025)),
    rgba(255, 255, 255, 0.035);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  transition:
    transform 0.2s cubic-bezier(0.2, 0.7, 0.2, 1),
    border-color 0.2s ease,
    background 0.2s ease;
}

.ppmx-support-channel:hover {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.105), rgba(255, 255, 255, 0.038)),
    rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.16);
  transform: translateY(-2px);
}

.ppmx-support-channel__icon {
  display: grid;
  flex: 0 0 auto;
  width: 46px;
  height: 46px;
  place-items: center;
  font-size: 21px;
  border-radius: 16px;
}

.ppmx-support-channel.is-telegram .ppmx-support-channel__icon {
  color: #3aa9e0;
  background: rgba(34, 158, 217, 0.15);
}

.ppmx-support-channel.is-whatsapp .ppmx-support-channel__icon {
  color: #39d97f;
  background: rgba(37, 211, 102, 0.15);
}

.ppmx-support-channel__body {
  display: grid;
  flex: 1 1 auto;
  min-width: 0;
  gap: 4px;
}

.ppmx-support-channel__body strong {
  overflow: hidden;
  font-size: 0.98rem;
  font-weight: 900;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ppmx-support-channel__body small {
  overflow: hidden;
  color: rgba(245, 245, 247, 0.5);
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ppmx-support-channel__arrow {
  flex: 0 0 auto;
  color: rgba(245, 245, 247, 0.36);
  font-size: 12px;
}

.ppmx-support-safe {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 12px 13px;
  color: rgba(245, 245, 247, 0.58);
  font-size: 0.76rem;
  font-weight: 700;
  line-height: 1.45;
  background: rgba(255, 215, 0, 0.055);
  border: 1px solid rgba(255, 215, 0, 0.12);
  border-radius: 16px;
}

.ppmx-support-safe i {
  margin-top: 2px;
  color: #ffd700;
}

.ppmx-support-fade-enter-active,
.ppmx-support-fade-leave-active {
  transition: opacity 0.22s ease;
}

.ppmx-support-fade-enter-active .ppmx-support-card,
.ppmx-support-fade-leave-active .ppmx-support-card {
  transition:
    transform 0.28s cubic-bezier(0.2, 0.8, 0.2, 1),
    opacity 0.22s ease;
}

.ppmx-support-fade-enter-from,
.ppmx-support-fade-leave-to {
  opacity: 0;
}

.ppmx-support-fade-enter-from .ppmx-support-card,
.ppmx-support-fade-leave-to .ppmx-support-card {
  opacity: 0;
  transform: translateY(14px) scale(0.97);
}

html[data-theme="light"] .ppmx-support-modal {
  color: #1a1a1a;
}

html[data-theme="light"] .ppmx-support-modal__backdrop {
  background:
    radial-gradient(circle at 50% 18%, rgba(20, 184, 166, 0.16), transparent 42%),
    radial-gradient(circle at 78% 72%, rgba(200, 16, 46, 0.12), transparent 42%),
    rgba(245, 245, 247, 0.68);
}

html[data-theme="light"] .ppmx-support-card {
  background:
    radial-gradient(circle at 14% 0%, rgba(20, 184, 166, 0.11), transparent 38%),
    radial-gradient(circle at 100% 12%, rgba(255, 215, 0, 0.14), transparent 40%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.98), rgba(255, 248, 248, 0.92));
  border-color: rgba(200, 16, 46, 0.12);
  box-shadow:
    0 28px 76px rgba(25, 25, 30, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.92);
}

html[data-theme="light"] .ppmx-support-card h2,
html[data-theme="light"] .ppmx-support-channel {
  color: #1a1a1a;
}

html[data-theme="light"] .ppmx-support-channel__body strong {
  color: #17171b;
}

html[data-theme="light"] .ppmx-support-card p,
html[data-theme="light"] .ppmx-support-channel__body small,
html[data-theme="light"] .ppmx-support-state,
html[data-theme="light"] .ppmx-support-safe {
  color: #5a5a62;
}

html[data-theme="light"] .ppmx-support-state {
  background: rgba(255, 255, 255, 0.72);
  border-radius: 16px;
}

html[data-theme="light"] .ppmx-support-channel {
  background: rgba(255, 255, 255, 0.72);
  border-color: rgba(23, 23, 27, 0.08);
}

html[data-theme="light"] .ppmx-support-channel:hover {
  background: #fff;
  border-color: rgba(255, 26, 26, 0.16);
  box-shadow: 0 18px 42px -32px rgba(200, 16, 46, 0.28);
}

html[data-theme="light"] .ppmx-support-channel__arrow {
  color: #9a9aa2;
}

html[data-theme="light"] .ppmx-support-safe {
  background: rgba(255, 26, 26, 0.06);
  border-color: rgba(255, 26, 26, 0.12);
}

html[data-theme="light"] .ppmx-support-safe i {
  color: var(--app-mode-accent, #ff1a1a);
}

html[data-theme="light"] .ppmx-support-close {
  color: #76767e;
  background: rgba(255, 255, 255, 0.72);
  border-color: rgba(23, 23, 27, 0.08);
}

html[data-theme="light"] .ppmx-support-close:hover {
  color: #17171b;
  background: #fff;
}

@media (max-width: 480px) {
  .ppmx-support-modal {
    align-items: end;
    padding: 12px;
  }

  .ppmx-support-card {
    width: 100%;
    padding: 26px 20px 20px;
    border-radius: 26px;
  }

  .ppmx-support-channel {
    min-height: 72px;
    padding: 14px;
  }
}
</style>
