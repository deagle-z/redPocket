import { http } from "@/utils/http";

export type PlatformProfitLedger = {
  id: number;
  createdAt: string;
  updatedAt: string;
  tenantId: number;
  userId: number;
  userUid: string;
  userName: string;
  sourceChannelId?: number | null;
  sourceType: string;
  sourceId: string;
  incomeAmount: number;
  expenseAmount: number;
  rebateAmount: number;
  actualIncomeAmount: number;
  netAmount: number;
  remark: string;
};

export type PlatformProfitLedgerSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  userId?: number;
  sourceType?: string;
  sourceId?: string;
  minNet?: number;
  maxNet?: number;
};

export type PlatformProfitLedgerListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: PlatformProfitLedger[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export const getPlatformProfitLedgerList = (
  data: PlatformProfitLedgerSearch
) => {
  return http.request<PlatformProfitLedgerListResult>(
    "post",
    "/api/v1/admin/platformProfitLedger/list",
    { data }
  );
};
