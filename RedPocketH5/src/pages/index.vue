<script setup lang="ts">
import { showToast } from 'vant'
import { useRouter } from 'vue-router'
import { getLuckyPacketList, getLuckyRecentWinners } from '@/api/user'
import LuckyGrabModal from '@/components/LuckyGrabModal.vue'
import SendPacketForm from '@/components/SendPacketForm.vue'
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
const sendPacketModalVisible = ref(false)
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

function formatActionLabel(isOngoing: boolean, isGrabbed: boolean, amount: number, seqNo: number) {
  if (!isGrabbed && isOngoing)
    return t('homeLucky.grabAction', { seq: seqNo })
  if (isGrabbed && isOngoing && amount <= 0)
    return t('homeLucky.loadingLabel')
  return Number(amount || 0).toFixed(2)
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
        thunderText: t('homeLucky.thunderNo', { no: Number(packet?.thunder ?? 0) }),
        timeText: t('homeLucky.statusDone'),
        packetImage: imgRedpacketJpg,
        actions: (packet.actions || []).map((action: any) => ({
          ...action,
          label: formatActionLabel(false, Boolean(action?.isGrabbed), Number(action?.amount || 0), Number(action?.seqNo || 0)),
        })),
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
      label: formatActionLabel(
        isOngoing,
        Number(it?.isGrabbed) === 1,
        Number(it?.amount || 0),
        Number(it?.seqNo || 0),
      ),
    }))
  if (actions.length === 0 && isOngoing) {
    const number = Number(item?.number || 0)
    actions = Array.from({ length: number }, (_, idx) => ({
      seqNo: idx + 1,
      isGrabbed: false,
      isGrabMine: false,
      amount: 0,
      thunder: 0,
      label: formatActionLabel(true, false, 0, idx + 1),
    }))
  }

  return {
    id: Number(item?.id || 0),
    username: item?.senderName || 'User',
    avatar: item?.senderAvatar || DEFAULT_AVATAR,
    amount: formatAmount(item?.amount),
    thunder: Number(item?.thunder || 0),
    status: isOngoing ? 'ongoing' : 'done',
    statusText: isOngoing ? t('homeLucky.statusOngoing') : t('homeLucky.statusDone'),
    gameText: t('homeLucky.game'),
    progressText: t('homeLucky.progress', { grabbed: Number(item?.grabbedCount || 0), total: Number(item?.number || 0) }),
    thunderText: t('homeLucky.thunderNo', { no: Number(item?.thunder || 0) }),
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
  if (packet?.status !== 'ongoing')
    return
  if (action.isGrabbed)
    return
  pendingGrabTarget.value = { packet, action }
  grabModalVisible.value = true
}

function openSendPacketDialog() {
  if (!loggedIn.value) {
    showToast(t('homeLucky.loginFirst'))
    return
  }
  sendPacketModalVisible.value = true
}

function closeSendPacketDialog() {
  sendPacketModalVisible.value = false
}

