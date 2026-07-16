<script setup lang="ts">
import { showFailToast, showToast } from 'vant'

import { checkRegisterPhone } from '@/api/user'
import { useUserStore } from '@/stores/user'
import {
  createThirdPartyEventId,
  trackAttributionEvent,
} from '@/utils/attribution'
import { getDeviceFingerprint } from '@/utils/deviceFingerprint'
import {
  captureFacebookPixelId,
  initFacebookPixel,
  trackCompleteRegistration,
} from '@/utils/facebookPixel'
import {
  buildMxPhoneWithDialCode,
  isValidMxNationalPhone,
  normalizeMxNationalPhone,
  MX_COUNTRY_CODE,
} from '@/utils/mxPhone'
import { APP_DIAL_CODE } from '@/config/market'
import {
  captureInviteCode,
  captureSourceChannelCode,
  getSourceChannelCode,
  getStoredInviteCode,
} from '@/utils/sourceChannel'
import type { PpmxAuthMode } from '../types'
import PpmxButton from '@/components/PpmxButton.vue'
import PpmxLogo from './PpmxLogo.vue'
import PpmxModal from './PpmxModal.vue'

defineOptions({ name: 'PpmxAuthModal' })

const props = defineProps<{
  mode: PpmxAuthMode
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:mode': [value: PpmxAuthMode]
  'update:modelValue': [value: boolean]
}>()

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const registerStep = ref(1)
const submitting = ref(false)
const submitError = ref('')
const AUTH_SUCCESS_REDIRECT = '/profile'
const LOGIN_REMEMBER_PHONE_KEY = 'ppmx_login_phone'
const loginForm = reactive({
  phone: getRememberedPhone(),
  password: '',
  remember: true,
})
const registerForm = reactive({
  phone: '',
  firstName: '',
  password: '',
  confirmPassword: '',
  inviteCode: '',
  accepted: false,
})
const modalOpen = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})
const isLogin = computed(() => props.mode === 'login')

watch(
  () => [props.modelValue, props.mode] as const,
  ([open, mode]) => {
    if (!open) return

    syncAttributionFromRoute()
    submitError.value = ''
    if (mode === 'register') registerStep.value = 1
  },
)

onMounted(() => {
  syncAttributionFromRoute()
})

function getRememberedPhone() {
  if (typeof window === 'undefined') return ''

  return window.localStorage.getItem(LOGIN_REMEMBER_PHONE_KEY) ?? ''
}

function saveRememberedPhone(phone: string) {
  if (typeof window === 'undefined') return

  if (loginForm.remember) {
    window.localStorage.setItem(LOGIN_REMEMBER_PHONE_KEY, phone)
  } else {
    window.localStorage.removeItem(LOGIN_REMEMBER_PHONE_KEY)
  }
}

function syncAttributionFromRoute() {
  const inviteCode = captureInviteCode(route.query)
  captureSourceChannelCode(route.query)
  captureFacebookPixelId(route.query)
  initFacebookPixel()

  if (inviteCode && !registerForm.inviteCode) {
    registerForm.inviteCode = inviteCode
  } else if (!registerForm.inviteCode) {
    registerForm.inviteCode = getStoredInviteCode()
  }
}

function closeAuthModal() {
  submitError.value = ''
  emit('update:modelValue', false)
}

function switchAuthMode(mode: PpmxAuthMode) {
  emit('update:mode', mode)
  submitError.value = ''
  registerStep.value = 1
}

async function openForgotPassword() {
  const nationalPhone = normalizeMxNationalPhone(loginForm.phone)

  closeAuthModal()
  await router.push({
    path: '/sec-password',
    query: nationalPhone ? { phone: nationalPhone } : undefined,
  })
}

function normalizeLoginPhone() {
  loginForm.phone = normalizeMxNationalPhone(loginForm.phone).slice(0, 10)
}

function normalizeRegisterPhone() {
  registerForm.phone = normalizeMxNationalPhone(registerForm.phone).slice(0, 10)
}

function getSubmitErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : t('error.requestFailed')
}

function getRegisterPhoneErrorMessage(error: unknown) {
  const message = getSubmitErrorMessage(error)

  return message === 'phone_registered' ? t('auth.phoneRegistered') : message
}

function validateLoginForm() {
  const nationalPhone = normalizeMxNationalPhone(loginForm.phone)

  if (!nationalPhone) {
    showFailToast(t('auth.requiredPhone'))
    return false
  }

  if (!isValidMxNationalPhone(nationalPhone)) {
    showFailToast(t('auth.invalidPhone'))
    return false
  }

  if (!loginForm.password) {
    showFailToast(t('auth.requiredPassword'))
    return false
  }

  return true
}

function validateRegisterStep(step = registerStep.value) {
  const nationalPhone = normalizeMxNationalPhone(registerForm.phone)

  if (step === 1) {
    if (!nationalPhone) {
      showFailToast(t('auth.requiredPhone'))
      return false
    }

    if (!isValidMxNationalPhone(nationalPhone)) {
      showFailToast(t('auth.invalidPhone'))
      return false
    }

    if (!registerForm.firstName.trim()) {
      showFailToast(t('auth.requiredNickname'))
      return false
    }

    return true
  }

  if (step === 2) {
    if (registerForm.password.length < 8) {
      showFailToast(t('auth.passwordTooShort'))
      return false
    }

    if (!registerForm.confirmPassword) {
      showFailToast(t('auth.requiredConfirmPassword'))
      return false
    }

    if (registerForm.password !== registerForm.confirmPassword) {
      showFailToast(t('auth.passwordMismatch'))
      return false
    }

    return true
  }

  if (!registerForm.accepted) {
    showFailToast(t('auth.requiredTerms'))
    return false
  }

  return true
}

function validateRegisterForm() {
  return validateRegisterStep(1) && validateRegisterStep(2) && validateRegisterStep(3)
}

async function checkRegisterPhoneAvailable() {
  if (!validateRegisterStep(1)) return false

  submitting.value = true
  submitError.value = ''
  const nationalPhone = normalizeMxNationalPhone(registerForm.phone)
  const phone = buildMxPhoneWithDialCode(nationalPhone)

  try {
    const result = await checkRegisterPhone({
      phone,
      country: MX_COUNTRY_CODE,
    })

    if (!result.available) {
      const message = t('auth.phoneRegistered')
      submitError.value = message
      showFailToast(message)
      return false
    }

    return true
  } catch (error) {
    const message = getRegisterPhoneErrorMessage(error)
    submitError.value = message
    showFailToast(message)
    return false
  } finally {
    submitting.value = false
  }
}

async function goRegisterStep(delta: number) {
  submitError.value = ''
  if (submitting.value) return
  if (delta > 0 && registerStep.value === 1) {
    if (!await checkRegisterPhoneAvailable()) return
    registerStep.value = 2
    return
  }

  if (delta > 0 && !validateRegisterStep(registerStep.value)) return
  registerStep.value = Math.min(3, Math.max(1, registerStep.value + delta))
}

async function submitLogin() {
  if (submitting.value || !validateLoginForm()) return

  submitting.value = true
  submitError.value = ''
  const nationalPhone = normalizeMxNationalPhone(loginForm.phone)
  const phone = buildMxPhoneWithDialCode(nationalPhone)
  const eventId = createThirdPartyEventId('login_submit')

  void trackAttributionEvent({
    eventName: 'login_submit',
    eventId,
    metadata: { country: MX_COUNTRY_CODE },
  })

  try {
    await userStore.loginByPhone({
      phone,
      password: loginForm.password,
      country: MX_COUNTRY_CODE,
    }, {
      remember: loginForm.remember,
    })
    saveRememberedPhone(nationalPhone)

    void trackAttributionEvent({
      eventName: 'login_success',
      eventId: createThirdPartyEventId('login_success'),
      metadata: { country: MX_COUNTRY_CODE },
    })

    showToast(t('auth.loginSuccess'))
    closeAuthModal()
    await router.push(AUTH_SUCCESS_REDIRECT)
  } catch (error) {
    const message = getSubmitErrorMessage(error)
    submitError.value = message
    showFailToast(message)
  } finally {
    submitting.value = false
  }
}

