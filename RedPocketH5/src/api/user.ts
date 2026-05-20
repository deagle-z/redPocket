import request from '@/utils/request'
import { i18n } from '@/utils/i18n'
import { getSourceChannelCode, normalizeSourceChannelCode } from '@/utils/source-channel'

export interface ApiResult<T> {
  code: number
  msg: string
  data: T
}

export interface LoginData {
  phone: string
  password: string
  country?: string
}

export interface LoginRes {
  token?: string
  accessToken?: string
}

export interface TgAuthLoginData {
  id: number
  first_name?: string
  last_name?: string
  username?: string
  photo_url?: string
  auth_date: number
  hash: string
  sourceChannelCode?: string
  channelCode?: string
}

export interface TgAuthLoginRes {
  accessToken: string
  userType: number
  expiresIn: number
  tgUser: {
    id: number
    tgId: number
    username?: string
    firstName?: string
    avatar?: string
  }
}

export interface TgCurrentUserInfo {
  avatar?: string
  tenantId?: number
  balance?: number
  country?: string
  uid?: string
  username?: string
  tgName?: string
  firstName?: string
  email?: string
  phone?: string
  tg_id?: number
  gift_amount?: number
  rebate_amount?: number
  freeLotteryCount?: number
  flowLotteryTotalCount?: number
  flowLotteryAvailableCount?: number
  flowLotteryBaseFlow?: number
  flowLotteryBaseRecordId?: number
  vip_level?: number | null
  vip_level_name?: string | null
  audioOpen?: 0 | 1
}

export interface TgWithdrawSummary {
  balance: number
  nonWithdrawableAmount: number
  todayWithdrawCount?: number
}

export interface VipLevelSimple {
  level: number
  levelName: string
  upgradeBonusAmount: number
}

export interface VipProgressInfo {
  currentLevel: VipLevelSimple | null
  prevLevel: VipLevelSimple | null
  nextLevel: VipLevelSimple | null
  levels?: VipLevelSimple[]
  progress: number
  currentValue: number
  targetValue: number
  nextBonusAmount: number
}

export interface VipRewardLog {
  id: number
  vipLevel: number
  levelName: string
  rewardType: number
  bonusAmount: number
  status: number
  createdAt: string
}

export interface TgInviteStats {
  inviteCode?: string
  inviteCount?: number
  todayInviteCount?: number
  rechargeUsers?: number
  todayRechargeUsers?: number
  totalCommission?: number
  availableCommission?: number
  todayCommission?: number
}

export interface TgInviteRuleConfig {
  luckySendCommission?: number
  luckyGrabbingCommission?: number
  inviteFirstRechargeReward?: number
  inviteLuckyRebateRate?: number
  inviteThunderRebateRate?: number
  sendMinAmount?: number
  sendMaxAmount?: number
}

export interface UserState {
  uid?: number
  name?: string
  avatar?: string
  country?: string
  tenantId?: number
}

export interface RechargeOrderAppReq {
  amount: number
  channel: string
  payMethod?: string
  currency?: string
  countryCode?: string
  merchantOrderNo?: string
  extraFields?: Record<string, string>
  activityType?: 0 | 1 | 2
  activityCode?: '' | 'first_recharge_3day' | 'today_first_recharge'
  confirmUnfinishedActivityCycle?: boolean
}

export function getRechargeIsFirst() {
  return request.get<ApiResult<{ hasFirst: boolean, hasTodayFirst: boolean }>>('/api/v1/app/recharge/isFirst')
}

export interface RechargePromotionDayRate {
  day: number
  rate: number
  status: 'available' | 'pending' | 'done' | 'expired' | string
}

export interface RechargeFirstRecharge3DayPromotion {
  visible: boolean
  selectable: boolean
  activityCode: 'first_recharge_3day'
  currentDay: number
  validFrom: string
  validTo: string
  rates: RechargePromotionDayRate[]
  todayRate: number
  title: string
}

export interface RechargeTodayFirstPromotion {
  visible: boolean
  selectable: boolean
  activityCode: 'today_first_recharge'
  rate: number
}

export interface RechargePromotionsResp {
  firstRecharge3Day: RechargeFirstRecharge3DayPromotion
  todayFirstRecharge: RechargeTodayFirstPromotion
}

export function getRechargePromotions() {
  return request.get<ApiResult<RechargePromotionsResp>>('/api/v1/app/recharge/promotions')
}

export function getAppConfig(key: string) {
  return request.get<ApiResult<{ configKey: string, configValue: string, configDesc: string }>>(`/api/v1/app/config/${key}`)
}

