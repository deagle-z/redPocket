import { http } from "@/utils/http";

export type SysTenantUser = {
  id: number;
  createdAt: string;
  updatedAt: string;
  tenantId: number;
  username: string;
  passwordHash: string;
  passwordAlgo: string;
  email?: string | null;
  mobile?: string | null;
  roleCode: string;
  isOwner: boolean;
  status: number;
  lastLoginAt?: string | null;
  lastLoginIp?: string | null;
  lastLoginUa?: string | null;
  loginFailCount: number;
  lockedUntil?: string | null;
  require2fa: boolean;
  twofaSecret?: string | null;
  remark?: string | null;
};

export type SysTenantUserSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  username?: string;
  email?: string;
  mobile?: string;
  roleCode?: string;
  isOwner?: boolean;
  status?: number;
  require2fa?: boolean;
};

export type SysTenantUserSet = {
  id: number;
  tenantId: number;
  username: string;
  passwordHash: string;
  passwordAlgo: string;
  email?: string | null;
  mobile?: string | null;
  roleCode: string;
  isOwner: boolean;
  status: number;
  require2fa: boolean;
  twofaSecret?: string | null;
  remark?: string | null;
};

export type SysTenantUserListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: SysTenantUser[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export type SysTenantUserResult = {
  code: number;
  message: string;
  success: boolean;
  data?: SysTenantUser;
};

export const getSysTenantUserList = (data: SysTenantUserSearch) => {
  return http.request<SysTenantUserListResult>(
    "post",
    "/api/v1/admin/tenantUser/list",
    { data }
  );
};

export const setSysTenantUser = (data: SysTenantUserSet) => {
  return http.request<SysTenantUserResult>(
    "post",
    "/api/v1/admin/tenantUser",
    { data }
  );
};

export const delSysTenantUser = (id: number) => {
  return http.request<SysTenantUserResult>(
    "delete",
    "/api/v1/admin/tenantUser/" + id
  );
};
