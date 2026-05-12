import { http } from "@/utils/http";

export type TrialLuckyMoney = {
  id: number;
  createdAt: string;
  senderId: number;
  senderType: string;
  senderName: string;
  amount: number;
  received: number;
  number: number;
  thunder: number;
  gameMode: number;
  loseRate: number;
  status: number;
  expireTime: string;
};

export type TrialCashHistory = {
  id: number;
  createdAt: string;
  userId: number;
  actorType: string;
  amount: number;
  startAmount: number;
  endAmount: number;
  cashMark: string;
  cashDesc: string;
  type: number;
  isThunder: number;
  luckyId: number;
};

export type TrialLuckySearch = {
  currentPage: number;
  pageSize: number;
  senderId?: number;
  status?: number;
};

export type TrialCashHistorySearch = {
  currentPage: number;
  pageSize: number;
  userId?: number;
  actorType?: string;
  cashMark?: string;
};

export type TrialPageResult<T> = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: T[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export const getTrialLuckyListAdmin = (data: TrialLuckySearch) => {
  return http.request<TrialPageResult<TrialLuckyMoney>>(
    "post",
    "/api/v1/admin/trial/lucky/list",
    { data }
  );
};

export const getTrialCashHistoryListAdmin = (data: TrialCashHistorySearch) => {
  return http.request<TrialPageResult<TrialCashHistory>>(
    "post",
    "/api/v1/admin/trial/cashHistory/list",
    { data }
  );
};
