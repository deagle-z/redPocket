import { http } from "@/utils/http";

/** 余额变动记录信息 */
export type CashHistory = {
  createdAt: string;
  userId: number;
  amount: number;
  startAmount: number;
  endAmount: number;
  cashMark: string;
  cashDesc: string;
  fromUserId: number;
};

/** 余额变动记录搜索参数 */
export type CashHistorySearch = {
  currentPage: number;
  pageSize: number;
  userId?: number;
  cashMark?: string;
};

/** 余额变动记录列表响应 */
export type CashHistoryListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: CashHistory[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

/** 获取余额变动记录列表（管理端） */
export const getCashHistoryListAdmin = (data: CashHistorySearch) => {
  return http.request<CashHistoryListResult>("post", "/api/v1/admin/cashHistory/list", { data });
};
