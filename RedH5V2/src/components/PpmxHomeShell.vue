<script setup lang="ts">
import { usePpmxLocale } from '@/pages/ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '@/pages/ppmx-home/types'

defineOptions({
  name: 'PpmxHomeShell',
})

const route = useRoute()
const router = useRouter()
const { pickText } = usePpmxLocale()

const navItems: Array<{
  page: string
  path: string
  icon: string
  label: LocalizedText
}> = [
  {
    page: 'home',
    path: '/',
    icon: 'fa-house',
    label: { es: 'Inicio', en: 'Home', zh: '首页' },
  },
  {
    page: 'wheel',
    path: '/wheel',
    icon: 'fa-dharmachakra',
    label: { es: 'Ruleta', en: 'Wheel', zh: '幸运转盘' },
  },
  {
    page: 'team',
    path: '/team',
    icon: 'fa-users',
    label: { es: 'Mi equipo', en: 'Team', zh: '我的团队' },
  },
  {
    page: 'promo',
    path: '/promo',
    icon: 'fa-gift',
    label: { es: 'Promos', en: 'Promos', zh: '优惠' },
  },
  {
    page: 'account',
    path: '/profile',
    icon: 'fa-user',
    label: { es: 'Cuenta', en: 'Account', zh: '我的账户' },
  },
]

function goTo(path: string) {
  if (route.path !== path) {
    void router.push(path)
  }
}
</script>

<template>
  <div class="ppmx-shell">
    <slot />

    <nav
      class="ppmx-mobile-tabbar tabbar lg:hidden fixed bottom-0 inset-x-0 z-40 glass-strong border-t border-white/8 grid grid-cols-5 text-[11px]"
      aria-label="Navegación móvil PP.MX"
    >
      <a
        v-for="item in navItems"
        :key="item.path"
        href="#"
        :data-page="item.page"
        class="ppmx-mobile-tabbar__item tabbar btn flex flex-col items-center gap-1 py-2.5"
        :class="route.path === item.path ? 'text-oro-400 active' : 'text-zinc-500'"
        @click.prevent="goTo(item.path)"
      >
        <span class="ppmx-mobile-tabbar__icon">
          <i class="fa-solid text-lg" :class="item.icon" />
        </span>
        <span>{{ pickText(item.label) }}</span>
      </a>
    </nav>
  </div>
</template>
