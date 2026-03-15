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

export type SysPayChannelSet = {
  id?: number;
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

export type SysPayChannelMethod = {
  id: number;
  channelId: number;
  methodId: number;
  createdAt: string;
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

type SysPayChannelResult = {
  code: number;
  message: string;
  success: boolean;
  data: SysPayChannel;
};

type SysPayChannelMethodListResult = {
  code: number;
  message: string;
  success: boolean;
  data: SysPayChannelMethod[];
};

export const getSysPayChannelList = (data: SysPayChannelSearch) => {
  return http.request<SysPayChannelListResult>(
    "post",
    "/api/v1/admin/sysPayChannel/list",
    { data }
  );
};

export const getSysPayChannelById = (id: number) => {
  return http.request<SysPayChannelResult>(
    "get",
    `/api/v1/admin/sysPayChannel/${id}`
  );
};

export const setSysPayChannel = (data: SysPayChannelSet) => {
  return http.request<SysPayChannelResult>(
    "post",
    "/api/v1/admin/sysPayChannel",
    { data }
  );
};

export const delSysPayChannel = (id: number) => {
  return http.request<SysPayChannelResult>(
    "delete",
    `/api/v1/admin/sysPayChannel/${id}`
  );
};

export const getSysPayChannelMethods = (channelId: number) => {
  return http.request<SysPayChannelMethodListResult>(
    "get",
    `/api/v1/admin/sysPayChannelMethod/${channelId}`
  );
};

export const setSysPayChannelMethods = (data: {
  channelId: number;
  methodIds: number[];
}) => {
  return http.request("post", "/api/v1/admin/sysPayChannelMethod", { data });
};
