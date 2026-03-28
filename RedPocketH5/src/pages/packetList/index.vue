<script setup lang="ts">
import { showConfirmDialog, showToast } from 'vant'
import type { LuckyPlayType, ParityChoice } from '@/utils/lucky-play'
import { getLuckyPacketList, getPrizePoolBalance } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import LuckyGrabModal from '@/components/LuckyGrabModal.vue'
import ParityChoiceDialog from '@/components/ParityChoiceDialog.vue'
import SendPacketForm from '@/components/SendPacketForm.vue'
import { isLogin } from '@/utils/auth'
import { formatCurrency } from '@/utils/currency'
import { isParityPlayType, resolveLuckyPlayType } from '@/utils/lucky-play'
import wsClient from '@/plugins/websocket'
import imgAvatarPlaceholder from '@/assets/images/avatar-placeholder.png'
import imgRedpacketGif from '@/assets/images/redpacket.gif'
import imgRedpacketJpg from '@/assets/images/redpacket.jpg'
import imgCoin from '@/assets/svg/coin.svg'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const DEFAULT_AVATAR = imgAvatarPlaceholder
const currentMode = ref<0 | 1>(0)
const packetList = ref<any[]>([])
const packetLoading = ref(false)
const grabModalVisible = ref(false)
const parityChoiceVisible = ref(false)
const sendPacketModalVisible = ref(false)
const pendingGrabTarget = ref<{ packet: any, action: any, choice?: ParityChoice | null } | null>(null)
const pendingParityTarget = ref<{ packet: any, action: any } | null>(null)
const prizePoolBalance = ref<number>(0)
let countdownTimer: number | undefined

const modePlayType = computed<LuckyPlayType>(() => currentMode.value === 1 ? 'parity' : 'thunder')
const pageTitle = computed(() => currentMode.value === 1 ? t('packetListPage.modeParity') : t('packetListPage.modeThunder'))
const showPacketEmpty = computed(() => !packetLoading.value && packetList.value.length === 0)
const showPacketLoading = computed(() => packetLoading.value && packetList.value.length === 0)

function syncModeFromRoute() {
  currentMode.value = String(route.query.mode || '0') === '1' ? 1 : 0
}

function goBack() {
  router.back()
}

