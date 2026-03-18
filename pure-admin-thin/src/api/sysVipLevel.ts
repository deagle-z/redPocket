import { http } from "@/utils/http";

export type SysVipLevel = {
  id: number;
  createdAt: string;
  updatedAt: string;
  tenantId: number;
  level: number;
  levelName: string;
  agentTag?: string | null;
  totalRechargeCount?: number | null;
  totalRechargeAmount?: number | null;
  totalValidBet?: number | null;
  monthRechargeAmount?: number | null;
  monthValidBet?: number | null;
  upgradeBonusAmount: number;
  upgradeType?: number | null;
  keepLevelCondition?: number | null;
  sort: number;
  status: number;
  remark?: string | null;
};

export type SysVipLevelSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  status?: number | null;
};

export type SysVipLevelSet = {
  id?: number;
  tenantId: number;
  level: number;
  levelName: string;
  agentTag?: string | null;
  totalRechargeCount?: number | null;
  totalRechargeAmount?: number | null;
  totalValidBet?: number | null;
  monthRechargeAmount?: number | null;
  monthValidBet?: number | null;
  upgradeBonusAmount: number;
  upgradeType?: number | null;
  keepLevelCondition?: number | null;
  sort: number;
  status: number;
  remark?: string | null;
};

type SysVipLevelListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: SysVipLevel[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

type SysVipLevelResult = {
  code: number;
  message: string;
  success: boolean;
  data: SysVipLevel;
};

export const getSysVipLevelList = (data: SysVipLevelSearch) => {
  return http.request<SysVipLevelListResult>(
    "post",
    "/api/v1/admin/sysVipLevel/list",
    { data }
  );
};

export const getSysVipLevelById = (id: number) => {
  return http.request<SysVipLevelResult>(
    "get",
    `/api/v1/admin/sysVipLevel/${id}`
  );
};

export const setSysVipLevel = (data: SysVipLevelSet) => {
  return http.request<SysVipLevelResult>(
    "post",
    "/api/v1/admin/sysVipLevel",
    { data }
  );
};

export const delSysVipLevel = (id: number) => {
  return http.request<SysVipLevelResult>(
    "delete",
    `/api/v1/admin/sysVipLevel/${id}`
  );
};
