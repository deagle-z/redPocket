import { http } from "@/utils/http";

export type PrizePool = {
  id: number;
  tenantId: number;
  poolCode: string;
  poolName: string;
  currency: string;
  balance: number;
  frozenBalance: number;
  minBalance: number;
  maxBalance: number | null;
  rtpRate: number;
  pumpRate: number;
  injectRate: number;
  status: number;
  remark?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type PrizePoolSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  poolCode?: string;
  status?: number | null;
};

export type PrizePoolSet = {
  id?: number;
  poolCode: string;
  poolName: string;
  currency: string;
  minBalance: number;
  maxBalance: number | null;
  rtpRate: number;
  pumpRate: number;
  injectRate: number;
  status: number;
  remark?: string | null;
};

type PrizePoolListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: PrizePool[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

type PrizePoolResult = {
  code: number;
  message: string;
  success: boolean;
  data: PrizePool;
};

export const getPrizePoolList = (data: PrizePoolSearch) => {
  return http.request<PrizePoolListResult>(
    "post",
    "/api/v1/admin/prizePool/list",
    { data }
  );
};

export const setPrizePool = (data: PrizePoolSet) => {
  return http.request<PrizePoolResult>(
    "post",
    "/api/v1/admin/prizePool",
    { data }
  );
};

export const delPrizePool = (id: number) => {
  return http.request<PrizePoolResult>(
    "delete",
    `/api/v1/admin/prizePool/${id}`
  );
};
