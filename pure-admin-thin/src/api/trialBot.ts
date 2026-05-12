import { http } from "@/utils/http";

export type TrialBotUser = {
  id: number;
  createdAt: string;
  updatedAt: string;
  username?: string | null;
  firstName?: string | null;
  avatar?: string | null;
  balance: number;
  status: number;
  tenantId?: number;
};

export type TrialBotUserSearch = {
  currentPage: number;
  pageSize: number;
  username?: string;
  firstName?: string;
  status?: number;
};

export type TrialBotBatchCreateReq = {
  num: number;
  randomName: boolean;
  nameFile: string;
  avatarLinks: string[];
  balance?: number;
};

export type TrialBotBatchUpdateReq = {
  ids: number[];
  randomName: boolean;
  nameFile: string;
  avatarLinks: string[];
  status?: number;
  balance?: number;
};

export type TrialBotResult<T> = {
  code: number;
  message: string;
  success: boolean;
  data: T;
};

export type TrialBotListResult = TrialBotResult<{
  list: TrialBotUser[];
  total: number;
  pageSize: number;
  currentPage: number;
}>;

export type TrialBotBatchResult = TrialBotResult<{
  count: number;
  list: TrialBotUser[];
}>;

export const getTrialBotList = (data: TrialBotUserSearch) => {
  return http.request<TrialBotListResult>("post", "/api/v1/admin/trialBot/list", {
    data
  });
};

export const batchCreateTrialBots = (data: TrialBotBatchCreateReq) => {
  return http.request<TrialBotBatchResult>(
    "post",
    "/api/v1/admin/trialBot/batchCreate",
    { data }
  );
};

export const batchUpdateTrialBots = (data: TrialBotBatchUpdateReq) => {
  return http.request<TrialBotBatchResult>(
    "post",
    "/api/v1/admin/trialBot/batchUpdate",
    { data }
  );
};

export const setTrialBotStatus = (data: { id: number; status: number }) => {
  return http.request<TrialBotResult<TrialBotUser>>(
    "post",
    "/api/v1/admin/trialBot/status",
    { data }
  );
};

export const delTrialBot = (id: number) => {
  return http.request<TrialBotResult<string>>(
    "delete",
    `/api/v1/admin/trialBot/${id}`
  );
};
