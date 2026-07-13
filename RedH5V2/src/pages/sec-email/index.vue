<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { showFailToast, showSuccessToast } from 'vant'
import PpmxButton from '@/components/PpmxButton.vue'
import PpmxSubpageLayout from '@/components/PpmxSubpageLayout.vue'
import { bindCurrentTgEmail, getCurrentUserInfo } from '@/api/user'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '../ppmx-home/types'

definePage({
  name: 'sec-email',
  meta: {
    titleKey: 'security.email',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: true,
  },
})

const router = useRouter()
const userStore = useUserStore()
const { pickText } = usePpmxLocale()

const secEmailText = {
  eyebrow: { es: 'Cuenta protegida', en: 'Account protected', zh: '账户保护' },
  title: { es: 'Vincular correo', en: 'Link email', zh: '绑定邮箱' },
  sub: {
    es: 'Vincula un correo para proteger y recuperar tu cuenta.',
    en: 'Link an email to protect and recover your account.',
    zh: '绑定邮箱用于保护和找回你的账号。',
  },
  label: { es: 'Correo electrónico', en: 'Email address', zh: '电子邮箱' },
  placeholder: { es: 'Ingresa tu correo', en: 'Enter your email', zh: '请输入邮箱' },
  current: { es: 'Correo actual', en: 'Current email', zh: '当前邮箱' },
  unbound: { es: 'Sin vincular', en: 'Not linked', zh: '未绑定' },
  submit: { es: 'Vincular', en: 'Link', zh: '绑定' },
  invalid: { es: 'Ingresa un correo válido.', en: 'Enter a valid email.', zh: '请输入有效的邮箱。' },
  done: { es: 'Correo vinculado con éxito.', en: 'Email linked successfully.', zh: '邮箱绑定成功。' },
} satisfies Record<string, LocalizedText>

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const secEmail = ref('')
const boundEmail = ref('')

const currentEmailLabel = computed(() => boundEmail.value || pickText(secEmailText.unbound))

function syncEmail(email?: string | null) {
  const nextEmail = String(email || '').trim()
  if (!nextEmail) return

  boundEmail.value = nextEmail
  if (!secEmail.value) secEmail.value = nextEmail
}

watch(
  () => userStore.userInfo?.email,
  email => syncEmail(email),
  { immediate: true },
)

async function loadCurrentEmail() {
  try {
    const currentUser = await getCurrentUserInfo()
    syncEmail(currentUser.email)
  } catch {}
}

onMounted(() => {
  void loadCurrentEmail()
})

async function submitBindEmail() {
  const email = secEmail.value.trim().toLowerCase()
  secEmail.value = email

  if (!EMAIL_PATTERN.test(email)) {
    showFailToast(pickText(secEmailText.invalid))
    return
  }

  await bindCurrentTgEmail({ email })
  boundEmail.value = email

  try {
    await userStore.loadUserInfo()
  } catch {}

  showSuccessToast(pickText(secEmailText.done))
  await router.push('/security')
}

const submitLock = useSubmitLock(submitBindEmail, { minLockMs: 600 })
const submitLoading = submitLock.locked

async function handleSubmit() {
  try {
    await submitLock.run()
  } catch (error) {
    showFailToast(error instanceof Error ? error.message : pickText(secEmailText.submit))
  }
}
</script>

<template>
  <PpmxSubpageLayout
    page-id="page-sec-email"
    page-class="ppmx-sec-email-page"
    stack-class="ppmx-sec-email-stack"
    head-class="ppmx-security-head reveal"
    :eyebrow="pickText(secEmailText.eyebrow)"
    :title="pickText(secEmailText.title)"
    :description="pickText(secEmailText.sub)"
  >
    <form class="ppmx-sec-email-form reveal" @submit.prevent="handleSubmit">
      <div class="ppmx-sec-email-current">
        <span>{{ pickText(secEmailText.current) }}</span>
        <strong>{{ currentEmailLabel }}</strong>
      </div>

      <label class="ppmx-bank-field" for="secEmail">
        <span>{{ pickText(secEmailText.label) }} <em>*</em></span>
        <input
          id="secEmail"
          v-model.trim="secEmail"
          type="email"
          inputmode="email"
          autocomplete="email"
          autocapitalize="none"
          autocorrect="off"
          spellcheck="false"
          :placeholder="pickText(secEmailText.placeholder)"
        >
      </label>

      <PpmxButton
        class="ppmx-sec-email-submit"
        type="submit"
        variant="gold"
        size="lg"
        block
        :loading="submitLoading"
      >
        <i v-if="!submitLoading" class="fa-solid fa-link" />
        <span>{{ pickText(secEmailText.submit) }}</span>
      </PpmxButton>
    </form>
  </PpmxSubpageLayout>
</template>
