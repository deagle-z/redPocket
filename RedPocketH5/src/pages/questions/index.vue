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
const { t } = useI18n()
const activeTab = ref<CategoryKey>('hot')

const tabs = computed<Array<{ key: CategoryKey, label: string, icon: string }>>(() => [
  { key: 'hot', label: t('questionsPage.tabHot'), icon: '★' },
  { key: 'finance', label: t('questionsPage.tabFinance'), icon: '💳' },
  { key: 'game', label: t('questionsPage.tabGame'), icon: '🎮' },
  { key: 'security', label: t('questionsPage.tabSecurity'), icon: '👤' },
])

function buildFaqItem(group: CategoryKey, key: string): FaqItem {
  return {
    key,
    question: t(`questionsPage.${group}.${key}.q`),
    answer: t(`questionsPage.${group}.${key}.a`),
  }
}

const faqGroups = computed<Array<{ key: CategoryKey, title: string, items: FaqItem[] }>>(() => [
  {
    key: 'hot',
    title: t('questionsPage.tabHot'),
    items: ['h1', 'h2', 'h3', 'h4', 'h5'].map(id => buildFaqItem('hot', id)),
  },
  {
    key: 'finance',
    title: t('questionsPage.tabFinance'),
    items: ['f1', 'f2', 'f3', 'f4'].map(id => buildFaqItem('finance', id)),
  },
  {
    key: 'game',
    title: t('questionsPage.tabGame'),
    items: ['g1', 'g2', 'g3', 'g4', 'g5', 'g6', 'g7', 'g8'].map(id => buildFaqItem('game', id)),
  },
  {
    key: 'security',
    title: t('questionsPage.tabSecurity'),
    items: ['s1', 's2', 's3', 's4'].map(id => buildFaqItem('security', id)),
  },
])

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
    <AppPageHeader :title="t('questionsPage.title')" @back="goBack" />

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
      <p>{{ t('questionsPage.footerHint') }}</p>
      <button type="button" class="cs-btn" @click="goCs">
        {{ t('questionsPage.contactCs') }}
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
