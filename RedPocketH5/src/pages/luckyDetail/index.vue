<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getLuckyDetail } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import LuckyGrabModal from '@/components/LuckyGrabModal.vue'
import wsClient from '@/plugins/websocket'
import { formatCurrency } from '@/utils/currency'
import { isLogin } from '@/utils/auth'
import imgAvatarPlaceholder from '@/assets/images/avatar-placeholder.png'
import imgRedpacketGif from '@/assets/images/redpacket.gif'
import imgRedpacketJpg from '@/assets/images/redpacket.jpg'

const { t } = useI18n()

const route = useRoute()
const router = useRouter()

const DEFAULT_AVATAR = imgAvatarPlaceholder
const detail = ref<any | null>(null)
const loading = ref(false)
const grabModalVisible = ref(false)
const pendingGrabTarget = ref<{ seqNo: number } | null>(null)

const luckyId = computed(() => Number(route.query.id || 0))
const loggedIn = computed(() => isLogin())

const overview = computed(() => {
  const data = detail.value
  const summary = data?.summary
  const sender = data?.sender
  const finance = data?.finance
  if (!summary)
    return null

  const isOngoing = Number(summary?.status) === 1
  const amount = Number(summary?.amount || 0)
  const received = Number(finance?.receivedAmount || 0)
  const hitCount = Number(finance?.hitCount || 0)
  const number = Number(summary?.number || 0)
  const grabbedCount = Number(summary?.grabbedCount || 0)
  const remain = Math.max(amount - received, 0)
  const expireAt = new Date(summary?.expireTime || '')
  const remainSeconds = Math.max(0, Number.isNaN(expireAt.getTime()) ? 0 : Math.floor((expireAt.getTime() - Date.now()) / 1000))

  return {
    id: Number(summary?.id || luckyId.value),
    status: isOngoing ? 'ongoing' : 'done',
    statusText: isOngoing ? t('homeLucky.statusOngoing') : t('homeLucky.statusDone'),
    senderName: sender?.senderName || t('luckyDetailPage.defaultSender'),
    senderAvatar: sender?.senderAvatar || DEFAULT_AVATAR,
    sentTimeText: formatTime(sender?.sendTime || ''),
    amountText: formatCurrency(amount),
    gameText: summary?.gameText || t('homeLucky.game'),
    progressText: t('homeLucky.progress', { grabbed: grabbedCount, total: number }),
    thunderText: isOngoing
      ? t('luckyDetailPage.thunderUnknown')
      : t('homeLucky.thunderNo', { no: Number(summary?.thunder || 0) }),
    oddsText: t('luckyDetailPage.odds', { rate: Number(summary?.loseRate || 0).toFixed(1) }),
    timeText: isOngoing
      ? t('homeLucky.remainingTime', { time: formatRemainText(remainSeconds) })
      : t('luckyDetailPage.timeExpired', { time: formatTime(summary?.expireTime || '') }),
    unitText: summary?.unitAmount || t('luckyDetailPage.unitAmountRandom'),
    packetImage: isOngoing ? imgRedpacketGif : imgRedpacketJpg,
    summaryRows: [
      { label: t('luckyDetailPage.rowSendAmount'), value: formatCurrency(amount) },
      { label: t('luckyDetailPage.rowGrabbedAmount'), value: formatCurrency(received) },
      { label: t('luckyDetailPage.rowRemainAmount'), value: formatCurrency(remain) },
      { label: t('luckyDetailPage.rowHitCount'), value: `${hitCount}` },
      { label: t('luckyDetailPage.rowThunderIncome'), value: formatCurrency(Number(finance?.thunderIncome || 0)), highlight: true },
      { label: t('luckyDetailPage.rowUnclaimed'), value: `${Math.max(number - grabbedCount, 0)}` },
    ],
  }
})