export interface RechargeOrderAppBack {
  orderNo: string
  merchantOrderNo?: string
  channel: string
  payMethod?: string
  currency: string
  amount: number
  status: number
  creditAmount?: number
  payUrl?: string
  devCallback?: boolean
  needConfirmUnfinishedActivityCycle?: boolean
  activeActivityMultiplier?: number
}

export interface RechargeSuccessNotification {
  orderNo: string
  channel: string
  currency: string
  amount: number
  creditAmount?: number | null
  bonusAmount: number
  status: number
  isFirstRecharge?: boolean
  payTime?: string | null
  frontendNotifyStatus: number
  frontendNotifyCount: number
  frontendNotifyAt?: string | null
  frontendNotifyAckAt?: string | null
}

export interface AppOrderHistoryReq {
  currentPage: number
  pageSize: number
}

export interface AppOrderHistoryItem {
  orderNo: string
  amount: number
  netAmount: number
  bonusAmount?: number
  fee?: number
  rejectReason?: string
  currency: string
  currencySymbol: string
  time: string
  status: number
}

export interface AppOrderHistoryResp {
  list: AppOrderHistoryItem[]
  total: number
  pageSize: number
  currentPage: number
}

export interface LuckyMoneySendReq {
  amount: number
  gameMode: 0 | 1
  thunder?: number
}

export interface LuckyMoneyBack {
  id: number
  amount: number
  thunder: number
  number: number
  status: number
}

export interface LuckyPacketListReq {
  currentPage: number
  pageSize: number
  gameMode?: 0 | 1
  chatId?: number
  status?: number
  luckyId?: number
}

export interface LuckyPacketListItem {
  id: number
  game_mode?: 0 | 1
  playType?: string | number
  gameType?: string | number
  ruleType?: string | number
  mode?: string | number
  senderId: number
  senderName: string
  senderAvatar?: string
  amount: number
  received: number
  number: number
  grabbedCount: number
  thunder: number
  hitCount: number
  loseRate: number
  status: number
  expireTime?: string
  remainingSeconds: number
  remainingText: string
  items: Array<{
    seqNo: number
    amount: number
    thunderAmount?: number
    isGrabbed: number
    isGrabMine: number
  }>
  createdAt: string
}

export interface LuckyPacketListResp {
  list: LuckyPacketListItem[]
  total: number
  pageSize: number
  currentPage: number
}

export interface LuckyGrabReq {
  luckyId: number
  grabIndex?: number
  oddEvenGuess?: 0 | 1
}

export interface LuckyGrabResp {
  amount: number
  grabAmount?: number
  actualAmount: number
  isThunder: number | string
  loseMoney: number
  openNum: number
  grabIndex: number
  isAmountHidden: number | string
  gameMode: 0 | 1
  guess: number
  luckyNumsHit: boolean
  luckyNums: string
  luckyNumsAmount: number
  message: string
  luckyInfo?: LuckyMoneyBack
}

export interface LuckyDetailReq {
  luckyId: number
}

export interface LuckyDetailSummary {
  id: number
  playType?: string | number
  gameType?: string | number
  ruleType?: string | number
  mode?: string | number
  status: number
  statusText: string
  amount: number
  thunder: number
  loseRate: number
  expireTime: string
  grabbedCount: number
  number: number
  gameText: string
  roomText: string
  unitAmount: string
}

export interface LuckyDetailSender {
  senderId: number
  senderName: string
  senderAvatar?: string
  sendTime: string
}

export interface LuckyDetailFinance {
  sendAmount: number
  receivedAmount: number
  refundAmount: number
  thunderIncome: number
  hitCount: number
  finalProfit: number
}

export interface LuckyDetailParticipant {
  seqNo: number
  isGrabbed: number
  userId: number
  firstName: string
  avatar?: string
  amount: number
  thunderAmount?: number
  isThunder: number
  createdAt: string
}

export interface LuckyDetailResp {
  summary: LuckyDetailSummary
  sender: LuckyDetailSender
  finance: LuckyDetailFinance
  participantCount: number
  participants: LuckyDetailParticipant[]
}

export interface LuckyRecentWinnerReq {
  limit?: number
}

export interface LuckyRecentWinnerItem {
  id: number
  userId: number
  firstName: string
  avatar?: string
  amount: number
  luckyId: number
  createdAt: string
  timeText: string
}

