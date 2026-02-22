import { http } from "@/utils/http";

export type WithdrawOrderBr = {
  id: number;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string | null;
  tenantId: number;
  appId?: number | null;
  userId: number;
  accountId?: string | null;
  orderNo: string;
  merchantOrderNo?: string | null;
  currency: string;
  amount: number;
  fee: number;
  netAmount: number;
  channel: string;
  payMethod?: string | null;
  status: number;
  reviewedBy?: number | null;
  reviewedAt?: string | null;
  paidAt?: string | null;
  failCode?: string | null;
  failMsg?: string | null;
  receiverName?: string | null;
  receiverDocument?: string | null;
  receiverDocumentType?: string | null;
  pixKeyType?: string | null;
  pixKey?: string | null;
  bankCode?: string | null;
  bankName?: string | null;
  branchNumber?: string | null;
  accountNumber?: string | null;
  accountType?: string | null;
  provider?: string | null;
  providerPayoutNo?: string | null;
  providerStatus?: string | null;
  notifyTime?: string | null;
  notifyCount: number;
  idempotencyKey?: string | null;
  riskLevel: number;
  remark?: string | null;
  extra?: string | null;
};

export type WithdrawOrderBrSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  userId?: number;
  status?: number;
  orderNo?: string;
  merchantOrderNo?: string;
  providerPayoutNo?: string;
  channel?: string;
  payMethod?: string;
  receiverDocumentType?: string;
  receiverDocument?: string;
};

export type WithdrawOrderBrSet = {
  id: number;
  status?: number;
  failMsg?: string;
  reviewedAt?: string;
};

export type WithdrawOrderBrListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: WithdrawOrderBr[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export type WithdrawOrderBrResult = {
  code: number;
  message: string;
  success: boolean;
  data?: WithdrawOrderBr;
};

export const getWithdrawOrderBrListAdmin = (data: WithdrawOrderBrSearch) => {
  return http.request<WithdrawOrderBrListResult>(
    "post",
    "/api/v1/tenant/withdrawOrderBr/list",
    { data }
  );
};

export const setWithdrawOrderBr = (data: WithdrawOrderBrSet) => {
  return http.request<WithdrawOrderBrResult>(
    "post",
    "/api/v1/tenant/withdrawOrderBr",
    { data }
  );
};
