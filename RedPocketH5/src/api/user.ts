import request from '@/utils/request'

export interface ApiResult<T> {
  code: number
  msg: string
  data: T
}

export interface LoginData {
  email: string
  password: string
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
  balance?: number
  uid?: string
  username?: string
  email?: string
  tg_id?: number
  gift_amount?: number
  rebate_amount?: number
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
}

export interface RechargeOrderAppReq {
  amount: number
  channel: string
  payMethod?: string
  currency?: string
  merchantOrderNo?: string
  extraFields?: Record<string, string>
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
}

export interface LuckyMoneySendReq {
  amount: number
  thunder: number
  number?: number
  chatId?: number
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
  chatId?: number
  status?: number
  luckyId?: number
}

export interface LuckyPacketListItem {
  id: number
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
}

export interface LuckyDetailReq {
  luckyId: number
}

export interface LuckyDetailSummary {
  id: number
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
  loseMoney: number
  isThunder: number
  thunder: number
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
  return request.post<ApiResult<TgAuthLoginRes>>('/api/v1/app/tg/login', data)
}

export function loginByEmail(data: LoginData) {
  return request.post<ApiResult<LoginRes>>('/api/v1/app/tg/loginByEmail', data)
}

export function logout() {
  return request.post('/user/logout')
}

export function getUserInfo() {
  return request<ApiResult<UserState>>('/user/me')
}

export interface RegisterData {
  email: string
  code: string
  password: string
  confirmPassword: string
  inviteCode: string
}

export interface ForgotPasswordData {
  email: string
  code: string
  newPassword: string
}

export function getEmailCode(): Promise<any> {
  return request.get('/user/email-code')
}

export function sendRegisterEmailCode(email: string): Promise<any> {
  return request.post('/api/v1/app/tg/sendEmailCode', { email })
}

export function resetPassword(): Promise<any> {
  return request.post('/user/reset-password')
}

export function forgotPasswordByEmail(data: ForgotPasswordData): Promise<any> {
  return request.post('/api/v1/app/tg/forgotPasswordByEmail', data)
}

export function getCurrentTgUserInfo() {
  return request.get<ApiResult<TgCurrentUserInfo>>('/api/v1/app/tg/currentUserInfo')
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
  code: string
}

export function transferRebateToBalance() {
  return request.post<ApiResult<TgRebateTransferResp>>('/api/v1/app/tg/rebate/transfer')
}

export function bindCurrentTgEmail(data: TgBindEmailReq) {
  return request.post<ApiResult<{ email: string }>>('/api/v1/app/tg/bindEmail', data)
}

export function createRechargeOrder(data: RechargeOrderAppReq) {
  return request.post<ApiResult<RechargeOrderAppBack>>('/api/v1/app/rechargeOrder', data)
}

export function sendLuckyPacket(data: LuckyMoneySendReq) {
  return request.post<ApiResult<LuckyMoneyBack>>('/api/v1/app/lucky/send', data)
}

export function getLuckyPacketList(data: LuckyPacketListReq) {
  return request.post<ApiResult<LuckyPacketListResp>>('/api/v1/app/lucky/list', data)
}

export function grabLuckyPacket(data: LuckyGrabReq) {
  return request.post<ApiResult<any>>('/api/v1/app/lucky/grab', data)
}

export function getLuckyDetail(data: LuckyDetailReq) {
  return request.post<ApiResult<LuckyDetailResp>>('/api/v1/app/lucky/detail', data)
}

export function getLuckyRecentWinners(data: LuckyRecentWinnerReq = {}) {
  return request.post<ApiResult<LuckyRecentWinnerItem[]>>('/api/v1/app/lucky/recentWinners', data)
}

export function getLuckyAppHistory(data: LuckyAppHistoryReq) {
  return request.post<ApiResult<LuckyAppHistoryResp>>('/api/v1/app/lucky/history', data)
}

export function getAppCashHistoryList(data: AppCashHistoryReq) {
  return request.post<ApiResult<AppCashHistoryResp>>('/api/v1/app/cashHistory/list', data)
}

export function register(data: RegisterData): Promise<any> {
  return request.post('/api/v1/app/tg/registerByEmail', {
    email: data.email,
    code: data.code,
    password: data.password,
  })
}

export interface AppCountryItem {
  id: number
  countryCode: string
  countryNameEn: string
  countryNameCn: string
  currencyCode: string
  currencySymbol?: string | null
  sort: number
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

export interface RechargeField {
  fieldKey: string
  fieldLabel: string
  fieldPlaceholder?: string | null
  fieldType: 'input' | 'textarea' | 'number'
  dataType: string
  isRequired: number
  defaultValue?: string | null
  maxLength?: number | null
  minLength?: number | null
  errorTips?: string | null
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