async function handleSendPacketSuccess() {
  sendPacketModalVisible.value = false
  await loadPacketList()
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
        label: formatActionLabel(true, true, rawAmount, grabIndex),
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
      if (grabbedSeqNo > 0) {
        const idx = nextActions.findIndex((it: any) => Number(it?.seqNo) === grabbedSeqNo)
        if (idx >= 0) {
          const nextAmount = Number(lucky?.grabAmount || nextActions[idx]?.amount || 0)
          nextActions[idx] = {
            ...nextActions[idx],
            isGrabbed: true,
            thunder: Number(lucky?.isThunder || 0),
            amount: nextAmount,
            label: formatActionLabel(true, true, nextAmount, Number(nextActions[idx]?.seqNo || grabbedSeqNo)),
          }
        }
      }
      else {
        const idx = nextActions.findIndex((it: any) => !it.isGrabbed)
        if (idx >= 0) {
          const seqNo = Number(nextActions[idx]?.seqNo || 0)
          nextActions[idx] = {
            ...nextActions[idx],
            isGrabbed: true,
            label: formatActionLabel(true, true, Number(nextActions[idx]?.amount || 0), seqNo),
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
      thunder: Number(lucky?.thunder || packet?.thunder || 0),
      rebateText: Number.isFinite(totalThunderAmount) ? t('homeLucky.rebate', { amount: formatAmount(totalThunderAmount) }) : packet.rebateText,
      thunderText: t('homeLucky.thunderNo', { no: Number(lucky?.thunder || 0) }),
      progressText: t('homeLucky.progress', { grabbed: grabbedCount, total: packetNumber }),
      timeText: nextStatus === 'ongoing' ? packet.timeText : t('homeLucky.statusDone'),
      packetImage: nextStatus === 'ongoing' ? imgRedpacketGif : imgRedpacketJpg,
      actions: nextActions.map((action: any) => ({
        ...action,
        label: formatActionLabel(
          nextStatus === 'ongoing',
          Boolean(action?.isGrabbed),
          Number(action?.amount || 0),
          Number(action?.seqNo || 0),
        ),
      })),
    }
  })
  refreshPacketCountdowns()
}

function applyLuckyFinished(message: any) {
  const detail = message?.data || message
  const summary = detail?.summary
  const finance = detail?.finance
  const participants: any[] = detail?.participants || []

  const luckyId = Number(summary?.id || 0)
  if (!luckyId)
    return

  // 按 seqNo 建立参与记录索引
  const participantMap = new Map<number, any>()
  for (const p of participants)
    participantMap.set(Number(p.seqNo), p)

  packetList.value = packetList.value.map((packet) => {
    if (Number(packet.id) !== luckyId)
      return packet

    const total = Number(summary?.number || packet.actions?.length || 0)
    const existingActions: any[] = Array.isArray(packet.actions) ? [...packet.actions] : []

    // 确保每个 seqNo 都有槽位
    if (existingActions.length === 0) {
      for (let i = 1; i <= total; i++)
        existingActions.push({ seqNo: i, isGrabbed: false, isGrabMine: false, amount: 0, thunder: 0, label: '' })
    }

    const updatedActions = existingActions.map((action: any) => {
      const seqNo = Number(action.seqNo)
      const participant = participantMap.get(seqNo)
      if (participant) {
        const amount = Number(participant.amount || 0)
        return {
          ...action,
          isGrabbed: true,
          amount,
          thunder: Number(participant.isThunder || 0),
          label: formatActionLabel(false, true, amount, seqNo),
        }
      }
      return {
        ...action,
        isGrabbed: false,
        label: formatActionLabel(false, false, Number(action.amount || 0), seqNo),
      }
    })

    const grabbedCount = Number(summary?.grabbedCount ?? updatedActions.filter((a: any) => a.isGrabbed).length)
    const thunderIncome = Number(finance?.thunderIncome || 0)

    return {
      ...packet,
      status: 'done',
      statusText: t('homeLucky.statusDone'),
      amount: formatAmount(Number(summary?.amount || 0)),
      thunder: Number(summary?.thunder ?? packet.thunder),
      progressText: t('homeLucky.progress', { grabbed: grabbedCount, total }),
      hitsText: t('homeLucky.hitsCount', { count: Number(finance?.hitCount || 0) }),
      rebateText: t('homeLucky.rebate', { amount: formatAmount(thunderIncome) }),
      thunderText: t('homeLucky.thunderNo', { no: Number(summary?.thunder || 0) }),
      timeText: t('homeLucky.statusDone'),
      packetImage: imgRedpacketJpg,
      actions: updatedActions,
    }
  })
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
  wsClient.on('lucky_finished', applyLuckyFinished)
})

onBeforeUnmount(() => {
  if (countdownTimer)
    window.clearInterval(countdownTimer)
  wsClient.off('lucky_sent', applyLuckySent)
  wsClient.off('lucky_grabbed', applyLuckyBroadcast)
  wsClient.off('lucky_finished', applyLuckyFinished)
})
</script>

