import type { SysCustomField } from "@/api/customField";

interface FormItemProps {
  title: string;
  id: number;
  countryCode: string;
  countryNameCn: string;
  countryNameEn: string;
  currencyCode: string;
  currencySymbol: string;
  timezone: string;
  languageCode: string;
  withdrawFields: SysCustomField[];
  rechargeFields: SysCustomField[];
  sort: number;
  status: number;
  remark: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
