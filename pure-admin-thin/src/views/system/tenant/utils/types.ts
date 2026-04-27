interface FormItemProps {
  title: string;
  id: number;
  tenantCode: string;
  tenantName: string;
  tenantType: number;
  status: number;
  loginPassword?: string;
  ownerUserId?: number;
  planCode?: string;
  bindDomain?: string;
  timezone: string;
  locale: string;
  remark?: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
