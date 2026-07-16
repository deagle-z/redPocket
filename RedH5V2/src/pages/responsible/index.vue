<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import { APP_CURRENCY } from '@/config/market'
import { showToast } from 'vant'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { LocalizedText } from '../ppmx-home/types'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'

definePage({
  name: 'responsible',
  meta: {
    title: 'Juego responsable',
    shell: 'ppmx',
    tabbar: false,
    requiresAuth: false,
  },
})

const { pickText } = usePpmxLocale()
const depositLimit = ref('')
const sessionTime = ref('30')

const responsibleText = {
  eyebrow: { es: 'Juego responsable', en: 'Responsible gaming', zh: '责任博彩' },
  title: { es: 'Tú tienes el control', en: 'You are in control', zh: '你掌控节奏' },
  intro: {
    es: 'El juego debe ser entretenimiento. Si sientes que se sale de control, pausa y contacta a soporte. No estás solo.',
    en: 'Gaming should be entertainment. If you feel it is getting out of control, pause and contact support. You are not alone.',
    zh: '游戏应当只是娱乐。如果你感觉正在失控，请暂停并联系客服。你并不孤单。',
  },
  limitTitle: { es: 'Límite de depósito', en: 'Deposit limit', zh: '存款限额' },
  limitDesc: {
    es: 'Define cuánto puedes depositar al día, semana o mes.',
    en: 'Set how much you can deposit per day, week, or month.',
    zh: '设置每日、每周或每月可存款金额。',
  },
  timeTitle: { es: 'Tiempo de sesión', en: 'Session time', zh: '会话时长' },
  timeDesc: {
    es: 'Te avisamos cuando lleves el tiempo que elijas.',
    en: 'We remind you when you reach the time you choose.',
    zh: '达到你选择的时长后，我们会提醒你。',
  },
  excludeTitle: { es: 'Autoexclusión', en: 'Self-exclusion', zh: '自我排除' },
  excludeDesc: {
    es: 'Cierra tu cuenta temporal o permanentemente. No te dejaremos volver hasta cumplir el plazo.',
    en: 'Temporarily or permanently close your account. Access stays blocked until the selected period ends.',
    zh: '你可以临时或永久关闭账户。在期限结束前，系统不会允许你重新进入。',
  },
  apply: { es: 'Aplicar', en: 'Apply', zh: '应用' },
  activate: { es: 'Activar', en: 'Activate', zh: '启用' },
  applied: { es: 'Configuración aplicada', en: 'Setting applied', zh: '设置已应用' },
  currency: { es: APP_CURRENCY, en: APP_CURRENCY, zh: APP_CURRENCY },
  days: { es: 'días', en: 'days', zh: '天' },
  month: { es: 'mes', en: 'month', zh: '个月' },
  months: { es: 'meses', en: 'months', zh: '个月' },
  permanent: { es: 'Permanente', en: 'Permanent', zh: '永久' },
  helpTitle: { es: '¿Necesitas ayuda?', en: 'Need help?', zh: '需要帮助？' },
  helpSub: {
    es: 'Si sientes que el juego se sale de control, pausa y contacta a soporte.',
    en: 'If gambling feels out of control, pause and contact support.',
    zh: '如果你感觉娱乐失控，请暂停并联系客服。',
  },
  contact: { es: 'Soporte PP.MX', en: 'PP.MX support', zh: 'PP.MX 客服' },
} satisfies Record<string, LocalizedText>

const exclusionOptions = [
  { id: '7d', amount: '7', label: responsibleText.days },
  { id: '1m', amount: '1', label: responsibleText.month },
  { id: '6m', amount: '6', label: responsibleText.months },
  { id: 'permanent', amount: '', label: responsibleText.permanent },
]

function applyResponsibleSetting() {
  showToast(pickText(responsibleText.applied))
}
</script>

<template>
  <main id="page-responsible" class="ppmx-page ppmx-subpage ppmx-responsible-page">
    <PpmxPageChrome>
      <section class="ppmx-subpage__compact ppmx-responsible-stack">
        <header class="ppmx-subpage-head ppmx-responsible-head reveal">
          <div class="ppmx-eyebrow">{{ pickText(responsibleText.eyebrow) }}</div>
          <h1 class="ppmx-display">{{ pickText(responsibleText.title) }}</h1>
          <p>{{ pickText(responsibleText.intro) }}</p>
        </header>

        <section class="ppmx-responsible-tools" aria-label="Responsible gaming controls">
          <article class="ppmx-responsible-card ppmx-responsible-card--limit reveal">
            <span class="ppmx-responsible-card__icon">
              <i class="fa-solid fa-sliders" />
            </span>
            <h2>{{ pickText(responsibleText.limitTitle) }}</h2>
            <p>{{ pickText(responsibleText.limitDesc) }}</p>
            <div class="ppmx-responsible-control">
              <label>
                <span>{{ pickText(responsibleText.currency) }}</span>
                <input
                  v-model="depositLimit"
                  type="number"
                  inputmode="numeric"
                  min="0"
                  :placeholder="APP_CURRENCY"
                >
              </label>
              <button class="ppmx-btn-red" type="button" @click="applyResponsibleSetting">
                {{ pickText(responsibleText.apply) }}
              </button>
            </div>
          </article>

          <article class="ppmx-responsible-card ppmx-responsible-card--time reveal">
            <span class="ppmx-responsible-card__icon">
              <i class="fa-solid fa-hourglass-half" />
            </span>
            <h2>{{ pickText(responsibleText.timeTitle) }}</h2>
            <p>{{ pickText(responsibleText.timeDesc) }}</p>
            <div class="ppmx-responsible-control">
              <label>
                <span>{{ pickText(responsibleText.timeTitle) }}</span>
                <select v-model="sessionTime">
                  <option value="30">30 min</option>
                  <option value="60">1 h</option>
                  <option value="120">2 h</option>
                </select>
              </label>
              <button class="ppmx-btn-red" type="button" @click="applyResponsibleSetting">
                {{ pickText(responsibleText.activate) }}
              </button>
            </div>
          </article>

          <article class="ppmx-responsible-card ppmx-responsible-card--exclude reveal">
            <span class="ppmx-responsible-card__icon is-danger">
              <i class="fa-solid fa-circle-stop" />
            </span>
            <h2>{{ pickText(responsibleText.excludeTitle) }}</h2>
            <p>{{ pickText(responsibleText.excludeDesc) }}</p>
            <div class="ppmx-responsible-exclusion">
              <button
                v-for="option in exclusionOptions"
                :key="option.id"
                class="ppmx-responsible-exclusion__item"
                type="button"
                @click="applyResponsibleSetting"
              >
                <span v-if="option.amount">{{ option.amount }}</span>
                {{ pickText(option.label) }}
              </button>
            </div>
          </article>
        </section>

      </section>
    </PpmxPageChrome>
  </main>
</template>
