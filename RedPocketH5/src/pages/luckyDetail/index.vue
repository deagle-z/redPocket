<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getLuckyDetail, grabLuckyPacket } from '@/api/user'
import AppPageHeader from '@/components/AppPageHeader.vue'
import wsClient from '@/plugins/websocket'

const DEFAULT_AVATAR = 'https://game.luckypacket.me/images/avatar-placeholder.png'
const ONGOING_PACKET_IMAGE = 'https://game.luckypacket.me/images/redpacket.gif'
const DONE_PACKET_IMAGE = 'https://game.luckypacket.me/images/redpacket.jpg'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const grabbingSeq = ref(0)
const GRAB_THROTTLE_MS = 1000
const lastGrabAt = ref(0)
const detail = ref<any | null>(null)

const luckyId = computed(() => Number(route.query.id || 0))

const positionList = computed(() => {
  const summary = detail.value?.summary || {}
  const participants = detail.value?.participants || []
  const participantMap = new Map<number, any>()
  participants.forEach((p: any) => {
    const seq = Number(p?.seqNo || 0)
    if (seq > 0)
      participantMap.set(seq, p)
  })
  const number = Number(summary?.number || 0)
  return Array.from({ length: number }, (_, idx) => {
    const seqNo = idx + 1
    const participant = participantMap.get(seqNo)
    const isGrabbed = !!participant
    const isMine = false
    return {
      seqNo,
      isGrabbed,
      isMine,
      amount: Number(participant?.amount || 0),
      isThunder: Number(participant?.isThunder || 0) === 1,
    }
  })
})

const recordList = computed(() => {
  const participants = detail.value?.participants || []
  return participants.map((it: any) => ({
    id: `${it?.seqNo || 0}_${it?.userId || 0}_${it?.createdAt || ''}`,
    seqNo: Number(it?.seqNo || 0),
    avatar: it?.avatar || DEFAULT_AVATAR,
    name: it?.firstName || `用户 #${Number(it?.seqNo || 0)}`,
    time: it?.createdAt || '',
    amount: Number(it?.amount || 0),
    status: Number(it?.isThunder || 0) === 1 ? '中雷' : '已参与',
  }))
})

const overview = computed(() => {
  const data = detail.value
  const summary = data?.summary
  const sender = data?.sender
  const finance = data?.finance
  if (!summary)
    return null
  const status = Number(summary?.status) === 1 ? 'ongoing' : 'done'
  const grabbedCount = Number(summary?.grabbedCount || 0)
  const number = Number(summary?.number || 0)
  const amount = Number(summary?.amount || 0)
  const received = Number(finance?.receivedAmount || 0)
  const hitCount = Number(finance?.hitCount || 0)
  const loseRate = Number(summary?.loseRate || 0)
  const unclaimed = Math.max(number - grabbedCount, 0)
  const remain = Math.max(amount - received, 0)
  const expireAt = new Date(summary?.expireTime || '')
  const now = Date.now()
  const diff = Math.max(0, Number.isNaN(expireAt.getTime()) ? 0 : Math.floor((expireAt.getTime() - now) / 1000))
  const mm = String(Math.floor(diff / 60)).padStart(2, '0')
  const ss = String(diff % 60).padStart(2, '0')
  const remainingText = `${mm}:${ss}`
  return {
    status,
    statusText: status === 'ongoing' ? '进行中' : '已结束',
    packetImage: status === 'ongoing' ? ONGOING_PACKET_IMAGE : DONE_PACKET_IMAGE,
    senderName: sender?.senderName || 'User',
    senderAvatar: sender?.senderAvatar || DEFAULT_AVATAR,
    amountText: `₱${amount.toFixed(2)}`,
    progressText: `已抢 ${grabbedCount}/${number} 个`,
    thunderText: status === 'ongoing' ? '雷号: ?' : `雷号: ${Number(summary?.thunder || 0)}`,
    oddsText: `赔率: ${loseRate.toFixed(1)} x`,
    timeText: status === 'ongoing' ? `剩余: ${remainingText}` : `过期: ${formatTime(summary?.expireTime || '')}`,
    sentTime: sender?.sendTime || '-',
    summaryRows: [
      { label: '发包金额', value: `₱${amount.toFixed(2)}` },
      { label: '已抢金额', value: `₱${received.toFixed(2)}` },
      { label: '剩余金额', value: `₱${remain.toFixed(2)}` },
      { label: '中雷次数', value: `${hitCount}` },
      { label: '中雷收益', value: `₱${Number(finance?.thunderIncome || 0).toFixed(2)}`, highlight: true },
      { label: '未抢数量', value: `${unclaimed}` },
    ],
  }
})

