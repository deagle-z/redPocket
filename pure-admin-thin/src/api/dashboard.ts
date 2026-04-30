import { http } from "@/utils/http";

export type AdminDashboardPeriodStats = {
  rechargeAmount: number;
  betAmount: number;
  withdrawAmount: number;
  rebateAmount: number;
  platformPumpAmount: number;
  rechargeUsers: number;
};

export type AdminDashboardStats = {
  today: AdminDashboardPeriodStats;
  month: AdminDashboardPeriodStats;
  totalPlatformPumpAmount: number;
  onlineUsers: number;
};

export type AdminDashboardStatsResult = {
  code: number;
  message: string;
  success: boolean;
  data: AdminDashboardStats;
};

export type AdminDashboardDetailSearch = {
  currentPage: number;
  pageSize: number;
  period?: "today" | "month";
};

export type AdminDashboardUserDetail = {
  id: number;
  tenantId: number;
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

export const getAdminDashboardStats = () => {
  return http.request<AdminDashboardStatsResult>(
    "get",
    "/api/v1/admin/dashboard/stats"
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
