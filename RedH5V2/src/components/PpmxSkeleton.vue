<script setup lang="ts">
defineOptions({ name: 'PpmxSkeleton' })

export type SkeletonVariant =
  | 'page'
  | 'card-list'
  | 'game-grid'
  | 'form'
  | 'modal'
  | 'checkin'
  | 'play'
  | 'inline'
  | 'home-games'
  | 'profile-page'
  | 'bank-page'
  | 'recharge-page'
  | 'team-page'
  | 'casino-page'
  | 'records-page'
  | 'wheel-page'
  | 'user-profile-page'
  | 'checkin-modal'
  | 'support-modal'
  | 'withdraw-modal'
  | 'commission-withdraw-modal'
  | 'bank-account-list'
  | 'casino-more'
  | 'play-launch'
  | 'play-frame'

const props = withDefaults(
  defineProps<{
    variant?: SkeletonVariant
    items?: number
    rows?: number
    compact?: boolean
    ariaLabel?: string
  }>(),
  {
    variant: 'page',
    items: 3,
    rows: 3,
    compact: false,
    ariaLabel: 'Loading',
  },
)

const itemCount = computed(() => Math.max(1, props.items))
const rowCount = computed(() => Math.max(1, props.rows))

const semanticPageVariants = new Set<SkeletonVariant>([
  'home-games',
  'profile-page',
  'bank-page',
  'recharge-page',
  'team-page',
  'casino-page',
  'records-page',
  'wheel-page',
  'user-profile-page',
])

const semanticModalVariants = new Set<SkeletonVariant>([
  'checkin-modal',
  'support-modal',
  'withdraw-modal',
  'commission-withdraw-modal',
  'bank-account-list',
  'casino-more',
  'play-launch',
  'play-frame',
])

const isSemanticPageVariant = computed(() => semanticPageVariants.has(props.variant))
const isSemanticModalVariant = computed(() => semanticModalVariants.has(props.variant))
</script>

