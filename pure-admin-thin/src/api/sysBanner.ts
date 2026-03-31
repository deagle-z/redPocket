import { http } from "@/utils/http";

export type SysBanner = {
  id: number;
  createdAt: string;
  updatedAt: string;
  tenantId: number;
  bannerName: string;
  position: string;
  platform: string;
  imageUrl: string;
  thumbUrl?: string | null;
  jumpType: string;
  jumpValue?: string | null;
  sort: number;
  status: number;
  startTime?: string | null;
  endTime?: string | null;
  clickCount: number;
  showCount: number;
  remark?: string | null;
};

export type SysBannerSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  bannerName?: string;
  position?: string;
  platform?: string;
  jumpType?: string;
  status?: number | null;
};

export type SysBannerSet = {
  id?: number;
  tenantId: number;
  bannerName: string;
  position: string;
  platform: string;
  imageUrl: string;
  thumbUrl?: string | null;
  jumpType: string;
  jumpValue?: string | null;
  sort: number;
  status: number;
  startTime?: number | null;
  endTime?: number | null;
  remark?: string | null;
};

type SysBannerListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: SysBanner[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

type SysBannerResult = {
  code: number;
  message: string;
  success: boolean;
  data: SysBanner;
};

export const getSysBannerList = (data: SysBannerSearch) => {
  return http.request<SysBannerListResult>("post", "/api/v1/admin/sysBanner/list", {
    data
  });
};

export const getSysBannerById = (id: number) => {
  return http.request<SysBannerResult>("get", `/api/v1/admin/sysBanner/${id}`);
};

export const setSysBanner = (data: SysBannerSet) => {
  return http.request<SysBannerResult>("post", "/api/v1/admin/sysBanner", {
    data
  });
};

export const delSysBanner = (id: number) => {
  return http.request<SysBannerResult>("delete", `/api/v1/admin/sysBanner/${id}`);
};