<template>
  <div class="home-page">
    <!-- Banner Carousel -->
    <section class="home-carousel-card">
      <van-swipe class="home-swipe" :autoplay="3200" lazy-render indicator-color="#d4af37" @change="onSwipeChange">
        <van-swipe-item v-for="(item, idx) in bannerList" :key="`${item.img}-${idx}`">
          <img :src="item.img" class="banner-image" :alt="`banner-${idx + 1}`">
        </van-swipe-item>
      </van-swipe>
      <div class="banner-stripe" />
      <van-notice-bar class="home-marquee" :scrollable="true" :text="marqueeText" />
    </section>

    <section class="send-promo-card">
      <div class="send-promo-copy">
        <p class="send-promo-eyebrow">
          {{ t('homeLucky.sendQuickEyebrow') }}
        </p>
        <p class="send-promo-text">
          {{ t('homeLucky.sendQuickText') }}
        </p>
      </div>
      <button type="button" class="send-promo-btn" @click="openSendPacketDialog">
        <van-icon name="gift-o" />
        <span>{{ t('homeLucky.sendQuickAction') }}</span>
      </button>
    </section>

    <!-- Hot Packets Section -->
    <section class="packet-section">
      <!-- <header class="packet-header">
        <div class="section-divider" />
        <div class="packet-title-wrap">
          <van-icon name="fire-o" />
          <span>{{ t('homeLucky.hotPackets') }}</span>
        </div>
      </header> -->

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
          <div class="packet-main">
            <div class="packet-top">
              <div class="user-wrap">
                <img :src="packet.avatar" alt="" class="user-avatar">
                <strong class="user-name">{{ packet.username }}</strong>
              </div>
              <button type="button" class="amount-wrap" @click="goLuckyDetail(packet)">
                <CoinAmount :text="packet.amount" class="packet-amount" />
                <van-icon name="arrow" />
              </button>
            </div>

            <div class="packet-body">
              <div class="packet-image-wrap">
                <span class="status-badge">{{ packet.statusText }}</span>
                <img :src="packet.packetImage" alt="" class="packet-image">
              </div>

              <div class="packet-info">
                <div class="tags-row">
                  <span class="tag game">🎮 {{ packet.gameText }}</span>
                  <span v-if="packet.thunderText" class="tag meta-tag">
                    {{ packet.thunderText }}
                  </span>
                  <span class="tag progress">{{ packet.progressText }}</span>
                  <span class="rebate-text">
                    {{ packet.rebateText }}
                  </span>
                  <span class="tag meta-tag">
                    {{ packet.hitsText }}
                  </span>
                </div>
                <div v-if="packet.actions?.length" class="packet-actions-inline">
                  <button
                    v-for="action in packet.actions" :key="`${packet.id}-${action.seqNo}`" type="button"
                    class="action-pill"
                    :class="{ grabbed: action.isGrabbed, mined: action.isGrabMine, locked: packet.status !== 'ongoing' && !action.isGrabbed }"
                    :disabled="packet.status !== 'ongoing' || action.isGrabbed"
                    @click="openGrabDialog(packet, action)"
                  >
                    <span v-if="action.thunder" aria-hidden="true">💣</span>
                    <span v-else-if="action.isGrabMine" class="mine-text">🎁 </span>
                    {{ action.label }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div class="packet-actions">
            <span v-if="packet.status === 'ongoing'" class="time-text time-text--footer">
              {{ packet.timeText }}
            </span>
            <button v-else type="button" class="action-pill done">
              {{ t('homeLucky.statusDone') }}
            </button>
          </div>
        </article>
      </template>
    </section>

    <!-- Recent Winners Section -->
    <section class="winner-section">
      <header class="packet-header">
        <div class="section-divider" />
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
                {{ t('homeLucky.gotPrefix') }} <strong><CoinAmount :text="item.amount" class="coin-amount--winner" /></strong>
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
          <span class="winner-end-line" />
          <span class="winner-end-text">{{ t('homeLucky.emptyWinners') }}</span>
          <span class="winner-end-line" />
        </div>
      </template>
    </section>

    <LuckyGrabModal
      v-model:show="grabModalVisible"
      :lucky-id="Number(pendingGrabTarget?.packet?.id || 0)"
      :grab-index="Number(pendingGrabTarget?.action?.seqNo || 0)"
      :sender-name="pendingGrabTarget?.packet?.username || t('grabModal.defaultSender')"
      :show-result-toast="false"
      @success="handleGrabSuccess"
      @close="closeGrabDialog"
    />

    <van-popup
      v-model:show="sendPacketModalVisible"
      round
      position="bottom"
      closeable
      class="send-packet-popup"
      close-icon-position="top-right"
      @closed="closeSendPacketDialog"
    >
      <section class="send-packet-modal">
        <div class="send-packet-modal__hero">
          <p class="send-packet-modal__eyebrow">
            {{ t('homeLucky.sendQuickEyebrow') }}
          </p>
          <h3 class="send-packet-modal__title">
            {{ t('sendPacketPage.packetTypeTitle') }}
          </h3>
          <p class="send-packet-modal__sub">
            {{ t('sendPacketPage.packetTypeSub') }}
          </p>
        </div>

        <SendPacketForm
          variant="modal"
          :show-intro="false"
          :show-tips="false"
          auto-reset
          @success="handleSendPacketSuccess"
        />
      </section>
    </van-popup>
  </div>
</template>

<style scoped>
/* ─── Page Shell ──────────────────────────────────── */
.home-page {
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
  padding: 12px 12px 90px;
  width: 100%;
  max-width: 100%;
  overflow: auto;
  color: #f9e8c6;
}

/* ─── Banner / Carousel ───────────────────────────── */
.home-carousel-card {
  position: relative;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgba(212, 175, 55, 0.55);
  box-shadow:
    0 12px 28px rgba(0, 0, 0, 0.45),
    inset 0 0 0 1px rgba(255, 248, 214, 0.15),
    0 0 0 1px rgba(128, 0, 0, 0.6);
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

/* Thai diagonal stripe overlay on banner */
.banner-stripe {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: repeating-linear-gradient(
    45deg,
    transparent,
    transparent 16px,
    rgba(212, 175, 55, 0.05) 16px,
    rgba(212, 175, 55, 0.05) 18px
  );
  z-index: 1;
}

:deep(.home-marquee.van-notice-bar) {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 26px;
  min-height: 26px;
  background: linear-gradient(180deg, rgba(133, 0, 0, 0) 0%, rgba(80, 0, 0, 0.96) 100%) !important;
  color: #ffd98b;
  padding: 0 10px;
  font-size: 11px;
  line-height: 26px;
  z-index: 2;
  letter-spacing: 0.03em;
}

:deep(.home-marquee .van-notice-bar__wrap) {
  height: 26px;
  background: transparent !important;
}

:deep(.home-marquee .van-notice-bar__content) {
  font-size: 11px;
  line-height: 26px;
}

/* ─── Section Header ──────────────────────────────── */
.packet-section {
  margin-top: 14px;
}

.send-promo-card {
  position: relative;
  margin-top: 14px;
  padding: 10px 12px;
  border-radius: 16px;
  border: 1px solid rgba(212, 175, 55, 0.4);
  background:
    radial-gradient(rgba(212, 175, 55, 1) 1px, transparent 1px),
    linear-gradient(155deg, rgba(116, 0, 0, 0.96) 0%, rgba(70, 0, 0, 0.96) 100%);
  background-size:
    18px 18px,
    100% 100%;
  box-shadow:
    0 10px 24px rgba(0, 0, 0, 0.35),
    inset 0 0 0 1px rgba(255, 248, 214, 0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.send-promo-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  border-radius: 16px 16px 0 0;
  background: linear-gradient(90deg, transparent 0%, #b8860b 16%, #ffd700 50%, #b8860b 84%, transparent 100%);
}

.send-promo-copy {
  min-width: 0;
  flex: 1;
}

.send-promo-eyebrow {
  margin: 0 0 2px;
  color: #ffd98b;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.send-promo-text {
  margin: 0;
  color: rgba(255, 232, 186, 0.82);
  font-size: 11px;
  line-height: 1.35;
}

.send-promo-btn {
  position: relative;
  overflow: hidden;
  flex: 0 0 auto;
  min-width: 98px;
  height: 34px;
  padding: 0 12px;
  border: 1px solid rgba(255, 248, 214, 0.42);
  border-radius: 999px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 11px;
  font-weight: 800;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  box-shadow:
    0 10px 18px rgba(75, 25, 0, 0.28),
    0 3px 0 rgba(126, 62, 0, 0.6),
    inset 0 1px 0 rgba(255, 255, 255, 0.34);
  transition:
    transform 0.16s ease,
    box-shadow 0.16s ease,
    filter 0.16s ease;
}

.send-promo-btn::before,
.send-promo-btn::after {
  content: '';
  position: absolute;
  left: 10px;
  right: 10px;
  pointer-events: none;
}

.send-promo-btn::before {
  top: 5px;
  height: 9px;
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.42), rgba(255, 255, 255, 0));
  opacity: 0.9;
}

.send-promo-btn::after {
  bottom: 4px;
  height: 8px;
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(126, 62, 0, 0), rgba(126, 62, 0, 0.28));
}

.send-promo-btn:active {
  transform: translateY(2px);
  filter: saturate(0.98);
  box-shadow:
    0 6px 12px rgba(75, 25, 0, 0.22),
    0 1px 0 rgba(126, 62, 0, 0.55),
    inset 0 1px 0 rgba(255, 255, 255, 0.26);
}

.send-promo-btn :deep(.van-icon) {
  font-size: 14px;
}

.packet-skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.packet-header {
  margin-bottom: 10px;
}

/* Gold gradient divider line */
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

.packet-title-wrap {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #ffd98b;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  text-shadow:
    0 1px 6px rgba(0, 0, 0, 0.5),
    0 0 10px rgba(212, 175, 55, 0.25);
}

.packet-title-wrap :deep(.van-icon) {
  font-size: 18px;
  color: #d4af37;
  filter: drop-shadow(0 0 5px rgba(212, 175, 55, 0.6));
}

/* ─── Packet Card ─────────────────────────────────── */
@keyframes cardGlow {
  0%,
  100% {
    box-shadow:
      0 10px 24px rgba(0, 0, 0, 0.38),
      inset 0 0 0 1px rgba(255, 248, 214, 0.12),
      0 0 0 1px rgba(212, 175, 55, 0.35);
  }
  50% {
    box-shadow:
      0 10px 28px rgba(0, 0, 0, 0.42),
      inset 0 0 0 1px rgba(255, 248, 214, 0.18),
      0 0 0 1px rgba(212, 175, 55, 0.65),
      0 0 12px rgba(212, 175, 55, 0.12);
  }
}

.packet-card {
  position: relative;
  border-radius: 16px;
  border: 1px solid rgba(212, 175, 55, 0.45);
  background: linear-gradient(160deg, rgba(140, 0, 0, 0.98) 0%, rgba(90, 0, 0, 0.97) 55%, rgba(55, 0, 0, 0.97) 100%);
  overflow: hidden;
  margin-bottom: 6px;
  box-shadow:
    0 8px 18px rgba(0, 0, 0, 0.34),
    inset 0 0 0 1px rgba(255, 248, 214, 0.12),
    0 0 0 1px rgba(212, 175, 55, 0.35);
}

/* Ongoing cards get a subtle pulsing gold border */
.packet-card.ongoing {
  animation: cardGlow 2.4s ease-in-out infinite;
}

/* Gold top accent stripe (like grap.html flap border-bottom: 4px solid #d4af37) */
.packet-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    #b8860b 15%,
    #ffd700 40%,
    #d4af37 60%,
    #b8860b 85%,
    transparent 100%
  );
  pointer-events: none;
  z-index: 4;
  border-radius: 16px 16px 0 0;
}

