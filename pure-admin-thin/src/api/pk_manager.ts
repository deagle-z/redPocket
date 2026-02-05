import { http } from "@/utils/http";
import type { PageResult, Result } from "@/api/system";

/** 获取PK管理器列表 */
export const getPkManagers = (data?: object) => {
  return http.request<PageResult>("post", "/api/v1/outside/pkManagers", {
    data
  });
};

/** 根据ID获取PK管理器 */
export const getPkManagerById = (id?: number) => {
  return http.request<Result>("get", "/api/v1/outside/pkManager/" + id);
};

/** 创建或更新PK管理器 */
export const setPkManager = (data?: object) => {
  return http.request<Result>("post", "/api/v1/manager/pkManager", { data });
};

/** 删除PK管理器 */
export const delPkManager = (id?: number) => {
  return http.request<Result>(
    "delete",
    "/api/v1/manager/pkManager/" + id
  );
};