<template>
  <section
    class="ppmx-skeleton"
    :class="[`ppmx-skeleton--${variant}`, { 'is-compact': compact }]"
    :data-variant="variant"
    role="status"
    aria-live="polite"
    aria-busy="true"
    :aria-label="ariaLabel"
  >
    <span class="ppmx-skeleton__sr">{{ ariaLabel }}</span>

    <template v-if="isSemanticPageVariant">
      <div class="ppmx-skeleton__page-shell">
        <div class="ppmx-skeleton__page-head">
          <span class="ppmx-skeleton-shimmer is-eyebrow" />
          <span class="ppmx-skeleton-shimmer is-title" />
          <span class="ppmx-skeleton-shimmer is-text" />
        </div>

        <template v-if="variant === 'home-games' || variant === 'casino-page'">
          <div class="ppmx-skeleton__category-strip">
            <span
              v-for="item in 5"
              :key="`cat-${item}`"
              class="ppmx-skeleton-shimmer is-chip"
            />
          </div>
          <div class="ppmx-skeleton__grid">
            <article
              v-for="item in itemCount"
              :key="`home-game-${item}`"
              class="ppmx-skeleton__game"
            >
              <span class="ppmx-skeleton-shimmer is-thumb" />
              <span class="ppmx-skeleton-shimmer is-line" />
              <span class="ppmx-skeleton-shimmer is-line is-short" />
            </article>
          </div>
        </template>

        <template v-else-if="variant === 'profile-page'">
          <div class="ppmx-skeleton__account-card">
            <span class="ppmx-skeleton-shimmer is-avatar" />
            <span class="ppmx-skeleton-shimmer is-title" />
            <span class="ppmx-skeleton-shimmer is-text" />
            <div class="ppmx-skeleton__action-row">
              <span class="ppmx-skeleton-shimmer is-button" />
              <span class="ppmx-skeleton-shimmer is-button" />
              <span class="ppmx-skeleton-shimmer is-button" />
            </div>
          </div>
          <div class="ppmx-skeleton__cards">
            <article
              v-for="item in 4"
              :key="`profile-card-${item}`"
              class="ppmx-skeleton__card"
            >
              <span class="ppmx-skeleton-shimmer is-icon" />
              <span class="ppmx-skeleton-shimmer is-line" />
              <span class="ppmx-skeleton-shimmer is-line is-short" />
            </article>
          </div>
        </template>

        <template v-else-if="variant === 'bank-page'">
          <div class="ppmx-skeleton__country-row">
            <span class="ppmx-skeleton-shimmer is-chip" />
            <span class="ppmx-skeleton-shimmer is-mini" />
          </div>
          <article
            v-for="item in 2"
            :key="`bank-page-${item}`"
            class="ppmx-skeleton__card is-wide"
          >
            <span class="ppmx-skeleton-shimmer is-icon" />
            <div>
              <span class="ppmx-skeleton-shimmer is-line" />
              <span class="ppmx-skeleton-shimmer is-line is-short" />
            </div>
          </article>
          <div class="ppmx-skeleton__form-panel">
            <span class="ppmx-skeleton-shimmer is-label" />
            <span class="ppmx-skeleton-shimmer is-control" />
            <span class="ppmx-skeleton-shimmer is-label" />
            <span class="ppmx-skeleton-shimmer is-control" />
          </div>
        </template>

        <template v-else-if="variant === 'recharge-page'">
          <div class="ppmx-skeleton__split">
            <div class="ppmx-skeleton__balance-panel">
              <span class="ppmx-skeleton-shimmer is-icon" />
              <span class="ppmx-skeleton-shimmer is-title" />
              <span class="ppmx-skeleton-shimmer is-text" />
            </div>
            <div class="ppmx-skeleton__form-panel">
              <span class="ppmx-skeleton-shimmer is-label" />
              <span class="ppmx-skeleton-shimmer is-control" />
              <div class="ppmx-skeleton__amount-grid">
                <span
                  v-for="item in 6"
                  :key="`amount-${item}`"
                  class="ppmx-skeleton-shimmer is-amount"
                />
              </div>
              <span class="ppmx-skeleton-shimmer is-button" />
            </div>
          </div>
        </template>

        <template v-else-if="variant === 'team-page'">
          <div class="ppmx-skeleton__metric-row">
            <article
              v-for="item in 3"
              :key="`team-metric-${item}`"
              class="ppmx-skeleton__metric"
            >
              <span class="ppmx-skeleton-shimmer is-icon" />
              <span class="ppmx-skeleton-shimmer is-title" />
              <span class="ppmx-skeleton-shimmer is-line is-short" />
            </article>
          </div>
          <div class="ppmx-skeleton__invite-panel">
            <span class="ppmx-skeleton-shimmer is-title" />
            <span class="ppmx-skeleton-shimmer is-control" />
            <span class="ppmx-skeleton-shimmer is-qr" />
          </div>
        </template>

        <template v-else-if="variant === 'records-page'">
          <div class="ppmx-skeleton__filter-row">
            <span
              v-for="item in 4"
              :key="`filter-${item}`"
              class="ppmx-skeleton-shimmer is-chip"
            />
          </div>
          <article
            v-for="item in itemCount"
            :key="`record-${item}`"
            class="ppmx-skeleton__record-row"
          >
            <span class="ppmx-skeleton-shimmer is-icon" />
            <span class="ppmx-skeleton-shimmer is-line" />
            <span class="ppmx-skeleton-shimmer is-mini" />
          </article>
        </template>

        <template v-else-if="variant === 'wheel-page'">
          <div class="ppmx-skeleton__wheel-stage">
            <span class="ppmx-skeleton-shimmer is-wheel" />
            <span class="ppmx-skeleton-shimmer is-button" />
          </div>
          <div class="ppmx-skeleton__cards">
            <article
              v-for="item in 3"
              :key="`wheel-card-${item}`"
              class="ppmx-skeleton__card"
            >
              <span class="ppmx-skeleton-shimmer is-line" />
              <span class="ppmx-skeleton-shimmer is-line is-short" />
            </article>
          </div>
        </template>

        <template v-else>
          <div class="ppmx-skeleton__account-card">
            <span class="ppmx-skeleton-shimmer is-avatar" />
            <span class="ppmx-skeleton-shimmer is-title" />
            <span class="ppmx-skeleton-shimmer is-text" />
          </div>
          <div class="ppmx-skeleton__form-panel">
            <span
              v-for="row in 4"
              :key="`user-profile-${row}`"
              class="ppmx-skeleton-shimmer is-control"
            />
          </div>
        </template>
      </div>
    </template>

    <template v-else-if="isSemanticModalVariant">
      <div class="ppmx-skeleton__modal-shell">
        <template v-if="variant === 'checkin-modal'">
          <span class="ppmx-skeleton-shimmer is-modal-icon" />
          <span class="ppmx-skeleton-shimmer is-title" />
          <div class="ppmx-skeleton__checkin">
            <article
              v-for="item in itemCount"
              :key="`checkin-modal-${item}`"
              class="ppmx-skeleton__checkin-tile"
            >
              <span class="ppmx-skeleton-shimmer is-line is-short" />
              <span class="ppmx-skeleton-shimmer is-coin" />
              <span class="ppmx-skeleton-shimmer is-line" />
            </article>
          </div>
        </template>

        <template v-else-if="variant === 'support-modal'">
          <article
            v-for="item in 2"
            :key="`support-channel-${item}`"
            class="ppmx-skeleton__support-channel"
          >
            <span class="ppmx-skeleton-shimmer is-icon" />
            <span class="ppmx-skeleton-shimmer is-line" />
            <span class="ppmx-skeleton-shimmer is-mini" />
          </article>
          <span class="ppmx-skeleton-shimmer is-support-safe" />
        </template>

        <template v-else-if="variant === 'withdraw-modal' || variant === 'commission-withdraw-modal'">
          <div v-if="compact && rowCount > 1" class="ppmx-skeleton__withdraw-balance-compact">
            <span class="ppmx-skeleton-shimmer is-label" />
            <span class="ppmx-skeleton-shimmer is-amount-inline" />
          </div>
          <div v-else-if="compact" class="ppmx-skeleton__withdraw-compact">
            <span
              v-for="row in rowCount"
              :key="`withdraw-compact-${row}`"
              class="ppmx-skeleton-shimmer"
              :class="row === 1 && rowCount > 1 ? 'is-label' : 'is-control'"
            />
          </div>
          <template v-else>
            <div class="ppmx-skeleton__withdraw-summary">
              <span class="ppmx-skeleton-shimmer is-label" />
              <span class="ppmx-skeleton-shimmer is-title" />
            </div>
            <span class="ppmx-skeleton-shimmer is-control" />
            <span class="ppmx-skeleton-shimmer is-control" />
            <div class="ppmx-skeleton__fee-panel">
              <span
                v-for="row in 3"
                :key="`withdraw-fee-${row}`"
                class="ppmx-skeleton-shimmer is-line"
              />
            </div>
          </template>
        </template>

        <template v-else-if="variant === 'bank-account-list'">
          <article
            v-for="item in itemCount"
            :key="`bank-list-${item}`"
            class="ppmx-skeleton__card is-wide"
          >
            <span class="ppmx-skeleton-shimmer is-icon" />
            <div>
              <span class="ppmx-skeleton-shimmer is-line" />
              <span class="ppmx-skeleton-shimmer is-line is-short" />
            </div>
          </article>
        </template>

        <template v-else-if="variant === 'casino-more'">
          <div class="ppmx-skeleton__grid">
            <article
              v-for="item in itemCount"
              :key="`casino-more-${item}`"
              class="ppmx-skeleton__game"
            >
              <span class="ppmx-skeleton-shimmer is-thumb" />
              <span class="ppmx-skeleton-shimmer is-line" />
            </article>
          </div>
        </template>

        <template v-else-if="variant === 'play-launch' || variant === 'play-frame'">
          <div class="ppmx-skeleton__play-room">
            <div class="ppmx-skeleton__play-room-bar">
              <span class="ppmx-skeleton-shimmer is-mini" />
              <span class="ppmx-skeleton-shimmer is-chip" />
              <span class="ppmx-skeleton-shimmer is-chip" />
            </div>
            <span class="ppmx-skeleton-shimmer is-frame" />
            <div class="ppmx-skeleton__play-room-status">
              <span class="ppmx-skeleton-shimmer is-modal-icon" />
              <span class="ppmx-skeleton-shimmer is-title" />
              <span class="ppmx-skeleton-shimmer is-line is-short" />
            </div>
          </div>
        </template>

        <template v-else>
          <div class="ppmx-skeleton__play">
            <span class="ppmx-skeleton-shimmer is-title" />
            <span class="ppmx-skeleton-shimmer is-frame" />
          </div>
        </template>
      </div>
    </template>

    <template v-else-if="variant === 'page'">
      <div class="ppmx-skeleton__hero">
        <span class="ppmx-skeleton-shimmer is-eyebrow" />
        <span class="ppmx-skeleton-shimmer is-title" />
        <span class="ppmx-skeleton-shimmer is-text" />
      </div>
      <div class="ppmx-skeleton__cards">
        <article
          v-for="item in itemCount"
          :key="`page-${item}`"
          class="ppmx-skeleton__card"
        >
          <span class="ppmx-skeleton-shimmer is-icon" />
          <span class="ppmx-skeleton-shimmer is-line" />
          <span class="ppmx-skeleton-shimmer is-line is-short" />
        </article>
      </div>
    </template>

    <template v-else-if="variant === 'card-list'">
      <article
        v-for="item in itemCount"
        :key="`card-${item}`"
        class="ppmx-skeleton__card is-wide"
      >
        <span class="ppmx-skeleton-shimmer is-icon" />
        <div>
          <span class="ppmx-skeleton-shimmer is-line" />
          <span class="ppmx-skeleton-shimmer is-line is-short" />
        </div>
      </article>
    </template>

    <template v-else-if="variant === 'game-grid'">
      <div class="ppmx-skeleton__grid">
        <article
          v-for="item in itemCount"
          :key="`game-${item}`"
          class="ppmx-skeleton__game"
        >
          <span class="ppmx-skeleton-shimmer is-thumb" />
          <span class="ppmx-skeleton-shimmer is-line" />
          <span class="ppmx-skeleton-shimmer is-line is-short" />
        </article>
      </div>
    </template>

    <template v-else-if="variant === 'form'">
      <div
        v-for="row in rowCount"
        :key="`form-${row}`"
        class="ppmx-skeleton__field"
      >
        <span class="ppmx-skeleton-shimmer is-label" />
        <span class="ppmx-skeleton-shimmer is-control" />
      </div>
    </template>

    <template v-else-if="variant === 'modal'">
      <div class="ppmx-skeleton__modal">
        <span class="ppmx-skeleton-shimmer is-title" />
        <span class="ppmx-skeleton-shimmer is-control" />
        <span class="ppmx-skeleton-shimmer is-control" />
      </div>
    </template>

    <template v-else-if="variant === 'checkin'">
      <div class="ppmx-skeleton__checkin">
        <article
          v-for="item in itemCount"
          :key="`checkin-${item}`"
          class="ppmx-skeleton__checkin-tile"
        >
          <span class="ppmx-skeleton-shimmer is-line is-short" />
          <span class="ppmx-skeleton-shimmer is-coin" />
          <span class="ppmx-skeleton-shimmer is-line" />
        </article>
      </div>
    </template>

    <template v-else-if="variant === 'play'">
      <div class="ppmx-skeleton__play">
        <span class="ppmx-skeleton-shimmer is-title" />
        <span class="ppmx-skeleton-shimmer is-frame" />
      </div>
    </template>

    <template v-else>
      <div class="ppmx-skeleton__inline">
        <span
          v-for="row in rowCount"
          :key="`inline-${row}`"
          class="ppmx-skeleton-shimmer is-line"
          :class="{ 'is-short': row === rowCount }"
        />
      </div>
    </template>
  </section>
