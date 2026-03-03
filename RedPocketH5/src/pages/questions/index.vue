<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { ComponentPublicInstance } from 'vue'
import AppPageHeader from '@/components/AppPageHeader.vue'

type CategoryKey = 'hot' | 'finance' | 'game' | 'security'

interface FaqItem {
  key: string
  question: string
  answer: string
}

const router = useRouter()
const activeTab = ref<CategoryKey>('hot')

const tabs: Array<{ key: CategoryKey, label: string, icon: string }> = [
  { key: 'hot', label: '热门问题', icon: '★' },
  { key: 'finance', label: '充值提现', icon: '💳' },
  { key: 'game', label: '游戏规则', icon: '🎮' },
  { key: 'security', label: '账户安全', icon: '👤' },
]

const faqGroups: Array<{ key: CategoryKey, title: string, items: FaqItem[] }> = [
  {
    key: 'hot',
    title: '热门问题',
    items: [
      { key: 'h1', question: '什么是红包雷游戏？', answer: '红包雷是抢红包玩法，用户设置雷号后参与发包/抢包，按规则结算收益。' },
      { key: 'h2', question: '如何充值？', answer: '进入“充值”页面选择通道和金额，提交订单后完成支付即可到账。' },
      { key: 'h3', question: '如何提现？', answer: '进入“提现”页面输入金额并提交，审核通过后将自动打款。' },
      { key: 'h4', question: '如何参与游戏？', answer: '先充值到账，然后在首页选择红包参与抢包即可。' },
      { key: 'h5', question: '什么是雷数？', answer: '雷数是用于判定中雷的尾数规则，不同红包会显示对应雷号。' },
    ],
  },
  {
    key: 'finance',
    title: '充值提现',
    items: [
      { key: 'f1', question: '支持哪些充值方式？', answer: '当前支持页面内展示的充值通道，实际可用通道以页面为准。' },
      { key: 'f2', question: '充值有最低限额吗？', answer: '有，最低充值金额按充值页的最小选项执行。' },
      { key: 'f3', question: '提现需要多长时间？', answer: '一般较快到账，若遇高峰或风控审核会有延迟。' },
      { key: 'f4', question: '提现有手续费吗？', answer: '是否收取及费率以当前提现页提示为准。' },
    ],
  },
  {
    key: 'game',
    title: '游戏规则',
    items: [
      { key: 'g1', question: '如何设置红包雷？', answer: '发红包时选择雷号即可，雷号会影响中雷判定。' },
      { key: 'g2', question: '红包金额如何分配？', answer: '系统会按规则随机拆分为多个子红包。' },
      { key: 'g3', question: '如何判断中雷？', answer: '根据子红包金额尾数是否命中雷号进行判定。' },
      { key: 'g4', question: '中雷后会发生什么？', answer: '中雷后会按当前规则扣除对应金额并结算给发包方。' },
      { key: 'g5', question: '除了中雷外，还有其他输赢规则吗？', answer: '还受平台抽成、活动规则等影响，具体以当期规则为准。' },
      { key: 'g6', question: '抢红包需要支付手续费吗？', answer: '若有手续费会在页面明示，未明示则默认按当前规则执行。' },
      { key: 'g7', question: '发红包需要支付手续费吗？', answer: '平台可能对发包收益进行抽成，具体比例以页面展示为准。' },
      { key: 'g8', question: '红包有效期是多久？', answer: '红包超时未抢完会自动结束，时间以红包卡片倒计时为准。' },
    ],
  },
  {
    key: 'security',
    title: '账户安全',
    items: [
      { key: 's1', question: '如何保护账户安全？', answer: '请勿泄露验证码、密码与登录信息，谨防钓鱼链接。' },
      { key: 's2', question: '忘记密码怎么办？', answer: '可在登录页点击“忘记密码”并通过邮箱验证码重置。' },
      { key: 's3', question: '账户被盗怎么办？', answer: '立即修改密码并联系在线客服协助处理。' },
      { key: 's4', question: '如何开启两步验证？', answer: '按平台后续支持入口开启，提升账号安全等级。' },
    ],
  },
]

