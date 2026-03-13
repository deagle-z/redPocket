import { http } from "@/utils/http";

export type SysCustomField = {
  id: number;
  createdAt: string;
  updatedAt: string;
  fieldKey: string;
  fieldLabel: string;
  fieldPlaceholder?: string | null;
  fieldType: string;
  dataType: string;
  defaultValue?: string | null;
  isRequired: number;
  isSensitive: number;
  maxLength?: number | null;
  minLength?: number | null;
  regexRule?: string | null;
  errorTips?: string | null;
  optionsJson?: string | null;
  status: number;
  remark?: string | null;
};

export type SysCustomFieldSearch = {
  currentPage: number;
  pageSize: number;
  fieldKey?: string;
  fieldLabel?: string;
  fieldType?: string;
  dataType?: string;
  status?: number;
};

export type SysCustomFieldSet = {
  id?: number;
  fieldKey: string;
  fieldLabel: string;
  fieldPlaceholder?: string | null;
  fieldType: string;
  dataType: string;
  defaultValue?: string | null;
  isRequired: number;
  isSensitive: number;
  maxLength?: number | null;
  minLength?: number | null;
  regexRule?: string | null;
  errorTips?: string | null;
  optionsJson?: string | null;
  status: number;
  remark?: string | null;
};

type SysCustomFieldListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: SysCustomField[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

type SysCustomFieldResult = {
  code: number;
  message: string;
  success: boolean;
  data: SysCustomField;
};

export const getSysCustomFieldList = (data: SysCustomFieldSearch) => {
  return http.request<SysCustomFieldListResult>(
    "post",
    "/api/v1/admin/sysCustomField/list",
    {
      data
    }
  );
};

export const getSysCustomFieldById = (id: number) => {
  return http.request<SysCustomFieldResult>(
    "get",
    `/api/v1/admin/sysCustomField/${id}`
  );
};

export const setSysCustomField = (data: SysCustomFieldSet) => {
  return http.request<SysCustomFieldResult>("post", "/api/v1/admin/sysCustomField", {
    data
  });
};

export const delSysCustomField = (id: number) => {
  return http.request<SysCustomFieldResult>(
    "delete",
    `/api/v1/admin/sysCustomField/${id}`
  );
};
