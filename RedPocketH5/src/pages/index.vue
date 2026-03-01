<script setup lang="ts">
import { showToast } from 'vant'
import { getLuckyPacketList, getLuckyRecentWinners, grabLuckyPacket } from '@/api/user'
import { isLogin } from '@/utils/auth'

const bannerList = [
  {
    img: 'https://bbgimage.s3.ap-south-1.amazonaws.com/tutorial.png',
    text: '新手必看：快速了解玩法，轻松开启红包收益。',
  },
  {
    img: 'https://bbgimage.s3.ap-south-1.amazonaws.com/activity-channel300.jpg',
    text: '活动频道上线：参与每日任务，领取专属奖励。',
  },
  {
    img: 'https://bbgimage.s3.ap-south-1.amazonaws.com/luckykita2.jpg',
    text: '幸运对战进行中：邀请好友一起冲榜赢金币。',
  },
]

const DEFAULT_AVATAR = 'https://game.luckypacket.me/images/avatar-placeholder.png'

const activeIndex = ref(0)

const packetList = ref<any[]>([])
const packetLoading = ref(false)
const grabbingKey = ref('')
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

const showWinnerEmpty = computed(() => {
  return !loggedIn.value || visibleWinners.value.length === 0
})

const marqueeText = computed(() => {
  const current = bannerList[activeIndex.value]
  return current?.text || ''
})

function onSwipeChange(index: number) {
  activeIndex.value = index
}

function formatAmount(value: number) {
  const num = Number(value || 0)
  if (Number.isNaN(num))
    return '₱0.00'
  return `₱${num.toFixed(2)}`
}

function mapPacket(item: any) {
  const isOngoing = Number(item?.status) === 1
  const actions = (item?.items || [])
    .map((it: any) => ({
      seqNo: Number(it?.seqNo || 0),
      isGrabbed: Number(it?.isGrabbed) === 1,
      isGrabMine: Number(it?.isGrabMine) === 1,
      amount: Number(it?.amount || 0),
      thunder: Number(it?.thunder || 0),
      label: Number(it?.isGrabbed) === 1 ? formatAmount(it?.amount) : `抢红包 #${it?.seqNo}`,
    }))

  return {
    id: Number(item?.id || 0),
    username: item?.senderName || 'User',
    avatar: item?.senderAvatar || DEFAULT_AVATAR,
    amount: formatAmount(item?.amount),
    status: isOngoing ? 'ongoing' : 'done',
    statusText: isOngoing ? '进行中' : '已结束',
    gameText: 'Game',
    progressText: `已抢: ${Number(item?.grabbedCount || 0)} / ${Number(item?.number || 0)} 个`,
    thunderText: `雷号: ${Number(item?.thunder || 0)}`,
    hitsText: `撞雷次数: ${Number(item?.hitCount || 0)}`,
    rebateText: '发包者获赢: ₱0.00',
    timeText: isOngoing ? `剩余时间: ${item?.remainingText || '00:00'}` : '已结束',
    packetImage: isOngoing ? 'https://game.luckypacket.me/images/redpacket.gif' : 'https://game.luckypacket.me/images/redpacket.jpg',
    actions,
  }
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
  }
  catch {
    showToast('红包列表加载失败')
  }
  finally {
    packetLoading.value = false
  }
}

async function handleGrab(packet: any, action: { seqNo: number, label: string }) {
  if (!loggedIn.value) {
    showToast('请先登录')
    return
  }
  const key = `${packet.id}_${action.seqNo}`
  if (grabbingKey.value === key)
    return
  grabbingKey.value = key
  try {
    const { data } = await grabLuckyPacket({
      luckyId: Number(packet.id),
      grabIndex: Number(action.seqNo),
    })
    showToast(data?.message || '抢红包成功')
    await loadPacketList()
  }
  catch {
    showToast('抢红包失败')
  }
  finally {
    grabbingKey.value = ''
  }
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
      time: item?.timeText || '刚刚',
    }))
  }
  catch {
    showToast('最新中奖加载失败')
  }
  finally {
    recentWinnersLoading.value = false
  }
}

