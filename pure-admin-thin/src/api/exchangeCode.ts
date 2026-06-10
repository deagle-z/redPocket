import { http } from "@/utils/http";
import type { PageResult, Result } from "@/api/system";

export type ExchangeCodeStatus = 1 | 0 | -1;

export type ExchangeCode = {
  id: number;
  createdAt: string;
  updatedAt: string;
  code: string;
  amount: number;
  maxRedeemCount: number;
  redeemCount: number;
  status: ExchangeCodeStatus;
  remark: string;
  createdBy: number;
};

export type ExchangeCodeSearch = {
  currentPage: number;
  pageSize: number;
  code?: string;
  status?: ExchangeCodeStatus | null;
};

export type ExchangeCodeSet = {
  id?: number;
  code?: string;
  amount?: number;
  maxRedeemCount?: number;
  generateCount?: number;
  status?: 1 | 0;
  remark?: string;
};

export const getExchangeCodeListAdmin = (data: ExchangeCodeSearch) => {
  return http.request<PageResult>("post", "/api/v1/admin/exchangeCode/list", {
    data
  });
};

export const getExchangeCodeByIdAdmin = (id: number) => {
  return http.request<Result>("get", `/api/v1/admin/exchangeCode/${id}`);
};

export const setExchangeCodeAdmin = (data: ExchangeCodeSet) => {
  return http.request<Result>("post", "/api/v1/admin/exchangeCode", { data });
};

export const delExchangeCodeAdmin = (id: number) => {
  return http.request<Result>("delete", `/api/v1/admin/exchangeCode/${id}`);
};
