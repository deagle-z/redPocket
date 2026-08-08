import { del, get, post, upload } from '@/utils/http'
import { APP_COUNTRY_CODE, type AppCountryCode } from '@/config/market'
import { getSourceChannelCode, normalizeSourceChannelCode } from '@/utils/sourceChannel'
import type { DeviceInfoPayload } from '@/utils/deviceInfo'
import { getDeviceInfo } from '@/utils/deviceInfo'

export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  token?: string
  accessToken?: string
}

export interface LoginByPhoneParams {
  phone: string
  password: string
  country?: AppCountryCode
}

export interface RegisterByPhoneParams {
  phone: string
  country?: AppCountryCode
  firstName?: string
  password: string
  inviteCode?: string
  sourceChannelCode?: string
  channelCode?: string
  deviceFingerprint?: string
}

export interface CheckRegisterPhoneParams {
  phone: string
  country?: AppCountryCode
}

export interface CheckRegisterPhoneResp {
  phone: string
  country: AppCountryCode
  available: boolean
}

export interface ForgotPasswordByPhoneParams {
  phone: string
  country: AppCountryCode
  code: string
  newPassword: string
}

export interface TgBindEmailReq {
  email: string
}

export interface TgBindPhoneReq {
  phone: string
  country: AppCountryCode
  code: string
}

export interface CurrentUserInfo {
  id?: number | string
  uid?: number | string
  userId?: number | string
  username?: string
  nickname?: string
  tgName?: string
  firstName?: string
  lastName?: string
  avatar?: string
  country?: string
  balance?: number | string
  sportBalance?: number | string
  hasWithdrawAccount?: boolean
  phone?: string
  email?: string
  rechargeCount?: number
  rebateWithdrawDisabled?: number
}

export interface TgInviteStats {
  inviteCode?: string
  inviteCount?: number
  todayInviteCount?: number
  validUsers?: number
  todayValidUsers?: number
  rechargeUsers?: number
  todayRechargeUsers?: number
  totalCommission?: number
  availableCommission?: number
  todayCommission?: number
  rebateWithdrawDisabled?: number
}

export interface TgInviteRuleConfig {
  luckySendCommission?: number
  luckyGrabbingCommission?: number
  inviteFirstRechargeReward?: number
  inviteLuckyRebateRate?: number
  inviteThunderRebateRate?: number
  inviteBetRebateRate?: number
  inviteValidMinRecharge?: number
  inviteValidMinBet?: number
  sendMinAmount?: number
  sendMaxAmount?: number
  /** 佣金提现固定手续费（如 4.25） */
  rebateWithdrawFeeFixed?: number
  /** 佣金提现百分比手续费(%)（如 1.5）。手续费=固定+金额×百分比 */
  rebateWithdrawFeePercent?: number
  /** 当前登录商户绑定域名（分享链接用，为空则回退当前 location） */
  bindDomain?: string
}

