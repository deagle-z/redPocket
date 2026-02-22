import { http } from "@/utils/http";

export type SysTenant = {
  id: number;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string | null;
  tenantCode: string;
  tenantName: string;
  tenantType: number;
  status: number;
  ownerUserId?: number | null;
  planCode?: string | null;
  timezone: string;
  locale: string;
  remark?: string | null;
};

export type SysTenantSearch = {
  currentPage: number;
  pageSize: number;
  tenantCode?: string;
  tenantName?: string;
  tenantType?: number;
  status?: number;
  ownerUserId?: number;
  planCode?: string;
};

export type SysTenantSet = {
  id: number;
  tenantCode: string;
  tenantName: string;
  tenantType: number;
  status: number;
  ownerUserId?: number | null;
  planCode?: string | null;
  timezone: string;
  locale: string;
  remark?: string | null;
};

export type SysTenantListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: SysTenant[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export type SysTenantResult = {
  code: number;
  message: string;
  success: boolean;
  data?: SysTenant;
};

export type SysTenantResetPassword = {
  tenantId: number;
  password: string;
};

export const getSysTenantList = (data: SysTenantSearch) => {
  return http.request<SysTenantListResult>("post", "/api/v1/admin/tenant/list", {
    data
  });
};

export const setSysTenant = (data: SysTenantSet) => {
  return http.request<SysTenantResult>("post", "/api/v1/admin/tenant", { data });
};

export const resetSysTenantPassword = (data: SysTenantResetPassword) => {
  return http.request<SysTenantResult>("post", "/api/v1/admin/tenant/resetPassword", {
    data
  });
};

export const delSysTenant = (id: number) => {
  return http.request<SysTenantResult>(
    "delete",
    "/api/v1/admin/tenant/" + id
  );
};
