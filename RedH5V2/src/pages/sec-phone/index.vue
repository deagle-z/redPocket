<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { showFailToast, showSuccessToast } from 'vant'
import PpmxButton from '@/components/PpmxButton.vue'
import PpmxSubpageLayout from '@/components/PpmxSubpageLayout.vue'
import { bindCurrentTgPhone, getCurrentUserInfo, sendTgSmsCode } from '@/api/user'
import {
  buildPePhoneWithDialCode,
  isValidPeNationalPhone,
  normalizePeNationalPhone,
  PE_COUNTRY_CODE,
} from '@/utils/pePhone'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '../ppmx-home/types'

definePage({
  name: 'sec-phone',
  meta: {
    titleKey: 'security.phone',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: true,
  },
})

const router = useRouter()
const userStore = useUserStore()
const { pickText } = usePpmxLocale()

const secPhoneText = {
  eyebrow: { es: 'Cuenta protegida', en: 'Account protected', zh: '账户保护' },
  title: { es: 'Cambiar celular', en: 'Change mobile number', zh: '修改手机号' },
  sub: {
    es: 'Actualiza el número asociado a tu cuenta.',
    en: 'Update the mobile number associated with your account.',
    zh: '更新当前账号绑定的手机号。',
  },
  currentPhone: { es: 'Celular actual', en: 'Current mobile number', zh: '当前手机号' },
  currentPlaceholder: { es: 'Sin vincular', en: 'Not linked', zh: '未绑定' },
  newPhone: { es: 'Nuevo celular', en: 'New mobile number', zh: '新手机号' },
  newPhonePlaceholder: { es: 'Ingresa tu nuevo celular', en: 'Enter your new number', zh: '请输入新手机号' },
  code: { es: 'Código de verificación', en: 'Verification code', zh: '验证码' },
  codePlaceholder: { es: 'Código de 6 dígitos', en: '6-digit code', zh: '6 位验证码' },
  getCode: { es: 'Obtener código', en: 'Get code', zh: '获取验证码' },
  sending: { es: 'Enviando', en: 'Sending', zh: '发送中' },
  submit: { es: 'Confirmar cambio', en: 'Confirm change', zh: '确认修改' },
  invalidNew: { es: 'Ingresa un celular válido de Perú.', en: 'Enter a valid Peru mobile number.', zh: '请输入有效的秘鲁手机号。' },
  invalidCode: { es: 'Ingresa un código de 6 dígitos.', en: 'Enter a 6-digit code.', zh: '请输入 6 位验证码。' },
  sent: { es: 'Código enviado por SMS.', en: 'SMS code sent.', zh: '验证码短信已发送。' },
  done: { es: 'Celular actualizado.', en: 'Mobile number updated.', zh: '手机号修改成功。' },
} satisfies Record<string, LocalizedText>

const secPhOld = ref('')
const secPhNew = ref('')
const secPhCode = ref('')
const countdown = ref(0)
let countdownTimer: ReturnType<typeof window.setInterval> | null = null

function syncCurrentPhone(phone?: string | null) {
  const nextPhone = normalizePeNationalPhone(String(phone || '')).slice(0, 9)
  if (!nextPhone) return

  secPhOld.value = nextPhone
}

watch(
  () => userStore.userInfo?.phone,
  phone => syncCurrentPhone(phone),
  { immediate: true },
)

onMounted(async () => {
  try {
    const currentUser = await getCurrentUserInfo()
    syncCurrentPhone(currentUser.phone)
  } catch {}
})

onUnmounted(() => {
  if (countdownTimer) window.clearInterval(countdownTimer)
})

function normalizeNewPhoneInput() {
  secPhNew.value = normalizePeNationalPhone(secPhNew.value).slice(0, 9)
}

function resolveNewPhoneWithCountryCode() {
  const nationalPhone = normalizePeNationalPhone(secPhNew.value)
  secPhNew.value = nationalPhone.slice(0, 9)

  if (!isValidPeNationalPhone(nationalPhone)) {
    showFailToast(pickText(secPhoneText.invalidNew))
    return ''
  }

  return buildPePhoneWithDialCode(nationalPhone)
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
  const phone = resolveNewPhoneWithCountryCode()
  if (!phone) return

  await sendTgSmsCode(phone, PE_COUNTRY_CODE)
  startCountdown()
  showSuccessToast(pickText(secPhoneText.sent))
}

