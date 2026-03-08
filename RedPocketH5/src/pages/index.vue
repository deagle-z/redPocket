<script setup lang="ts">
import { showToast } from 'vant'
import { useRouter } from 'vue-router'
import { getLuckyPacketList, getLuckyRecentWinners } from '@/api/user'
import LuckyGrabModal from '@/components/LuckyGrabModal.vue'
import { isLogin } from '@/utils/auth'
import { formatCurrency } from '@/utils/currency'
import wsClient from '@/plugins/websocket'
import imgAvatarPlaceholder from '@/assets/images/avatar-placeholder.png'
import imgRedpacketGif from '@/assets/images/redpacket.gif'
import imgRedpacketJpg from '@/assets/images/redpacket.jpg'
import imgTutorial from '@/assets/images/tutorial.png'
import imgActivityChannel from '@/assets/images/activity-channel300.jpg'
import imgLuckykita from '@/assets/images/luckykita2.jpg'

const { t } = useI18n()

const bannerList = computed(() => [
  {
    img: imgTutorial,
    text: t('homeLucky.bannerText1'),
  },
  {
    img: imgActivityChannel,
    text: t('homeLucky.bannerText2'),
  },
  {
    img: imgLuckykita,
    text: t('homeLucky.bannerText3'),
  },
])

const DEFAULT_AVATAR = imgAvatarPlaceholder

const activeIndex = ref(0)
const router = useRouter()

const packetList = ref<any[]>([])
const packetLoading = ref(false)
const grabModalVisible = ref(false)
const pendingGrabTarget = ref<{ packet: any, action: any } | null>(null)
let countdownTimer: number | undefined
const loggedIn = computed(() => isLogin())
const recentWinnersLoading = ref(false)

const recentWinners = ref<any[]>([])

const visibleWinners = computed(() => {
  if (!loggedIn.value)
    return []
  return recentWinners.value
})

const showPacketEmpty = computed(() => {
  return !loggedIn.value || (!packetLoading.value && packetList.value.length === 0)
})
const showPacketLoading = computed(() => {
  return loggedIn.value && packetLoading.value && packetList.value.length === 0
})
const showWinnerLoading = computed(() => {
  return loggedIn.value && recentWinnersLoading.value && visibleWinners.value.length === 0
})

const showWinnerEmpty = computed(() => {
  return !loggedIn.value || visibleWinners.value.length === 0
})

const marqueeText = computed(() => {
  const current = bannerList.value[activeIndex.value]
  return current?.text || ''
})

function onSwipeChange(index: number) {
  activeIndex.value = index
}

function goLuckyDetail(packet: any) {
  const id = Number(packet?.id || 0)
  if (!id)
    return
  router.push({
    path: '/luckyDetail',
    query: { id: String(id) },
  })
}

function formatAmount(value: number) {
  return formatCurrency(Number(value || 0))
}

function formatActionLabel(isGrabbed: boolean, amount: number, seqNo: number) {
  if (!isGrabbed)
    return t('homeLucky.grabAction', { seq: seqNo })
  if (Number(amount) <= 0)
    return t('homeLucky.loadingLabel')
  return formatAmount(amount)
}

function formatRemainText(seconds: number) {
  const s = Math.max(0, Math.floor(seconds))
  return `${String(Math.floor(s / 60)).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`
}

function refreshPacketCountdowns() {
  const now = Date.now()
  packetList.value = packetList.value.map((packet) => {
    if (packet.status !== 'ongoing')
      return packet
    const expireAtMs = Number(packet.expireAtMs || 0)
    if (!expireAtMs)
      return packet
    const remainSec = Math.max(0, Math.floor((expireAtMs - now) / 1000))
    if (remainSec <= 0) {
      return {
        ...packet,
        status: 'done',
        statusText: t('homeLucky.statusDone'),
        timeText: t('homeLucky.statusDone'),
        packetImage: imgRedpacketJpg,
      }
    }
    return {
      ...packet,
      timeText: t('homeLucky.remainingTime', { time: formatRemainText(remainSec) }),
    }
  })
}

