<script setup lang="ts">
export interface PpmxSelectInputOption {
  label: string
  value: string
  disabled?: boolean
}

const modelValue = defineModel<string>({ default: '' })

const props = withDefaults(
  defineProps<{
    options: PpmxSelectInputOption[]
    placeholder?: string
    title?: string
    disabled?: boolean
  }>(),
  {
    placeholder: '',
    title: '',
    disabled: false,
  },
)

const { isPc } = useResponsiveLayout()
const isOpen = ref(false)
const selectedOption = computed(() =>
  props.options.find(option => option.value === modelValue.value) ?? null,
)
const displayText = computed(() => selectedOption.value?.label ?? modelValue.value)
const popupTitle = computed(() => props.title || props.placeholder)
const showFieldPicker = computed({
  get: () => !isPc.value && isOpen.value,
  set: (value: boolean) => {
    if (!value) isOpen.value = false
  },
})

function toggleSelect() {
  if (props.disabled || !props.options.length) return

  isOpen.value = !isOpen.value
}

function chooseSelectOption(option: PpmxSelectInputOption) {
  if (option.disabled) return

  modelValue.value = option.value
  isOpen.value = false
}

function closeSelectOnOutsidePointerDown(event: PointerEvent) {
  if (!isPc.value || !isOpen.value) return

  const target = event.target

  if (!(target instanceof Element) || !target.closest('.ppmx-select-input')) {
    isOpen.value = false
  }
}

function closeOnEscape(event: KeyboardEvent) {
  if (event.key === 'Escape') isOpen.value = false
}

onMounted(() => {
  document.addEventListener('pointerdown', closeSelectOnOutsidePointerDown)
  document.addEventListener('keydown', closeOnEscape)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', closeSelectOnOutsidePointerDown)
  document.removeEventListener('keydown', closeOnEscape)
})

watch(isPc, () => {
  isOpen.value = false
})
</script>

<template>
  <div
    class="ppmx-select-input"
    :class="{ 'is-open': isOpen, 'is-disabled': disabled }"
  >
    <button
      class="ppmx-select-input__button"
      :class="{ 'is-empty': !modelValue }"
      type="button"
      :disabled="disabled"
      @click="toggleSelect"
    >
      <span>{{ displayText || placeholder }}</span>
      <i class="fa-solid fa-chevron-down" />
    </button>

    <div
      v-if="isPc && isOpen"
      class="ppmx-select-input__menu"
    >
      <button
        v-for="option in options"
        :key="option.value"
        type="button"
        :disabled="option.disabled"
        :class="{ 'is-active': modelValue === option.value }"
        @click="chooseSelectOption(option)"
      >
        <span>{{ option.label }}</span>
        <i
          v-if="modelValue === option.value"
          class="fa-solid fa-check"
        />
      </button>
    </div>

    <van-popup
      v-if="!isPc"
      v-model:show="showFieldPicker"
      class="ppmx-select-input__popup"
      position="bottom"
      round
    >
      <section class="ppmx-select-input__popup-body">
        <h3>{{ popupTitle }}</h3>
        <div class="ppmx-select-input__popup-list">
          <button
            v-for="option in options"
            :key="option.value"
            type="button"
            :disabled="option.disabled"
            :class="{ 'is-active': modelValue === option.value }"
            @click="chooseSelectOption(option)"
          >
            <span>{{ option.label }}</span>
            <i
              v-if="modelValue === option.value"
              class="fa-solid fa-check"
            />
          </button>
        </div>
      </section>
    </van-popup>
  </div>
</template>

<style>
.ppmx-select-input {
  position: relative;
  z-index: 1;
  min-width: 0;
}

.ppmx-select-input.is-open {
  z-index: 6;
}

.ppmx-select-input__button {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  min-width: 0;
  height: 50px;
  padding: 0 14px;
  color: var(--ppmx-text, #f5f5f7);
  font: inherit;
  font-size: 0.92rem;
  text-align: left;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.055);
  border: 1px solid rgba(255, 255, 255, 0.09);
  border-radius: 14px;
  outline: none;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background 0.18s ease;
}

.ppmx-select-input__button:focus {
  border-color: rgba(240, 188, 61, 0.5);
  box-shadow: 0 0 0 3px rgba(240, 188, 61, 0.1);
}

.ppmx-select-input.is-open .ppmx-select-input__button {
  background:
    linear-gradient(180deg, rgba(240, 188, 61, 0.1), rgba(255, 255, 255, 0.055)),
    rgba(255, 255, 255, 0.055);
  border-color: rgba(240, 188, 61, 0.45);
  box-shadow: 0 0 0 3px rgba(240, 188, 61, 0.1);
}

.ppmx-select-input.is-disabled .ppmx-select-input__button {
  cursor: not-allowed;
  opacity: 0.56;
}

.ppmx-select-input__button span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ppmx-select-input__button.is-empty span {
  color: rgba(245, 245, 247, 0.32);
}

.ppmx-select-input__button i {
  flex: 0 0 auto;
  color: rgba(245, 245, 247, 0.46);
  transition: transform 0.18s ease;
}

.ppmx-select-input.is-open .ppmx-select-input__button > i {
  color: var(--ppmx-gold);
  transform: rotate(180deg);
}

