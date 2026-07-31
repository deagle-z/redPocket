<script setup lang="ts">
import type { CSSProperties } from 'vue'

import type { AppPopupAnnouncement } from '@/api/banner'

defineOptions({ name: 'PpmxAnnouncementPopup' })

const props = defineProps<{
  announcement: AppPopupAnnouncement
  closeText: string
  confirmText: string
  modelValue: boolean
}>()

const emit = defineEmits<{
  close: []
  confirm: []
}>()

const cardStyle = computed<CSSProperties>(() => ({
  background: props.announcement.bgColor || undefined,
  color: props.announcement.textColor || undefined,
}))

const buttonStyle = computed<CSSProperties>(() => ({
  background: props.announcement.buttonColor || undefined,
}))

const displayTitle = computed(
  () => props.announcement.title?.trim() || props.announcement.bannerName,
)

const displayImage = computed(
  () =>
    props.announcement.imageUrl?.trim() ||
    props.announcement.bgImageUrl?.trim() ||
    props.announcement.iconUrl?.trim() ||
    '',
)
</script>

<template>
  <Teleport to="body">
    <Transition name="ppmx-announcement-fade">
      <div
        v-if="modelValue"
        class="ppmx-announcement"
        role="dialog"
        aria-modal="true"
        :aria-label="displayTitle"
      >
        <button
          class="ppmx-announcement__backdrop"
          type="button"
          :aria-label="closeText"
          @click="emit('close')"
        />

        <article class="ppmx-announcement__card" :style="cardStyle">
          <button
            class="ppmx-announcement__close"
            type="button"
            :aria-label="closeText"
            @click="emit('close')"
          >
            <i class="fa-solid fa-xmark" />
          </button>

          <img
            v-if="displayImage"
            class="ppmx-announcement__image"
            :src="displayImage"
            :alt="displayTitle"
          >

          <div class="ppmx-announcement__content">
            <span v-if="announcement.subTitle" class="ppmx-announcement__eyebrow">
              {{ announcement.subTitle }}
            </span>
            <h2>{{ displayTitle }}</h2>
            <p v-if="announcement.description">
              {{ announcement.description }}
            </p>
          </div>

          <div class="ppmx-announcement__actions">
            <button
              class="ppmx-announcement__button ppmx-announcement__button--close"
              type="button"
              @click="emit('close')"
            >
              {{ closeText }}
            </button>
            <button
              class="ppmx-announcement__button ppmx-announcement__button--confirm"
              type="button"
              :style="buttonStyle"
              @click="emit('confirm')"
            >
              {{ announcement.buttonText?.trim() || confirmText }}
            </button>
          </div>
        </article>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.ppmx-announcement {
  position: fixed;
  inset: 0;
  z-index: 10020;
  display: grid;
  min-height: var(--app-visual-height, 100dvh);
  padding: 20px;
  place-items: center;
  color: #f7f7f8;
  font-family: "Montserrat", system-ui, -apple-system, "Segoe UI", sans-serif;
}

.ppmx-announcement__backdrop {
  position: absolute;
  inset: 0;
  cursor: default;
  background: rgba(0, 0, 0, 0.78);
  border: 0;
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
}

.ppmx-announcement__card {
  position: relative;
  width: min(100%, 420px);
  max-height: calc(var(--app-visual-height, 100dvh) - 32px);
  overflow: auto;
  background:
    radial-gradient(circle at 85% 0%, rgba(255, 215, 0, 0.15), transparent 38%),
    linear-gradient(145deg, #17171c, #09090c);
  border: 1px solid rgba(255, 255, 255, 0.13);
  border-radius: 8px;
  box-shadow: 0 30px 90px rgba(0, 0, 0, 0.76);
}

.ppmx-announcement__close {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 1;
  display: grid;
  width: 36px;
  height: 36px;
  color: #fff;
  cursor: pointer;
  background: rgba(0, 0, 0, 0.55);
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 50%;
  place-items: center;
}

.ppmx-announcement__image {
  display: block;
  width: 100%;
  max-height: min(44dvh, 360px);
  object-fit: cover;
  background: rgba(255, 255, 255, 0.04);
}

.ppmx-announcement__content {
  display: grid;
  gap: 10px;
  padding: 24px 24px 18px;
}

.ppmx-announcement__eyebrow {
  color: #f1c75b;
  font-size: 0.76rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  line-height: 1.4;
  text-transform: uppercase;
}

.ppmx-announcement__content h2 {
  margin: 0;
  color: inherit;
  font-size: clamp(1.35rem, 5vw, 1.85rem);
  font-weight: 900;
  line-height: 1.15;
}

.ppmx-announcement__content p {
  margin: 0;
  color: inherit;
  font-size: 0.92rem;
  font-weight: 600;
  line-height: 1.65;
  opacity: 0.72;
  white-space: pre-line;
}

.ppmx-announcement__actions {
  display: grid;
  grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.2fr);
  gap: 10px;
  padding: 0 24px 24px;
}

.ppmx-announcement__button {
  min-height: 48px;
  padding: 10px 14px;
  font: inherit;
  font-size: 0.9rem;
  font-weight: 900;
  cursor: pointer;
  border-radius: 8px;
}

.ppmx-announcement__button--close {
  color: rgba(255, 255, 255, 0.78);
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.13);
}

.ppmx-announcement__button--confirm {
  color: #211403;
  background: linear-gradient(180deg, #ffe889, #efbc3e);
  border: 1px solid rgba(255, 255, 255, 0.18);
}

.ppmx-announcement__button:active {
  transform: translateY(1px);
}

.ppmx-announcement-fade-enter-active,
.ppmx-announcement-fade-leave-active {
  transition: opacity 0.2s ease;
}

.ppmx-announcement-fade-enter-active .ppmx-announcement__card,
.ppmx-announcement-fade-leave-active .ppmx-announcement__card {
  transition: opacity 0.2s ease, transform 0.24s ease;
}

.ppmx-announcement-fade-enter-from,
.ppmx-announcement-fade-leave-to {
  opacity: 0;
}

.ppmx-announcement-fade-enter-from .ppmx-announcement__card,
.ppmx-announcement-fade-leave-to .ppmx-announcement__card {
  opacity: 0;
  transform: translateY(12px) scale(0.98);
}

@media (max-width: 480px) {
  .ppmx-announcement {
    padding: 16px;
  }

  .ppmx-announcement__content {
    padding: 22px 20px 16px;
  }

  .ppmx-announcement__actions {
    grid-template-columns: 1fr;
    padding: 0 20px 20px;
  }
}
</style>
