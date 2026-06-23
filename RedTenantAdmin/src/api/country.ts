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
  rate: number;
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

export type AppPayMethodItem = {
  id: number;
  methodCode: string;
  methodName: string;
  icon?: string | null;
  sort: number;
};

export type AppRechargeChannelItem = {
  id: number;
  channelCode: string;
  channelName: string;
  providerType: string;
  icon?: string | null;
  sort: number;
  methods: AppPayMethodItem[];
};

export type AppCountryRechargeInfo = {
  rechargeFields: unknown[];
  channels: AppRechargeChannelItem[];
  minAmount?: number;
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

export const getSysCountryList = (data: SysCountrySearch) => {
  return http.request<SysCountryListResult>(
    "post",
    "/api/v1/tenant/sysCountry/list",
    {
      data
    }
  );
};

export const getTenantCountryRechargeInfo = (code: string) => {
  return http.request<{
    code: number;
    message: string;
    success: boolean;
    data: AppCountryRechargeInfo;
  }>("get", `/api/v1/tenant/sysCountryRecharge/${code}`);
};