function mapPacket(item: any) {
  const isOngoing = Number(item?.status) === 1
  const senderWinAmount = (item?.items || []).reduce((sum: number, it: any) => {
    return sum + Number(it?.thunderAmount || 0)
  }, 0)
  let actions = (item?.items || [])
    .map((it: any) => ({
      seqNo: Number(it?.seqNo || 0),
      isGrabbed: Number(it?.isGrabbed) === 1,
      isGrabMine: Number(it?.isGrabMine) === 1,
      amount: Number(it?.amount || 0),
      thunder: Number(it?.thunder || 0),
      label: formatActionLabel(Number(it?.isGrabbed) === 1, Number(it?.amount || 0), Number(it?.seqNo || 0)),
    }))
  if (actions.length === 0 && isOngoing) {
    const number = Number(item?.number || 0)
    actions = Array.from({ length: number }, (_, idx) => ({
      seqNo: idx + 1,
      isGrabbed: false,
      isGrabMine: false,
      amount: 0,
      thunder: 0,
      label: t('homeLucky.grabAction', { seq: idx + 1 }),
    }))
  }

  return {
    id: Number(item?.id || 0),
    username: item?.senderName || 'User',
    avatar: item?.senderAvatar || DEFAULT_AVATAR,
    amount: formatAmount(item?.amount),
    status: isOngoing ? 'ongoing' : 'done',
    statusText: isOngoing ? t('homeLucky.statusOngoing') : t('homeLucky.statusDone'),
    gameText: t('homeLucky.game'),
    progressText: t('homeLucky.progress', { grabbed: Number(item?.grabbedCount || 0), total: Number(item?.number || 0) }),
    thunderText: isOngoing ? '' : t('homeLucky.thunderNo', { no: Number(item?.thunder || 0) }),
    hitsText: t('homeLucky.hitsCount', { count: Number(item?.hitCount || 0) }),
    rebateText: t('homeLucky.rebate', { amount: formatAmount(senderWinAmount) }),
    timeText: isOngoing ? t('homeLucky.remainingTime', { time: item?.remainingText || '00:00' }) : t('homeLucky.statusDone'),
    packetImage: isOngoing ? imgRedpacketGif : imgRedpacketJpg,
    expireAtMs: item?.expireTime ? new Date(item.expireTime).getTime() : 0,
    actions,
  }
}

function applyLuckySent(message: any) {
  const lucky = message?.data || message
  const mapped = mapPacket(lucky)
  if (!mapped.id)
    return

  const current = packetList.value
  const existingIndex = current.findIndex(item => Number(item.id) === Number(mapped.id))
  if (existingIndex >= 0) {
    const next = [...current]
    next[existingIndex] = {
      ...next[existingIndex],
      ...mapped,
    }
    packetList.value = next
    return
  }

  packetList.value = [mapped, ...current].slice(0, 20)
  refreshPacketCountdowns()
}

async function loadPacketList() {
  if (!loggedIn.value) {
    packetList.value = []
    return
  }
  if (packetLoading.value)
    return
  try {
    packetLoading.value = true
    const { data } = await getLuckyPacketList({
      currentPage: 0,
      pageSize: 20,
    })
    packetList.value = (data?.list || []).map((item: any) => mapPacket(item))
    refreshPacketCountdowns()
  }
  catch {
    showToast(t('homeLucky.packetLoadFailed'))
  }
  finally {
    packetLoading.value = false
  }
}

function openGrabDialog(packet: any, action: any) {
  if (!loggedIn.value) {
    showToast(t('homeLucky.loginFirst'))
    return
  }
  if (action.isGrabbed)
    return
  pendingGrabTarget.value = { packet, action }
  grabModalVisible.value = true
}

function closeGrabDialog() {
  grabModalVisible.value = false
  pendingGrabTarget.value = null
}

