import { http } from "@/utils/http";

export type SysCountry = {
  id: number;
  createdAt: string;
  updatedAt: string;
  countryCode: string;
  countryNameCn: string;
  countryNameEn: string;
  currencyCode: string;
  currencySymbol?: string | null;
  timezone?: string | null;
  languageCode?: string | null;
  withdrawFields?: string | null;
  rechargeFields?: string | null;
  sort: number;
  status: number;
  remark?: string | null;
};

export type SysCountrySearch = {
  currentPage: number;
  pageSize: number;
  countryCode?: string;
  countryName?: string;
  currencyCode?: string;
  status?: number;
};

export type SysCountrySet = {
  id?: number;
  countryCode: string;
  countryNameCn: string;
  countryNameEn: string;
  currencyCode: string;
  currencySymbol?: string | null;
  timezone?: string | null;
  languageCode?: string | null;
  withdrawFields?: string | null;
  rechargeFields?: string | null;
  sort: number;
  status: number;
  remark?: string | null;
};

type SysCountryListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: SysCountry[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

type SysCountryResult = {
  code: number;
  message: string;
  success: boolean;
  data: SysCountry;
};

export const getSysCountryList = (data: SysCountrySearch) => {
  return http.request<SysCountryListResult>("post", "/api/v1/admin/sysCountry/list", {
    data
  });
};

export const getSysCountryById = (id: number) => {
  return http.request<SysCountryResult>("get", `/api/v1/admin/sysCountry/${id}`);
};

export const setSysCountry = (data: SysCountrySet) => {
  return http.request<SysCountryResult>("post", "/api/v1/admin/sysCountry", {
    data
  });
};

export const delSysCountry = (id: number) => {
  return http.request<SysCountryResult>("delete", `/api/v1/admin/sysCountry/${id}`);
};