/* Gold dot pattern watermark (from grap.html gift-card::before) */
.packet-card::before {
  content: '';
  position: absolute;
  inset: 3px 0 0;
  background-image: radial-gradient(rgba(212, 175, 55, 1) 1px, transparent 1px);
  background-size: 18px 18px;
  opacity: 0.05;
  pointer-events: none;
  z-index: 0;
  border-radius: 16px;
}

/* Thai corner bracket decorations (from grap.html .corner) */
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
  width: 60px;
  height: 60px;
  border-radius: 6px;
  overflow: hidden;
}

.packet-main {
  position: relative;
  z-index: 1;
  padding: 8px 9px 5px;
}

.packet-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.user-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(212, 175, 55, 0.65);
  box-shadow:
    0 0 8px rgba(212, 175, 55, 0.25),
    0 4px 8px rgba(0, 0, 0, 0.28);
}

.user-name {
  font-size: 12px;
  line-height: 1;
  color: #fff0c9;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 160px;
  letter-spacing: 0.02em;
}

.amount-wrap {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 0;
  border: 0;
  background: transparent;
  color: rgba(255, 232, 186, 0.7);
  cursor: pointer;
}

.amount-wrap:active {
  transform: translateY(1px);
}

/* CoinAmount size overrides */
.packet-amount :deep(.coin-amount-icon) {
  width: 15px;
  height: 15px;
}