function handleGrabSuccess(payload: { luckyId: number, grabIndex: number, data: any }) {
  const luckyId = Number(payload?.luckyId || 0)
  const grabIndex = Number(payload?.grabIndex || 0)
  const rawAmount = Number(payload?.data?.amount ?? payload?.data?.grabAmount ?? 0)
  if (!luckyId || !grabIndex)
    return

  packetList.value = packetList.value.map((packet) => {
    if (Number(packet.id) !== luckyId)
      return packet

    const nextActions = Array.isArray(packet.actions) ? [...packet.actions] : []
    const idx = nextActions.findIndex((it: any) => Number(it?.seqNo) === grabIndex)
    if (idx >= 0) {
      nextActions[idx] = {
        ...nextActions[idx],
        isGrabbed: true,
        amount: rawAmount,
        label: formatActionLabel(true, rawAmount, grabIndex),
      }
    }
    const grabbedCount = nextActions.filter((it: any) => it.isGrabbed).length
    const packetNumber = Number(packet?.actions?.length || 0)
    return {
      ...packet,
      progressText: t('homeLucky.progress', { grabbed: grabbedCount, total: packetNumber }),
      actions: nextActions,
    }
  })
}

function applyLuckyBroadcast(message: any) {
  const lucky = message?.data || message
  const luckyId = Number(lucky?.id || 0)
  const grabbedSeqNo = Number(lucky?.grabIndex || 0)
  const totalThunderAmount = Number(lucky?.totalThunderAmount)
  if (!luckyId)
    return

  packetList.value = packetList.value.map((packet) => {
    if (Number(packet.id) !== luckyId)
      return packet

    const nextStatus = Number(lucky?.status) === 1 ? 'ongoing' : 'done'
    const nextActions = Array.isArray(packet.actions) ? [...packet.actions] : []

    if (nextStatus === 'ongoing') {
      // 广播带序号时，精确更新对应子红包。
      if (grabbedSeqNo > 0) {
        const idx = nextActions.findIndex((it: any) => Number(it?.seqNo) === grabbedSeqNo)
        if (idx >= 0) {
          const nextAmount = Number(lucky?.grabAmount || nextActions[idx]?.amount || 0)
          nextActions[idx] = {
            ...nextActions[idx],
            isGrabbed: true,
            thunder: Number(lucky?.isThunder || 0),
            amount: nextAmount,
            label: formatActionLabel(true, nextAmount, Number(nextActions[idx]?.seqNo || grabbedSeqNo)),
          }
        }
      }
      else {
        // 兼容旧广播：按顺序推进一个未抢子红包，避免再拉列表。
        const idx = nextActions.findIndex((it: any) => !it.isGrabbed)
        if (idx >= 0) {
          const seqNo = Number(nextActions[idx]?.seqNo || 0)
          nextActions[idx] = {
            ...nextActions[idx],
            isGrabbed: true,
            label: formatActionLabel(true, Number(nextActions[idx]?.amount || 0), seqNo),
          }
        }
      }
    }

    const grabbedCount = nextActions.filter((it: any) => it.isGrabbed).length
    const packetNumber = Number(lucky?.number || packet?.actions?.length || 0)

    return {
      ...packet,
      status: nextStatus,
      statusText: nextStatus === 'ongoing' ? t('homeLucky.statusOngoing') : t('homeLucky.statusDone'),
      amount: formatAmount(lucky?.amount),
      rebateText: Number.isFinite(totalThunderAmount) ? t('homeLucky.rebate', { amount: formatAmount(totalThunderAmount) }) : packet.rebateText,
      thunderText: nextStatus === 'ongoing' ? '' : t('homeLucky.thunderNo', { no: Number(lucky?.thunder || 0) }),
      progressText: t('homeLucky.progress', { grabbed: grabbedCount, total: packetNumber }),
      timeText: nextStatus === 'ongoing' ? packet.timeText : t('homeLucky.statusDone'),
      packetImage: nextStatus === 'ongoing' ? imgRedpacketGif : imgRedpacketJpg,
      actions: nextActions,
    }
  })
  refreshPacketCountdowns()
}

