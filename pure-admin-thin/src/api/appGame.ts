import { http } from "@/utils/http";

export type AppGame = {
  gameId: number;
  gameName?: string | null;
  categoryCode?: string | null;
  showIndex?: number | null;
  hot?: number | null;
  homeShow?: number | null;
  type?: number | null;
  parentId?: number | null;
  platformCode?: string | null;
  thirdGameId?: string | null;
  thirdGameName?: string | null;
  thirdGameCategory?: string | null;
  horizontalImage?: string | null;
  gameIcon?: string | null;
  typeIcon?: string | null;
  typeActiveIcon?: string | null;
  tenantId?: number | null;
  sort?: number | null;
  disabledFlag?: number | null;
  deletedFlag?: number | null;
  remark?: string | null;
  updateTime?: string | null;
  createTime?: string | null;
};

export type AppGameSearch = {
  currentPage: number;
  pageSize: number;
  gameId?: number;
  gameName?: string;
  categoryCode?: string;
  type?: number;
  parentId?: number;
  platformCode?: string;
  thirdGameId?: string;
  thirdGameName?: string;
  thirdGameCategory?: string;
  tenantId?: number;
  hot?: number;
  homeShow?: number;
  disabledFlag?: number;
  deletedFlag?: number;
};

export type AppGameSyncReq = {
  language?: string;
  platformCode?: string;
};

export type AppGameThirdCategorySearch = {
  categoryCode?: string;
};

export type AppGameSet = {
  gameId?: number;
  gameName?: string | null;
  categoryCode?: string | null;
  showIndex?: number | null;
  hot?: number | null;
  homeShow?: number | null;
  type?: number | null;
  parentId?: number | null;
  platformCode?: string | null;
  thirdGameId?: string | null;
  thirdGameName?: string | null;
  thirdGameCategory?: string | null;
  horizontalImage?: string | null;
  gameIcon?: string | null;
  typeIcon?: string | null;
  typeActiveIcon?: string | null;
  tenantId?: number | null;
  sort?: number | null;
  disabledFlag?: number | null;
  deletedFlag?: number | null;
  remark?: string | null;
};

export type AppGameListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: AppGame[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export type AppGameSyncResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    total: number;
    created: number;
    updated: number;
    skipped: number;
  };
};

type AppGameResult = {
  code: number;
  message: string;
  success: boolean;
  data: AppGame;
};

type AppGameThirdCategoryResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: string[];
  };
};

export const getAppGameList = (data: AppGameSearch) => {
  return http.request<AppGameListResult>("post", "/api/v1/admin/appGame/list", {
    data
  });
};

export const getAppGameThirdCategories = (data: AppGameThirdCategorySearch) => {
  return http.request<AppGameThirdCategoryResult>(
    "post",
    "/api/v1/admin/appGame/thirdCategories",
    { data }
  );
};

export const syncAppGames = (data: AppGameSyncReq) => {
  return http.request<AppGameSyncResult>("post", "/api/v1/admin/appGame/sync", {
    data
  });
};

export const setAppGame = (data: AppGameSet) => {
  return http.request<AppGameResult>("post", "/api/v1/admin/appGame", {
    data
  });
};
