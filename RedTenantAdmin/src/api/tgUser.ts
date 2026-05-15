import { http } from "@/utils/http";

export type TgUser = {
  id: number;
  createdAt: string;
  updatedAt: string;
  username?: string | null;
  firstName?: string | null;
  avatar?: string | null;
  phone?: string | null;
  isBot?: boolean;
  tgId: number;
  uid?: string;
  balance: number;
  trialBalance: number;
  giftAmount: number;
  giftTotal: number;
  subRechargeAmount?: number;
  subFlowAmount?: number;
  subProfitAmount?: number;
  subWithdrawAmount?: number;
  rebateAmount: number;
  rebateTotalAmount: number;
  rebateRate: number;
  status: number;
  parentId?: number | null;
  parentUid?: string | null;
  inviteCode?: string | null;
  remark?: string | null;
  ip?: string | null;
  region?: string | null;
};

export type TgUserSearch = {
  currentPage: number;
  pageSize: number;
  tgId?: number;
  uid?: string;
  username?: string;
  firstName?: string;
  isBot?: boolean;
  status?: number;
  parentId?: number;
  parentUid?: string;
  inviteCode?: string;
};

export type TgUserStatusSet = {
  id: number;
  status: number;
};

export type TgUserRebateRateSet = {
  id: number;
  rebateRate: number;
};

export type TgUserRemarkSet = {
  id: number;
  remark: string;
};

export type TgUserSubStatsSummarySearch = {
  tgId?: number;
  uid?: string;
  username?: string;
  firstName?: string;
  isBot?: boolean;
  status?: number;
  parentId?: number;
  parentUid?: string;
  inviteCode?: string;
};

export type TgUserBatchCreateBotReq = {
  num: number;
  randomName: boolean;
  nameFile: string;
  avatarLinks: string[];
};

export type TgUserListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: TgUser[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export type TgUserResult = {
  code: number;
  message: string;
  success: boolean;
  data: TgUser;
};

export type TgUserSubStatsSummaryResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    subRechargeAmount: number;
    subFlowAmount: number;
    subProfitAmount: number;
    subWithdrawAmount: number;
  };
};

export type TgUserBatchCreateBotResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    count: number;
    list: TgUser[];
  };
};

export const getTgUserList = (data: TgUserSearch) => {
  return http.request<TgUserListResult>(
    "post",
    "/api/v1/tenant/tgUser/listWithSubStats",
    {
      data
    }
  );
};

export const getTgUserSubStatsSummary = (data: TgUserSubStatsSummarySearch) => {
  return http.request<TgUserSubStatsSummaryResult>(
    "post",
    "/api/v1/tenant/tgUser/subStatsSummary",
    {
      data
    }
  );
};

export const setTgUserStatus = (data: TgUserStatusSet) => {
  return http.request<TgUserResult>("post", "/api/v1/tenant/tgUser/status", {
    data
  });
};

export const setTgUserRebateRate = (data: TgUserRebateRateSet) => {
  return http.request<TgUserResult>(
    "post",
    "/api/v1/tenant/tgUser/rebateRate",
    {
      data
    }
  );
};

export const setTgUserRemark = (data: TgUserRemarkSet) => {
  return http.request<TgUserResult>("post", "/api/v1/tenant/tgUser/remark", {
    data
  });
};

export const getAdminBotUserList = (data: TgUserSearch) => {
  return http.request<TgUserListResult>("post", "/api/v1/admin/tgUser/list", {
    data: { ...data, isBot: true }
  });
};

export const batchCreateBotUsers = (data: TgUserBatchCreateBotReq) => {
  return http.request<TgUserBatchCreateBotResult>(
    "post",
    "/api/v1/admin/tgUser/batchCreateBot",
    {
      data
    }
  );
};

export const setAdminTgUserStatus = (data: TgUserStatusSet) => {
  return http.request<TgUserResult>("post", "/api/v1/admin/tgUser/status", {
    data
  });
};
