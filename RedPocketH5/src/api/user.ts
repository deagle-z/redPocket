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
  tg_id?: number
  gift_amount?: number
  rebate_amount?: number
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
  remainingSeconds: number
  remainingText: string
  items: Array<{
    seqNo: number
    amount: number
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

export function tgLogout() {
  return request.post('/api/v1/app/tg/logout')
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

export function getLuckyRecentWinners(data: LuckyRecentWinnerReq = {}) {
  return request.post<ApiResult<LuckyRecentWinnerItem[]>>('/api/v1/app/lucky/recentWinners', data)
}

export function register(data: RegisterData): Promise<any> {
  return request.post('/api/v1/app/tg/registerByEmail', {
    email: data.email,
    code: data.code,
    password: data.password,
  })
}
