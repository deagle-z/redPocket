<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { showFailToast, showSuccessToast } from 'vant'
import PpmxButton from '@/components/PpmxButton.vue'
import PpmxSubpageLayout from '@/components/PpmxSubpageLayout.vue'
import { forgotPasswordByPhone, sendTgSmsCode } from '@/api/user'
import type { LocationQueryValue } from 'vue-router'
import {
  buildMxPhoneWithDialCode,
  isValidMxNationalPhone,
  normalizeMxNationalPhone,
  MX_COUNTRY_CODE,
} from '@/utils/mxPhone'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '../ppmx-home/types'

definePage({
  name: 'sec-password',
  meta: {
    titleKey: 'security.password',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: false,
  },
})

const userStore = useUserStore()
const route = useRoute()
const router = useRouter()
const { pickText } = usePpmxLocale()

const secPasswordText = {
  eyebrow: { es: 'Cuenta protegida', en: 'Account protected', zh: '账户保护' },
  title: { es: 'Cambiar contraseña', en: 'Change password', zh: '修改密码' },
  sub: {
    es: 'Actualiza tu contraseña de acceso con verificación por SMS.',
    en: 'Update your account password with SMS verification.',
    zh: '通过短信验证码更新你的登录密码。',
  },
  phone: { es: 'Celular', en: 'Mobile number', zh: '手机号' },
  phonePlaceholder: { es: 'Ingresa tu número', en: 'Enter your number', zh: '请输入手机号' },
  code: { es: 'Código de verificación', en: 'Verification code', zh: '验证码' },
  codePlaceholder: { es: 'Código de 6 dígitos', en: '6-digit code', zh: '6 位验证码' },
  sendCode: { es: 'Enviar', en: 'Send', zh: '发送验证码' },
  sending: { es: 'Enviando', en: 'Sending', zh: '发送中' },
  newPassword: { es: 'Nueva contraseña', en: 'New password', zh: '新密码' },
  newPasswordPlaceholder: { es: 'Define tu nueva contraseña', en: 'Set your new password', zh: '请设置新密码' },
  confirmPassword: { es: 'Confirmar nueva contraseña', en: 'Confirm new password', zh: '确认新密码' },
  confirmPasswordPlaceholder: { es: 'Repite la nueva contraseña', en: 'Re-enter the new password', zh: '请再次输入新密码' },
  submit: { es: 'Cambiar', en: 'Change', zh: '修改' },
  invalidPhone: { es: 'Ingresa un celular válido de México.', en: 'Enter a valid Mexico mobile number.', zh: '请输入有效的墨西哥手机号。' },
  requiredCode: { es: 'Ingresa el código de verificación.', en: 'Enter the verification code.', zh: '请输入验证码。' },
  shortPassword: { es: 'La contraseña debe tener al menos 6 caracteres.', en: 'Password must be at least 6 characters.', zh: '密码至少需 6 位字符。' },
  mismatch: { es: 'Las contraseñas no coinciden.', en: 'Passwords do not match.', zh: '两次输入的密码不一致。' },
  sent: { es: 'Código enviado por SMS.', en: 'SMS code sent.', zh: '验证码短信已发送。' },
  done: { es: 'Contraseña actualizada.', en: 'Password updated.', zh: '密码修改成功。' },
} satisfies Record<string, LocalizedText>

const secPwPhone = ref('')
const secPwCode = ref('')
const secPwNew = ref('')
const secPwConfirm = ref('')
const countdown = ref(0)
let countdownTimer: ReturnType<typeof window.setInterval> | null = null

function firstQueryValue(value: LocationQueryValue | LocationQueryValue[]) {
  return Array.isArray(value) ? value[0] : value
}

function syncUserPhone(phone?: string) {
  if (secPwPhone.value || !phone) return
  secPwPhone.value = normalizeMxNationalPhone(phone).slice(0, 10)
}

function syncRoutePhone() {
  const phone = firstQueryValue(route.query.phone)
  if (!phone) return

  secPwPhone.value = normalizeMxNationalPhone(phone).slice(0, 10)
}

watch(
  () => userStore.userInfo?.phone,
  phone => syncUserPhone(phone),
  { immediate: true },
)

watch(
  () => route.query.phone,
  syncRoutePhone,
  { immediate: true },
)

onUnmounted(() => {
  if (countdownTimer) window.clearInterval(countdownTimer)
})

function normalizePhoneInput() {
  secPwPhone.value = normalizeMxNationalPhone(secPwPhone.value).slice(0, 10)
}

function resolvePhoneWithCountryCode() {
  const nationalPhone = normalizeMxNationalPhone(secPwPhone.value)
  secPwPhone.value = nationalPhone.slice(0, 10)

  if (!isValidMxNationalPhone(nationalPhone)) {
    showFailToast(pickText(secPasswordText.invalidPhone))
    return ''
  }

  return buildMxPhoneWithDialCode(nationalPhone)
}

