import { http } from "@/utils/http";

export type AdminDashboardPeriodStats = {
  rechargeAmount: number;
  betAmount: number;
  withdrawAmount: number;
  rebateAmount: number;
  platformPumpAmount: number;
  rechargeUsers: number;
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

export const getAdminDashboardStats = (tenantId?: number) => {
  return http.request<AdminDashboardStatsResult>(
    "get",
    "/api/v1/admin/dashboard/stats",
    tenantId ? { params: { tenantId } } : undefined
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
