import { http } from "@/utils/http";

export type TenantDashboardPeriodStats = {
  rechargeAmount: number;
  betAmount: number;
  withdrawAmount: number;
  rebateAmount: number;
  platformPumpAmount: number;
  rechargeUsers: number;
};

export type TenantDashboardStats = {
  today: TenantDashboardPeriodStats;
  month: TenantDashboardPeriodStats;
  totalPlatformPumpAmount: number;
  onlineUsers: number;
};

export type TenantDashboardStatsResult = {
  code: number;
  message: string;
  success: boolean;
  data: TenantDashboardStats;
};

export type TenantDashboardDetailSearch = {
  currentPage: number;
  pageSize: number;
  period?: "today" | "month";
};

export type TenantDashboardUserDetail = {
  id: number;
  uid: string;
  tgId: number;
  username?: string | null;
  firstName?: string | null;
  phone?: string | null;
  balance: number;
  status: number;
  rechargeAmount: number;
  rechargeCount: number;
  lastRechargeAt?: string | null;
  lastActiveAt?: string | null;
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

export const getTenantDashboardStats = () => {
  return http.request<TenantDashboardStatsResult>(
    "get",
    "/api/v1/tenant/dashboard/stats"
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