const positionList = computed(() => {
  const summary = detail.value?.summary || {}
  const participants = detail.value?.participants || []
  const participantMap = new Map<number, any>()
  participants.forEach((item: any) => {
    const seqNo = Number(item?.seqNo || 0)
    if (seqNo > 0)
      participantMap.set(seqNo, item)
  })

  const total = Number(summary?.number || 0)
  const isOngoing = Number(summary?.status) === 1

  return Array.from({ length: total }, (_, idx) => {
    const seqNo = idx + 1
    const participant = participantMap.get(seqNo)
    const isGrabbed = !!participant
    const isThunder = Number(participant?.isThunder || 0) === 1
    const amount = Number(participant?.amount || 0)

    let label = ''
    if (isGrabbed)
      label = amount > 0 ? formatCurrency(amount) : t('homeLucky.loadingLabel')
    else if (isOngoing)
      label = t('homeLucky.grabAction', { seq: seqNo })
    else
      label = t('homeLucky.statusDone')

    return {
      seqNo,
      isGrabbed,
      isThunder,
      label,
    }
  })
})

const recordList = computed(() => {
  const participants = detail.value?.participants || []
  return participants.map((item: any) => ({
    id: `${item?.seqNo || 0}_${item?.userId || 0}_${item?.createdAt || ''}`,
    seqNo: Number(item?.seqNo || 0),
    avatar: item?.avatar || DEFAULT_AVATAR,
    name: item?.firstName || t('luckyDetailPage.positionUser', { seq: Number(item?.seqNo || 0) }),
    time: formatTime(item?.createdAt || ''),
    amountText: Number(item?.amount || 0) <= 0 ? t('homeLucky.loadingLabel') : formatCurrency(Number(item?.amount || 0)),
    statusText: Number(item?.isThunder || 0) === 1 ? t('luckyDetailPage.positionThunder') : t('luckyDetailPage.positionJoined'),
  }))
})

function formatRemainText(seconds: number) {
  const safeSeconds = Math.max(0, Math.floor(seconds))
  return `${String(Math.floor(safeSeconds / 60)).padStart(2, '0')}:${String(safeSeconds % 60).padStart(2, '0')}`
}

function formatTime(raw: string) {
  const date = new Date(raw)
  if (Number.isNaN(date.getTime()))
    return raw || '-'
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hh = String(date.getHours()).padStart(2, '0')
  const mm = String(date.getMinutes()).padStart(2, '0')
  return `${year}/${month}/${day} ${hh}:${mm}`
}

async function loadDetail() {
  if (!luckyId.value) {
    showToast(t('luckyDetailPage.toastInvalidParam'))
    return
  }
  if (loading.value)
    return

  loading.value = true
  try {
    const { data } = await getLuckyDetail({ luckyId: luckyId.value })
    detail.value = data || null
  }
  catch {
    showToast(t('luckyDetailPage.toastLoadFailed'))
  }
  finally {
    loading.value = false
  }
}

function openGrabDialog(item: { seqNo: number, isGrabbed: boolean }) {
  if (!loggedIn.value) {
    showToast(t('homeLucky.loginFirst'))
    return
  }
  if (overview.value?.status !== 'ongoing' || item.isGrabbed)
    return
  pendingGrabTarget.value = { seqNo: item.seqNo }
  grabModalVisible.value = true
}

function closeGrabDialog() {
  grabModalVisible.value = false
  pendingGrabTarget.value = null
}

function handleGrabSuccess() {
  void loadDetail()
}

function handleLuckyWsMessage(message: any) {
  const payload = message?.data || message
  if (Number(payload?.id || 0) !== luckyId.value)
    return
  void loadDetail()
}

function goBack() {
  router.back()
}

function goInvite() {
  router.push('/invite')
}

onMounted(() => {
  void loadDetail()
  wsClient.on('lucky_sent', handleLuckyWsMessage)
  wsClient.on('lucky_grabbed', handleLuckyWsMessage)
})

onBeforeUnmount(() => {
  wsClient.off('lucky_sent', handleLuckyWsMessage)
  wsClient.off('lucky_grabbed', handleLuckyWsMessage)
})
</script>

