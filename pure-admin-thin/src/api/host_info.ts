import { http } from "@/utils/http";
import type { PageResult, Result } from "@/api/system";

export const getHostInfos = (data?: object) => {
  return http.request<PageResult>("post", "/api/v1/admin/host_infos", {
    data
  });
};

export const delHostInfo = (hostInfoId?: number) => {
  return http.request<Result>(
    "delete",
    "/api/v1/admin/host_info/" + hostInfoId
  );
};

export const setHostInfo = (data?: object) => {
  return http.request<Result>("put", "/api/v1/admin/host_info", { data });
};
