import { http } from "@/utils/http";

export type WithdrawOrderBr = {
  id: number;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string | null;
  tenantId: number;
  appId?: number | null;
  userId: number;
  userUid?: string;
  accountId?: string | null;
  orderNo: string;
  merchantOrderNo?: string | null;
  currency: string;
  countryCode?: string;
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
  userUid?: string;
  status?: number;
  orderNo?: string;
  merchantOrderNo?: string;
  providerPayoutNo?: string;
  countryCode?: string;
  channel?: string;
  payMethod?: string;
  receiverDocumentType?: string;
  receiverDocument?: string;
};

export type WithdrawOrderBrSet = {
  id: number;
  status?: number;
  channel?: string;
  payMethod?: string;
  provider?: string;
  countryCode?: string;
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

export type WithdrawActivityFlowCycle = {
  id: number;
  createdAt: string;
  updatedAt: string;
  activityCode: string;
  activityType: number;
  status: number;
  multiplier: number;
  baseAmount: number;
  requiredFlow: number;
  flowStartValue: number;
  flowConsumed: number;
  currentFlow: number;
  availableFlow: number;
  remainingFlow: number;
  progressPercent: number;
  balanceThreshold: number;
  lastRechargeNo: string;
  endReason: string;
  startedAt?: string | null;
  endedAt?: string | null;
};

export type WithdrawActivityFlow = {
  userId: number;
  balance: number;
  totalFlow: number;
  hasActivity: boolean;
  activeActivity?: WithdrawActivityFlowCycle | null;
  activities: WithdrawActivityFlowCycle[];
};

export type WithdrawActivityFlowResult = {
  code: number;
  message: string;
  success: boolean;
  data: WithdrawActivityFlow;
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

export const getTgUserWithdrawActivityFlowTenant = (userId: number) => {
  return http.request<WithdrawActivityFlowResult>(
    "get",
    `/api/v1/tenant/tgUser/${userId}/withdrawActivityFlow`
  );
};
