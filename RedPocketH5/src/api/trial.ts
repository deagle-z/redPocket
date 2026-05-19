import request from '@/utils/request'
import type { ApiResult } from './user'

export interface TrialMeResp {
  trialBalance: number
}

export interface TrialBalanceRefreshResp {
  trialBalance: number
  refreshed: boolean
}

export interface TrialLuckySendReq {
  amount: number
  gameMode: 0 | 1
  thunder?: number
  number?: number
  chatId?: number
}

export interface TrialLuckyBack {
  id: number
  amount: number
  thunder: number
  number: number
  status: number
}

export interface TrialLuckyListReq {
  currentPage: number
  pageSize: number
  gameMode?: 0 | 1
  status?: number
  luckyId?: number
}

export interface TrialLuckyItem {
  seqNo: number
  amount: number
  isGrabbed: number
  thunder: number
  isGrabMine: number
}

export interface TrialLuckyListItem {
  id: number
  senderId: number
  senderType: 'user' | 'bot'
  senderName: string
  senderAvatar?: string
  amount: number
  received: number
  number: number
  grabbedCount: number
  thunder: number
  gameMode: 0 | 1
  hitCount: number
  loseRate: number
  status: number
  remainingSeconds: number
  remainingText: string
  items: TrialLuckyItem[]
  createdAt: string
}

export interface TrialLuckyListResp {
  list: TrialLuckyListItem[]
  total: number
  pageSize: number
  currentPage: number
}

export interface TrialLuckyGrabReq {
  luckyId: number
  grabIndex?: number
  oddEvenGuess?: 0 | 1
}

export interface TrialLuckyGrabResp {
  luckyId: number
  amount: number
  actualAmount: number
  loseMoney: number
  isThunder: number
  openNum: number
  balance: number
  message: string
  lotteryRewardCount?: number
  trialFlowLotteryRewarded?: boolean
}

export interface TrialCashHistoryReq {
  currentPage: number
  pageSize: number
  cashMark?: string
}

export interface TrialCashHistoryItem {
  id: number
  createdAt: string
  amount: number
  startAmount: number
  endAmount: number
  cashMark: string
  cashDesc: string
  type: number
  isThunder: number
  luckyId: number
}

export interface TrialCashHistoryResp {
  list: TrialCashHistoryItem[]
  total: number
  pageSize: number
  currentPage: number
}

export interface TrialLuckyFlowLotteryRewardProgress {
  enabled: boolean
  thresholdAmount: number
  rewardCount: number
  totalFlow: number
  remainingFlow: number
  progressPercent: number
  rewarded: boolean
  canReward: boolean
  availableRewardCount: number
  drawn: boolean
  freeLotteryCount: number
}

export function getTrialMe() {
  return request.get<ApiResult<TrialMeResp>>('/api/v1/app/trial/me')
}

export function refreshTrialBalance() {
  return request.post<ApiResult<TrialBalanceRefreshResp>>('/api/v1/app/trial/balance/refresh')
}

export function getTrialLuckyList(data: TrialLuckyListReq) {
  return request.post<ApiResult<TrialLuckyListResp>>('/api/v1/app/trial/lucky/list', data)
}

export function sendTrialLucky(data: TrialLuckySendReq) {
  return request.post<ApiResult<TrialLuckyBack>>('/api/v1/app/trial/lucky/send', data)
}

export function grabTrialLucky(data: TrialLuckyGrabReq) {
  return request.post<ApiResult<TrialLuckyGrabResp>>('/api/v1/app/trial/lucky/grab', data)
}

export function getTrialLuckyHistory(data: TrialCashHistoryReq) {
  return request.post<ApiResult<TrialCashHistoryResp>>('/api/v1/app/trial/lucky/history', data)
}

export function getTrialLuckyFlowLotteryReward() {
  return request.get<ApiResult<TrialLuckyFlowLotteryRewardProgress>>('/api/v1/app/trial/lucky/flowLotteryReward')
}