async function submitRegister() {
  if (submitting.value || !validateRegisterForm()) return

  submitting.value = true
  submitError.value = ''
  const nationalPhone = normalizeMxNationalPhone(registerForm.phone)
  const phone = buildMxPhoneWithDialCode(nationalPhone)
  const sourceChannelCode = getSourceChannelCode()
  const inviteCode = registerForm.inviteCode.trim() || getStoredInviteCode()
  const submitEventId = createThirdPartyEventId('register_submit')

  void trackAttributionEvent({
    eventName: 'register_submit',
    eventId: submitEventId,
    metadata: { country: MX_COUNTRY_CODE, inviteCode, sourceChannelCode },
  })

  try {
    const deviceFingerprint = await getDeviceFingerprint()

    await userStore.registerByPhone({
      phone,
      country: MX_COUNTRY_CODE,
      firstName: registerForm.firstName.trim(),
      password: registerForm.password,
      inviteCode,
      sourceChannelCode,
      channelCode: sourceChannelCode,
      deviceFingerprint,
    })

    const successEventId = createThirdPartyEventId('register_success')

    void trackAttributionEvent({
      eventName: 'register_success',
      eventId: successEventId,
      metadata: { country: MX_COUNTRY_CODE, inviteCode, sourceChannelCode },
    })
    trackCompleteRegistration(successEventId)

    showToast(t('auth.registerSuccess'))
    closeAuthModal()
    await router.push(AUTH_SUCCESS_REDIRECT)
  } catch (error) {
    const message = getSubmitErrorMessage(error)
    submitError.value = message
    showFailToast(message)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <PpmxModal v-model="modalOpen" aria-label="PP.MX account access" card-class="ppmx-auth-card">
    <template v-if="isLogin">
      <PpmxLogo />
      <h2>{{ t('auth.welcomeTitle') }}</h2>
      <p>{{ t('auth.welcomeDescription') }}</p>

      <div v-if="submitError" class="ppmx-auth-error" role="alert">
        <i class="fa-solid fa-circle-exclamation" />
        <span>{{ submitError }}</span>
      </div>

      <form class="ppmx-auth-form" @submit.prevent="submitLogin">
        <label>
          <span>{{ t('auth.phone') }}</span>
          <div class="ppmx-auth-phone">
            <strong>+{{ APP_DIAL_CODE }}</strong>
            <input
              v-model="loginForm.phone"
              class="ppmx-auth-field"
              name="phone"
              type="tel"
              inputmode="numeric"
              required
              :placeholder="t('auth.phonePlaceholder')"
              @input="normalizeLoginPhone"
            >
          </div>
        </label>

        <label>
          <span>{{ t('auth.password') }}</span>
          <div class="ppmx-auth-password">
            <input
              v-model="loginForm.password"
              class="ppmx-auth-field"
              name="password"
              type="password"
              required
              placeholder="••••••••"
            >
            <i class="fa-solid fa-eye" />
          </div>
        </label>

        <div class="ppmx-auth-row">
          <label class="ppmx-auth-check">
            <input v-model="loginForm.remember" type="checkbox">
            <span>{{ t('auth.remember') }}</span>
          </label>
          <button type="button" @click="openForgotPassword">
            {{ t('auth.forgot') }}
          </button>
        </div>

        <PpmxButton class="ppmx-auth-submit" type="submit" variant="red" block :loading="submitting">
          {{ t('auth.loginSubmit') }}
        </PpmxButton>
      </form>

      <p class="ppmx-auth-foot">
        {{ t('auth.noAccount') }}
        <button type="button" @click="switchAuthMode('register')">{{ t('auth.registerFree') }}</button>
      </p>
    </template>

    <template v-else>
      <PpmxLogo />
      <div class="ppmx-auth-steps" aria-label="Registro en 3 pasos">
        <span
          v-for="step in 3"
          :key="step"
          class="ppmx-auth-step-dot"
          :class="{ 'is-active': step <= registerStep }"
        />
      </div>

      <h2>{{ t(`auth.registerTitle${registerStep}`) }}</h2>
      <p>{{ t(`auth.registerSub${registerStep}`) }}</p>

      <div v-if="submitError" class="ppmx-auth-error" role="alert">
        <i class="fa-solid fa-circle-exclamation" />
        <span>{{ submitError }}</span>
      </div>

      <form class="ppmx-auth-form" @submit.prevent="submitRegister">
        <div v-if="registerStep === 1" class="ppmx-auth-stack">
          <label>
            <span>{{ t('auth.phone') }}</span>
            <div class="ppmx-auth-phone">
              <strong>+{{ APP_DIAL_CODE }}</strong>
              <input
                v-model="registerForm.phone"
                class="ppmx-auth-field"
                name="phone"
                type="tel"
                inputmode="numeric"
                required
                :placeholder="t('auth.phonePlaceholder')"
                @input="normalizeRegisterPhone"
              >
            </div>
          </label>
          <label>
            <span>{{ t('auth.firstName') }}</span>
            <input
              v-model.trim="registerForm.firstName"
              class="ppmx-auth-field"
              name="firstName"
              required
              :placeholder="t('auth.firstNamePlaceholder')"
            >
          </label>
        </div>

        <div v-else-if="registerStep === 2" class="ppmx-auth-stack">
          <label>
            <span>{{ t('auth.createPassword') }}</span>
            <input
              v-model="registerForm.password"
              class="ppmx-auth-field"
              name="password"
              type="password"
              minlength="8"
              required
              :placeholder="t('auth.minPassword')"
            >
          </label>
          <label>
            <span>{{ t('auth.confirmPassword') }}</span>
            <input
              v-model="registerForm.confirmPassword"
              class="ppmx-auth-field"
              name="confirmPassword"
              type="password"
              minlength="8"
              required
              :placeholder="t('auth.confirmPasswordPlaceholder')"
            >
          </label>
        </div>

        <div v-else class="ppmx-auth-stack">
          <label>
            <span>{{ t('auth.inviteCode') }}</span>
            <input
              v-model.trim="registerForm.inviteCode"
              class="ppmx-auth-field"
              name="inviteCode"
              autocomplete="off"
              :placeholder="t('auth.inviteCodePlaceholder')"
            >
          </label>
          <label class="ppmx-auth-terms">
            <input v-model="registerForm.accepted" type="checkbox" required>
            <span>{{ t('auth.terms') }}</span>
          </label>
        </div>

        <div class="ppmx-auth-actions">
          <PpmxButton
            v-if="registerStep > 1"
            class="ppmx-auth-back"
            type="button"
            variant="ghost"
            :disabled="submitting"
            @click="goRegisterStep(-1)"
          >
            {{ t('auth.back') }}
          </PpmxButton>
          <PpmxButton
            v-if="registerStep < 3"
            class="ppmx-auth-submit"
            type="button"
            variant="red"
            block
            :disabled="submitting"
            :loading="submitting && registerStep === 1"
            @click="goRegisterStep(1)"
          >
            {{ t('auth.next') }}
          </PpmxButton>
          <PpmxButton
            v-else
            class="ppmx-auth-submit ppmx-auth-finish"
            type="submit"
            variant="red"
            block
            :loading="submitting"
          >
            {{ t('auth.finish') }}
          </PpmxButton>
        </div>
      </form>

      <p class="ppmx-auth-foot">
        {{ t('auth.hasAccount') }}
        <button type="button" @click="switchAuthMode('login')">{{ t('auth.signIn') }}</button>
      </p>
    </template>
  </PpmxModal>
</template>