export interface LuckyHistoryUserFlowReq {
  currentPage: number
  pageSize: number
  luckyId?: number
}

export interface LuckyHistoryUserFlowItem {
  userId: number
  avatar?: string | null
  firstName: string
  flowAmount: number
}

export interface LuckyHistoryUserFlowResp {
  list: LuckyHistoryUserFlowItem[]
  total?: number
  pageSize?: number
  currentPage?: number
}

export interface LuckyAppHistoryReq {
  currentPage: number
  pageSize: number
  actionType?: 0 | 1 | 2
  resultType?: 0 | 1 | 2
  startTime?: number
  endTime?: number
}

export interface LuckyAppHistoryItem {
  recordType: 'send' | 'grab'
  actionType: number
  recordId: number
  luckyId: number
  luckyAmount: number
  grabAmount: number
  actualAmount: number
  loseMoney: number
  isThunder: number
  thunder: number
  gameMode: 0 | 1
  senderId: number
  senderName: string
  avatar?: string
  income: number
  expense: number
  netProfit: number
  createdAt: string
}

export interface LuckyAppHistoryResp {
  list: LuckyAppHistoryItem[]
  total: number
  pageSize: number
  currentPage: number
  totalIncome: number
  totalExpense: number
  netProfitLoss: number
}

export interface AppCashHistoryReq {
  currentPage: number
  pageSize: number
  cashMark?: string
}

export interface AppCashHistoryItem {
  createdAt: string
  userId: number
  amount: number
  startAmount: number
  endAmount: number
  cashMark: string
  cashDesc: string
  type: number
  isGift: number
  fromUserId: number
}

export interface AppCashHistoryResp {
  list: AppCashHistoryItem[]
  total: number
  pageSize: number
  currentPage: number
}

export function login(data: LoginData) {
  return request.post<ApiResult<LoginRes>>('/auth/login', data)
}

export function tgLogin(data: TgAuthLoginData) {
  const sourceChannelCode = getRequestSourceChannelCode(data.sourceChannelCode, data.channelCode)
  return request.post<ApiResult<TgAuthLoginRes>>('/api/v1/app/tg/login', {
    ...data,
    sourceChannelCode,
    channelCode: sourceChannelCode,
  })
}

export function loginByPhone(data: LoginData) {
  return request.post<ApiResult<LoginRes>>('/api/v1/app/tg/phoneLogin', data)
}

export function logout() {
  return request.post('/user/logout')
}

export function getUserInfo() {
  return request<ApiResult<UserState>>('/user/me')
}

export interface RegisterData {
  phone: string
  country: string
  firstName?: string
  code?: string
  inviteCode?: string
  password: string
  sourceChannelCode?: string
  channelCode?: string
}

export interface EmailRegisterData {
  email: string
  firstName?: string
  code: string
  inviteCode?: string
  password: string
  sourceChannelCode?: string
  channelCode?: string
}

export interface ForgotPasswordData {
  phone: string
  country: string
  code: string
  newPassword: string
}

export function getEmailCode(): Promise<any> {
  return request.get('/user/email-code')
}

export function sendRegisterEmailCode(email: string): Promise<any> {
  return request.post('/api/v1/app/tg/sendEmailCode', { email })
}

export function sendRegisterSMSCode(phone: string, country: string): Promise<any> {
  return request.post('/api/v1/app/tg/sendSMSCode', { phone, country })
}

export function resetPassword(): Promise<any> {
  return request.post('/user/reset-password')
}

export function forgotPasswordByPhone(data: ForgotPasswordData): Promise<any> {
  return request.post('/api/v1/app/tg/forgotPasswordByPhone', data)
}

export function getCurrentTgUserInfo() {
  return request.get<ApiResult<TgCurrentUserInfo>>('/api/v1/app/tg/currentUserInfo')
}

export function getCurrentTgWithdrawSummary() {
  return request.get<ApiResult<TgWithdrawSummary>>('/api/v1/app/tg/withdrawSummary')
}

export function getVipProgress() {
  return request.get<ApiResult<VipProgressInfo>>('/api/v1/app/vip/progress')
}

export function getClaimableVipRewards() {
  return request.get<ApiResult<VipRewardLog[]>>('/api/v1/app/vip/rewards')
}

export function claimVipReward(id: number) {
  return request.post<ApiResult<string>>(`/api/v1/app/vip/rewards/${id}/claim`)
}

export function getCurrentTgInviteStats() {
  return request.get<ApiResult<TgInviteStats>>('/api/v1/app/tg/inviteStats')
}