onMounted(() => {
  loadPacketList()
  loadRecentWinners()
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
          <span>热门红包</span>
        </div>
      </header>

      <AppEmpty v-if="showPacketEmpty" text="暂无热门红包" :min-height="120" />

      <template v-else>
        <article v-for="packet in packetList" :key="packet.id" class="packet-card" :class="packet.status">
          <div class="packet-main">
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
                  <span>{{ packet.thunderText }}</span>
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
                :disabled="action.isGrabbed || grabbingKey === `${packet.id}_${action.seqNo}`"
                @click="handleGrab(packet, action)"
              >
                <span v-if="action.thunder" aria-hidden="true">💣</span>
                <span v-else-if="action.isGrabMine" class="mine-text">🎁 </span>
                {{ action.label }}
              </button>
            </template>
            <button v-else type="button" class="action-pill done">
              已结束
            </button>
          </div>
        </article>
      </template>
    </section>

    <section class="winner-section">
      <header class="packet-header">
        <div class="packet-title-wrap">
          <van-icon name="trophy-o" />
          <span>最新中奖</span>
        </div>
      </header>

      <AppEmpty v-if="showWinnerEmpty" text="没有更多了" :min-height="120" />

      <template v-else>
        <div class="winner-card">
          <article v-for="item in visibleWinners" :key="item.id" class="winner-item">
            <img :src="item.avatar" alt="" class="winner-avatar">
            <div class="winner-main">
              <p class="winner-amount">
                获得 <strong>{{ item.amount }}</strong>
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

        <AppEmpty text="没有更多了" :min-height="120" />
      </template>
    </section>
  </div>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
  background: var(--bg-secondary);
  padding: var(--space-md);
  width: 100%;
  max-width: 100%;
  overflow: auto;
}

.home-carousel-card {
  position: relative;
  border-radius: 5px;
  overflow: hidden;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.14);
  background: transparent;
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
  background: linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, 0.56) 100%) !important;
  color: #fff;
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

.packet-header {
  border-top: 1px solid #e5e7eb;
  padding-top: 8px;
  margin-bottom: 8px;
}

.packet-title-wrap {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 14px;
  color: #111827;
  font-weight: 700;
}

.packet-title-wrap :deep(.van-icon) {
  color: #22c55e;
}

.packet-card {
  border-radius: 10px;
  border: 1px solid #e5e7eb;
  background: #fff;
  overflow: hidden;
  margin-bottom: 10px;
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
  width: 42px;
  height: 42px;
  border-radius: 50%;
  object-fit: cover;
}

.user-name {
  font-size: 20px;
  line-height: 1;
  color: #111827;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 160px;
}

.amount-wrap {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: #d1d5db;
}

.packet-amount {
  font-size: 16px;
  line-height: 1;
  font-weight: 700;
  color: #f1b91b;
}

.packet-card.done .packet-amount {
  color: #111827;
}

.packet-body {
  margin-top: 8px;
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.packet-image-wrap {
  width: 108px;
  flex: 0 0 108px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 56px;
  height: 24px;
  border-radius: 4px;
  background: #f4c33e;
  color: #111827;
  font-size: 14px;
  line-height: 1;
  margin-bottom: 6px;
}

.packet-card.done .status-badge {
  background: #9ca3af;
  color: #fff;
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
  height: 24px;
  display: inline-flex;
  align-items: center;
  border-radius: 4px;
  padding: 0 8px;
  font-size: 13px;
  line-height: 1;
}

.tag.game {
  color: #4b5563;
  background: #f7ca4b;
}

.tag.progress {
  color: #fff;
  background: #4cad68;
}

.packet-card.done .tag.game {
  background: #9ca3af;
  color: #fff;
}

.packet-card.done .tag.progress {
  background: #9ca3af;
}

.meta-row {
  margin-top: 7px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #374151;
  font-size: 12px;
}

.rebate-text {
  margin: 8px 0 0;
  color: #374151;
  font-size: 12px;
}

.time-text {
  margin: 8px 0 0;
  color: #2ea34f;
  font-size: 18px;
  line-height: 1;
}

.packet-card.done .time-text {
  color: #6b7280;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 30px;
  border-radius: 4px;
  background: #9ca3af;
  padding: 0 8px;
  font-size: 15px;
}

.packet-actions {
  border-top: 1px solid #edf0f4;
  padding: 8px 10px;
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 6px;
}

.action-pill {
  border: none;
  border-radius: 999px;
  background: #45a75f;
  color: #fff;
  font-size: 8px;
  line-height: 1;
  min-height: 30px;
}

.action-pill.grabbed {
  background: #eef1f4;
  color: #b1b7bf;
  border: 1px solid #d9dee5;
  text-decoration: line-through;
}

.action-pill.mined {
  color: #9b8ac9;
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
  color: #8b5cf6;
  text-decoration: none;
}

.action-pill.done {
  grid-column: 1 / -1;
  background: #f3f4f6;
  color: #9ca3af;
}

.winner-section {
  margin-top: 12px;
}

.winner-card {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #fff;
  overflow: hidden;
}

.winner-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
}

.winner-item + .winner-item {
  border-top: 1px solid #eef1f5;
}

.winner-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
  flex: 0 0 auto;
}

.winner-main {
  min-width: 0;
  flex: 1;
}

.winner-amount {
  margin: 0;
  color: #111827;
  font-size: 18px;
  line-height: 1.2;
}

.winner-amount strong {
  font-weight: 700;
}

.winner-name {
  margin: 8px 0 0;
  color: #9ca3af;
  font-size: 15px;
  line-height: 1;
}

.winner-right {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #b8dcc8;
}

.winner-time {
  color: #6b7280;
  font-size: 16px;
  line-height: 1;
}
</style>

<route lang="json5">
{
  name: 'Home'
}
</route>
