<script setup lang="ts">
import type { PageStateStatus } from '@/types/business'
import PpmxSkeleton from '@/components/PpmxSkeleton.vue'
import type { SkeletonVariant } from '@/components/PpmxSkeleton.vue'

const props = withDefaults(
  defineProps<{
    status?: PageStateStatus
    title?: string
    message?: string
    actionText?: string
    fullscreen?: boolean
    skeletonVariant?: SkeletonVariant
  }>(),
  {
    status: 'empty',
    title: '',
    message: '',
    actionText: '',
    fullscreen: false,
    skeletonVariant: undefined,
  },
)

defineEmits<{
  retry: []
}>()

const preset = computed(() => {
  const map: Record<PageStateStatus, { icon: string; titleKey: string; messageKey: string }> = {
    idle: { icon: 'info-o', titleKey: 'state.idle.title', messageKey: 'state.idle.message' },
    loading: { icon: 'underway-o', titleKey: 'state.loading.title', messageKey: 'state.loading.message' },
    ready: { icon: 'passed', titleKey: 'state.ready.title', messageKey: 'state.ready.message' },
    empty: { icon: 'orders-o', titleKey: 'state.empty.title', messageKey: 'state.empty.message' },
    network: { icon: 'warning-o', titleKey: 'state.network.title', messageKey: 'state.network.message' },
    rateLimit: { icon: 'clock-o', titleKey: 'state.rateLimit.title', messageKey: 'state.rateLimit.message' },
    maintenance: { icon: 'setting-o', titleKey: 'state.maintenance.title', messageKey: 'state.maintenance.message' },
    fallback: { icon: 'failure', titleKey: 'state.fallback.title', messageKey: 'state.fallback.message' },
    error: { icon: 'warning-o', titleKey: 'state.error.title', messageKey: 'state.error.message' },
  }

  return map[props.status]
})

const { t } = useI18n()
</script>

<template>
  <section
    class="flex flex-col items-center justify-center px-6 py-10 text-center"
    :class="fullscreen ? 'min-h-screen bg-gray-50' : 'min-h-48'"
  >
    <PpmxSkeleton
      v-if="status === 'loading'"
      :variant="skeletonVariant || 'page'"
      :aria-label="title || t(preset.titleKey)"
    />
    <van-icon v-else :name="preset.icon" size="44" class="text-gray-400" />

    <h2 v-if="status !== 'loading'" class="mt-4 text-base font-semibold text-gray-950">
      {{ title || t(preset.titleKey) }}
    </h2>
    <p v-if="status !== 'loading'" class="mt-2 max-w-72 text-sm leading-6 text-gray-500">
      {{ message || t(preset.messageKey) }}
    </p>

    <van-button
      v-if="!['loading', 'ready', 'maintenance'].includes(status)"
      class="mt-5"
      size="small"
      type="primary"
      @click="$emit('retry')"
    >
      {{ actionText || t('common.retry') }}
    </van-button>
  </section>
</template>