async function loadRecentWinners() {
  if (!loggedIn.value) {
    recentWinners.value = []
    return
  }
  if (recentWinnersLoading.value)
    return
  try {
    recentWinnersLoading.value = true
    const { data } = await getLuckyRecentWinners({ limit: 10 })
    recentWinners.value = (data || []).map((item: any) => ({
      id: Number(item?.id || 0),
      avatar: item?.avatar || DEFAULT_AVATAR,
      amount: formatAmount(item?.amount),
      name: item?.firstName || 'User',
      time: item?.timeText || t('homeLucky.timeJustNow'),
    }))
  }
  catch {
    showToast(t('homeLucky.winnerLoadFailed'))
  }
  finally {
    recentWinnersLoading.value = false
  }
}

onMounted(() => {
  if (!loggedIn.value)
    return
  loadPacketList()
  loadRecentWinners()
  countdownTimer = window.setInterval(refreshPacketCountdowns, 1000)
  wsClient.on('lucky_sent', applyLuckySent)
  wsClient.on('lucky_grabbed', applyLuckyBroadcast)
})

onBeforeUnmount(() => {
  if (countdownTimer)
    window.clearInterval(countdownTimer)
  wsClient.off('lucky_sent', applyLuckySent)
  wsClient.off('lucky_grabbed', applyLuckyBroadcast)
})
</script>

