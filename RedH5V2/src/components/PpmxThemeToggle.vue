<script setup lang="ts">
import { usePpmxTheme } from '@/composables/usePpmxTheme'

defineOptions({ name: 'PpmxThemeToggle' })

withDefaults(defineProps<{
  iconOnly?: boolean
  block?: boolean
  showToggleLabel?: boolean
}>(), {
  iconOnly: false,
  block: false,
  showToggleLabel: false,
})

const {
  theme,
  themeIconClass,
  themeLabel,
  themeToggleLabel,
  toggleTheme,
} = usePpmxTheme()
</script>

<template>
  <button
    class="ppmx-theme-toggle theme-btn"
    :class="{
      'is-light': theme === 'light',
      'is-icon-only': iconOnly,
      'is-block': block,
    }"
    type="button"
    :aria-label="themeToggleLabel"
    @click="toggleTheme"
  >
    <span class="ppmx-theme-toggle__icon tb-ico" aria-hidden="true">
      <i class="fa-solid theme-ico" :class="themeIconClass" />
    </span>
    <span v-if="showToggleLabel && !iconOnly" class="ppmx-theme-toggle__label">
      {{ themeToggleLabel }}
    </span>
    <span v-if="!iconOnly" class="ppmx-theme-toggle__state tb-state">
      {{ themeLabel }}
    </span>
  </button>
</template>
