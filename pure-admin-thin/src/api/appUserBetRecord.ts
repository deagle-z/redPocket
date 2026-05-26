import { http } from "@/utils/http";

export type AppUserBetRecord = {
  id: number;
  uid?: number | null;
  userId?: number | null;
  gameId?: string | null;
  gameName?: string | null;
  platformCode?: string | null;
  betAmount?: number | null;
  winAmount?: number | null;
  roundId?: string | null;
  traceId?: string | null;
  roundEnd?: number | null;
  date?: string | null;
  remark?: string | null;
  createTime?: string | null;
  updateTime?: string | null;
};

export type AppUserBetRecordSearch = {
  currentPage: number;
  pageSize: number;
  id?: number;
  uid?: number;
  userId?: number;
  gameId?: string;
  gameName?: string;
  platformCode?: string;
  roundId?: string;
  traceId?: string;
  roundEnd?: number;
  startTime?: number;
  endTime?: number;
};

export type AppUserBetRecordListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: AppUserBetRecord[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export const getAppUserBetRecordList = (data: AppUserBetRecordSearch) => {
  return http.request<AppUserBetRecordListResult>(
    "post",
    "/api/v1/admin/appUserBetRecord/list",
    { data }
  );
};
