import { http } from "@/utils/http";

export type SysPayMethod = {
  id: number;
  createdAt: string;
  updatedAt: string;
  methodCode: string;
  methodName: string;
  icon?: string | null;
  sort: number;
  status: number;
  remark?: string | null;
};

export type SysPayMethodSearch = {
  currentPage: number;
  pageSize: number;
  methodCode?: string;
  methodName?: string;
  status?: number | null;
};

export type SysPayMethodSet = {
  id?: number;
  methodCode: string;
  methodName: string;
  icon?: string | null;
  sort: number;
  status: number;
  remark?: string | null;
};

type SysPayMethodListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: SysPayMethod[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

type SysPayMethodResult = {
  code: number;
  message: string;
  success: boolean;
  data: SysPayMethod;
};

export const getSysPayMethodList = (data: SysPayMethodSearch) => {
  return http.request<SysPayMethodListResult>(
    "post",
    "/api/v1/admin/sysPayMethod/list",
    { data }
  );
};

export const getSysPayMethodById = (id: number) => {
  return http.request<SysPayMethodResult>(
    "get",
    `/api/v1/admin/sysPayMethod/${id}`
  );
};

export const setSysPayMethod = (data: SysPayMethodSet) => {
  return http.request<SysPayMethodResult>(
    "post",
    "/api/v1/admin/sysPayMethod",
    { data }
  );
};

export const delSysPayMethod = (id: number) => {
  return http.request<SysPayMethodResult>(
    "delete",
    `/api/v1/admin/sysPayMethod/${id}`
  );
};
