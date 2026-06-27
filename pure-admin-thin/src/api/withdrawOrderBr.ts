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
  autoReviewed: number;
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

export type WithdrawFlowBatch = {
  id: number;
  createdAt: string;
  updatedAt: string;
  tenantId: number;
  userId: number;
  sourceType: string;
  sourceOrderId: number;
  sourceOrderNo: string;
  activityType: number;
  activityCode: string;
  creditAmount: number;
  bonusAmount: number;
  baseAmount: number;
  withdrawMultiplier: number;
  giftMultiplier: number;
  requiredFlow: number;
  completedFlow: number;
  remainingFlow: number;
  progressPercent: number;
  status: number;
  closedReason: string;
  completedAt?: string | null;
  closedAt?: string | null;
  lastFlowAt?: string | null;
};

export type WithdrawFlowBatchOverview = {
  userId: number;
  balance: number;
  totalFlow: number;
  totalRequiredFlow: number;
  totalCompletedFlow: number;
  hasUnfinishedBatch: boolean;
  unfinishedRequiredFlow: number;
  unfinishedCompletedFlow: number;
  unfinishedRemainingFlow: number;
  batchCount: number;
  unfinishedBatchCount: number;
  batches: WithdrawFlowBatch[];
};

export type WithdrawFlowBatchOverviewResult = {
  code: number;
  message: string;
  success: boolean;
  data: WithdrawFlowBatchOverview;
};

export const getWithdrawOrderBrListAdmin = (data: WithdrawOrderBrSearch) => {
  return http.request<WithdrawOrderBrListResult>(
    "post",
    "/api/v1/admin/withdrawOrderBr/list",
    { data }
  );
};

export const setWithdrawOrderBr = (data: WithdrawOrderBrSet) => {
  return http.request<WithdrawOrderBrResult>(
    "post",
    "/api/v1/admin/withdrawOrderBr",
    { data }
  );
};

export const getTgUserWithdrawActivityFlowAdmin = (userId: number) => {
  return http.request<WithdrawActivityFlowResult>(
    "get",
    `/api/v1/admin/tgUser/${userId}/withdrawActivityFlow`
  );
};

export const getTgUserWithdrawFlowBatchOverviewAdmin = (userId: number) => {
  return http.request<WithdrawFlowBatchOverviewResult>(
    "get",
    `/api/v1/admin/tgUser/${userId}/withdrawFlowBatchOverview`
  );
};
