interface FormItemProps {
  title: string;
  id: number;
  tenantCode: string;
  tenantName: string;
  tenantType: number;
  status: number;
  ownerUserId?: number;
  planCode?: string;
  timezone: string;
  locale: string;
  remark?: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