<template>
  <div class="lucky-detail-page">
    <AppPageHeader :title="t('luckyDetailPage.title')" @back="goBack" @right-click="goInvite">
      <template #right>
        <van-icon name="share-o" />
      </template>
    </AppPageHeader>

    <div class="detail-content">
      <template v-if="loading && !overview">
        <section class="detail-card skeleton-card">
          <van-skeleton title avatar :row="4" :loading="true" avatar-size="72px" />
        </section>
        <section class="detail-card skeleton-card">
          <van-skeleton title :row="4" :loading="true" />
        </section>
        <section class="detail-card skeleton-card">
          <van-skeleton title :row="6" :loading="true" />
        </section>
      </template>

      <template v-else-if="overview">
        <section class="detail-card hero-card" :class="overview.status">
          <div class="hero-main">
            <div class="hero-top">
              <div class="user-wrap">
                <img :src="overview.senderAvatar" alt="" class="user-avatar">
                <div class="user-copy">
                  <strong class="user-name">{{ overview.senderName }}</strong>
                  <p class="sender-time">{{ t('luckyDetailPage.sendTime', { time: overview.sentTimeText }) }}</p>
                </div>
              </div>
              <div class="amount-wrap">
                <span class="packet-amount">{{ overview.amountText }}</span>
              </div>
            </div>

            <div class="hero-body">
              <div class="packet-image-wrap">
                <span class="status-badge">{{ overview.statusText }}</span>
                <img :src="overview.packetImage" alt="" class="packet-image">
              </div>

              <div class="packet-info">
                <div class="tags-row">
                  <span class="tag game">{{ overview.gameText }}</span>
                  <span class="tag progress">{{ overview.progressText }}</span>
                </div>
                <div class="meta-row">
                  <span>{{ overview.thunderText }}</span>
                  <span>{{ overview.oddsText }}</span>
                </div>
                <p class="rebate-text">
                  {{ overview.unitText }}
                </p>
                <p class="time-text">
                  {{ overview.timeText }}
                </p>
              </div>
            </div>
          </div>
        </section>

        <section class="detail-card summary-card">
          <header class="section-header">
            <div class="section-divider" />
            <div class="section-title-wrap">
              <van-icon name="bar-chart-o" />
              <span>{{ t('luckyDetailPage.panelFinance') }}</span>
            </div>
          </header>
          <div class="summary-list">
            <div v-for="row in overview.summaryRows" :key="row.label" class="summary-row">
              <span class="summary-label">{{ row.label }}</span>
              <span class="summary-value" :class="{ highlight: row.highlight }">{{ row.value }}</span>
            </div>
          </div>
        </section>

        <section class="detail-card positions-card">
          <header class="section-header">
            <div class="section-divider" />
            <div class="section-title-wrap">
              <van-icon name="apps-o" />
              <span>{{ t('luckyDetailPage.panelChoosePosition') }}</span>
            </div>
          </header>

          <div class="packet-actions detail-actions">
            <button
              v-for="item in positionList"
              :key="item.seqNo"
              type="button"
              class="action-pill"
              :class="{
                grabbed: item.isGrabbed,
                ended: overview.status !== 'ongoing' && !item.isGrabbed,
                thunder: item.isThunder,
              }"
              :disabled="overview.status !== 'ongoing' || item.isGrabbed"
              @click="openGrabDialog(item)"
            >
              <span v-if="item.isThunder" aria-hidden="true">💣</span>
              {{ item.label }}
            </button>
          </div>
        </section>

        <section class="detail-card records-card">
          <header class="section-header">
            <div class="section-divider" />
            <div class="section-title-wrap">
              <van-icon name="friends-o" />
              <span>{{ t('luckyDetailPage.panelParticipants') }}</span>
            </div>
          </header>

          <AppEmpty v-if="recordList.length === 0" :text="t('luckyDetailPage.emptyText')" :min-height="120" />

          <div v-else class="record-card">
            <article v-for="item in recordList" :key="item.id" class="record-item">
              <img :src="item.avatar" alt="" class="record-avatar">
              <div class="record-main">
                <p class="record-amount">
                  #{{ item.seqNo }} {{ item.name }}
                </p>
                <p class="record-name">
                  {{ item.time }}
                </p>
              </div>
              <div class="record-right">
                <p class="record-total">
                  {{ item.amountText }}
                </p>
                <p class="record-status">
                  {{ item.statusText }}
                </p>
              </div>
            </article>
          </div>
        </section>
      </template>

      <AppEmpty v-else-if="!loading" :text="t('luckyDetailPage.emptyText')" :min-height="160" />
    </div>

    <LuckyGrabModal
      v-model:show="grabModalVisible"
      :lucky-id="overview?.id || luckyId"
      :grab-index="Number(pendingGrabTarget?.seqNo || 0)"
      :sender-name="overview?.senderName || t('grabModal.defaultSender')"
      :show-result-toast="false"
      @success="handleGrabSuccess"
      @close="closeGrabDialog"
    />
  </div>