const sendCodeLock = useSubmitLock(sendCode, { minLockMs: 400 })
const sendCodeLoading = sendCodeLock.locked

async function handleSendCode() {
  try {
    await sendCodeLock.run()
  } catch (error) {
    showFailToast(error instanceof Error ? error.message : pickText(secPhoneText.getCode))
  }
}

async function submitBindPhone() {
  const phone = resolveNewPhoneWithCountryCode()
  if (!phone) return

  const code = secPhCode.value.trim()
  if (!/^\d{6}$/.test(code)) {
    showFailToast(pickText(secPhoneText.invalidCode))
    return
  }

  await bindCurrentTgPhone({
    phone,
    country: PE_COUNTRY_CODE,
    code,
  })

  secPhOld.value = secPhNew.value
  secPhNew.value = ''
  secPhCode.value = ''

  try {
    await userStore.loadUserInfo()
  } catch {}

  showSuccessToast(pickText(secPhoneText.done))
  await router.push('/security')
}

const submitLock = useSubmitLock(submitBindPhone, { minLockMs: 600 })
const submitLoading = submitLock.locked

async function handleSubmit() {
  try {
    await submitLock.run()
  } catch (error) {
    showFailToast(error instanceof Error ? error.message : pickText(secPhoneText.submit))
  }
}
</script>

<template>
  <PpmxSubpageLayout
    page-id="page-sec-phone"
    page-class="ppmx-sec-phone-page"
    stack-class="ppmx-sec-phone-stack"
    head-class="ppmx-security-head reveal"
    :eyebrow="pickText(secPhoneText.eyebrow)"
    :title="pickText(secPhoneText.title)"
    :description="pickText(secPhoneText.sub)"
  >
    <form class="ppmx-sec-phone-form reveal" @submit.prevent="handleSubmit">
      <label class="ppmx-bank-field" for="secPhOld">
        <span>{{ pickText(secPhoneText.currentPhone) }}</span>
        <input
          id="secPhOld"
          v-model="secPhOld"
          type="tel"
          inputmode="tel"
          readonly
          :placeholder="pickText(secPhoneText.currentPlaceholder)"
        >
      </label>

      <label class="ppmx-bank-field" for="secPhNew">
        <span>{{ pickText(secPhoneText.newPhone) }} <em>*</em></span>
        <input
          id="secPhNew"
          v-model="secPhNew"
          type="tel"
          inputmode="tel"
          autocomplete="tel"
          :placeholder="pickText(secPhoneText.newPhonePlaceholder)"
          @input="normalizeNewPhoneInput"
        >
      </label>

      <label class="ppmx-bank-field" for="secPhCode">
        <span>{{ pickText(secPhoneText.code) }} <em>*</em></span>
        <div class="ppmx-sec-phone-code">
          <input
            id="secPhCode"
            v-model.trim="secPhCode"
            type="text"
            inputmode="numeric"
            maxlength="6"
            autocomplete="one-time-code"
            :placeholder="pickText(secPhoneText.codePlaceholder)"
          >
          <button
            class="ppmx-sec-phone-send"
            type="button"
            :disabled="sendCodeLoading || countdown > 0"
            @click="handleSendCode"
          >
            <i v-if="sendCodeLoading" class="fa-solid fa-circle-notch fa-spin" />
            <span v-else-if="countdown > 0">{{ countdown }}s</span>
            <span v-else>{{ pickText(secPhoneText.getCode) }}</span>
          </button>
        </div>
      </label>

      <PpmxButton
        class="ppmx-sec-phone-submit"
        type="submit"
        variant="gold"
        size="lg"
        block
        :loading="submitLoading"
      >
        <i v-if="!submitLoading" class="fa-solid fa-check" />
        <span>{{ pickText(secPhoneText.submit) }}</span>
      </PpmxButton>
    </form>
  </PpmxSubpageLayout>
</template>