<template>
  <div class="home-page">
    <section class="home-carousel-card">
      <van-swipe class="home-swipe" :autoplay="3200" lazy-render indicator-color="#ffffff" @change="onSwipeChange">
        <van-swipe-item v-for="(item, idx) in bannerList" :key="`${item.img}-${idx}`">
          <img :src="item.img" class="banner-image" :alt="`banner-${idx + 1}`">
        </van-swipe-item>
      </van-swipe>

      <van-notice-bar class="home-marquee" :scrollable="true" :text="marqueeText" />
    </section>

    <section class="packet-section">
      <header class="packet-header">
        <div class="packet-title-wrap">
          <van-icon name="fire-o" />
          <span>{{ t('homeLucky.hotPackets') }}</span>
        </div>
      </header>

      <div v-if="showPacketLoading" class="packet-skeleton-list">
        <article v-for="idx in 3" :key="`packet-skeleton-${idx}`" class="packet-card skeleton-card">
          <div class="packet-main">
            <div class="packet-top">
              <div class="user-wrap">
                <van-skeleton-avatar avatar-size="34px" />
                <van-skeleton title :row="0" class="skeleton-user-name" />
              </div>
              <van-skeleton title :row="0" class="skeleton-amount" />
            </div>
            <div class="packet-body">
              <van-skeleton-image class="skeleton-packet-image" />
              <div class="packet-info">
                <van-skeleton title :row="3" />
              </div>
            </div>
          </div>
          <div class="packet-actions packet-actions-skeleton">
            <van-skeleton v-for="pill in 6" :key="`pill-${idx}-${pill}`" title :row="0" class="skeleton-pill" />
          </div>
        </article>
      </div>

      <AppEmpty v-else-if="showPacketEmpty" :text="t('homeLucky.emptyPackets')" :min-height="120" />

      <template v-else>
        <article v-for="packet in packetList" :key="packet.id" class="packet-card" :class="packet.status">
          <div class="packet-main" @click="goLuckyDetail(packet)">
            <div class="packet-top">
              <div class="user-wrap">
                <img :src="packet.avatar" alt="" class="user-avatar">
                <strong class="user-name">{{ packet.username }}</strong>
              </div>
              <div class="amount-wrap">
                <span class="packet-amount">{{ packet.amount }}</span>
                <van-icon name="arrow" />
              </div>
            </div>

            <div class="packet-body">
              <div class="packet-image-wrap">
                <span class="status-badge">{{ packet.statusText }}</span>
                <img :src="packet.packetImage" alt="" class="packet-image">
              </div>

              <div class="packet-info">
                <div class="tags-row">
                  <span class="tag game">🎮 {{ packet.gameText }}</span>
                  <span class="tag progress">{{ packet.progressText }}</span>
                </div>
                <div class="meta-row">
                  <span v-if="packet.thunderText">{{ packet.thunderText }}</span>
                  <span>{{ packet.hitsText }}</span>
                </div>
                <p class="rebate-text">
                  {{ packet.rebateText }}
                </p>
                <p class="time-text">
                  {{ packet.timeText }}
                </p>
              </div>
            </div>
          </div>

          <div class="packet-actions">
            <template v-if="packet.status === 'ongoing'">
              <button
                v-for="action in packet.actions" :key="`${packet.id}-${action.seqNo}`" type="button"
                class="action-pill" :class="{ grabbed: action.isGrabbed, mined: action.isGrabMine }"
                :disabled="action.isGrabbed"
                @click="openGrabDialog(packet, action)"
              >
                <span v-if="action.thunder" aria-hidden="true">💣</span>
                <span v-else-if="action.isGrabMine" class="mine-text">🎁 </span>
                {{ action.label }}
              </button>
            </template>
            <button v-else type="button" class="action-pill done">
              {{ t('homeLucky.statusDone') }}
            </button>
          </div>
        </article>
      </template>
    </section>

    <section class="winner-section">
      <header class="packet-header">
        <div class="packet-title-wrap">
          <van-icon name="trophy-o" />
          <span>{{ t('homeLucky.latestWinners') }}</span>
        </div>
      </header>

      <div v-if="showWinnerLoading" class="winner-skeleton-card">
        <article v-for="idx in 4" :key="`winner-skeleton-${idx}`" class="winner-item">
          <van-skeleton-avatar avatar-size="44px" />
          <div class="winner-main">
            <van-skeleton title :row="1" />
          </div>
          <van-skeleton title :row="0" class="winner-skeleton-time" />
        </article>
      </div>

      <AppEmpty v-else-if="showWinnerEmpty" :text="t('homeLucky.emptyWinners')" :min-height="120" />

      <template v-else>
        <div class="winner-card">
          <article v-for="item in visibleWinners" :key="item.id" class="winner-item">
            <img :src="item.avatar" alt="" class="winner-avatar">
            <div class="winner-main">
              <p class="winner-amount">
                {{ t('homeLucky.gotPrefix') }} <strong>{{ item.amount }}</strong>
              </p>
              <p class="winner-name">
                {{ item.name }}
              </p>
            </div>
            <div class="winner-right">
              <span class="winner-time">{{ item.time }}</span>
              <van-icon name="arrow" />
            </div>
          </article>
        </div>

        <div class="winner-end">
          {{ t('homeLucky.emptyWinners') }}
        </div>
      </template>
    </section>

    <LuckyGrabModal
      v-model:show="grabModalVisible"
      :lucky-id="Number(pendingGrabTarget?.packet?.id || 0)"
      :grab-index="Number(pendingGrabTarget?.action?.seqNo || 0)"
      :sender-name="pendingGrabTarget?.packet?.username || t('grabModal.defaultSender')"
      @success="handleGrabSuccess"
      @close="closeGrabDialog"
    />
  </div>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
  background:
    radial-gradient(circle at 20% 10%, rgba(212, 175, 55, 0.18), transparent 30%),
    radial-gradient(circle at 80% 90%, rgba(255, 215, 0, 0.12), transparent 28%),
    linear-gradient(180deg, #3e0000 0%, #230000 62%, #160000 100%);
  padding: 12px;
  width: 100%;
  max-width: 100%;
  overflow: auto;
  color: #f9e8c6;
}

.home-carousel-card {
  position: relative;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(212, 175, 55, 0.45);
  box-shadow:
    0 10px 24px rgba(0, 0, 0, 0.35),
    inset 0 0 0 1px rgba(255, 248, 214, 0.12);
  background: #5a0000;
}

.home-swipe {
  width: 100%;
}

.banner-image {
  display: block;
  width: 100%;
  height: auto;
  aspect-ratio: 16 / 6;
  object-fit: cover;
}