function openSendPacketDialog() {
  if (!isLogin()) {
    promptLogin()
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

function promptLogin() {
  showConfirmDialog({
    title: t('homeLucky.loginDialogTitle'),
    message: t('homeLucky.loginDialogMsg'),
    confirmButtonText: t('homeLucky.loginDialogConfirm'),
    cancelButtonText: t('common.cancel'),
  }).then(() => {
    router.push('/login')
  }).catch(() => {})
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

function getPlayTypeText(playType: LuckyPlayType) {
  return playType === 'parity' ? t('homeLucky.playTypeParity') : t('homeLucky.playTypeThunder')
}

function formatActionLabel(isOngoing: boolean, isGrabged: boolean, amount: number, seqNo: number) {
  if (!isGrabged && isOngoing)
    return t('homeLucky.grabAction', { seq: seqNo })
  if (isGrabged && isOngoing && amount <= 0)
    return t('homeLucky.loadingLabel')
  if (!isGrabged && !isOngoing)
    return amount > 0 ? Number(amount).toFixed(2) : '—'
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
        ruleText: packet.playType === 'parity'
          ? t('homeLucky.paritySelectHint')
          : t('homeLucky.thunderNo', { no: Number(packet?.thunder ?? 0) }),
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
  const playType = resolveLuckyPlayType(item, modePlayType.value)
  const isOngoing = Number(item?.status) === 1
  const senderWinAmount = (item?.items || []).reduce((sum: number, it: any) => sum + Number(it?.thunderAmount || 0), 0)
  const parityChoiceCount = Number(item?.choiceCount ?? item?.selectedCount ?? item?.settledCount ?? 0)
  let actions = (item?.items || []).map((it: any) => ({
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
    playType,
    playTypeText: getPlayTypeText(playType),
    username: item?.senderName || 'User',
    avatar: item?.senderAvatar || DEFAULT_AVATAR,
    amount: formatAmount(item?.amount),
    thunder: Number(item?.thunder || 0),
    status: isOngoing ? 'ongoing' : 'done',
    statusText: isOngoing ? t('homeLucky.statusOngoing') : t('homeLucky.statusDone'),
    gameText: t('homeLucky.game'),
    progressText: t('homeLucky.progress', { grabbed: Number(item?.grabbedCount || 0), total: Number(item?.number || 0) }),
    ruleText: playType === 'parity'
      ? t('homeLucky.paritySelectHint')
      : t('homeLucky.thunderNo', { no: Number(item?.thunder || 0) }),
    statText: playType === 'parity'
      ? (parityChoiceCount > 0 ? t('homeLucky.paritySelectedCount', { count: parityChoiceCount }) : '')
      : t('homeLucky.rebate', { amount: formatAmount(senderWinAmount) }),
    timeText: isOngoing ? t('homeLucky.remainingTime', { time: item?.remainingText || '00:00' }) : t('homeLucky.statusDone'),
    packetImage: isOngoing ? imgRedpacketGif : imgRedpacketJpg,
    expireAtMs: item?.expireTime ? new Date(item.expireTime).getTime() : 0,
    actions,
  }
}

async function loadPacketList() {
  if (packetLoading.value)
    return
  try {
    packetLoading.value = true
    const { data } = await getLuckyPacketList({
      currentPage: 0,
      pageSize: 20,
      gameMode: currentMode.value,
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
  if (!isLogin()) {
    promptLogin()
    return
  }
  if (packet?.status !== 'ongoing' || action.isGrabbed)
    return
  if (isParityPlayType(packet.playType)) {
    pendingParityTarget.value = { packet, action }
    parityChoiceVisible.value = true
    return
  }
  pendingGrabTarget.value = { packet, action, choice: null }
  grabModalVisible.value = true
}

function handleParityChoiceConfirm(choice: ParityChoice) {
  if (!pendingParityTarget.value)
    return
  pendingGrabTarget.value = { ...pendingParityTarget.value, choice }
  parityChoiceVisible.value = false
  pendingParityTarget.value = null
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
    if (idx >= 0)
      nextActions[idx] = { ...nextActions[idx], isGrabbed: true, amount: rawAmount }
    const grabbedCount = nextActions.filter((it: any) => it.isGrabbed).length
    const packetNumber = Number(packet?.actions?.length || 0)
    const remaining = packetNumber - grabbedCount
    return {
      ...packet,
      progressText: t('homeLucky.progress', { grabbed: grabbedCount, total: packetNumber }),
      actions: nextActions.map((action: any) => {
        const isNewGrab = Number(action?.seqNo) === grabIndex && action.isGrabbed
        const showLoading = isNewGrab && remaining === 1
        return {
          ...action,
          displayLoading: showLoading,
          label: showLoading
            ? t('homeLucky.loadingLabel')
            : formatActionLabel(true, Boolean(action?.isGrabbed), Number(action?.amount || 0), Number(action?.seqNo || 0)),
        }
      }),
    }
  })
}

function applyLuckySent(message: any) {
  const lucky = message?.data || message
  if (resolveLuckyPlayType(lucky, 'thunder') !== modePlayType.value)
    return
  const mapped = mapPacket(lucky)
  if (!mapped.id)
    return
  const existingIndex = packetList.value.findIndex(item => Number(item.id) === Number(mapped.id))
  if (existingIndex >= 0) {
    const next = [...packetList.value]
    next[existingIndex] = { ...next[existingIndex], ...mapped }
    packetList.value = next
    return
  }
  packetList.value = [mapped, ...packetList.value].slice(0, 20)
  refreshPacketCountdowns()
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
    }

    const grabbedCount = nextActions.filter((it: any) => it.isGrabbed).length
    const packetNumber = Number(lucky?.number || packet?.actions?.length || 0)
    return {
      ...packet,
      status: nextStatus,
      statusText: nextStatus === 'ongoing' ? t('homeLucky.statusOngoing') : t('homeLucky.statusDone'),
      amount: formatAmount(lucky?.amount),
      thunder: Number(lucky?.thunder || packet?.thunder || 0),
      statText: packet.playType === 'parity'
        ? packet.statText
        : (Number.isFinite(totalThunderAmount) ? t('homeLucky.rebate', { amount: formatAmount(totalThunderAmount) }) : packet.statText),
      ruleText: packet.playType === 'parity'
        ? t('homeLucky.paritySelectHint')
        : t('homeLucky.thunderNo', { no: Number(lucky?.thunder || 0) }),
      progressText: t('homeLucky.progress', { grabbed: grabbedCount, total: packetNumber }),
      timeText: nextStatus === 'ongoing' ? packet.timeText : t('homeLucky.statusDone'),
      packetImage: nextStatus === 'ongoing' ? imgRedpacketGif : imgRedpacketJpg,
      actions: nextActions,
    }
  })
}

function applyLuckyFinished(message: any) {
  const detail = message?.data || message
  const summary = detail?.summary
  const finance = detail?.finance
  const participants: any[] = detail?.participants || []
  const luckyId = Number(summary?.id || 0)
  if (!luckyId)
    return

  const participantMap = new Map<number, any>()
  for (const p of participants)
    participantMap.set(Number(p.seqNo), p)

  packetList.value = packetList.value.map((packet) => {
    if (Number(packet.id) !== luckyId)
      return packet
    const total = Number(summary?.number || packet.actions?.length || 0)
    const existingActions: any[] = Array.isArray(packet.actions) ? [...packet.actions] : []
    if (existingActions.length === 0) {
      for (let i = 1; i <= total; i++)
        existingActions.push({ seqNo: i, isGrabbed: false, isGrabMine: false, amount: 0, thunder: 0, label: '' })
    }
    const updatedActions = existingActions.map((action: any) => {
      const participant = participantMap.get(Number(action.seqNo))
      if (!participant)
        return { ...action, isGrabbed: false, displayLoading: false, label: formatActionLabel(false, false, Number(action.amount || 0), Number(action.seqNo)) }
      const grabbed = Number(participant.isGrabbed ?? 1) === 1
      const amount = Number(participant.amount || 0)
      return {
        ...action,
        isGrabbed: grabbed,
        amount,
        thunder: grabbed ? Number(participant.isThunder || 0) : 0,
        displayLoading: false,
        label: formatActionLabel(false, grabbed, amount, Number(action.seqNo)),
      }
    })
    const grabbedCount = Number(summary?.grabbedCount ?? updatedActions.filter((a: any) => a.isGrabbed).length)
    return {
      ...packet,
      status: 'done',
      statusText: t('homeLucky.statusDone'),
      amount: formatAmount(Number(summary?.amount || 0)),
      thunder: Number(summary?.thunder ?? packet.thunder),
      progressText: t('homeLucky.progress', { grabbed: grabbedCount, total }),
      statText: packet.playType === 'parity' ? packet.statText : t('homeLucky.rebate', { amount: formatAmount(Number(finance?.thunderIncome || 0)) }),
      ruleText: packet.playType === 'parity' ? t('homeLucky.paritySelectHint') : t('homeLucky.thunderNo', { no: Number(summary?.thunder || 0) }),
      timeText: t('homeLucky.statusDone'),
      packetImage: imgRedpacketJpg,
      actions: updatedActions,
    }
  })
}

async function loadPrizePoolBalance() {
  try {
    const { data } = await getPrizePoolBalance('lucky')
    prizePoolBalance.value = Number(data?.balance ?? 0)
  }
  catch {}
}

function applyPrizePoolBalance(message: any) {
  const payload = message?.data || message
  if (payload?.poolCode && payload.poolCode !== 'lucky')
    return
  prizePoolBalance.value = Number(payload?.balance ?? 0)
}

// Jackpot rolling display
const displayBalance = ref(0)
const isFlashing = ref(false)
let jpRafId: number | undefined

function rollToBalance(target: number) {
  cancelAnimationFrame(jpRafId!)
  const step = () => {
    const diff = target - displayBalance.value
    if (Math.abs(diff) > 0.5) {
      displayBalance.value += diff * 0.12
      jpRafId = requestAnimationFrame(step)
    }
    else {
      displayBalance.value = target
    }
  }
  jpRafId = requestAnimationFrame(step)
}

watch(prizePoolBalance, (val) => {
  isFlashing.value = false
  nextTick(() => { isFlashing.value = true })
  setTimeout(() => { isFlashing.value = false }, 600)
  rollToBalance(val)
}, { immediate: true })

watch(() => route.query.mode, async () => {
  syncModeFromRoute()
  await loadPacketList()
}, { immediate: true })

onMounted(() => {
  countdownTimer = window.setInterval(refreshPacketCountdowns, 1000)
  wsClient.on('lucky_sent', applyLuckySent)
  wsClient.on('lucky_grabbed', applyLuckyBroadcast)
  wsClient.on('lucky_finished', applyLuckyFinished)
  wsClient.on('prize_pool_balance', applyPrizePoolBalance)
  loadPrizePoolBalance()
})

onBeforeUnmount(() => {
  cancelAnimationFrame(jpRafId!)
  if (countdownTimer)
    window.clearInterval(countdownTimer)
  wsClient.off('lucky_sent', applyLuckySent)
  wsClient.off('lucky_grabbed', applyLuckyBroadcast)
  wsClient.off('lucky_finished', applyLuckyFinished)
  wsClient.off('prize_pool_balance', applyPrizePoolBalance)
})
</script>

<template>
  <div class="packet-list-page">
    <AppPageHeader :title="pageTitle" @back="goBack" @right-click="openSendPacketDialog">
      <template #right><van-icon name="gift-o" /></template>
    </AppPageHeader>

    <div class="jackpot-panel" :class="{ 'jackpot-panel--flash': isFlashing }">
      <div class="jp-label">{{ t('packetListPage.prizePoolLabel') }}</div>
      <div class="jp-amount-box">
        <img :src="imgCoin" class="jp-coin-icon" alt="">
        <span class="jp-amount-num">{{ Math.floor(displayBalance).toLocaleString('en-US') }}</span>
      </div>
    </div>

    <button type="button" class="send-entry-btn" @click="openSendPacketDialog">
      <span class="send-entry-btn__badge">{{ t('homeLucky.sendQuickEyebrow') }}</span>
      <span class="send-entry-btn__main">
        <span class="send-entry-btn__icon"><van-icon name="gift-o" /></span>
        <span class="send-entry-btn__copy">
          <strong>{{ t('homeLucky.sendQuickAction') }}</strong>
          <small>{{ pageTitle }}</small>
        </span>
      </span>
      <van-icon name="arrow" class="send-entry-btn__arrow" />
    </button>

    <section class="packet-section">
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
                  <span class="tag play-type-tag" :class="packet.playType">{{ packet.playTypeText }}</span>
                  <span class="tag progress">{{ packet.progressText }}</span>
                </div>
                <div class="meta-row">
                  <span v-if="packet.ruleText">{{ packet.ruleText }}</span>
                  <span v-if="packet.statText" class="meta-chip--accent"><CurrencyText :text="packet.statText" /></span>
                </div>
                <div v-if="packet.actions?.length" class="packet-actions-inline">
                  <button
                    v-for="action in packet.actions"
                    :key="`${packet.id}-${action.seqNo}`"
                    type="button"
                    class="action-pill"
                    :class="{ grabbed: action.isGrabbed, mined: action.isGrabMine, locked: packet.status !== 'ongoing' && !action.isGrabbed }"
                    :disabled="packet.status !== 'ongoing' || action.isGrabbed"
                    @click="openGrabDialog(packet, action)"
                  >
                    <span v-if="action.thunder" aria-hidden="true">💣</span>
                    <span v-else-if="action.isGrabMine" class="mine-text">🎁 </span>
                    <span v-else-if="packet.playType === 'parity' && !action.isGrabbed && packet.status === 'ongoing'" class="choice-mark">奇/偶</span>
                    <CoinAmount v-if="action.amount > 0 && !action.thunder && !action.displayLoading && (action.isGrabbed || packet.status !== 'ongoing')" :text="`${action.amount.toFixed(2)}`" />
                    <template v-else>{{ action.label }}</template>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div class="packet-actions">
            <span v-if="packet.status === 'ongoing'" class="time-text time-text--footer">{{ packet.timeText }}</span>
            <button v-else type="button" class="action-pill done">{{ t('homeLucky.statusDone') }}</button>
          </div>
        </article>
      </template>
    </section>

    <LuckyGrabModal
      v-model:show="grabModalVisible"
      :lucky-id="Number(pendingGrabTarget?.packet?.id || 0)"
      :grab-index="Number(pendingGrabTarget?.action?.seqNo || 0)"
      :choice="pendingGrabTarget?.choice || ''"
      :sender-name="pendingGrabTarget?.packet?.username || t('grabModal.defaultSender')"
      :show-result-toast="false"
      @success="handleGrabSuccess"
      @close="closeGrabDialog"
    />

    <ParityChoiceDialog
      v-model:show="parityChoiceVisible"
      :sender-name="pendingParityTarget?.packet?.username || t('grabModal.defaultSender')"
      @confirm="handleParityChoiceConfirm"
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
          <p class="send-packet-modal__eyebrow">{{ t('homeLucky.sendQuickEyebrow') }}</p>
          <h3 class="send-packet-modal__title">{{ t('sendPacketPage.playTypeTitle') }}</h3>
          <p class="send-packet-modal__sub">{{ t('sendPacketPage.packetTypeSub') }}</p>
        </div>

        <SendPacketForm
          variant="modal"
          :show-intro="false"
          :show-play-type="false"
          :show-tips="false"
          :default-play-type="modePlayType"
          lock-play-type
          auto-reset
          @success="handleSendPacketSuccess"
        />
      </section>
    </van-popup>
  </div>
</template>

<style scoped>
.packet-list-page {
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
  padding: 8px 12px calc(90px + env(safe-area-inset-bottom));
}

.jackpot-panel {
  margin-top: 12px;
  padding: 14px 20px 16px;
  border-radius: 18px;
  background: rgba(0, 0, 0, 0.58);
  border: 2px solid #ffbb00;
  box-shadow: 0 0 22px rgba(0, 0, 0, 0.75), inset 0 0 16px rgba(255, 184, 0, 0.1);
  text-align: center;
  animation: jpPanelPulse 2s infinite ease-in-out;
  position: relative;
  overflow: hidden;
}

.jackpot-panel--flash {
  animation: jpFlash 0.5s ease-out forwards, jpPanelPulse 2s 0.5s infinite ease-in-out;
}

@keyframes jpPanelPulse {
  0%, 100% { box-shadow: 0 0 22px rgba(0, 0, 0, 0.75), inset 0 0 16px rgba(255, 184, 0, 0.1); }
  50% { box-shadow: 0 0 38px rgba(255, 184, 0, 0.28), inset 0 0 20px rgba(255, 184, 0, 0.15); border-color: #fff8c0; }
}

@keyframes jpFlash {
  0%   { transform: scale(1);    filter: brightness(1); }
  30%  { transform: scale(1.025); filter: brightness(1.7) drop-shadow(0 0 14px #ffbb00); }
  100% { transform: scale(1);    filter: brightness(1); }
}

.jp-label {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.55);
  letter-spacing: 4px;
  text-transform: uppercase;
  margin-bottom: 10px;
}

.jp-amount-box {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.jp-coin-icon {
  width: 34px;
  height: 34px;
  flex-shrink: 0;
}

.jp-amount-num {
  font-size: 38px;
  font-weight: 900;
  line-height: 1;
  background: linear-gradient(180deg, #fff5c3 0%, #ffbb00 45%, #e19d00 58%, #c87000 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  filter: drop-shadow(0 2px 6px rgba(0, 0, 0, 0.8));
  font-variant-numeric: tabular-nums;
}

.send-entry-btn {
  width: 100%;
  margin-top: 12px;
  padding: 12px 14px;
  border: 1px solid rgba(255, 227, 153, 0.58);
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(255, 244, 205, 0.14), rgba(255, 244, 205, 0) 38%),
    linear-gradient(135deg, #f6d978 0%, #e3b84a 48%, #c88a1f 100%);
  box-shadow:
    0 14px 26px rgba(0, 0, 0, 0.26),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  color: #592400;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.send-entry-btn__badge {
  flex: 0 0 auto;
  height: 22px;
  padding: 0 8px;
  border-radius: 999px;
  background: rgba(124, 44, 0, 0.14);
  color: rgba(89, 36, 0, 0.8);
  font-size: 10px;
  font-weight: 800;
  display: inline-flex;
  align-items: center;
}

.send-entry-btn__main {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}

.send-entry-btn__icon {
  width: 40px;
  height: 40px;
  border-radius: 14px;
  background: rgba(124, 44, 0, 0.12);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
}

.send-entry-btn__copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.send-entry-btn__copy strong {
  font-size: 17px;
  line-height: 1.1;
  font-weight: 900;
}

.send-entry-btn__copy small {
  margin-top: 4px;
  color: rgba(89, 36, 0, 0.72);
  font-size: 11px;
  line-height: 1.2;
}

.send-entry-btn__arrow {
  flex: 0 0 auto;
  font-size: 18px;
  color: rgba(89, 36, 0, 0.72);
}

.packet-section {
  margin-top: 14px;
}

.packet-skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
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

.packet-card.parity {
  background: linear-gradient(
    160deg,
    rgba(19, 74, 114, 0.98) 0%,
    rgba(12, 52, 89, 0.97) 55%,
    rgba(9, 35, 68, 0.97) 100%
  );
}

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
}

.packet-card::before {
  content: '';
  position: absolute;
  inset: 3px 0 0;
  background-image: radial-gradient(rgba(212, 175, 55, 1) 1px, transparent 1px);
  background-size: 18px 18px;
  opacity: 0.05;
}

.packet-main {
  position: relative;
  z-index: 1;
  padding: 8px 9px 5px;
}

.packet-top,
.user-wrap,
.amount-wrap,
.tags-row,
.meta-row {
  display: flex;
  align-items: center;
}

.packet-top {
  justify-content: space-between;
}

.user-wrap {
  gap: 6px;
  min-width: 0;
}

.user-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(212, 175, 55, 0.65);
}

.user-name {
  font-size: 12px;
  line-height: 1;
  color: #fff0c9;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 160px;
}

.amount-wrap {
  gap: 3px;
  padding: 0;
  border: 0;
  background: transparent;
  color: rgba(255, 232, 186, 0.7);
}

.packet-amount {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  font-weight: 700;
  background: linear-gradient(to bottom, #cfb53b, #ffd700, #d4af37);
  -webkit-background-clip: text;
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
  margin: 0 auto 2px;
  font-weight: 700;
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
  border: 1px solid transparent;
  white-space: nowrap;
  flex-shrink: 0;
}

.tag.game,
.tag.play-type-tag.thunder {
  color: #ffe7bf;
  background: rgba(212, 175, 55, 0.14);
  border-color: rgba(212, 175, 55, 0.4);
}

.tag.play-type-tag {
  color: #fff3de;
  background: rgba(255, 248, 214, 0.1);
  border-color: rgba(255, 248, 214, 0.2);
}

.tag.play-type-tag.parity {
  background: rgba(74, 163, 226, 0.16);
  border-color: rgba(74, 163, 226, 0.32);
}

.tag.progress {
  color: #ffeecf;
  background: rgba(255, 248, 214, 0.1);
  border-color: rgba(255, 248, 214, 0.28);
}

.meta-row {
  margin-top: 1px;
  gap: 4px;
  flex-wrap: wrap;
}

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
}

.meta-chip--accent {
  background: rgba(212, 175, 55, 0.1) !important;
  border-color: rgba(212, 175, 55, 0.28) !important;
  color: rgba(255, 232, 160, 0.85) !important;
}

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
  word-break: break-word;
  min-height: 28px;
}

.action-pill.grabbed,
.action-pill.locked {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 248, 214, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.14);
}

.action-pill.done {
  width: 100%;
  background: rgba(212, 175, 55, 0.08);
  color: rgba(255, 248, 214, 0.6);
  border: 1px solid rgba(212, 175, 55, 0.28);
  font-size: 9px;
  min-height: 24px;
}

.mine-text {
  margin-right: 3px;
  color: #ffd45d;
}

.choice-mark {
  color: #8fd5ff;
}

.time-text {
  color: #ffd87f;
  font-size: 10px;
  font-weight: 700;
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
  width: 60px;
  height: 60px;
  border-radius: 6px;
  overflow: hidden;
}

:deep(.send-packet-popup.van-popup) {
  max-height: calc(100vh - 56px);
  background:
    radial-gradient(circle at 12% 10%, rgba(212, 175, 55, 0.18), transparent 22%),
    linear-gradient(180deg, #540000 0%, #280000 100%);
  border-radius: 24px 24px 0 0;
  border: 1px solid rgba(212, 175, 55, 0.34);
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
}

.send-packet-modal__sub {
  margin: 6px 0 0;
  color: rgba(255, 229, 186, 0.7);
  font-size: 12px;
  line-height: 1.45;
}
</style>

<route lang="json5">
{
  name: 'PacketList',
}
</route>
