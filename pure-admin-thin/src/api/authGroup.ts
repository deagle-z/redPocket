import { http } from "@/utils/http";

/** 授权群组信息 */
export type AuthGroup = {
  id: number;
  createdAt: string;
  updatedAt: string;
  groupId: number;
  groupName: string;
  status: number;
  serviceUrl: string;
  rechargeUrl: string;
  channelUrl: string;
  sendPacketImage?: string;
  loseRate?: number;
  numConfig?: string;
  sendCommission?: number;
  grabbingCommission?: number;
  deleteMsg?: number;
  whiteIds?: string;
};

/** 授权群组搜索参数 */
export type AuthGroupSearch = {
  currentPage: number;
  pageSize: number;
  groupId?: number;
  status?: number;
};

/** 授权群组列表响应 */
export type AuthGroupListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: AuthGroup[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

/** 授权群组响应 */
export type AuthGroupResult = {
  code: number;
  message: string;
  success: boolean;
  data: AuthGroup;
};

/** 获取授权群组列表 */
export const getAuthGroups = (data: AuthGroupSearch) => {
  return http.request<AuthGroupListResult>("post", "/api/v1/admin/authGroup/list", { data });
};

/** 创建或更新授权群组 */
export const setAuthGroup = (data: AuthGroup) => {
  return http.request<AuthGroupResult>("post", "/api/v1/admin/authGroup", { data });
};

/** 删除授权群组 */
export const delAuthGroup = (id: number) => {
  return http.request<any>("delete", `/api/v1/admin/authGroup/${id}`);
};
