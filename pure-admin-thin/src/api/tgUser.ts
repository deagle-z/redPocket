import { http } from "@/utils/http";

export type TgUser = {
  id: number;
  createdAt: string;
  updatedAt: string;
  username?: string | null;
  firstName?: string | null;
  avatar?: string | null;
  isBot?: boolean;
  tgId: number;
  balance: number;
  giftAmount: number;
  giftTotal: number;
  subRechargeAmount?: number;
  subFlowAmount?: number;
  subWithdrawAmount?: number;
  rebateAmount: number;
  rebateTotalAmount: number;
  status: number;
  parentId?: number | null;
  inviteCode?: string | null;
  tenantId?: number;
};

export type TgUserSearch = {
  currentPage: number;
  pageSize: number;
  tgId?: number;
  username?: string;
  firstName?: string;
  isBot?: boolean;
  status?: number;
  parentId?: number;
  inviteCode?: string;
};

export type TgUserStatusSet = {
  id: number;
  status: number;
};

export type TgUserSubStatsSummarySearch = {
  tenantId?: number;
  parentId?: number;
  tgId?: number;
  isBot?: boolean;
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
  return http.request<TgUserListResult>("post", "/api/v1/admin/tgUser/list", {
    data
  });
};

export const getTgUserListWithSubStats = (data: TgUserSearch) => {
  return http.request<TgUserListResult>("post", "/api/v1/admin/tgUser/listWithSubStats", {
    data
  });
};

export const getTgUserSubStatsSummary = (data: TgUserSubStatsSummarySearch) => {
  return http.request<TgUserSubStatsSummaryResult>("post", "/api/v1/admin/tgUser/subStatsSummary", {
    data
  });
};

export const setTgUserStatus = (data: TgUserStatusSet) => {
  return http.request<TgUserResult>("post", "/api/v1/admin/tgUser/status", {
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

export const delAdminTgUser = (id: number) => {
  return http.request<TgUserResult>("delete", `/api/v1/admin/tgUser/${id}`);
};