function formatTime(raw: string) {
  const d = new Date(raw)
  if (Number.isNaN(d.getTime()))
    return raw || '-'
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${y}/${m}/${day} ${hh}:${mm}`
}

async function loadDetail() {
  if (!luckyId.value) {
    showToast('参数错误')
    return
  }
  if (loading.value)
    return
  loading.value = true
  try {
    const { data } = await getLuckyDetail({
      luckyId: luckyId.value,
    })
    detail.value = data || null
  }
  catch {
    showToast('详情加载失败')
  }
  finally {
    loading.value = false
  }
}

function goInvite() {
  router.push('/invite')
}

function goBack() {
  router.back()
}

async function handleGrab(seqNo: number) {
  const current = detail.value
  const summary = current?.summary
  if (!current)
    return
  if (Number(summary?.status) !== 1)
    return
  const target = positionList.value.find(item => item.seqNo === seqNo)
  if (!target || target.isGrabbed)
    return
  if (grabbingSeq.value === seqNo)
    return
  const now = Date.now()
  if (now - lastGrabAt.value < GRAB_THROTTLE_MS)
    return
  lastGrabAt.value = now

  grabbingSeq.value = seqNo
  try {
    const { data } = await grabLuckyPacket({
      luckyId: Number(summary?.id || luckyId.value),
      grabIndex: seqNo,
    })
    showToast(data?.message || '抢红包成功')
    await loadDetail()
  }
  catch {
    showToast('抢红包失败')
  }
  finally {
    grabbingSeq.value = 0
  }
}

function handleLuckyWsMessage(message: any) {
  const payload = message?.data || message
  const changedLuckyId = Number(payload?.id || 0)
  if (!changedLuckyId || changedLuckyId !== luckyId.value)
    return
  void loadDetail()
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
    <AppPageHeader title="红包详情" @back="goBack" @right-click="goInvite">
      <template #right>
        <van-icon name="share-o" />
      </template>
    </AppPageHeader>

    <div class="detail-content">
      <template v-if="loading && !overview">
        <section class="card panel-card skeleton-card">
          <van-skeleton title avatar :row="4" :loading="true" avatar-size="82px" />
        </section>
        <section class="card panel-card skeleton-card">
          <van-skeleton title avatar :row="2" :loading="true" avatar-size="36px" />
        </section>
        <section class="card panel-card skeleton-card">
          <van-skeleton title :row="6" :loading="true" />
        </section>
        <section class="card panel-card skeleton-card">
          <van-skeleton title :row="4" :loading="true" />
        </section>
      </template>

      <template v-else-if="overview">
        <section class="card hero-card">
          <div class="hero-left">
            <span class="status-chip">{{ overview.statusText }}</span>
            <img :src="overview.packetImage" alt="" class="packet-image">
          </div>
          <div class="hero-main">
            <div class="hero-title-row">
              <h2 class="packet-amount">
                红包金额 {{ overview.amountText }}
              </h2>
              <span class="progress-chip">{{ overview.progressText }}</span>
            </div>
            <p class="hero-meta">
              {{ overview.thunderText }} | {{ overview.oddsText }}
            </p>
            <p class="hero-meta">
              {{ overview.timeText }}
            </p>
            <p class="hero-meta">
              单个金额: Random
            </p>
          </div>
        </section>

        <section class="card sender-card">
          <img :src="overview.senderAvatar" alt="" class="sender-avatar">
          <div class="sender-main">
            <p class="sender-name">
              {{ overview.senderName }}
            </p>
            <p class="sender-time">
              发包时间: {{ formatTime(overview.sentTime) }}
            </p>
          </div>
        </section>

        <section class="card panel-card">
          <div class="panel-title-wrap">
            <h3 class="panel-title">
              财务明细
            </h3>
            <p class="panel-subtitle">
              中雷{{ detail?.finance?.hitCount || 0 }}次
            </p>
          </div>
          <div class="summary-list">
            <div v-for="row in overview.summaryRows" :key="row.label" class="summary-row">
              <span>{{ row.label }}</span>
              <span :class="{ highlight: row.highlight }">{{ row.value }}</span>
            </div>
          </div>
        </section>

        <section class="card panel-card">
          <div class="panel-title-row">
            <p class="panel-title with-icon">
              <van-icon name="apps-o" />
              选择位置抢红包
            </p>
            <span class="panel-count">{{ overview.progressText.replace('已抢 ', '') }}</span>
          </div>
          <div class="position-grid">
            <button
              v-for="item in positionList"
              :key="item.seqNo"
              type="button"
              class="position-item"
              :class="{
                grabbed: item.isGrabbed,
                mine: item.isMine,
                loading: grabbingSeq === item.seqNo,
                ended: overview.status !== 'ongoing' && !item.isGrabbed,
              }"
              :disabled="overview.status !== 'ongoing' || item.isGrabbed || grabbingSeq > 0"
              @click="handleGrab(item.seqNo)"
            >
              <p class="position-no">
                #{{ item.seqNo }}
              </p>
              <p class="position-status">
                {{ grabbingSeq === item.seqNo ? '抢包中' : (item.isGrabbed ? (item.isThunder ? '中雷' : '已抢') : (overview.status !== 'ongoing' ? '已结束' : '可抢')) }}
              </p>
            </button>
          </div>
        </section>

        <section class="card panel-card record-card">
          <div class="panel-title-row">
            <p class="panel-title with-icon">
              <van-icon name="friends-o" />
              参与记录
            </p>
            <span class="panel-count">共{{ recordList.length }}人参与</span>
          </div>
          <div v-if="recordList.length > 0">
            <article v-for="item in recordList" :key="item.id" class="record-item">
              <img :src="item.avatar" alt="" class="record-avatar">
              <div class="record-main">
                <p class="record-name">
                  {{ item.name }}（#{{ item.seqNo }}）
                </p>
                <p class="record-time">
                  {{ formatTime(item.time) }}
                </p>
              </div>
              <div class="record-right">
                <p class="record-amount">
                  ₱{{ item.amount.toFixed(2) }}
                </p>
                <p class="record-status">
                  {{ item.status }}
                </p>
              </div>
            </article>
          </div>
          <p class="no-more">
            没有更多了
          </p>
        </section>
      </template>

      <AppEmpty v-else-if="!loading" text="暂无详情数据" :min-height="160" />
    </div>

    <div class="share-bar">
      <button type="button" class="share-btn" @click="goInvite">
        <van-icon name="share-o" />
        <span>转发</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.lucky-detail-page {
  min-height: 100vh;
  background: #f5f6fa;
  padding-bottom: calc(56px + env(safe-area-inset-bottom));
}

.detail-content {
  padding: 10px 10px 0;
}

.card {
  background: #fff;
  border-radius: 10px;
  margin-bottom: 8px;
}

.hero-card {
  display: flex;
  gap: 10px;
  padding: 10px;
}

.hero-left {
  width: 82px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-chip {
  background: var(--color-primary);
  color: #fff;
  border-radius: 4px;
  font-size: 10px;
  line-height: 16px;
  text-align: center;
}

.packet-image {
  width: 82px;
  height: 82px;
  border-radius: 8px;
  object-fit: cover;
}

.hero-main {
  flex: 1;
  min-width: 0;
}

.hero-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.packet-amount {
  margin: 0;
  color: #1a1a2e;
  font-size: 16px;
  font-weight: 700;
}

.progress-chip {
  border: 1px solid var(--color-primary);
  color: var(--color-primary);
  border-radius: 12px;
  padding: 0 8px;
  font-size: 10px;
  line-height: 18px;
  white-space: nowrap;
}

.hero-meta {
  margin: 0;
  color: #6b7280;
  font-size: 11px;
  line-height: 1.6;
}

.sender-card {
  padding: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.sender-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
}

.sender-main {
  min-width: 0;
}

.sender-name {
  margin: 0;
  color: #1a1a2e;
  font-size: 13px;
  font-weight: 600;
}

.sender-time {
  margin: 2px 0 0;
  color: #9ca3af;
  font-size: 10px;
}

.panel-card {
  padding: 12px;
}

.panel-title-wrap {
  margin-bottom: 8px;
}

.panel-title {
  margin: 0;
  color: #1a1a2e;
  font-size: 14px;
  font-weight: 700;
}

.panel-subtitle {
  margin: 4px 0 0;
  color: #ff9500;
  font-size: 11px;
}

.summary-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #6b7280;
  font-size: 12px;
}

.summary-row .highlight {
  color: var(--color-primary);
  font-weight: 600;
}

.panel-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.panel-title.with-icon {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.panel-count {
  color: #9ca3af;
  font-size: 11px;
}

.position-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.position-item {
  border: none;
  background: var(--color-primary);
  border-radius: 8px;
  color: #fff;
  text-align: center;
  min-height: 40px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 2px;
  padding: 0;
  cursor: pointer;
}

.position-item:disabled {
  cursor: not-allowed;
}

.position-item.grabbed {
  background: #d1d5db;
  color: #6b7280;
}

.position-item.ended {
  background: #d1d5db;
  color: #6b7280;
}

.position-item.mine {
  background: #16a34a;
  color: #fff;
}

.position-item.loading {
  opacity: 0.8;
}

.position-no {
  margin: 0;
  font-size: 12px;
  font-weight: 700;
}

.position-status {
  margin: 0;
  font-size: 10px;
}

.record-card {
  margin-bottom: 0;
}

.record-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  border-top: 1px solid #f0f0f5;
}

.record-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
}

.record-main {
  flex: 1;
  min-width: 0;
}

.record-name {
  margin: 0;
  font-size: 12px;
  color: #1a1a2e;
  font-weight: 600;
}

.record-time {
  margin: 2px 0 0;
  font-size: 10px;
  color: #9ca3af;
}

.record-right {
  text-align: right;
}

.record-amount {
  margin: 0;
  font-size: 12px;
  color: #1a1a2e;
  font-weight: 600;
}

.record-status {
  margin: 2px 0 0;
  font-size: 10px;
  color: #9ca3af;
}

.no-more {
  margin: 8px 0 0;
  font-size: 11px;
  color: var(--color-primary);
  text-align: center;
}

.share-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: #fff;
  border-top: 1px solid #f0f0f5;
  padding-bottom: env(safe-area-inset-bottom);
}

.share-btn {
  width: 100%;
  height: 56px;
  border: 0;
  background: #fff;
  color: #1a1a2e;
  font-size: 14px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.skeleton-card {
  padding-top: 10px;
  padding-bottom: 10px;
}

.skeleton-card :deep(.van-skeleton__title),
.skeleton-card :deep(.van-skeleton__row) {
  background: #edf1f5;
}
</style>

<route lang="json5">
{
  name: 'LuckyDetail',
}
</route>