export interface AppCountryItem {
  id: number
  countryCode: string
  countryNameEn: string
  countryNameCn: string
  currencyCode: string
  currencySymbol?: string | null
  sort: number
  rate?: number
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
  rechargeFields: unknown[]
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

export interface RechargeOrderAppReq {
  amount: number
  channel: string
  payMethod?: string
  walletType?: 'balance' | 'sport'
  currency?: string
  countryCode?: string
  merchantOrderNo?: string
  extraFields?: Record<string, string>
  activityType?: 0 | 1 | 2
  activityCode?: '' | 'first_recharge_3day' | 'today_first_recharge'
  confirmUnfinishedActivityCycle?: boolean
}

export interface CryptoPayment {
  network: string
  token: string
  contractAddress: string
  receiveAddress: string
  expectedAmount: string
  expireTime: string
  qrContent: string
  platformRate: string
}

export interface CryptoRechargeOption {
  network: string
  token: string
  contractAddress: string
  receiveAddress: string
  estimatedAmount: string
  platformRate: string
  platformAmount: number
}

export interface CryptoRechargeOptionsResp {
  platformAmount: number
  options: CryptoRechargeOption[]
}

export interface CryptoRechargeOrderReq {
  amount: number
  network?: string
  token?: string
  merchantOrderNo?: string
  activityCode?: string
  confirmUnfinishedActivityCycle?: boolean
}

export interface CryptoRechargeOrderStatusResp {
  orderNo: string
  status: 0 | 1 | 2 | 3 | number
  rechargeStatus: number
  txId?: string
  payTime?: string | null
  expireTime: string
  remainingSeconds: number
  cryptoPayment?: CryptoPayment
}

export interface RechargeFirstStatusResp {
  hasFirst: boolean
  isFirstRecharge: boolean
  /** 下一次充值次序：1至11次分别对应档位，≥12沿用第11档 */
  rechargeNumber?: number
  /** 各档位赠送比例(%)：[第1次至第11次及以后] */
  giftRates?: number[]
  /** 赠送门槛(满此金额即送) */
  giftMinAmount?: number
}

export interface RechargeCountResp {
  rechargeCount: number
  /** 是否曾把佣金转入余额（佣金转账可解锁游戏；赠送余额不算） */
  rebateTransferred?: boolean
  /** 是否启用充值/佣金转入解锁游戏限制。 */
  rechargeGameUnlockEnabled: boolean
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
  firstRecharge3Day?: RechargeFirstRecharge3DayPromotion | null
  todayFirstRecharge?: RechargeTodayFirstPromotion | null
}

export interface RechargeOrderAppBack {
  orderNo: string
  merchantOrderNo?: string
  channel: string
  payMethod?: string
  currency: string
  amount: number
  netAmount?: number
  status: number
  creditAmount?: number
  bonusAmount?: number
  payUrl?: string
  devCallback?: boolean
  needConfirmUnfinishedActivityCycle?: boolean
  activeActivityMultiplier?: number
  cryptoPayment?: CryptoPayment
}

export interface AppOrderHistoryReq {
  currentPage: number
  pageSize: number
}

export interface AppOrderHistoryItem {
  orderNo: string
  amount: number
  netAmount?: number
  bonusAmount?: number
  fee?: number
  rejectReason?: string | null
  currency: string
  currencySymbol?: string
  time?: string
  createdAt?: string
  status: number
}

export interface AppOrderHistoryResp {
  list: AppOrderHistoryItem[]
  total: number
  pageSize: number
  currentPage: number
}

export interface AppUserBetRecordReq {
  categoryCode?: string
  startTime: number
  endTime: number
  currentPage: number
  pageSize: number
}

export interface AppUserBetRecordItem {
  id: number
  uid: number
  userId: number
  gameId: string
  gameName: string
  platformCode: string
  betAmount: number
  winAmount: number
  roundId: string
  traceId: string
  roundEnd: number
  date: string
  remark?: string | null
  updateTime?: string | null
  createTime?: string | null
}

export interface AppUserBetRecordResp {
  list: AppUserBetRecordItem[]
  total: number
  pageSize: number
  currentPage: number
}

export interface WithdrawOrderHistoryReq {
  currentPage: number
  pageSize: number
}

export interface WithdrawOrderHistoryItem {
  orderNo: string
  type: string
  currency: string
  amount: number
  fee: number
  netAmount: number
  status: number
  createdAt: string
  paidAt?: string | null
  rejectReason?: string | null
}

export interface WithdrawOrderHistoryResp {
  list: WithdrawOrderHistoryItem[]
  total: number
  pageSize: number
  currentPage: number
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

export interface ExchangeCodeRedeemResp {
  code: string
  amount: number
  balance: number
  redeemedAt: string
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

export interface PrizePoolBalanceResp {
  poolCode: string
  balance: number
}

export interface PrizePoolOutRecordItem {
  id: number
  tenantId?: number
  poolId?: number
  userId?: number
  userName?: string
  user_name?: string
  firstName?: string
  username?: string
  changeType?: 'out' | string
  amount: number
  beforeBalance?: number
  afterBalance?: number
  consumedAmount?: number
  createdAt: string
}

export interface PrizePoolOutRecordResp {
  list: PrizePoolOutRecordItem[]
  total: number
  pageSize: number
  currentPage: number
}

export interface RechargeSuccessNotification {
  orderNo: string
  channel: string
  currency: string
  amount: number
  creditAmount?: number | null
  bonusAmount?: number | null
  status: number
  isFirstRecharge?: boolean
  payTime?: string | null
  frontendNotifyStatus: number
  frontendNotifyCount: number
  frontendNotifyAt?: string | null
  frontendNotifyAckAt?: string | null
}

export interface WithdrawAccountItem {
  id: number
  tenantId?: number
  userId?: number
  countryCode: string
  accountData: string
  isDefault: number
  status?: number
  remark?: string | null
  createdAt?: string
  updatedAt?: string
}

export interface AddWithdrawAccountReq {
  countryCode: string
  accountData: string
  isDefault?: number
  remark?: string
}

export interface TgWithdrawSummary {
  balance: number | string
  nonWithdrawableAmount: number | string
  todayWithdrawCount?: number
}

export interface TgWithdrawFlowBatchSummary {
  balance: number | string
  hasUnfinishedBatch: boolean
  requiredFlow: number | string
  completedFlow: number | string
  remainingFlow: number | string
  unfinishedBatchCount: number
  /** 充值到账(本金)金额提现所需流水倍数（withdraw_limit） */
  withdrawMultiplier?: number
  /** 赠送金额提现所需流水倍数（withdraw_gift_limit） */
  giftMultiplier?: number
}

export interface TenantServiceLinks {
  tgServiceUrl?: string | null
  wsServiceUrl?: string | null
}

export interface CreateWithdrawOrderReq {
  amount: number
  countryCode: string
  accountId?: number
  fieldValues?: Record<string, string>
}

export interface CreateWithdrawOrderResp {
  orderNo: string
  fee?: number
}

export interface RebateWithdrawOrderResp {
  orderNo: string
  rebateAmount?: number
}

export interface TgRebateTransferReq {
  amount: number
}

export interface TgRebateTransferResp {
  transferAmount: number
  balance: number
  rebateAmount: number
}

export type TaskActivityStatus = 0 | 1 | 2 | 3 | 4 | number

export interface TaskActivityItem {
  configId: number
  recordId?: number
  title: string
  subTitle?: string
  levelCode: 'primary' | 'middle' | 'advanced' | 'custom' | string
  levelName?: string
  requiredInviteCount: number
  requiredRechargeAmount: number
  durationMinutes: number
  rewardAmount: number
  rewardCurrency?: string
  rewardTarget?: string
  claimed: boolean
  status?: TaskActivityStatus | null
  claimedAt?: string | null
  deadlineAt?: string | null
  remainingSeconds?: number
  progressInviteCount?: number
  progressRechargeCount?: number
  progressRechargeAmount?: number
  progressPercent?: number
}

export interface TaskActivityRecordsReq {
  currentPage?: number
  pageSize?: number
}

export interface TaskActivityClaimReq {
  levelCode?: string
  inviteCount?: number
}

export interface TaskActivityRecordsResp {
  list: TaskActivityItem[]
  total: number
  pageSize: number
  currentPage: number
}

export interface TgWalletTransferReq {
  amount: number
  fromWallet?: 'sport' | 'balance'
  toWallet?: 'sport' | 'balance'
}

export interface TgWalletTransferResp {
  fromWallet: string
  toWallet: string
  transferAmount: number
  balance: number
  sportBalance: number
}

export interface Profile {
  id: number
  username: string
  nickname?: string
}

export function loginApi(data: LoginParams) {
  return post<LoginResult>('/login', data)
}

export function profileApi() {
  return get<Profile>('/user/profile')
}

export async function loginByPhone(data: LoginByPhoneParams) {
  const deviceInfo = await getDeviceInfo()

  return post<LoginResult>('/v1/app/tg/phoneLogin', {
    country: APP_COUNTRY_CODE,
    ...data,
    ...deviceInfo,
  })
}

export function checkRegisterPhone(data: CheckRegisterPhoneParams) {
  return post<CheckRegisterPhoneResp>('/v1/app/tg/checkRegisterPhone', {
    phone: data.phone,
    country: data.country ?? APP_COUNTRY_CODE,
  }, {
    meta: {
      auth: false,
      showError: false,
    },
  })
}

export function sendTgSmsCode(phone: string, country: AppCountryCode = APP_COUNTRY_CODE) {
  return post('/v1/app/tg/sendSMSCode', {
    phone,
    country,
  }, {
    meta: {
      auth: false,
      showError: false,
    },
  })
}

export function forgotPasswordByPhone(data: ForgotPasswordByPhoneParams) {
  return post('/v1/app/tg/forgotPasswordByPhone', data, {
    meta: {
      auth: false,
      showError: false,
    },
  })
}

export function bindCurrentTgEmail(data: TgBindEmailReq) {
  return post<{ email: string }>('/v1/app/tg/bindEmail', data)
}

export function bindCurrentTgPhone(data: TgBindPhoneReq) {
  return post<{ phone: string; country: AppCountryCode }>('/v1/app/tg/bindPhone', data)
}

function resolveSourceChannelCode(data: RegisterByPhoneParams) {
  return (
    normalizeSourceChannelCode(data.sourceChannelCode) ||
    normalizeSourceChannelCode(data.channelCode) ||
    getSourceChannelCode()
  )
}

export async function registerByPhone(data: RegisterByPhoneParams) {
  const sourceChannelCode = resolveSourceChannelCode(data)
  const deviceInfo = await getDeviceInfo()
  if (data.deviceFingerprint?.trim()) {
    deviceInfo.deviceFingerprint = data.deviceFingerprint.trim()
  }

  return post<LoginResult>('/v1/app/tg/registerByPhone', {
    phone: data.phone,
    country: data.country ?? APP_COUNTRY_CODE,
    firstName: data.firstName,
    password: data.password,
    inviteCode: data.inviteCode,
    sourceChannelCode,
    channelCode: sourceChannelCode,
    ...deviceInfo,
  })
}

export function getCurrentUserInfo() {
  return get<CurrentUserInfo>('/v1/app/tg/currentUserInfo')
}

export function reportCurrentTgDeviceInfo(data: DeviceInfoPayload) {
  return post('/v1/app/tg/deviceInfo', data)
}

export interface UserRechargeCount {
  rechargeCount: number
}

export function getCurrentUserRechargeCount() {
  return get<UserRechargeCount>('/v1/app/recharge/count')
}

export interface VipWeeklySalary {
  level: number
  levelName: string
  weeklySalary: number
  claimable: boolean
  claimed: boolean
  nextResetAt: string
}

export function getVipWeeklySalary() {
  return get<VipWeeklySalary>('/v1/app/vip/weeklySalary')
}

export function claimVipWeeklySalary() {
  return post<string>('/v1/app/vip/weeklySalary/claim')
}

export function getCurrentTgInviteStats() {
  return get<TgInviteStats>('/v1/app/tg/inviteStats')
}

export function getCurrentTgInviteRuleConfig() {
  return get<TgInviteRuleConfig>('/v1/app/tg/inviteRuleConfig')
}

function withLangParam(path: string, lang?: string) {
  const normalized = String(lang || '').trim()
  if (!normalized) return path

  const joiner = path.includes('?') ? '&' : '?'
  return `${path}${joiner}lang=${encodeURIComponent(normalized)}`
}

export function getTaskActivityList(lang?: string) {
  return get<TaskActivityItem[]>(withLangParam('/v1/app/taskActivity/list', lang), {
    meta: { showError: false },
  })
}

export function claimTaskActivity(dataOrLang: TaskActivityClaimReq | string = {}, lang?: string) {
  const data = typeof dataOrLang === 'string' ? {} : dataOrLang
  const requestLang = typeof dataOrLang === 'string' ? dataOrLang : lang
  return post<TaskActivityItem>(withLangParam('/v1/app/taskActivity/claim', requestLang), data, {
    meta: { showError: false },
  })
}

export function rewardTaskActivity(lang?: string) {
  return post<TaskActivityItem>(withLangParam('/v1/app/taskActivity/reward', lang), {}, {
    meta: { showError: false },
  })
}

export function getTaskActivityRecords(data: TaskActivityRecordsReq = {}, lang?: string) {
  return post<TaskActivityRecordsResp>(withLangParam('/v1/app/taskActivity/records', lang), data, {
    meta: { showError: false },
  })
}

export function appUpload(file: File) {
  return upload<{ url: string }>('/v1/app/upload', { file })
}

export function updateCurrentTgAvatar(avatar: string) {
  return post<{ avatar: string }>('/v1/app/tg/avatar', { avatar })
}

export function updateCurrentTgName(username: string) {
  return post<{ username: string }>('/v1/app/tg/name', { username })
}

export function getAppCountries() {
  return get<AppCountryItem[]>('/v1/app/countries')
}

export function getCountryRechargeInfo(code: string) {
  return get<AppCountryRechargeInfo>(`/v1/app/country/${code}/recharge`)
}

export function getCountryRechargeFields(code: string) {
  return get<RechargeField[]>(`/v1/app/country/${code}/rechargeFields`)
}

export function getCountryWithdrawFields(code: string) {
  return get<RechargeField[]>(`/v1/app/country/${code}/withdrawFields`)
}

export function getRechargePromotions() {
  return get<RechargePromotionsResp>('/v1/app/recharge/promotions')
}

export function createRechargeOrder(data: RechargeOrderAppReq) {
  return post<RechargeOrderAppBack>('/v1/app/rechargeOrder', data)
}

export function getRechargeIsFirstV2() {
  return get<RechargeFirstStatusResp>('/v1/app/recharge/isFirst/v2')
}

export function getRechargeCount() {
  return get<RechargeCountResp>('/v1/app/recharge/count')
}

export function createRechargeOrderV2(data: RechargeOrderAppReq) {
  return post<RechargeOrderAppBack>('/v1/app/rechargeOrder/v2', data)
}

export function getCryptoRechargeOptions(amount: number) {
  return get<CryptoRechargeOptionsResp>('/v1/app/rechargeOrder/crypto/options', {
    params: { amount },
  })
}

export function createCryptoRechargeOrder(data: CryptoRechargeOrderReq) {
  return post<RechargeOrderAppBack>('/v1/app/rechargeOrder/crypto', data)
}

export function getCryptoRechargeOrderStatus(orderNo: string) {
  return get<CryptoRechargeOrderStatusResp>(`/v1/app/rechargeOrder/crypto/${orderNo}/status`, {
    meta: { showError: false },
  })
}

export function cancelCryptoRechargeOrder(orderNo: string) {
  return post<CryptoRechargeOrderStatusResp>(`/v1/app/rechargeOrder/crypto/${orderNo}/cancel`, {})
}

export function getRechargeOrderHistory(data: AppOrderHistoryReq) {
  return post<AppOrderHistoryResp>('/v1/app/rechargeOrder/list', data)
}

export function getAppUserBetRecordList(data: AppUserBetRecordReq) {
  return post<AppUserBetRecordResp>('/v1/app/appUserBetRecord/list', data)
}

export function getWithdrawOrderHistory(data: WithdrawOrderHistoryReq) {
  return post<WithdrawOrderHistoryResp>('/v1/app/withdraw/list', data)
}

export function getCurrentTgWithdrawSummary() {
  return get<TgWithdrawSummary>('/v1/app/tg/withdrawSummary')
}

export function getCurrentTgWithdrawFlowBatchSummary() {
  return get<TgWithdrawFlowBatchSummary>('/v1/app/withdraw/v2/summary')
}

export function getTenantServiceLinks() {
  return get<TenantServiceLinks>('/v1/app/domain/serviceLinks', {
    meta: { showError: false },
  })
}

export function createWithdrawOrder(data: CreateWithdrawOrderReq) {
  return post<CreateWithdrawOrderResp>('/v1/app/withdraw', data)
}

export function createWithdrawOrderV2(data: CreateWithdrawOrderReq) {
  return post<CreateWithdrawOrderResp>('/v1/app/withdraw/v2', data, {
    meta: { showError: false },
  })
}

export function createRebateWithdrawOrder(data: CreateWithdrawOrderReq) {
  return post<RebateWithdrawOrderResp>('/v1/app/tg/rebate/withdraw', data, {
    meta: { showError: false },
  })
}

export function transferRebateToBalance(data: TgRebateTransferReq) {
  return post<TgRebateTransferResp>('/v1/app/tg/rebate/transfer', data, {
    meta: { showError: false },
  })
}

export function transferTgWalletBalance(data: TgWalletTransferReq) {
  return post<TgWalletTransferResp>('/v1/app/tg/wallet/transfer', data, {
    meta: { showError: false },
  })
}

export function getLotteryChances() {
  return get<LotteryChancesResp>('/v1/app/lottery/chances')
}

export function drawLottery() {
  return post<LotteryDrawResp>('/v1/app/lottery/draw', {})
}

export function getCheckInStatus() {
  return get<CheckInStatusResp>('/v1/app/checkin/status')
}

export function doCheckIn() {
  return post<CheckInResp>('/v1/app/checkin', {})
}

export function redeemExchangeCode(code: string) {
  return post<ExchangeCodeRedeemResp>(
    '/v1/app/exchangeCode/redeem',
    { code },
    { meta: { showError: false } },
  )
}

export function getCheckInRecords(limit = 30) {
  return get<CheckInRecordItem[]>(`/v1/app/checkin/records?limit=${limit}`)
}

export function getPrizePoolBalance(poolCode = 'lucky') {
  return get<PrizePoolBalanceResp>(`/v1/app/prizePool/balance?poolCode=${poolCode}`)
}

export function getPrizePoolOutRecords(currentPage = 0, pageSize = 10) {
  return get<PrizePoolOutRecordResp>(`/v1/app/prizePool/outRecords?currentPage=${currentPage}&pageSize=${pageSize}`)
}

export function getPendingRechargeNotifications() {
  return get<RechargeSuccessNotification[]>('/v1/app/rechargeOrder/pendingNotifications')
}

export function ackRechargeNotification(orderNo: string) {
  return post<string>('/v1/app/rechargeOrder/notifyAck', { orderNo })
}

export function getWithdrawAccounts() {
  return get<WithdrawAccountItem[]>('/v1/app/withdrawAccount/list')
}

export function addWithdrawAccount(data: AddWithdrawAccountReq) {
  return post<WithdrawAccountItem>('/v1/app/withdrawAccount', data)
}

export function updateWithdrawAccount(id: number, data: AddWithdrawAccountReq) {
  return post<WithdrawAccountItem>(`/v1/app/withdrawAccount/${id}/update`, data)
}

export function deleteWithdrawAccount(id: number) {
  return del<string>(`/v1/app/withdrawAccount/${id}`)
}

export function setDefaultWithdrawAccount(id: number) {
  return post<string>(`/v1/app/withdrawAccount/${id}/setDefault`, {})
}
