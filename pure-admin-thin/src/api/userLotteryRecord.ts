import { http } from "@/utils/http";

export type UserLotteryRecord = {
  id: number;
  tenantId: number;
  userId: number;
  poolId: number;
  configId: number;
  peerAmount: number;
  awardAmount: number;
  beforeBalance: number;
  afterBalance: number;
  status: number; // 0待结算 1已发放 2未中奖
  remark?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type UserLotteryRecordSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  userId?: number;
  poolId?: number;
  status?: number | null;
  startTime?: number;
  endTime?: number;
};

type ListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: UserLotteryRecord[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

type SingleResult = {
  code: number;
  message: string;
  success: boolean;
  data: UserLotteryRecord;
};

export const getUserLotteryRecords = (data: UserLotteryRecordSearch) =>
  http.request<ListResult>("post", "/api/v1/admin/userLotteryRecord/list", { data });

export const getUserLotteryRecordById = (id: number) =>
  http.request<SingleResult>("get", `/api/v1/admin/userLotteryRecord/${id}`);