:deep(.home-marquee.van-notice-bar) {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 24px;
  min-height: 24px;
  background: linear-gradient(180deg, rgba(133, 0, 0, 0) 0%, rgba(105, 0, 0, 0.95) 100%) !important;
  color: #ffe5a8;
  padding: 0 8px;
  font-size: 11px;
  line-height: 24px;
  z-index: 2;
}

:deep(.home-marquee .van-notice-bar__wrap) {
  height: 24px;
  background: transparent !important;
}

:deep(.home-marquee .van-notice-bar__content) {
  font-size: 11px;
  line-height: 24px;
}

.packet-section {
  margin-top: 12px;
}

.packet-skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.packet-header {
  border-top: 1px solid rgba(212, 175, 55, 0.35);
  padding-top: 10px;
  margin-bottom: 8px;
}

.packet-title-wrap {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 15px;
  color: #ffd98b;
  font-weight: 700;
  letter-spacing: 0.02em;
  text-shadow: 0 1px 0 rgba(0, 0, 0, 0.25);
}

.packet-title-wrap :deep(.van-icon) {
  color: #f7c548;
}

.packet-card {
  border-radius: 14px;
  border: 1px solid rgba(212, 175, 55, 0.36);
  background: linear-gradient(160deg, rgba(124, 0, 0, 0.98) 0%, rgba(80, 0, 0, 0.96) 70%, rgba(56, 0, 0, 0.96) 100%);
  overflow: hidden;
  margin-bottom: 8px;
  box-shadow:
    0 8px 20px rgba(0, 0, 0, 0.32),
    inset 0 0 0 1px rgba(255, 248, 214, 0.1);
}

.skeleton-card {
  margin-bottom: 0;
}

.skeleton-user-name {
  width: 120px;
}

.skeleton-amount {
  width: 84px;
}

.skeleton-packet-image {
  width: 90px;
  height: 90px;
  border-radius: 6px;
  overflow: hidden;
}

.packet-main {
  padding: 10px 12px 8px;
}

.packet-top {
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

.user-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid rgba(255, 222, 141, 0.55);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.22);
}

.user-name {
  font-size: 16px;
  line-height: 1;
  color: #fff0c9;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 160px;
}

.amount-wrap {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: rgba(255, 232, 186, 0.7);
}

.packet-amount {
  font-size: 15px;
  line-height: 1;
  font-weight: 700;
  color: #ffd66e;
  text-shadow: 0 0 10px rgba(255, 214, 110, 0.35);
}

.packet-card.done .packet-amount {
  color: #efc57f;
}