const opened = reactive<Record<CategoryKey, string[]>>({
  hot: [],
  finance: [],
  game: [],
  security: [],
})

const sectionRefs = reactive<Partial<Record<CategoryKey, HTMLElement>>>({})

function setSectionRef(key: CategoryKey, el: Element | ComponentPublicInstance | null) {
  if (el instanceof HTMLElement) {
    sectionRefs[key] = el
    return
  }
  const maybeEl = (el as ComponentPublicInstance | null)?.$el
  if (maybeEl instanceof HTMLElement)
    sectionRefs[key] = maybeEl
}

function onTabClick(key: CategoryKey) {
  activeTab.value = key
  sectionRefs[key]?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function goCs() {
  router.push('/cs')
}

function goBack() {
  router.back()
}
</script>

<template>
  <div class="questions-page">
    <AppPageHeader title="常见问题" @back="goBack" />

    <div class="faq-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        class="faq-tab"
        :class="{ active: activeTab === tab.key }"
        @click="onTabClick(tab.key)"
      >
        <span class="tab-icon">{{ tab.icon }}</span>
        <span>{{ tab.label }}</span>
      </button>
    </div>

    <section
      v-for="group in faqGroups"
      :key="group.key"
      :ref="el => setSectionRef(group.key, el)"
      class="faq-section"
    >
      <h3 class="section-title">
        {{ group.title }}
      </h3>
      <van-collapse v-model="opened[group.key]" class="faq-collapse">
        <van-collapse-item
          v-for="item in group.items"
          :key="item.key"
          :name="item.key"
          :title="item.question"
          class="faq-item"
        >
          <p class="faq-answer">
            {{ item.answer }}
          </p>
        </van-collapse-item>
      </van-collapse>
    </section>

    <div class="faq-footer">
      <p>没有找到您需要的答案?</p>
      <button type="button" class="cs-btn" @click="goCs">
        联系客服
      </button>
    </div>
  </div>
</template>

<style scoped>
.questions-page {
  min-height: 100vh;
  background: #fff;
  padding: 10px 14px calc(16px + env(safe-area-inset-bottom));
}

.faq-tabs {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
  margin-bottom: 14px;
}

.faq-tab {
  height: 70px;
  border-radius: 0;
  border: 0;
  background: transparent;
  color: #0f172a;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 5px;
  font-size: 10px;
  font-weight: 700;
}

.faq-tab.active {
  background: #ff5a5f;
  color: #0b1220;
}

.tab-icon {
  font-size: 14px;
  line-height: 1;
}

.faq-section {
  margin-bottom: 14px;
  scroll-margin-top: 12px;
}

.section-title {
  margin: 0 0 8px;
  color: #0f2740;
  font-size: 13px;
  font-weight: 700;
}

:deep(.faq-collapse.van-collapse) {
  border: 0;
}

:deep(.faq-item.van-collapse-item) {
  background: #fff;
}

:deep(.faq-item .van-cell) {
  min-height: 64px;
  padding: 0 18px;
  align-items: center;
  background: #fff;
  color: #1f2937;
}

:deep(.faq-item .van-cell__title) {
  font-size: 12px;
  line-height: 1.3;
  font-weight: 500;
}

:deep(.faq-item .van-collapse-item__content) {
  padding: 0 18px 14px;
  background: #fff;
}

:deep(.faq-item .van-icon-arrow) {
  color: #a6b2c5;
  font-size: 16px;
}

.faq-answer {
  margin: 0;
  color: #6b7280;
  font-size: 11px;
  line-height: 1.5;
}

.faq-footer {
  margin: 18px 0 4px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.faq-footer p {
  margin: 0;
  color: #8c9ab0;
  font-size: 10px;
}

.cs-btn {
  width: 176px;
  height: 56px;
  border-radius: 30px;
  border: 2px solid #0f172a;
  background: transparent;
  color: #0f172a;
  font-size: 13px;
  font-weight: 700;
}
</style>

<route lang="json5">
{
  name: 'Questions'
}
</route>
