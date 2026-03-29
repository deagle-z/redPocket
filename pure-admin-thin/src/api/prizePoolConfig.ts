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

type PrizePoolConfigResult = {
  code: number;
  message: string;
  success: boolean;
  data: PrizePoolConfig;
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