.packet-body {
  margin-top: 6px;
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.packet-image-wrap {
  width: 90px;
  flex: 0 0 90px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 56px;
  height: 20px;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 12px;
  line-height: 1;
  margin-bottom: 4px;
  font-weight: 700;
  border: 1px solid rgba(255, 248, 214, 0.45);
}

.packet-card.done .status-badge {
  background: linear-gradient(180deg, #7f7061 0%, #64574c 100%);
  color: #f6e8d2;
  border-color: rgba(255, 255, 255, 0.18);
}

.packet-image {
  width: 100%;
  height: auto;
  display: block;
}

.packet-info {
  min-width: 0;
  flex: 1;
}

.tags-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.tag {
  height: 20px;
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0 6px;
  font-size: 11px;
  line-height: 1;
  border: 1px solid transparent;
}

.tag.game {
  color: #ffe7bf;
  background: rgba(255, 215, 0, 0.12);
  border-color: rgba(255, 215, 0, 0.35);
}

.tag.progress {
  color: #ffeecf;
  background: rgba(255, 248, 214, 0.1);
  border-color: rgba(255, 248, 214, 0.3);
}

.packet-card.done .tag.game {
  background: rgba(255, 255, 255, 0.1);
  color: #e8d5b2;
  border-color: rgba(255, 255, 255, 0.2);
}

.packet-card.done .tag.progress {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.15);
}

.meta-row {
  margin-top: 5px;
  display: flex;
  align-items: center;
  gap: 6px;
  color: rgba(255, 229, 186, 0.9);
  font-size: 11px;
}

.rebate-text {
  margin: 6px 0 0;
  color: rgba(255, 248, 214, 0.78);
  font-size: 11px;
}

.time-text {
  margin: 6px 0 0;
  color: #ffd87f;
  font-size: 14px;
  line-height: 1;
  font-weight: 600;
}

.packet-card.done .time-text {
  color: #f0dbc0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 24px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  padding: 0 8px;
  font-size: 12px;
}

.packet-actions {
  border-top: 1px solid rgba(255, 248, 214, 0.12);
  padding: 8px 9px;
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 5px;
}

.packet-actions-skeleton {
  grid-template-columns: repeat(6, minmax(0, 1fr));
}

.skeleton-pill {
  width: 100%;
}

.action-pill {
  border: none;
  border-radius: 999px;
  background: linear-gradient(180deg, #e24b2d 0%, #b12715 100%);
  color: #fff3de;
  font-size: 8px;
  line-height: 1;
  min-height: 25px;
  box-shadow:
    inset 0 1px 0 rgba(255, 248, 214, 0.35),
    0 2px 6px rgba(0, 0, 0, 0.25);
}

.action-pill.grabbed {
  background: rgba(255, 255, 255, 0.15);
  color: rgba(255, 248, 214, 0.52);
  border: 1px solid rgba(255, 255, 255, 0.18);
  box-shadow: none;
  text-decoration: line-through;
}

.action-pill.mined {
  color: #ffe088;
}

.mine-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #8b5cf6;
  margin-right: 4px;
  display: inline-block;
  vertical-align: 1px;
}

.mine-text {
  margin-right: 4px;
  color: #ffd45d;
  text-decoration: none;
}

.action-pill.done {
  grid-column: 1 / -1;
  background: rgba(255, 255, 255, 0.13);
  color: rgba(255, 248, 214, 0.7);
  box-shadow: none;
}

.winner-section {
  margin-top: 12px;
}

.winner-card {
  border: 1px solid rgba(212, 175, 55, 0.34);
  border-radius: 14px;
  background: linear-gradient(170deg, rgba(116, 0, 0, 0.95), rgba(68, 0, 0, 0.95));
  overflow: hidden;
  box-shadow:
    0 8px 20px rgba(0, 0, 0, 0.28),
    inset 0 0 0 1px rgba(255, 248, 214, 0.08);
}

.winner-skeleton-card {
  border: 1px solid rgba(212, 175, 55, 0.34);
  border-radius: 14px;
  background: linear-gradient(170deg, rgba(116, 0, 0, 0.95), rgba(68, 0, 0, 0.95));
  overflow: hidden;
}

.winner-skeleton-time {
  width: 64px;
}

.winner-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
}

.winner-item + .winner-item {
  border-top: 1px solid rgba(255, 248, 214, 0.13);
}

.winner-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
  flex: 0 0 auto;
  border: 1px solid rgba(255, 222, 141, 0.58);
}

.winner-main {
  min-width: 0;
  flex: 1;
}

.winner-amount {
  margin: 0;
  color: #ffefca;
  font-size: 18px;
  line-height: 1.2;
}

.winner-amount strong {
  font-weight: 700;
}

.winner-name {
  margin: 8px 0 0;
  color: rgba(255, 229, 186, 0.72);
  font-size: 15px;
  line-height: 1;
}

.winner-right {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: rgba(255, 229, 186, 0.7);
}

.winner-time {
  color: #f4d7aa;
  font-size: 16px;
  line-height: 1;
}

.winner-end {
  margin-top: 12px;
  height: 30px;
  border-radius: 6px;
  border: 1px solid rgba(212, 175, 55, 0.3);
  background: rgba(123, 0, 0, 0.58);
  color: rgba(255, 229, 186, 0.66);
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

@media (max-width: 390px) {
  .packet-main {
    padding: 10px 10px 8px;
  }

  .user-name {
    max-width: 132px;
  }

  .packet-actions {
    gap: 4px;
  }
}
</style>

<route lang="json5">
{
  name: 'Home'
}
</route>
