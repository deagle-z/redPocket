import { http } from "@/utils/http";

export type TgUserRebateRecord = {
  id: number;
  createdAt: string;
  updatedAt: string;
  tenantId?: number | null;
  subUserId: number;
  parentUserId: number;
  sourceType: number;
  sourceOrderId: string;
  sourceAmount: number;
  rebateRate: number;
  rebateAmount: number;
  currency: string;
  status: number;
  settledAt?: string | null;
  idempotencyKey: string;
  remark?: string | null;
};

export type TgUserRebateSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  subUserId?: number;
  parentUserId?: number;
  sourceType?: number;
  sourceOrderId?: string;
  status?: number;
  idempotencyKey?: string;
};

export type TgUserRebateListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: TgUserRebateRecord[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export const getTgUserRebateList = (data: TgUserRebateSearch) => {
  return http.request<TgUserRebateListResult>(
    "post",
    "/api/v1/admin/tgUserRebate/list",
    { data }
  );
};
