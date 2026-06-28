import { http } from "@/utils/http";

export type TenantDashboardPeriodStats = {
  rechargeAmount: number;
  betAmount: number;
  withdrawAmount: number;
  rebateAmount: number;
  platformPumpAmount: number;
  rechargeUsers: number;
  repeatRechargeUsers: number;
  registerUsers: number;
};

export type TenantDashboardStats = {
  today: TenantDashboardPeriodStats;
  yesterday: TenantDashboardPeriodStats;
  month: TenantDashboardPeriodStats;
  totalPlatformPumpAmount: number;
  totalRegisterUsers: number;
  onlineUsers: number;
};

export type TenantDashboardStatsResult = {
  code: number;
  message: string;
  success: boolean;
  data: TenantDashboardStats;
};

export type TenantDashboardMonthlyBalance = {
  month: string;
  rechargeAmount: number;
  withdrawAmount: number;
  balanceAmount: number;
};

export type TenantDashboardMonthlyBalanceResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    year: number;
    list: TenantDashboardMonthlyBalance[];
    totalRechargeAmount: number;
    totalWithdrawAmount: number;
    totalBalanceAmount: number;
  };
};

export type TenantDashboardDetailSearch = {
  currentPage: number;
  pageSize: number;
  period?: "today" | "yesterday" | "month" | "total";
};

export type TenantDashboardUserDetail = {
  id: number;
  uid: string;
  tgId: number;
  username?: string | null;
  firstName?: string | null;
  phone?: string | null;
  parentId?: number | null;
  parentUid?: string | null;
  balance: number;
  status: number;
  rechargeAmount: number;
  rechargeCount: number;
  lastRechargeAt?: string | null;
  lastActiveAt?: string | null;
  registeredAt?: string | null;
};

export type TenantDashboardUserDetailResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: TenantDashboardUserDetail[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export type TenantDashboardAgentRank = {
  id: number;
  tenantId: number;
  uid: string;
  tgId: number;
  username?: string | null;
  firstName?: string | null;
  phone?: string | null;
  balance: number;
  status: number;
  subRechargeAmount: number;
  subRechargeUsers: number;
  subRechargeCount: number;
  lastRechargeAt?: string | null;
};

export type TenantDashboardAgentRankResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: TenantDashboardAgentRank[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export const getTenantDashboardStats = () => {
  return http.request<TenantDashboardStatsResult>(
    "get",
    "/api/v1/tenant/dashboard/stats"
  );
};

export const getTenantDashboardMonthlyBalances = (year?: number) => {
  return http.request<TenantDashboardMonthlyBalanceResult>(
    "get",
    "/api/v1/tenant/dashboard/monthlyBalances",
    { params: { year } }
  );
};

export const getTenantDashboardOnlineUsers = (
  data: TenantDashboardDetailSearch
) => {
  return http.request<TenantDashboardUserDetailResult>(
    "post",
    "/api/v1/tenant/dashboard/onlineUsers",
    { data }
  );
};

export const getTenantDashboardRechargeUsers = (
  data: TenantDashboardDetailSearch
) => {
  return http.request<TenantDashboardUserDetailResult>(
    "post",
    "/api/v1/tenant/dashboard/rechargeUsers",
    { data }
  );
};

export const getTenantDashboardRegisterUsers = (
  data: TenantDashboardDetailSearch
) => {
  return http.request<TenantDashboardUserDetailResult>(
    "post",
    "/api/v1/tenant/dashboard/registerUsers",
    { data }
  );
};

export const getTenantDashboardAgentRanks = (
  data: TenantDashboardDetailSearch
) => {
  return http.request<TenantDashboardAgentRankResult>(
    "post",
    "/api/v1/tenant/dashboard/agentRanks",
    { data }
  );
};

export type TenantDashboardOrderDetail = {
  id: number;
  orderNo: string;
  tenantId: number;
  userId: number;
  uid: string;
  username?: string | null;
  firstName?: string | null;
  phone?: string | null;
  amount: number;
  fee: number;
  netAmount: number;
  channel: string;
  status: number;
  time?: string | null;
};

export type TenantDashboardOrderDetailResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: TenantDashboardOrderDetail[];
    total: number;
    pageSize: number;
    currentPage: number;
    totalAmount: number;
  };
};

export const getTenantDashboardRechargeOrders = (
  data: TenantDashboardDetailSearch
) => {
  return http.request<TenantDashboardOrderDetailResult>(
    "post",
    "/api/v1/tenant/dashboard/rechargeOrders",
    { data }
  );
};

export const getTenantDashboardWithdrawOrders = (
  data: TenantDashboardDetailSearch
) => {
  return http.request<TenantDashboardOrderDetailResult>(
    "post",
    "/api/v1/tenant/dashboard/withdrawOrders",
    { data }
  );
};