</template>

<style scoped>
.lucky-detail-page {
  min-height: 100vh;
  background-image:
    radial-gradient(circle at 20% 10%, rgba(212, 175, 55, 0.18), transparent 30%),
    radial-gradient(circle at 80% 90%, rgba(255, 215, 0, 0.12), transparent 28%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.04) 18px,
      rgba(212, 175, 55, 0.04) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #230000 62%, #160000 100%);
  padding-bottom: calc(18px + env(safe-area-inset-bottom));
}

.detail-content {
  padding: 10px 12px 0;
}

.detail-card {
  position: relative;
  border-radius: 16px;
  border: 1px solid rgba(212, 175, 55, 0.45);
  background: linear-gradient(160deg, rgba(140, 0, 0, 0.98) 0%, rgba(90, 0, 0, 0.97) 55%, rgba(55, 0, 0, 0.97) 100%);
  overflow: hidden;
  margin-bottom: 12px;
  box-shadow:
    0 10px 24px rgba(0, 0, 0, 0.38),
    inset 0 0 0 1px rgba(255, 248, 214, 0.12),
    0 0 0 1px rgba(212, 175, 55, 0.35);
}

.detail-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent 0%, #b8860b 15%, #ffd700 40%, #d4af37 60%, #b8860b 85%, transparent 100%);
  pointer-events: none;
}

.detail-card::before {
  content: '';
  position: absolute;
  inset: 3px 0 0;
  background-image: radial-gradient(rgba(212, 175, 55, 1) 1px, transparent 1px);
  background-size: 18px 18px;
  opacity: 0.05;
  pointer-events: none;
}

.hero-main,
.summary-card,
.positions-card,
.records-card {
  position: relative;
  z-index: 1;
}

.hero-card {
  padding: 12px;
}

.hero-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.user-copy {
  min-width: 0;
}

.user-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(212, 175, 55, 0.65);
  box-shadow: 0 0 8px rgba(212, 175, 55, 0.25), 0 4px 8px rgba(0, 0, 0, 0.28);
}

