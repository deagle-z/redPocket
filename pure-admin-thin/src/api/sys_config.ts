import { http } from "@/utils/http";
import type { PageResult, Result } from "@/api/system";

/** 获取系统配置列表 */
export const getSysConfigs = (data?: object) => {
  return http.request<PageResult>("post", "/api/v1/admin/sysConfig/list", {
    data
  });
};

/** 创建或更新系统配置 */
export const setSysConfig = (data?: object) => {
  return http.request<Result>("post", "/api/v1/admin/sysConfig", { data });
};

/** 删除系统配置 */
export const delSysConfig = (id: number) => {
  return http.request<Result>("delete", `/api/v1/admin/sysConfig/${id}`);
};
