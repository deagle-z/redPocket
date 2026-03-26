import { http } from "@/utils/http";

export type SysSourceChannel = {
  id: number;
  createdAt: string;
  updatedAt: string;
  tenantId: number;
  channelCode: string;
  channelName: string;
  parentId?: number | null;
  level: number;
  status: number;
  sort: number;
  remark?: string | null;
};

export type SysSourceChannelSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  channelCode?: string;
  channelName?: string;
  parentId?: number | null;
  level?: number | null;
  status?: number | null;
};

export type SysSourceChannelSet = {
  id?: number;
  tenantId: number;
  channelCode?: string;
  channelName: string;
  parentId?: number | null;
  level: number;
  status: number;
  sort: number;
  remark?: string | null;
};

type SysSourceChannelListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: SysSourceChannel[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

type SysSourceChannelResult = {
  code: number;
  message: string;
  success: boolean;
  data: SysSourceChannel;
};

export const getSysSourceChannelList = (data: SysSourceChannelSearch) => {
  return http.request<SysSourceChannelListResult>(
    "post",
    "/api/v1/admin/sysSourceChannel/list",
    { data }
  );
};

export const getSysSourceChannelById = (id: number) => {
  return http.request<SysSourceChannelResult>(
    "get",
    `/api/v1/admin/sysSourceChannel/${id}`
  );
};

export const setSysSourceChannel = (data: SysSourceChannelSet) => {
  return http.request<SysSourceChannelResult>(
    "post",
    "/api/v1/admin/sysSourceChannel",
    { data }
  );
};

export const delSysSourceChannel = (id: number) => {
  return http.request<SysSourceChannelResult>(
    "delete",
    `/api/v1/admin/sysSourceChannel/${id}`
  );
};