.ppmx-select-input__menu {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  display: grid;
  gap: 6px;
  width: 100%;
  max-height: min(330px, 44vh);
  padding: 8px;
  overflow-y: auto;
  color: var(--ppmx-text, #f5f5f7);
  background:
    linear-gradient(180deg, rgba(30, 30, 35, 0.98), rgba(12, 12, 16, 0.98)),
    #111116;
  border: 1px solid rgba(240, 188, 61, 0.18);
  border-radius: 16px;
  box-shadow:
    0 28px 68px rgba(0, 0, 0, 0.52),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
  overscroll-behavior: contain;
}

.ppmx-select-input__menu button,
.ppmx-select-input__popup-list button {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  min-height: 40px;
  padding: 0 10px;
  color: rgba(245, 245, 247, 0.88);
  font: inherit;
  font-size: 0.88rem;
  font-weight: 800;
  text-align: left;
  cursor: pointer;
  background: transparent;
  border: 0;
  border-radius: 10px;
}

.ppmx-select-input__menu button:hover,
.ppmx-select-input__menu button.is-active,
.ppmx-select-input__popup-list button:active,
.ppmx-select-input__popup-list button.is-active {
  color: var(--ppmx-text, #f5f5f7);
  background: rgba(240, 188, 61, 0.12);
}

.ppmx-select-input__menu button:disabled,
.ppmx-select-input__popup-list button:disabled {
  cursor: not-allowed;
  opacity: 0.48;
}

.ppmx-select-input__menu button span,
.ppmx-select-input__popup-list button span {
  min-width: 0;
  overflow-wrap: anywhere;
}

.ppmx-select-input__menu button i,
.ppmx-select-input__popup-list button i {
  flex: 0 0 auto;
  color: var(--ppmx-gold);
}

.ppmx-select-input__popup {
  overflow: hidden;
  color: var(--ppmx-text, #f5f5f7);
  background:
    linear-gradient(180deg, rgba(24, 24, 28, 0.98), rgba(7, 7, 10, 0.98)),
    #09090d;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.ppmx-select-input__popup-body {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  gap: 12px;
  max-height: min(64dvh, calc(var(--app-visual-height, 100dvh) - 24px), 520px);
  padding: 20px 18px calc(20px + env(safe-area-inset-bottom));
  overflow: hidden;
}

.ppmx-select-input__popup-body h3 {
  margin: 0;
  color: #fff;
  font-size: 1rem;
  font-weight: 900;
}

.ppmx-select-input__popup-list {
  display: grid;
  gap: 10px;
  min-height: 0;
  padding-right: 2px;
  overflow-y: auto;
  overscroll-behavior: contain;
}

.ppmx-select-input__popup-list button {
  min-height: 48px;
  padding: 0 14px;
  font-size: 0.92rem;
  background: rgba(255, 255, 255, 0.055);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 14px;
}

html[data-theme="light"] .ppmx-select-input__button {
  color: #17171b;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 249, 251, 0.92)),
    #fff;
  border-color: rgba(26, 26, 32, 0.12);
  box-shadow:
    0 14px 34px -28px rgba(36, 38, 45, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.96);
}

html[data-theme="light"] .ppmx-select-input__button:focus,
html[data-theme="light"] .ppmx-select-input.is-open .ppmx-select-input__button {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 1), rgba(255, 247, 248, 0.94)),
    #fff;
  border-color: rgba(200, 16, 46, 0.28);
  box-shadow:
    0 0 0 3px rgba(255, 26, 26, 0.08),
    0 16px 34px -28px rgba(200, 16, 46, 0.35),
    inset 0 1px 0 rgba(255, 255, 255, 0.96);
}

html[data-theme="light"] .ppmx-select-input__button.is-empty span {
  color: #9a9aa2;
}

html[data-theme="light"] .ppmx-select-input__button i {
  color: rgba(23, 23, 27, 0.46);
}

html[data-theme="light"] .ppmx-select-input.is-open .ppmx-select-input__button > i {
  color: var(--ppmx-mode-accent, #ff1a1a);
}

html[data-theme="light"] .ppmx-select-input__menu {
  color: #17171b;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 249, 251, 0.98)),
    #fff;
  border-color: rgba(200, 16, 46, 0.14);
  box-shadow:
    0 24px 58px -30px rgba(36, 38, 45, 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.94);
}

html[data-theme="light"] .ppmx-select-input__menu button,
html[data-theme="light"] .ppmx-select-input__popup-list button {
  color: rgba(23, 23, 27, 0.76);
}

html[data-theme="light"] .ppmx-select-input__menu button:hover,
html[data-theme="light"] .ppmx-select-input__menu button.is-active,
html[data-theme="light"] .ppmx-select-input__popup-list button:active,
html[data-theme="light"] .ppmx-select-input__popup-list button.is-active {
  color: var(--ppmx-mode-accent, #ff1a1a);
  background: rgba(200, 16, 46, 0.08);
}

html[data-theme="light"] .ppmx-select-input__menu button i,
html[data-theme="light"] .ppmx-select-input__popup-list button i {
  color: var(--ppmx-mode-accent, #ff1a1a);
}

html[data-theme="light"] .ppmx-select-input__popup {
  color: #17171b;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 249, 251, 0.98)),
    #fff;
  border-color: rgba(26, 26, 32, 0.1);
}

html[data-theme="light"] .ppmx-select-input__popup-body h3 {
  color: #17171b;
}

html[data-theme="light"] .ppmx-select-input__popup-list button {
  background: rgba(255, 255, 255, 0.84);
  border-color: rgba(26, 26, 32, 0.1);
}

@media (max-width: 1023px) and (pointer: coarse) {
  .ppmx-select-input__button {
    font-size: 16px;
  }
}
</style>