export function getCurrentTgInviteRuleConfig() {
  return request.get<ApiResult<TgInviteRuleConfig>>('/api/v1/app/tg/inviteRuleConfig')
}

export function appUpload(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post<ApiResult<{ url: string }>>('/api/v1/app/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
}

export function updateCurrentTgAvatar(avatar: string) {
  return request.post<ApiResult<{ avatar: string }>>('/api/v1/app/tg/avatar', { avatar })
}

export function updateCurrentTgName(username: string) {
  return request.post<ApiResult<{ username: string }>>('/api/v1/app/tg/name', { username })
}

export interface TgBindChannelNameResp {
  tgName: string
  freeLotteryCount: number
  awardedCount: number
}

export function bindCurrentTgChannelName(tgName: string) {
  return request.post<ApiResult<TgBindChannelNameResp>>('/api/v1/app/tg/channelName', { tgName })
}

export function tgLogout() {
  return request.post('/api/v1/app/tg/logout')
}

export interface TgRebateTransferResp {
  transferAmount: number
  balance: number
  rebateAmount: number
}

export interface TgBindEmailReq {
  email: string
}

export interface TgBindPhoneReq {
  phone: string
  country: string
  code: string
}

export interface TgRebateTransferReq {
  amount: number
}

export function transferRebateToBalance(data: TgRebateTransferReq) {
  return request.post<ApiResult<TgRebateTransferResp>>('/api/v1/app/tg/rebate/transfer', data)
}

export function bindCurrentTgEmail(data: TgBindEmailReq) {
  return request.post<ApiResult<{ email: string }>>('/api/v1/app/tg/bindEmail', data)
}

export function bindCurrentTgPhone(data: TgBindPhoneReq) {
  return request.post<ApiResult<{ phone: string, country: string }>>('/api/v1/app/tg/bindPhone', data)
}

export function createRechargeOrder(data: RechargeOrderAppReq) {
  return request.post<ApiResult<RechargeOrderAppBack>>('/api/v1/app/rechargeOrder', data)
}

export function getRechargeOrderHistory(data: AppOrderHistoryReq) {
  return request.post<ApiResult<AppOrderHistoryResp>>('/api/v1/app/rechargeOrder/list', data)
}

export function getPendingRechargeNotifications() {
  return request.get<ApiResult<RechargeSuccessNotification[]>>('/api/v1/app/rechargeOrder/pendingNotifications')
}

export function ackRechargeNotification(orderNo: string) {
  return request.post<ApiResult<string>>('/api/v1/app/rechargeOrder/notifyAck', { orderNo })
}

export function sendLuckyPacket(data: LuckyMoneySendReq) {
  return request.post<ApiResult<LuckyMoneyBack>>('/api/v1/app/lucky/send', data)
}

export function getLuckyPacketList(data: LuckyPacketListReq) {
  return request.post<ApiResult<LuckyPacketListResp>>('/api/v1/app/lucky/list', data)
}

export function grabLuckyPacket(data: LuckyGrabReq) {
  return request.post<ApiResult<LuckyGrabResp>>('/api/v1/app/lucky/grab', data)
}

export function getLuckyDetail(data: LuckyDetailReq) {
  return request.post<ApiResult<LuckyDetailResp>>('/api/v1/app/lucky/detail', data)
}

export function getLuckyRecentWinners(data: LuckyRecentWinnerReq = {}) {
  return request.post<ApiResult<LuckyRecentWinnerItem[]>>('/api/v1/app/lucky/recentWinners', data)
}

export function getLuckyHistoryUserFlow(data: LuckyHistoryUserFlowReq) {
  return request.post<ApiResult<LuckyHistoryUserFlowResp>>('/api/v1/admin/lucky/historyUserFlow', data)
}

export function getLuckyAppHistory(data: LuckyAppHistoryReq) {
  return request.post<ApiResult<LuckyAppHistoryResp>>('/api/v1/app/lucky/history', data)
}

export function getPrizePoolBalance(poolCode = 'lucky') {
  return request.get<ApiResult<{ poolCode: string, balance: number }>>(`/api/v1/app/prizePool/balance?poolCode=${poolCode}`)
}

export interface PrizePoolOutRecordItem {
  id: number
  tenantId: number
  poolId: number
  userId?: number
  userName?: string
  user_name?: string
  firstName?: string
  username?: string
  changeType: 'out'
  amount: number
  beforeBalance: number
  afterBalance: number
  consumedAmount?: number
  createdAt: string
}

export interface PrizePoolOutRecordResp {
  list: PrizePoolOutRecordItem[]
  total: number
  pageSize: number
  currentPage: number
}

export function getPrizePoolOutRecords(currentPage = 0, pageSize = 10) {
  return request.get<ApiResult<PrizePoolOutRecordResp>>(`/api/v1/app/prizePool/outRecords?currentPage=${currentPage}&pageSize=${pageSize}`)
}

export interface LotteryChancesResp {
  totalFlow: number
  currentFlow: number
  peerAmount: number
  remainingFlow: number
  earnedCount: number
  usedCount: number
  freeCount: number
  flowLotteryTotalCount: number
  flowLotteryAvailableCount: number
  availableCount: number
  amounts: number[]
}

export interface LotteryDrawResp {
  recordId: number
  awardAmount: number
}

export interface LotteryHistoryItem {
  name: string
  awardAmount: number
}

export function getLotteryChances() {
  return request.get<ApiResult<LotteryChancesResp>>('/api/v1/app/lottery/chances')
}

export function drawLottery() {
  return request.post<ApiResult<LotteryDrawResp>>('/api/v1/app/lottery/draw', {})
}

export function getLotteryHistory(limit = 10) {
  return request.get<ApiResult<LotteryHistoryItem[]>>(`/api/v1/app/lottery/history?limit=${limit}`)
}

export function getAppCashHistoryList(data: AppCashHistoryReq) {
  return request.post<ApiResult<AppCashHistoryResp>>('/api/v1/app/cashHistory/list', data)
}

export interface CheckInStatusResp {
  todayChecked: boolean
  totalCheckInDays: number
  nextSeq: number
  nextRewardAmount: number
  rewards: number[]
  completed: boolean
  timezone: string
}

export interface CheckInResp {
  recordId: number
  checkInSeq: number
  rewardAmount: number
  balance: number
  todayChecked: boolean
  checkInDate: string
}

export interface CheckInRecordItem {
  id: number
  checkInDate: string
  checkInSeq: number
  rewardAmount: number
  beforeBalance: number
  afterBalance: number
  createdAt: string
}

export function getCheckInStatus() {
  return request.get<ApiResult<CheckInStatusResp>>('/api/v1/app/checkin/status')
}

export function doCheckIn() {
  return request.post<ApiResult<CheckInResp>>('/api/v1/app/checkin', {})
}

export function getCheckInRecords(limit = 30) {
  return request.get<ApiResult<CheckInRecordItem[]>>(`/api/v1/app/checkin/records?limit=${limit}`)
}

export function register(data: RegisterData): Promise<any> {
  const sourceChannelCode = getRequestSourceChannelCode(data.sourceChannelCode, data.channelCode)
  return request.post('/api/v1/app/tg/registerByPhone', {
    phone: data.phone,
    country: data.country,
    firstName: data.firstName,
    ...(data.inviteCode ? { inviteCode: data.inviteCode } : {}),
    ...(data.code ? { code: data.code } : {}),
    password: data.password,
    sourceChannelCode,
    channelCode: sourceChannelCode,
  })
}

export function registerByEmail(data: EmailRegisterData): Promise<any> {
  const sourceChannelCode = getRequestSourceChannelCode(data.sourceChannelCode, data.channelCode)
  return request.post('/api/v1/app/tg/registerByEmail', {
    email: data.email,
    firstName: data.firstName,
    code: data.code,
    ...(data.inviteCode ? { inviteCode: data.inviteCode } : {}),
    password: data.password,
    sourceChannelCode,
    channelCode: sourceChannelCode,
  })
}

function getRequestSourceChannelCode(...values: Array<string | undefined>) {
  for (const value of values) {
    const normalizedCode = normalizeSourceChannelCode(value)
    if (normalizedCode)
      return normalizedCode
  }
  return getSourceChannelCode()
}

export interface AppCountryItem {
  id: number
  countryCode: string
  countryNameEn: string
  countryNameCn: string
  currencyCode: string
  currencySymbol?: string | null
  sort: number
  rate: number
}

export interface AppPayMethodItem {
  id: number
  methodCode: string
  methodName: string
  icon?: string | null
  sort: number
}

export interface AppRechargeChannelItem {
  id: number
  channelCode: string
  channelName: string
  providerType: string
  icon?: string | null
  sort: number
  methods: AppPayMethodItem[]
}

export interface AppCountryRechargeInfo {
  rechargeFields: any[]
  channels: AppRechargeChannelItem[]
}

export interface RechargeFieldOption {
  label: string
  value: string
}

export interface RechargeField {
  fieldKey: string
  fieldLabel: string
  fieldPlaceholder?: string | null
  fieldType: 'input' | 'textarea' | 'number' | 'select'
  dataType: string
  isRequired: number
  defaultValue?: string | null
  maxLength?: number | null
  minLength?: number | null
  regexRule?: string | null
  errorTips?: string | null
  optionsJson?: string | null
}

export function getAppCountries() {
  return request.get<ApiResult<AppCountryItem[]>>('/api/v1/app/countries')
}

export function getCountryRechargeInfo(code: string) {
  return request.get<ApiResult<AppCountryRechargeInfo>>(`/api/v1/app/country/${code}/recharge`)
}

export function getCountryRechargeFields(code: string) {
  return request.get<ApiResult<RechargeField[]>>(`/api/v1/app/country/${code}/rechargeFields`)
}

export function getCountryWithdrawFields(code: string) {
  return request.get<ApiResult<RechargeField[]>>(`/api/v1/app/country/${code}/withdrawFields`)
}

export interface WithdrawAccountItem {
  id: number
  tenantId: number
  userId: number
  countryCode: string
  accountData: string
  isDefault: number
  status: number
  remark?: string | null
  createdAt: string
  updatedAt: string
}

export interface AddWithdrawAccountReq {
  countryCode: string
  accountData: string
  isDefault?: number
  remark?: string
}

export function getWithdrawAccounts() {
  return request.get<ApiResult<WithdrawAccountItem[]>>('/api/v1/app/withdrawAccount/list')
}

export function addWithdrawAccount(data: AddWithdrawAccountReq) {
  return request.post<ApiResult<WithdrawAccountItem>>('/api/v1/app/withdrawAccount', data)
}

export function updateWithdrawAccount(id: number, data: AddWithdrawAccountReq) {
  return request.post<ApiResult<WithdrawAccountItem>>(`/api/v1/app/withdrawAccount/${id}/update`, data)
}

export function deleteWithdrawAccount(id: number) {
  return request.delete<ApiResult<string>>(`/api/v1/app/withdrawAccount/${id}`)
}

export function setDefaultWithdrawAccount(id: number) {
  return request.post<ApiResult<string>>(`/api/v1/app/withdrawAccount/${id}/setDefault`, {})
}

export interface CreateWithdrawOrderReq {
  amount: number
  countryCode: string
  accountId?: number
  fieldValues?: Record<string, string>
}

export function createWithdrawOrder(data: CreateWithdrawOrderReq) {
  return request.post<ApiResult<{ orderNo: string, fee?: number }>>('/api/v1/app/withdraw', data)
}

export function getWithdrawOrderHistory(data: AppOrderHistoryReq) {
  return request.post<ApiResult<AppOrderHistoryResp>>('/api/v1/app/withdraw/list', data)
}

export function createRebateWithdrawOrder(data: CreateWithdrawOrderReq) {
  return request.post<ApiResult<{ orderNo: string, rebateAmount?: number }>>('/api/v1/app/tg/rebate/withdraw', data)
}

export interface BannerItem {
  id: number
  bannerName: string
  bannerCode?: string
  position: string
  platform: string
  bannerType?: string
  displayType?: string
  openMode?: string
  languageCode?: string
  title?: string
  subTitle?: string
  description?: string
  buttonText?: string
  imageUrl?: string
  thumbUrl?: string
  bgImageUrl?: string
  iconUrl?: string
  videoUrl?: string
  jumpType: string
  jumpValue?: string
  textColor?: string
  buttonColor?: string
  bgColor?: string
  sort: number
  status: number
}

export interface BannersData {
  home: BannerItem[]
  popup: BannerItem[]
  activity: BannerItem[]
}

export interface TenantServiceLinks {
  tgServiceUrl?: string | null
  wsServiceUrl?: string | null
}

export function getBanners(data: { position?: string, countryCode?: string } = {}) {
  const lang = i18n.global.locale.value || 'en-US'
  return request.post<ApiResult<BannersData>>('/api/v1/app/banners', {
    platform: 'h5',
    lang,
    ...data,
  })
}

export function getTenantServiceLinks() {
  return request.get<ApiResult<TenantServiceLinks>>('/api/v1/app/domain/serviceLinks')
}

export function setAudioOpen(audioOpen: 0 | 1) {
  return request.post<ApiResult<null>>('/api/v1/app/tg/audioOpen', { audioOpen })
}
