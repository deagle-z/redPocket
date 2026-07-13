<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { showToast } from 'vant'
import PpmxActionCard from '@/components/PpmxActionCard.vue'
import PpmxSubpageLayout from '@/components/PpmxSubpageLayout.vue'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '../ppmx-home/types'

definePage({
  name: 'security',
  meta: {
    titleKey: 'security.title',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: true,
  },
})

interface SecurityItem {
  id: 'password' | 'email' | 'phone'
  icon: string
  tone: 'red' | 'blue' | 'green'
  title: LocalizedText
  description: ComputedRef<string>
}

const userStore = useUserStore()
const router = useRouter()
const { pickText } = usePpmxLocale()
const securityText = {
  eyebrow: { es: 'Cuenta protegida', en: 'Account protected', zh: '账户保护' },
  title: { es: 'Centro de seguridad', en: 'Security center', zh: '安全中心' },
  sub: {
    es: 'Gestiona tu contraseña y datos de acceso.',
    en: 'Manage your password and access details.',
    zh: '管理你的密码与登录信息。',
  },
  passwordTitle: { es: 'Cambiar contraseña', en: 'Change password', zh: '修改密码' },
  passwordDesc: {
    es: 'Última actualización hace 3 meses',
    en: 'Last updated 3 months ago',
    zh: '上次更新于 3 个月前',
  },
  emailTitle: { es: 'Correo electrónico', en: 'Email address', zh: '绑定邮箱' },
  phoneTitle: { es: 'Número de celular', en: 'Phone number', zh: '修改手机号' },
  verified: { es: 'verificado', en: 'verified', zh: '已验证' },
  unbound: { es: 'Sin vincular', en: 'Not linked', zh: '未绑定' },
  soon: {
    es: 'Función en construcción · próxima iteración',
    en: 'Feature under construction · next iteration',
    zh: '功能建设中 · 下一版本',
  },
} satisfies Record<string, LocalizedText>

const maskedEmail = computed(() => {
  const email = userStore.userInfo?.email?.trim()
  if (!email) return pickText(securityText.unbound)

  const [name, domain] = email.split('@')
  if (!domain) return email

  const visible = name.slice(0, 2)

  return `${visible}${'*'.repeat(Math.max(2, name.length - visible.length))}@${domain}`
})
const maskedPhone = computed(() => {
  const phone = userStore.userInfo?.phone?.trim()
  if (!phone) return pickText(securityText.unbound)

  const digits = phone.replace(/\D/g, '')
  if (digits.length < 6) return phone

  return `+${digits.slice(0, 2)} ${digits.slice(2, 4)} **** ${digits.slice(-4)}`
})
const emailDescription = computed(() => {
  if (!userStore.userInfo?.email?.trim()) return maskedEmail.value

  return `${maskedEmail.value} · ${pickText(securityText.verified)}`
})
const securityItems: SecurityItem[] = [
  {
    id: 'password',
    icon: 'fa-key',
    tone: 'red',
    title: securityText.passwordTitle,
    description: computed(() => pickText(securityText.passwordDesc)),
  },
  {
    id: 'email',
    icon: 'fa-envelope',
    tone: 'blue',
    title: securityText.emailTitle,
    description: emailDescription,
  },
  {
    id: 'phone',
    icon: 'fa-mobile-screen',
    tone: 'green',
    title: securityText.phoneTitle,
    description: maskedPhone,
  },
]

function showSoon() {
  showToast(pickText(securityText.soon))
}

function openSecurityItem(item: SecurityItem) {
  if (item.id === 'password') {
    void router.push('/sec-password')
    return
  }

  if (item.id === 'email') {
    void router.push('/sec-email')
    return
  }

  if (item.id === 'phone') {
    void router.push('/sec-phone')
    return
  }

  showSoon()
}
</script>

<template>
  <PpmxSubpageLayout
    page-class="ppmx-security-page"
    stack-class="ppmx-security-stack"
    head-class="ppmx-security-head reveal"
    :eyebrow="pickText(securityText.eyebrow)"
    :title="pickText(securityText.title)"
    :description="pickText(securityText.sub)"
  >
    <section class="ppmx-security-list" aria-label="Security actions">
      <PpmxActionCard
        v-for="item in securityItems"
        :key="item.id"
        base-class="ppmx-security-card reveal"
        body-class="ppmx-security-card__body"
        icon-class="ppmx-security-card__icon"
        :icon="item.icon"
        :tone="item.tone"
        :title="pickText(item.title)"
        :description="item.description.value"
        @click="openSecurityItem(item)"
      />
    </section>
  </PpmxSubpageLayout>
</template>
