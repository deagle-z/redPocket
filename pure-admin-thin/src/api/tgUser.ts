import { http } from "@/utils/http";

export type TgUser = {
  id: number;
  createdAt: string;
  updatedAt: string;
  username?: string | null;
  firstName?: string | null;
  avatar?: string | null;
  tgId: number;
  balance: number;
  giftAmount: number;
  giftTotal: number;
  rebateAmount: number;
  rebateTotalAmount: number;
  status: number;
  parentId?: number | null;
  inviteCode?: string | null;
};

export type TgUserSearch = {
  currentPage: number;
  pageSize: number;
  tgId?: number;
  username?: string;
  firstName?: string;
  status?: number;
  parentId?: number;
  inviteCode?: string;
};

export type TgUserStatusSet = {
  id: number;
  status: number;
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

export const getTgUserList = (data: TgUserSearch) => {
  return http.request<TgUserListResult>("post", "/api/v1/admin/tgUser/list", {
    data
  });
};

export const setTgUserStatus = (data: TgUserStatusSet) => {
  return http.request<TgUserResult>("post", "/api/v1/admin/tgUser/status", {
    data
  });
};
