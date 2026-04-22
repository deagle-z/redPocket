import { http } from "@/utils/http";

export type PrizePoolConfig = {
  id: number;
  tenantId: number;
  poolId: number;
  probabilities: string;
  amounts: string;
  totalProbability: number;
  count: number;
  peerAmount: number;
  status: number;
  remark?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type PrizePoolConfigSet = {
  id?: number;
  poolId: number;
  probabilities: string;
  amounts: string;
  totalProbability: number;
  count: number;
  peerAmount: number;
  status: number;
  remark?: string | null;
};

export type PrizePoolBalance = {
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

export type PrizePoolBalanceSet = {
  poolCode: string;
  balance: number;
  remark?: string | null;
};

type PrizePoolConfigResult = {
  code: number;
  message: string;
  success: boolean;
  data: PrizePoolConfig;
};

type PrizePoolBalanceResult = {
  code: number;
  message: string;
  success: boolean;
  data: PrizePoolBalance;
};

export const getPrizePoolConfig = (poolId: number) => {
  return http.request<PrizePoolConfigResult>(
    "get",
    `/api/v1/admin/prizePoolConfig/${poolId}`
  );
};

export const setPrizePoolConfig = (data: PrizePoolConfigSet) => {
  return http.request<PrizePoolConfigResult>(
    "post",
    "/api/v1/admin/prizePoolConfig",
    { data }
  );
};

export const delPrizePoolConfig = (id: number) => {
  return http.request<PrizePoolConfigResult>(
    "delete",
    `/api/v1/admin/prizePoolConfig/${id}`
  );
};

export const getPrizePoolBalance = (poolCode: string) => {
  return http.request<PrizePoolBalanceResult>(
    "get",
    `/api/v1/admin/prizePoolBalance/${poolCode}`
  );
};

export const setPrizePoolBalance = (data: PrizePoolBalanceSet) => {
  return http.request<PrizePoolBalanceResult>(
    "post",
    "/api/v1/admin/prizePoolBalance",
    { data }
  );
};