function startCountdown() {
  countdown.value = 60
  if (countdownTimer) window.clearInterval(countdownTimer)

  countdownTimer = window.setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0 && countdownTimer) {
      window.clearInterval(countdownTimer)
      countdownTimer = null
      countdown.value = 0
    }
  }, 1000)
}

async function sendCode() {
  const phone = resolvePhoneWithCountryCode()
  if (!phone) return

  await sendTgSmsCode(phone, MX_COUNTRY_CODE)
  startCountdown()
  showSuccessToast(pickText(secPasswordText.sent))
}

const sendCodeLock = useSubmitLock(sendCode, { minLockMs: 400 })
const sendCodeLoading = sendCodeLock.locked

async function handleSendCode() {
  try {
    await sendCodeLock.run()
  } catch (error) {
    showFailToast(error instanceof Error ? error.message : pickText(secPasswordText.sendCode))
  }
}

async function submitPasswordReset() {
  const phone = resolvePhoneWithCountryCode()
  if (!phone) return

  const code = secPwCode.value.trim()
  if (!code) {
    showFailToast(pickText(secPasswordText.requiredCode))
    return
  }

  if (secPwNew.value.length < 6) {
    showFailToast(pickText(secPasswordText.shortPassword))
    return
  }

  if (secPwNew.value !== secPwConfirm.value) {
    showFailToast(pickText(secPasswordText.mismatch))
    return
  }

  await forgotPasswordByPhone({
    phone,
    country: MX_COUNTRY_CODE,
    code,
    newPassword: secPwNew.value,
  })

  secPwCode.value = ''
  secPwNew.value = ''
  secPwConfirm.value = ''
  showSuccessToast(pickText(secPasswordText.done))
  await router.push('/security')
}

const submitLock = useSubmitLock(submitPasswordReset, { minLockMs: 600 })
const submitLoading = submitLock.locked

async function handleSubmit() {
  try {
    await submitLock.run()
  } catch (error) {
    showFailToast(error instanceof Error ? error.message : pickText(secPasswordText.submit))
  }
}
</script>

<template>
  <PpmxSubpageLayout
    page-id="page-sec-password"
    page-class="ppmx-sec-password-page"
    stack-class="ppmx-sec-password-stack"
    head-class="ppmx-security-head reveal"
    :eyebrow="pickText(secPasswordText.eyebrow)"
    :title="pickText(secPasswordText.title)"
    :description="pickText(secPasswordText.sub)"
  >
    <form class="ppmx-sec-password-form reveal" @submit.prevent="handleSubmit">
      <label class="ppmx-bank-field" for="secPwPhone">
        <span>{{ pickText(secPasswordText.phone) }} <em>*</em></span>
        <input
          id="secPwPhone"
          v-model="secPwPhone"
          type="tel"
          inputmode="tel"
          autocomplete="tel"
          :placeholder="pickText(secPasswordText.phonePlaceholder)"
          @input="normalizePhoneInput"
        >
      </label>

      <label class="ppmx-bank-field" for="secPwCode">
        <span>{{ pickText(secPasswordText.code) }} <em>*</em></span>
        <div class="ppmx-sec-password-code">
          <input
            id="secPwCode"
            v-model.trim="secPwCode"
            type="text"
            inputmode="numeric"
            maxlength="6"
            autocomplete="one-time-code"
            :placeholder="pickText(secPasswordText.codePlaceholder)"
          >
          <button
            class="ppmx-sec-password-send"
            type="button"
            :disabled="sendCodeLoading || countdown > 0"
            @click="handleSendCode"
          >
            <i v-if="sendCodeLoading" class="fa-solid fa-circle-notch fa-spin" />
            <span v-else-if="countdown > 0">{{ countdown }}s</span>
            <span v-else>{{ pickText(secPasswordText.sendCode) }}</span>
          </button>
        </div>
      </label>

      <label class="ppmx-bank-field" for="secPwNew">
        <span>{{ pickText(secPasswordText.newPassword) }} <em>*</em></span>
        <input
          id="secPwNew"
          v-model="secPwNew"
          type="password"
          minlength="6"
          autocomplete="new-password"
          :placeholder="pickText(secPasswordText.newPasswordPlaceholder)"
        >
      </label>

      <label class="ppmx-bank-field" for="secPwConfirm">
        <span>{{ pickText(secPasswordText.confirmPassword) }} <em>*</em></span>
        <input
          id="secPwConfirm"
          v-model="secPwConfirm"
          type="password"
          minlength="6"
          autocomplete="new-password"
          :placeholder="pickText(secPasswordText.confirmPasswordPlaceholder)"
        >
      </label>

      <PpmxButton
        class="ppmx-sec-password-submit"
        type="submit"
        variant="gold"
        size="lg"
        block
        :loading="submitLoading"
      >
        <span>{{ pickText(secPasswordText.submit) }}</span>
      </PpmxButton>
    </form>
  </PpmxSubpageLayout>
</template>