.coin-amount--winner :deep(.coin-amount-icon) {
  width: 18px;
  height: 18px;
}

/* Gold gradient amount text (like grap.html .amount) */
.packet-amount {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  line-height: 1;
  font-weight: 700;
  background: linear-gradient(to bottom, #cfb53b, #ffd700, #d4af37);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.25));
}

.packet-card.done .packet-amount {
  background: linear-gradient(to bottom, #c8a96e, #a88c58);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.packet-body {
  margin-top: 4px;
  display: flex;
  align-items: flex-start;
  gap: 6px;
}

.packet-image-wrap {
  width: 60px;
  flex: 0 0 60px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* Status badge — gold gradient from grap.html */
.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 42px;
  height: 16px;
  border-radius: 999px;
  background: linear-gradient(180deg, #ffdf87 0%, #d4af37 100%);
  color: #5a1b00;
  font-size: 8px;
  line-height: 1;
  margin: 0 auto 2px;
  font-weight: 700;
  letter-spacing: 0.03em;
  border: 1px solid rgba(255, 248, 214, 0.5);
  box-shadow:
    0 2px 6px rgba(0, 0, 0, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  text-align: center;
}

.packet-card.done .status-badge {
  background: linear-gradient(180deg, #7f7061 0%, #575049 100%);
  color: #f6e8d2;
  border-color: rgba(255, 255, 255, 0.15);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.packet-image {
  width: 100%;
  height: auto;
  display: block;
}

.packet-info {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.tags-row {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: nowrap;
  overflow: hidden;
}

.tag {
  height: 15px;
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0 3px;
  font-size: 7px;
  line-height: 1;
  border: 1px solid transparent;
  letter-spacing: 0.02em;
  white-space: nowrap;
  flex-shrink: 0;
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

.packet-card.done .tag.game {
  background: rgba(255, 255, 255, 0.09);
  color: #e8d5b2;
  border-color: rgba(255, 255, 255, 0.18);
}

.packet-card.done .tag.progress {
  background: rgba(255, 255, 255, 0.07);
  border-color: rgba(255, 255, 255, 0.14);
}

.tag.meta-tag {
  color: rgba(255, 229, 186, 0.82);
  background: rgba(255, 248, 214, 0.07);
  border-color: rgba(255, 248, 214, 0.18);
}

.meta-row {
  margin-top: 1px;
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

/* Each meta item as a refined mini badge */
.meta-row span {
  display: inline-flex;
  align-items: center;
  height: 14px;
  padding: 0 4px;
  border-radius: 999px;
  background: rgba(255, 248, 214, 0.07);
  border: 1px solid rgba(255, 248, 214, 0.18);
  color: rgba(255, 229, 186, 0.82);
  font-size: 7px;
  letter-spacing: 0.02em;
  line-height: 1;
}

.rebate-text {
  margin: 0;
  display: inline-flex;
  align-self: flex-start;
  align-items: center;
  width: fit-content;
  max-width: 100%;
  min-height: 14px;
  padding: 2px 5px;
  border-radius: 999px;
  background: rgba(212, 175, 55, 0.1);
  border: 1px solid rgba(212, 175, 55, 0.28);
  color: rgba(255, 232, 160, 0.85);
  font-size: 7px;
  letter-spacing: 0.01em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.time-text {
  margin: 1px 0 0;
  display: block;
  align-self: flex-start;
  width: fit-content;
  max-width: 100%;
  color: #ffd87f;
  font-size: 10px;
  line-height: 1.2;
  font-weight: 700;
  letter-spacing: 0.03em;
  text-shadow: 0 0 8px rgba(255, 216, 127, 0.35);
  white-space: normal;
  word-break: break-word;
}

.packet-card.done .time-text {
  margin-top: 0;
  color: #f0dbc0;
  display: inline-block;
  min-height: 16px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.11);
  padding: 2px 6px;
  font-size: 9px;
  border: 1px solid rgba(255, 255, 255, 0.12);
}

/* ─── Action Pills ────────────────────────────────── */
.packet-actions-inline {
  margin-top: 3px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 4px;
}

.packet-actions {
  position: relative;
  z-index: 1;
  border-top: 1px solid rgba(212, 175, 55, 0.22);
  padding: 6px 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.2) 0%, rgba(0, 0, 0, 0.35) 100%);
}

.packet-actions-skeleton {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.skeleton-pill {
  width: 100%;
}

/* From grap.html's pocket gradient (#900000 → #600000) */
.action-pill {
  border: none;
  border-radius: 999px;
  background: linear-gradient(180deg, #9e1010 0%, #6a0000 100%);
  color: #fff3de;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px 3px;
  font-size: 7px;
  line-height: 1.15;
  text-align: center;
  white-space: normal;
  word-break: break-word;
  min-height: 28px;
  box-shadow:
    inset 0 1px 0 rgba(212, 175, 55, 0.45),
    0 2px 6px rgba(0, 0, 0, 0.3);
  letter-spacing: 0.01em;
  transition:
    opacity 0.15s,
    transform 0.1s;
}


.action-pill:not(:disabled):active {
  transform: scale(0.96);
}

.action-pill.grabbed {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 248, 214, 0.45);
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: none;
  text-decoration: line-through;
}

.action-pill.locked {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 248, 214, 0.72);
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: none;
  cursor: not-allowed;
}

.action-pill.mined {
  color: #ffe088;
}

.mine-text {
  margin-right: 3px;
  color: #ffd45d;
  text-decoration: none;
}

.action-pill.done {
  width: 100%;
  background: rgba(212, 175, 55, 0.08);
  color: rgba(255, 248, 214, 0.6);
  border: 1px solid rgba(212, 175, 55, 0.28);
  box-shadow: none;
  letter-spacing: 0.06em;
  font-size: 9px;
  min-height: 24px;
}

.time-text--footer {
  margin: 0;
  align-self: auto;
  width: 100%;
  text-align: center;
}

/* ─── Winner Section ──────────────────────────────── */
.winner-section {
  margin-top: 14px;
}

.winner-card {
  border: 1px solid rgba(212, 175, 55, 0.45);
  border-radius: 16px;
  background: linear-gradient(170deg, rgba(125, 0, 0, 0.97), rgba(60, 0, 0, 0.97));
  overflow: hidden;
  isolation: isolate;
  box-shadow:
    0 10px 24px rgba(0, 0, 0, 0.35),
    inset 0 0 0 1px rgba(255, 248, 214, 0.09),
    0 0 0 1px rgba(212, 175, 55, 0.3);
  position: relative;
}

/* Gold top accent on winner card too */
.winner-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    #b8860b 15%,
    #ffd700 40%,
    #d4af37 60%,
    #b8860b 85%,
    transparent 100%
  );
  pointer-events: none;
  z-index: 0;
}

/* Gold dot watermark on winner card */
.winner-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(rgba(212, 175, 55, 1) 1px, transparent 1px);
  background-size: 18px 18px;
  opacity: 0.04;
  pointer-events: none;
  z-index: 0;
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
  position: relative;
  z-index: 1;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 13px 14px;
}

.winner-item + .winner-item {
  border-top: 1px solid rgba(212, 175, 55, 0.15);
}

.winner-avatar {
  width: 46px;
  height: 46px;
  border-radius: 50%;
  object-fit: cover;
  flex: 0 0 auto;
  border: 2px solid rgba(212, 175, 55, 0.7);
  box-shadow:
    0 0 10px rgba(212, 175, 55, 0.28),
    0 3px 8px rgba(0, 0, 0, 0.35);
}

.winner-main {
  min-width: 0;
  flex: 1;
}

.winner-amount {
  margin: 0;
  color: #ffefca;
  font-size: 16px;
  line-height: 1.3;
  letter-spacing: 0.01em;
}

/* Gold gradient on winning amount (like grap.html .amount) */
.winner-amount strong {
  display: inline-flex;
  align-items: center;
  font-weight: 700;
  background: linear-gradient(to bottom, #cfb53b, #ffd700, #d4af37);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  filter: drop-shadow(0 1px 1px rgba(0, 0, 0, 0.2));
}

.winner-name {
  margin: 6px 0 0;
  color: rgba(255, 229, 186, 0.68);
  font-size: 14px;
  line-height: 1;
}

.winner-right {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: rgba(255, 229, 186, 0.65);
}

.winner-time {
  color: #f4d7aa;
  font-size: 15px;
  line-height: 1;
}

/* Ornamental "end" divider */
.winner-end {
  margin-top: 12px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.winner-end-line {
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(212, 175, 55, 0.45), transparent);
}

.winner-end-text {
  color: rgba(255, 229, 186, 0.55);
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  white-space: nowrap;
}

:deep(.send-packet-popup.van-popup) {
  max-height: calc(100vh - 56px);
  background:
    radial-gradient(circle at 12% 10%, rgba(212, 175, 55, 0.18), transparent 22%),
    linear-gradient(180deg, #540000 0%, #280000 100%);
  border-radius: 24px 24px 0 0;
  border: 1px solid rgba(212, 175, 55, 0.34);
  box-shadow: 0 -12px 32px rgba(0, 0, 0, 0.48);
}

:deep(.send-packet-popup .van-popup__close-icon) {
  color: #ffd98b;
}

.send-packet-modal {
  padding: 18px 14px calc(14px + env(safe-area-inset-bottom));
}

.send-packet-modal__hero {
  margin-bottom: 12px;
  text-align: center;
}

.send-packet-modal__eyebrow {
  margin: 0 0 6px;
  color: #ffd98b;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.send-packet-modal__title {
  margin: 0;
  color: #fff0c9;
  font-size: 20px;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.send-packet-modal__sub {
  margin: 6px 0 0;
  color: rgba(255, 229, 186, 0.7);
  font-size: 12px;
  line-height: 1.45;
}

/* ─── Responsive ──────────────────────────────────── */
@media (max-width: 390px) {
  .send-promo-card {
    padding: 9px 10px;
    gap: 8px;
  }

  .send-promo-btn {
    min-width: 90px;
    padding: 0 10px;
  }

  .packet-main {
    padding: 7px 8px 5px;
  }

  .user-name {
    max-width: 104px;
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
