import { http } from "@/utils/http";

/** 红包信息 */
export type LuckyMoney = {
  id: number;
  createdAt: string;
  updatedAt: string;
  senderId: number;
  senderName: string;
  amount: number;
  received: number;
  number: number;
  lucky: number;
  thunder: number;
  chatId: number;
  redList: string;
  loseRate: number;
  status: number;
};

/** 红包领取历史 */
export type LuckyHistory = {
  id: number;
  createdAt: string;
  updatedAt: string;
  userId: number;
  firstName: string;
  luckyId: number;
  isThunder: number;
  amount: number;
  loseMoney: number;
};

/** 红包搜索参数 */
export type LuckyMoneySearch = {
  currentPage: number;
  pageSize: number;
  senderId?: number;
  chatId?: number;
  status?: number;
};

/** 发送红包参数 */
export type LuckyMoneySend = {
  amount: number;
  thunder: number;
  number?: number;
  chatId: number;
};

/** 抢红包参数 */
export type LuckyMoneyGrab = {
  luckyId: number;
};

/** 红包列表响应 */
export type LuckyMoneyListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: LuckyMoney[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

/** 红包详情响应 */
export type LuckyMoneyDetailResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    luckyMoney: LuckyMoney;
    history: LuckyHistory[];
  };
};

/** 领取历史搜索参数 */
export type LuckyHistorySearch = {
  currentPage: number;
  pageSize: number;
  luckyId?: number;
  userId?: number;
};

/** 领取历史列表响应 */
export type LuckyHistoryListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: LuckyHistory[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

/** 红包状态响应 */
export type LuckyMoneyStatusResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    luckyMoney: LuckyMoney;
    grabbedCount: number;
    openNum: number;
  };
};

/** 发送红包 */
export const sendRedPacket = (data: LuckyMoneySend) => {
  return http.request<LuckyMoneyListResult>("post", "/api/v1/outside/lucky/send", { data });
};

/** 抢红包 */
export const grabRedPacket = (data: LuckyMoneyGrab) => {
  return http.request<any>("post", "/api/v1/outside/lucky/grab", { data });
};

/** 获取红包列表（用户端） */
export const getRedPacketList = (data: LuckyMoneySearch) => {
  return http.request<LuckyMoneyListResult>("post", "/api/v1/outside/lucky/list", { data });
};

/** 获取红包详情 */
export const getRedPacketDetail = (id: number) => {
  return http.request<LuckyMoneyDetailResult>("get", `/api/v1/outside/lucky/${id}`);
};

/** 获取红包状态 */
export const getRedPacketStatus = (id: number) => {
  return http.request<LuckyMoneyStatusResult>("get", `/api/v1/outside/lucky/status/${id}`);
};

/** 检查抢包余额 */
export const checkGrabBalance = (data: LuckyMoneyGrab) => {
  return http.request<any>("post", "/api/v1/outside/lucky/checkBalance", { data });
};

/** 管理员 - 获取红包列表 */
export const getLuckyMoneyListAdmin = (data: LuckyMoneySearch) => {
  return http.request<LuckyMoneyListResult>("post", "/api/v1/tenant/lucky/list", { data });
};

/** 管理员 - 获取领取历史 */
export const getLuckyHistoryListAdmin = (data: LuckyHistorySearch) => {
  return http.request<LuckyHistoryListResult>("post", "/api/v1/tenant/lucky/history", { data });
};

/** 管理员 - 获取红包详情 */
export const getLuckyMoneyDetailAdmin = (id: number) => {
  return http.request<LuckyMoneyDetailResult>("get", `/api/v1/tenant/lucky/${id}`);
};