</template>

<style scoped>
.ppmx-skeleton {
  --skeleton-base: rgba(255, 255, 255, 0.05);
  --skeleton-strong: rgba(255, 255, 255, 0.105);
  --skeleton-panel: rgba(255, 255, 255, 0.035);
  --skeleton-panel-strong: rgba(255, 255, 255, 0.058);
  --skeleton-border: rgba(255, 255, 255, 0.082);
  --skeleton-gold: rgba(245, 190, 67, 0.12);
  --skeleton-red: rgba(205, 42, 55, 0.075);
  --skeleton-sheen: rgba(255, 222, 135, 0.24);
  --skeleton-shadow: rgba(0, 0, 0, 0.22);
  display: grid;
  gap: 14px;
  width: 100%;
  min-width: 0;
  color: var(--ppmx-text, var(--app-text, #f5f5f7));
}

.ppmx-skeleton.is-compact {
  gap: 10px;
}

.ppmx-skeleton--home-games {
  margin-top: 12px;
}

.ppmx-skeleton__sr {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.ppmx-skeleton-shimmer {
  position: relative;
  display: block;
  overflow: hidden;
  background:
    radial-gradient(circle at 18% 0%, var(--skeleton-gold), transparent 32%),
    linear-gradient(145deg, var(--skeleton-red), transparent 46%),
    linear-gradient(180deg, var(--skeleton-strong), var(--skeleton-base));
  border: 1px solid var(--skeleton-border);
  border-radius: 8px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.07),
    inset 0 -1px 0 rgba(0, 0, 0, 0.18),
    0 10px 24px var(--skeleton-shadow);
}

.ppmx-skeleton-shimmer::after {
  position: absolute;
  inset: 0;
  content: "";
  background:
    linear-gradient(
      108deg,
      transparent 0%,
      transparent 34%,
      rgba(255, 255, 255, 0.055) 42%,
      var(--skeleton-sheen) 49%,
      rgba(255, 255, 255, 0.075) 57%,
      transparent 70%
    );
  transform: translateX(-120%);
  animation: ppmxSkeletonShimmer 1.35s ease-in-out infinite;
}

.ppmx-skeleton__hero,
.ppmx-skeleton__card,
.ppmx-skeleton__modal,
.ppmx-skeleton__page-head,
.ppmx-skeleton__account-card,
.ppmx-skeleton__balance-panel,
.ppmx-skeleton__form-panel,
.ppmx-skeleton__metric,
.ppmx-skeleton__invite-panel,
.ppmx-skeleton__record-row,
.ppmx-skeleton__modal-shell,
.ppmx-skeleton__support-channel,
.ppmx-skeleton__fee-panel,
.ppmx-skeleton__withdraw-summary {
  background:
    radial-gradient(circle at 92% 0%, var(--skeleton-gold), transparent 32%),
    radial-gradient(circle at 0% 100%, var(--skeleton-red), transparent 38%),
    linear-gradient(180deg, var(--skeleton-panel-strong), var(--skeleton-panel));
  border: 1px solid var(--skeleton-border);
  border-radius: 18px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.055),
    0 18px 44px rgba(0, 0, 0, 0.18);
}

.ppmx-skeleton__hero {
  display: grid;
  gap: 12px;
  min-height: 152px;
  padding: 20px;
}

.ppmx-skeleton__cards {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.ppmx-skeleton__page-shell {
  display: grid;
  gap: 14px;
  width: 100%;
}

.ppmx-skeleton__page-head {
  display: grid;
  gap: 11px;
  min-height: 128px;
  padding: 18px;
}

.ppmx-skeleton__category-strip,
.ppmx-skeleton__filter-row,
.ppmx-skeleton__country-row,
.ppmx-skeleton__action-row,
.ppmx-skeleton__amount-grid,
.ppmx-skeleton__metric-row {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.ppmx-skeleton__category-strip {
  grid-template-columns: repeat(5, minmax(0, 1fr));
}

.ppmx-skeleton__filter-row {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.ppmx-skeleton__country-row {
  grid-template-columns: minmax(0, 1fr) 72px;
  align-items: center;
}

.ppmx-skeleton__action-row,
.ppmx-skeleton__amount-grid,
.ppmx-skeleton__metric-row {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.ppmx-skeleton__account-card,
.ppmx-skeleton__balance-panel,
.ppmx-skeleton__form-panel,
.ppmx-skeleton__metric,
.ppmx-skeleton__invite-panel,
.ppmx-skeleton__withdraw-summary,
.ppmx-skeleton__fee-panel {
  display: grid;
  gap: 12px;
  padding: 16px;
}

.ppmx-skeleton__account-card {
  min-height: 230px;
}

.ppmx-skeleton__split {
  display: grid;
  grid-template-columns: minmax(220px, 0.8fr) minmax(0, 1.2fr);
  gap: 14px;
}

.ppmx-skeleton__balance-panel {
  align-content: start;
  min-height: 220px;
}

.ppmx-skeleton__invite-panel {
  grid-template-columns: minmax(0, 1fr) 112px;
  align-items: center;
}

.ppmx-skeleton__invite-panel .is-title,
.ppmx-skeleton__invite-panel .is-control {
  grid-column: 1;
}

.ppmx-skeleton__invite-panel .is-qr {
  grid-row: 1 / span 2;
  grid-column: 2;
}

.ppmx-skeleton__record-row,
.ppmx-skeleton__support-channel {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) 72px;
  align-items: center;
  gap: 12px;
  min-height: 74px;
  padding: 14px;
}

.ppmx-skeleton__wheel-stage {
  display: grid;
  place-items: center;
  gap: 16px;
  min-height: 320px;
  padding: 18px;
  background:
    radial-gradient(circle, rgba(245, 190, 67, 0.14), transparent 58%),
    radial-gradient(circle at 50% 100%, rgba(205, 42, 55, 0.08), transparent 48%),
    linear-gradient(180deg, var(--skeleton-panel-strong), var(--skeleton-panel));
  border: 1px solid var(--skeleton-border);
  border-radius: 24px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.055),
    0 20px 50px rgba(0, 0, 0, 0.18);
}

.ppmx-skeleton__modal-shell {
  display: grid;
  gap: 12px;
  width: 100%;
  padding: 14px;
}

.ppmx-skeleton__withdraw-summary {
  min-height: 92px;
}

.ppmx-skeleton__card {
  display: grid;
  gap: 10px;
  min-height: 112px;
  padding: 16px;
}

.ppmx-skeleton__card.is-wide {
  grid-template-columns: auto minmax(0, 1fr);
  align-items: center;
  min-height: 88px;
}

.ppmx-skeleton__card.is-wide > div {
  display: grid;
  gap: 10px;
}

.ppmx-skeleton__grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.ppmx-skeleton__game,
.ppmx-skeleton__checkin-tile,
.ppmx-skeleton__field {
  display: grid;
  gap: 8px;
  min-width: 0;
}

.ppmx-skeleton__field {
  gap: 9px;
}

.ppmx-skeleton__modal {
  display: grid;
  gap: 12px;
  padding: 14px;
}

.ppmx-skeleton__checkin {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.ppmx-skeleton__checkin-tile {
  min-height: 126px;
  padding: 12px;
  background:
    radial-gradient(circle at 50% 0%, var(--skeleton-gold), transparent 46%),
    linear-gradient(180deg, var(--skeleton-panel-strong), var(--skeleton-panel));
  border: 1px solid var(--skeleton-border);
  border-radius: 16px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.055);
}

.ppmx-skeleton__play {
  display: grid;
  gap: 16px;
  width: min(100%, 760px);
  margin: auto;
}

.ppmx-skeleton__play-room {
  position: relative;
  display: grid;
  gap: 14px;
  width: 100%;
  height: 100%;
  min-height: min(68vh, 640px);
  padding: clamp(14px, 2.6vw, 22px);
  overflow: hidden;
  background:
    radial-gradient(circle at 50% 16%, var(--skeleton-gold), transparent 34%),
    radial-gradient(circle at 0% 100%, var(--skeleton-red), transparent 42%),
    linear-gradient(180deg, var(--skeleton-panel-strong), var(--skeleton-panel));
  border: 1px solid var(--skeleton-border);
  border-radius: 24px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.055),
    0 20px 50px rgba(0, 0, 0, 0.18);
}

.ppmx-skeleton__play-room-bar {
  display: grid;
  grid-template-columns: minmax(72px, 0.28fr) repeat(2, minmax(86px, 0.18fr));
  gap: 10px;
  align-items: center;
}

.ppmx-skeleton__play-room-bar .is-chip {
  height: 36px;
}

.ppmx-skeleton__play-room .is-frame {
  min-height: 0;
  height: 100%;
  border-radius: 20px;
}

.ppmx-skeleton__play-room-status {
  position: absolute;
  top: 50%;
  left: 50%;
  display: grid;
  justify-items: center;
  width: min(72%, 360px);
  gap: 12px;
  transform: translate(-50%, -50%);
}

.ppmx-skeleton__inline {
  display: grid;
  gap: 8px;
}

.is-eyebrow,
.is-label {
  width: 34%;
  height: 12px;
  border-radius: 999px;
}

.is-title {
  width: min(72%, 420px);
  height: 28px;
}

.is-text,
.is-line {
  width: 100%;
  height: 14px;
  border-radius: 999px;
}

.is-short {
  width: 58%;
}

.is-icon,
.is-coin {
  width: 42px;
  height: 42px;
  border-radius: 14px;
}

.is-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
}

.is-modal-icon {
  justify-self: center;
  width: 58px;
  height: 58px;
  border-radius: 18px;
}

.is-coin {
  justify-self: center;
  width: 48px;
  height: 48px;
  border-radius: 999px;
}

.is-thumb {
  width: 100%;
  aspect-ratio: 1 / 1;
  border-radius: 14px;
}

.is-control {
  width: 100%;
  height: 50px;
  border-radius: 14px;
}

.is-chip {
  width: 100%;
  height: 42px;
  border-radius: 999px;
}

.is-mini {
  width: 72px;
  height: 14px;
  border-radius: 999px;
}

.is-button,
.is-amount {
  width: 100%;
  height: 54px;
  border-radius: 16px;
}

.is-qr {
  width: 112px;
  height: 112px;
  border-radius: 18px;
}

.is-wheel {
  width: min(62vw, 260px);
  height: min(62vw, 260px);
  border-radius: 50%;
}

.is-frame {
  width: 100%;
  min-height: min(54vh, 460px);
  border-radius: 24px;
}

.ppmx-skeleton--inline {
  gap: 6px;
}

.ppmx-skeleton--inline .is-line {
  height: 16px;
}

.ppmx-skeleton--play {
  place-items: center;
  min-height: 100%;
  padding: 24px;
}

.ppmx-skeleton--play-launch,
.ppmx-skeleton--play-frame {
  place-items: center;
  width: 100%;
  min-height: 100%;
}

.ppmx-skeleton--play-launch .ppmx-skeleton__modal-shell,
.ppmx-skeleton--play-frame .ppmx-skeleton__modal-shell {
  width: 100%;
  height: 100%;
  padding: 0;
  background: transparent;
  border: 0;
  box-shadow: none;
}

.ppmx-skeleton--support-modal .ppmx-skeleton__modal-shell {
  gap: 10px;
  padding: 0;
  background: transparent;
  border: 0;
  box-shadow: none;
}

.ppmx-skeleton--support-modal .ppmx-skeleton__support-channel {
  min-height: 72px;
  border-radius: 20px;
}

.ppmx-skeleton.is-compact.ppmx-skeleton--withdraw-modal .ppmx-skeleton__modal-shell,
.ppmx-skeleton.is-compact.ppmx-skeleton--commission-withdraw-modal .ppmx-skeleton__modal-shell {
  gap: 0;
  padding: 0;
  background: transparent;
  border: 0;
  box-shadow: none;
}

.ppmx-skeleton__withdraw-compact {
  display: grid;
  gap: 8px;
  width: 100%;
}

.ppmx-skeleton__withdraw-balance-compact {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  min-width: 0;
}

.ppmx-skeleton__withdraw-compact .is-control {
  height: 46px;
  border-radius: 14px;
}

.ppmx-skeleton__withdraw-compact .is-label {
  width: min(42%, 160px);
}

.ppmx-skeleton__withdraw-balance-compact .is-label {
  flex: 1 1 42%;
  width: auto;
  max-width: 150px;
}

.is-amount-inline {
  flex: 0 0 min(42%, 178px);
  width: min(42%, 178px);
  height: 18px;
  border-radius: 999px;
}

.is-support-safe {
  width: 100%;
  height: 44px;
  border-radius: 16px;
}

@keyframes ppmxSkeletonShimmer {
  100% {
    transform: translateX(120%);
  }
}

@media (max-width: 720px) {
  .ppmx-skeleton__cards,
  .ppmx-skeleton__grid,
  .ppmx-skeleton__category-strip,
  .ppmx-skeleton__filter-row,
  .ppmx-skeleton__metric-row {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ppmx-skeleton__checkin {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ppmx-skeleton__split,
  .ppmx-skeleton__action-row,
  .ppmx-skeleton__amount-grid {
    grid-template-columns: 1fr;
  }

  .ppmx-skeleton__invite-panel {
    grid-template-columns: 1fr;
  }

  .ppmx-skeleton__invite-panel .is-qr {
    grid-row: auto;
    grid-column: auto;
  }

  .ppmx-skeleton__record-row,
  .ppmx-skeleton__support-channel {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .ppmx-skeleton__record-row .is-mini,
  .ppmx-skeleton__support-channel .is-mini {
    grid-column: 2;
  }
}

@media (prefers-reduced-motion: reduce) {
  .ppmx-skeleton-shimmer::after {
    animation: none;
  }
}

:global(html[data-theme="light"]) .ppmx-skeleton {
  --skeleton-base: rgba(24, 25, 31, 0.065);
  --skeleton-strong: rgba(24, 25, 31, 0.12);
  --skeleton-panel: rgba(255, 255, 255, 0.72);
  --skeleton-panel-strong: rgba(255, 255, 255, 0.9);
  --skeleton-border: rgba(155, 121, 42, 0.14);
  --skeleton-gold: rgba(214, 158, 38, 0.12);
  --skeleton-red: rgba(185, 28, 45, 0.07);
  --skeleton-sheen: rgba(214, 158, 38, 0.18);
  --skeleton-shadow: rgba(49, 24, 20, 0.08);
}

:global(html[data-theme="light"]) .ppmx-skeleton--wheel-page {
  --skeleton-base: rgba(255, 26, 26, 0.055);
  --skeleton-strong: rgba(255, 26, 26, 0.105);
  --skeleton-panel: rgba(255, 255, 255, 0.82);
  --skeleton-panel-strong: rgba(255, 255, 255, 0.96);
  --skeleton-border: rgba(255, 26, 26, 0.13);
  --skeleton-gold: rgba(255, 26, 26, 0.08);
  --skeleton-red: rgba(255, 26, 26, 0.06);
  --skeleton-sheen: rgba(255, 255, 255, 0.72);
  --skeleton-shadow: rgba(36, 38, 45, 0.08);
}

:global(html[data-theme="light"]) .ppmx-skeleton--casino-page,
:global(html[data-theme="light"]) .ppmx-skeleton--casino-more {
  --skeleton-base: rgba(226, 31, 51, 0.032);
  --skeleton-strong: rgba(226, 31, 51, 0.072);
  --skeleton-panel: rgba(255, 255, 255, 0.88);
  --skeleton-panel-strong: rgba(255, 255, 255, 0.98);
  --skeleton-border: rgba(226, 31, 51, 0.105);
  --skeleton-gold: rgba(255, 215, 0, 0.12);
  --skeleton-red: rgba(226, 31, 51, 0.04);
  --skeleton-sheen: rgba(255, 255, 255, 0.82);
  --skeleton-shadow: rgba(36, 38, 45, 0.045);
}

:global(html[data-theme="light"]) .ppmx-skeleton--home-games,
:global(html[data-theme="light"]) .ppmx-skeleton--profile-page,
:global(html[data-theme="light"]) .ppmx-skeleton--user-profile-page,
:global(html[data-theme="light"]) .ppmx-skeleton--bank-page,
:global(html[data-theme="light"]) .ppmx-skeleton--bank-account-list,
:global(html[data-theme="light"]) .ppmx-skeleton--recharge-page,
:global(html[data-theme="light"]) .ppmx-skeleton--team-page,
:global(html[data-theme="light"]) .ppmx-skeleton--records-page {
  --skeleton-base: rgba(34, 34, 40, 0.052);
  --skeleton-strong: rgba(34, 34, 40, 0.098);
  --skeleton-panel: rgba(255, 255, 255, 0.8);
  --skeleton-panel-strong: rgba(255, 255, 255, 0.96);
  --skeleton-border: rgba(226, 31, 51, 0.105);
  --skeleton-gold: rgba(226, 31, 51, 0.07);
  --skeleton-red: rgba(226, 31, 51, 0.052);
  --skeleton-sheen: rgba(255, 255, 255, 0.7);
  --skeleton-shadow: rgba(36, 38, 45, 0.07);
}

:global(html[data-theme="light"]) .ppmx-skeleton--support-modal,
:global(html[data-theme="light"]) .ppmx-skeleton--withdraw-modal,
:global(html[data-theme="light"]) .ppmx-skeleton--commission-withdraw-modal {
  --skeleton-base: rgba(34, 34, 40, 0.05);
  --skeleton-strong: rgba(34, 34, 40, 0.095);
  --skeleton-panel: rgba(255, 255, 255, 0.84);
  --skeleton-panel-strong: rgba(255, 255, 255, 0.97);
  --skeleton-border: rgba(226, 31, 51, 0.12);
  --skeleton-gold: rgba(226, 31, 51, 0.07);
  --skeleton-red: rgba(226, 31, 51, 0.05);
  --skeleton-sheen: rgba(255, 255, 255, 0.72);
  --skeleton-shadow: rgba(36, 38, 45, 0.08);
}

:global(html[data-theme="light"]) .ppmx-skeleton--play-launch,
:global(html[data-theme="light"]) .ppmx-skeleton--play-frame {
  --skeleton-base: rgba(25, 27, 34, 0.06);
  --skeleton-strong: rgba(25, 27, 34, 0.11);
  --skeleton-panel: rgba(255, 255, 255, 0.84);
  --skeleton-panel-strong: rgba(255, 255, 255, 0.98);
  --skeleton-border: rgba(226, 31, 51, 0.12);
  --skeleton-gold: rgba(226, 31, 51, 0.08);
  --skeleton-red: rgba(226, 31, 51, 0.06);
  --skeleton-sheen: rgba(255, 255, 255, 0.76);
  --skeleton-shadow: rgba(36, 38, 45, 0.08);
}
</style>
