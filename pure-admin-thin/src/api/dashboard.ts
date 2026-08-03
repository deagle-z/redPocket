import { http } from "@/utils/http";

export type DomainVisitStats = {
  date: string;
  pageViews: number;
  uniqueVisitors: number;
};

export type DomainVisitRow = {
  domain: string;
  pageViews: number;
  uniqueVisitors: number;
};

export type DomainVisitHourly = {
  hour: number;
  pageViews: number;
  uniqueVisitors: number;
};

export type DomainVisitStatsResult = {
  code: number;
  message: string;
  success: boolean;
  data: DomainVisitStats;
};

export type DomainVisitDetailResult = {
  code: number;
  message: string;
  success: boolean;
  data: DomainVisitStats & {
    list: DomainVisitRow[];
    total: number;
    pageSize: number;
    currentPage: number;
    hourly: DomainVisitHourly[];
  };
};

export const getAdminDomainVisitStats = () => {
  return http.request<DomainVisitStatsResult>(
    "get",
    "/api/v1/admin/dashboard/domainVisitStats"
  );
};

export const getAdminDomainVisits = (data: {
  currentPage: number;
  pageSize: number;
  domain?: string;
}) => {
  return http.request<DomainVisitDetailResult>(
    "post",
    "/api/v1/admin/dashboard/domainVisits",
    { data }
  );
};

export type AdminDashboardPeriodStats = {
  rechargeAmount: number;
  betAmount: number;
  withdrawAmount: number;
  rebateAmount: number;
  platformPumpAmount: number;
  rechargeUsers: number;
  repeatRechargeUsers: number;
  registerUsers: number;
};

export type AdminDashboardStats = {
  today: AdminDashboardPeriodStats;
  yesterday: AdminDashboardPeriodStats;
  month: AdminDashboardPeriodStats;
  totalPlatformPumpAmount: number;
  totalRegisterUsers: number;
  onlineUsers: number;
};

export type AdminDashboardStatsResult = {
  code: number;
  message: string;
  success: boolean;
  data: AdminDashboardStats;
};

export type AdminDashboardMonthlyBalance = {
  month: string;
  rechargeAmount: number;
  withdrawAmount: number;
  balanceAmount: number;
};

export type AdminDashboardMonthlyBalanceResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    year: number;
    list: AdminDashboardMonthlyBalance[];
    totalRechargeAmount: number;
    totalWithdrawAmount: number;
    totalBalanceAmount: number;
  };
};

export type AdminDashboardDetailSearch = {
  currentPage: number;
  pageSize: number;
  period?: "today" | "yesterday" | "month" | "total";
  tenantId?: number;
};

export type AdminDashboardUserDetail = {
  id: number;
  tenantId: number;
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

export type AdminDashboardUserDetailResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: AdminDashboardUserDetail[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export type AdminDashboardAgentRank = {
  id: number;
  tenantId: number;
  tenantName?: string | null;
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

export type AdminDashboardAgentRankResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: AdminDashboardAgentRank[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export const getAdminDashboardStats = (tenantId?: number) => {
  return http.request<AdminDashboardStatsResult>(
    "get",
    "/api/v1/admin/dashboard/stats",
    tenantId ? { params: { tenantId } } : undefined
  );
};

export const getAdminDashboardMonthlyBalances = (
  year?: number,
  tenantId?: number
) => {
  return http.request<AdminDashboardMonthlyBalanceResult>(
    "get",
    "/api/v1/admin/dashboard/monthlyBalances",
    { params: { year, tenantId } }
  );
};

export const getAdminDashboardOnlineUsers = (
  data: AdminDashboardDetailSearch
) => {
  return http.request<AdminDashboardUserDetailResult>(
    "post",
    "/api/v1/admin/dashboard/onlineUsers",
    { data }
  );
};

export const getAdminDashboardRechargeUsers = (
  data: AdminDashboardDetailSearch
) => {
  return http.request<AdminDashboardUserDetailResult>(
    "post",
    "/api/v1/admin/dashboard/rechargeUsers",
    { data }
  );
};

export const getAdminDashboardRegisterUsers = (
  data: AdminDashboardDetailSearch
) => {
  return http.request<AdminDashboardUserDetailResult>(
    "post",
    "/api/v1/admin/dashboard/registerUsers",
    { data }
  );
};

export const getAdminDashboardAgentRanks = (
  data: AdminDashboardDetailSearch
) => {
  return http.request<AdminDashboardAgentRankResult>(
    "post",
    "/api/v1/admin/dashboard/agentRanks",
    { data }
  );
};

export type AdminDashboardOrderDetail = {
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

export type AdminDashboardOrderDetailResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: AdminDashboardOrderDetail[];
    total: number;
    pageSize: number;
    currentPage: number;
    totalAmount: number;
  };
};

export const getAdminDashboardRechargeOrders = (
  data: AdminDashboardDetailSearch
) => {
  return http.request<AdminDashboardOrderDetailResult>(
    "post",
    "/api/v1/admin/dashboard/rechargeOrders",
    { data }
  );
};

export const getAdminDashboardWithdrawOrders = (
  data: AdminDashboardDetailSearch
) => {
  return http.request<AdminDashboardOrderDetailResult>(
    "post",
    "/api/v1/admin/dashboard/withdrawOrders",
    { data }
  );
};
