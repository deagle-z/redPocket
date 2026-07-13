<script setup lang="ts">
import PpmxPageChrome from '@/pages/ppmx-home/components/PpmxPageChrome.vue'

const props = withDefaults(defineProps<{
  centered?: boolean
  description?: string
  eyebrow?: string
  headClass?: string
  pageId?: string
  pageClass?: string
  stackClass?: string
  title?: string
  width?: 'narrow' | 'compact' | 'wide'
}>(), {
  centered: false,
  description: '',
  eyebrow: '',
  headClass: '',
  pageId: '',
  pageClass: '',
  stackClass: '',
  title: '',
  width: 'compact',
})
</script>

<template>
  <main :id="props.pageId || undefined" class="ppmx-page ppmx-subpage" :class="props.pageClass">
    <PpmxPageChrome>
      <section :class="[`ppmx-subpage__${props.width}`, props.stackClass]">
        <header
          v-if="props.eyebrow || props.title || props.description || $slots.head"
          class="ppmx-subpage-head"
          :class="[props.headClass, { 'ppmx-subpage-head--center': props.centered }]"
        >
          <slot name="head">
            <div v-if="props.eyebrow" class="ppmx-eyebrow">{{ props.eyebrow }}</div>
            <h1 v-if="props.title" class="ppmx-display">{{ props.title }}</h1>
            <p v-if="props.description">{{ props.description }}</p>
          </slot>
        </header>

        <slot />
      </section>
    </PpmxPageChrome>
  </main>
</template>
