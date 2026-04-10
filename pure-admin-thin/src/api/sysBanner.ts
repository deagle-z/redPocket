import { http } from "@/utils/http";

export type SysBannerI18n = {
  id: number;
  createdAt?: string;
  updatedAt?: string;
  tenantId: number;
  bannerId: number;
  languageCode: string;
  countryCode?: string | null;
  title?: string | null;
  subTitle?: string | null;
  description?: string | null;
  buttonText?: string | null;
  imageUrl: string;
  thumbUrl?: string | null;
  bgImageUrl?: string | null;
  iconUrl?: string | null;
  videoUrl?: string | null;
  jumpValue?: string | null;
  textColor?: string | null;
  buttonColor?: string | null;
  bgColor?: string | null;
  status: number;
  isDeleted?: number;
  remark?: string | null;
  createdBy?: number | null;
  updatedBy?: number | null;
};

export type SysBannerCountryRel = {
  id: number;
  createdAt?: string;
  updatedAt?: string;
  tenantId: number;
  bannerId: number;
  countryCode: string;
  status: number;
  remark?: string | null;
  createdBy?: number | null;
  updatedBy?: number | null;
};

export type SysBanner = {
  id: number;
  createdAt: string;
  updatedAt: string;
  tenantId: number;
  bannerName: string;
  bannerCode?: string | null;
  position: string;
  platform: string;
  bannerType: string;
  jumpType: string;
  displayType: string;
  openMode: string;
  sort: number;
  status: number;
  isDeleted: number;
  startTime?: string | null;
  endTime?: string | null;
  clickCount: number;
  showCount: number;
  version: number;
  remark?: string | null;
  createdBy?: number | null;
  updatedBy?: number | null;
  i18nList: SysBannerI18n[];
  countryList: SysBannerCountryRel[];
};

export type SysBannerSearch = {
  currentPage: number;
  pageSize: number;
  tenantId?: number;
  bannerName?: string;
  bannerCode?: string;
  position?: string;
  platform?: string;
  bannerType?: string;
  jumpType?: string;
  displayType?: string;
  languageCode?: string;
  countryCode?: string;
  status?: number | null;
};

export type SysBannerSet = {
  id?: number;
  tenantId: number;
  bannerName: string;
  bannerCode?: string | null;
  position: string;
  platform: string;
  bannerType: string;
  jumpType: string;
  displayType: string;
  openMode: string;
  sort: number;
  status: number;
  startTime?: number | null;
  endTime?: number | null;
  version?: number;
  remark?: string | null;
  i18nList: Array<{
    id?: number;
    tenantId?: number;
    bannerId?: number;
    languageCode: string;
    countryCode?: string | null;
    title?: string | null;
    subTitle?: string | null;
    description?: string | null;
    buttonText?: string | null;
    imageUrl: string;
    thumbUrl?: string | null;
    bgImageUrl?: string | null;
    iconUrl?: string | null;
    videoUrl?: string | null;
    jumpValue?: string | null;
    textColor?: string | null;
    buttonColor?: string | null;
    bgColor?: string | null;
    status: number;
    remark?: string | null;
  }>;
  countryList: Array<{
    id?: number;
    tenantId?: number;
    bannerId?: number;
    countryCode: string;
    status: number;
    remark?: string | null;
  }>;
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
  return http.request<SysBannerListResult>(
    "post",
    "/api/v1/admin/sysBanner/list",
    {
      data
    }
  );
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
  return http.request<SysBannerResult>(
    "delete",
    `/api/v1/admin/sysBanner/${id}`
  );
};
