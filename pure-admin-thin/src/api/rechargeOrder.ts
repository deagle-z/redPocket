import { http } from "@/utils/http";

export type RechargeOrder = {
  id: number;
  createdAt: string;
  updatedAt: string;
  tenantId: number;
  appId?: number | null;
  userId: number;
  accountId?: string | null;
  orderNo: string;
  merchantOrderNo?: string | null;
  channel: string;
  payMethod?: string | null;
  currency: string;
  amount: number;
  fee: number;
  netAmount: number;
  creditAmount?: number | null;
  bonusAmount: number;
  status: number;
  payTime?: string | null;
  expireTime?: string | null;
  notifyTime?: string | null;
  provider?: string | null;
  providerTradeNo?: string | null;
  providerStatus?: string | null;
  notifyCount: number;
  notifyLastAt?: string | null;
  idempotencyKey?: string | null;
  title?: string | null;
  remark?: string | null;
  extra?: string | null;
};

export type RechargeOrderSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  userId?: number;
  status?: number;
  orderNo?: string;
  merchantOrderNo?: string;
  providerTradeNo?: string;
  channel?: string;
  payMethod?: string;
};

export type RechargeOrderListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: RechargeOrder[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export const getRechargeOrderListAdmin = (data: RechargeOrderSearch) => {
  return http.request<RechargeOrderListResult>(
    "post",
    "/api/v1/admin/rechargeOrder/list",
    { data }
  );
};
