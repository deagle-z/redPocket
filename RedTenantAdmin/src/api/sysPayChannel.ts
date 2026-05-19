import { http } from "@/utils/http";

export type SysPayChannel = {
  id: number;
  createdAt: string;
  updatedAt: string;
  channelCode: string;
  channelName: string;
  channelType: string;
  providerType: string;
  countryCode?: string | null;
  icon?: string | null;
  remark?: string | null;
  sort: number;
  status: number;
};

export type SysPayChannelSearch = {
  currentPage: number;
  pageSize: number;
  channelCode?: string;
  channelName?: string;
  channelType?: string;
  providerType?: string;
  countryCode?: string;
  status?: number | null;
};

type SysPayChannelListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: SysPayChannel[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export const getSysPayChannelList = (data: SysPayChannelSearch) => {
  return http.request<SysPayChannelListResult>(
    "post",
    "/api/v1/tenant/sysPayChannel/list",
    { data }
  );
};