.user-name {
  display: block;
  color: #fff0c9;
  font-size: 15px;
  line-height: 1.1;
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sender-time {
  margin: 3px 0 0;
  color: rgba(255, 229, 186, 0.62);
  font-size: 10px;
  line-height: 1.3;
}

.amount-wrap {
  display: inline-flex;
  align-items: center;
}

.packet-amount {
  font-size: 16px;
  line-height: 1;
  font-weight: 700;
  background: linear-gradient(to bottom, #cfb53b, #ffd700, #d4af37);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.25));
}

.hero-body {
  margin-top: 8px;
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.packet-image-wrap {
  width: 84px;
  flex: 0 0 84px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 56px;
  height: 20px;
  margin-bottom: 4px;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 10px;
  font-weight: 700;
  border: 1px solid rgba(255, 248, 214, 0.5);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2), inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.packet-image {
  width: 100%;
  display: block;
}

.packet-info {
  flex: 1;
  min-width: 0;
}

.tags-row,
.meta-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.tag,
.meta-row span,
.rebate-text {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
}

.tag {
  height: 20px;
  padding: 0 7px;
  font-size: 11px;
  border: 1px solid transparent;
}

.tag.game {
  color: #ffe7bf;
  background: rgba(212, 175, 55, 0.14);
  border-color: rgba(212, 175, 55, 0.4);
}

.tag.progress {
  color: #ffeecf;
  background: rgba(255, 248, 214, 0.1);
  border-color: rgba(255, 248, 214, 0.28);
}

.meta-row {
  margin-top: 6px;
}

.meta-row span {
  height: 18px;
  padding: 0 6px;
  background: rgba(255, 248, 214, 0.07);
  border: 1px solid rgba(255, 248, 214, 0.18);
  color: rgba(255, 229, 186, 0.82);
  font-size: 10px;
}

.rebate-text {
  margin: 6px 0 0;
  min-height: 18px;
  padding: 0 7px;
  background: rgba(212, 175, 55, 0.1);
  border: 1px solid rgba(212, 175, 55, 0.28);
  color: rgba(255, 232, 160, 0.85);
  font-size: 10px;
}

.time-text {
  margin: 7px 0 0;
  color: #ffd87f;
  font-size: 14px;
  line-height: 1.2;
  font-weight: 700;
  text-shadow: 0 0 8px rgba(255, 216, 127, 0.35);
}

.section-header {
  padding: 12px 12px 0;
}

.section-divider {
  height: 1px;
  margin-bottom: 10px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(212, 175, 55, 0.6) 20%,
    rgba(212, 175, 55, 0.9) 50%,
    rgba(212, 175, 55, 0.6) 80%,
    transparent 100%
  );
}

.section-title-wrap {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #ffd98b;
  font-size: 15px;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.section-title-wrap :deep(.van-icon) {
  color: #d4af37;
}

.summary-list {
  padding: 10px 12px 12px;
}

.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.summary-row + .summary-row {
  margin-top: 8px;
}

.summary-label {
  color: rgba(255, 229, 186, 0.7);
  font-size: 12px;
}

.summary-value {
  color: #fff0c9;
  font-size: 13px;
  font-weight: 700;
}

.summary-value.highlight {
  color: #ffd87f;
}

.detail-actions {
  padding: 10px 12px 12px;
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 6px;
}

.action-pill {
  min-height: 28px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(180deg, #9e1010 0%, #6a0000 100%);
  color: #fff3de;
  font-size: 8px;
  box-shadow: inset 0 1px 0 rgba(212, 175, 55, 0.45), 0 2px 6px rgba(0, 0, 0, 0.3);
}

.action-pill.grabbed {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 248, 214, 0.45);
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: none;
}

.action-pill.ended {
  background: rgba(212, 175, 55, 0.08);
  color: rgba(255, 248, 214, 0.6);
  border: 1px solid rgba(212, 175, 55, 0.2);
  box-shadow: none;
}

.action-pill.thunder {
  color: #ffe088;
}

.record-card {
  margin: 0 12px 12px;
  border-radius: 16px;
  border: 1px solid rgba(212, 175, 55, 0.36);
  background: linear-gradient(170deg, rgba(125, 0, 0, 0.72), rgba(60, 0, 0, 0.72));
  overflow: hidden;
}

.record-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
}

.record-item + .record-item {
  border-top: 1px solid rgba(212, 175, 55, 0.14);
}

.record-avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(212, 175, 55, 0.7);
}

.record-main {
  flex: 1;
  min-width: 0;
}

.record-amount {
  margin: 0;
  color: #ffefca;
  font-size: 14px;
  line-height: 1.3;
  font-weight: 700;
}

.record-name {
  margin: 4px 0 0;
  color: rgba(255, 229, 186, 0.68);
  font-size: 12px;
}

.record-right {
  text-align: right;
}

.record-total {
  margin: 0;
  color: #ffd87f;
  font-size: 14px;
  font-weight: 700;
}

.record-status {
  margin: 4px 0 0;
  color: rgba(255, 229, 186, 0.62);
  font-size: 11px;
}

.skeleton-card {
  padding: 12px;
}

.skeleton-card :deep(.van-skeleton__title),
.skeleton-card :deep(.van-skeleton__row) {
  background: rgba(255, 248, 214, 0.14);
}

@media (max-width: 390px) {
  .detail-content {
    padding: 8px 10px 0;
  }

  .hero-card {
    padding: 10px;
  }

  .packet-image-wrap {
    width: 74px;
    flex-basis: 74px;
  }

  .detail-actions {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>

<route lang="json5">
{
  name: 'LuckyDetail',
}
</route>
