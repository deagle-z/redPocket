interface FormItemProps {
  title: string;
  id: number;
  tenantCode: string;
  tenantName: string;
  tenantType: number;
  status: number;
  enableWithdraw: number;
  withdrawAutoReviewEnabled: number;
  withdrawAutoReviewMaxAmount: number;
  withdrawAutoReviewDailyLimit: number;
  loginPassword?: string;
  ownerUserId?: number | null;
  planCode?: string;
  bindDomain?: string;
  tgServiceUrl?: string;
  wsServiceUrl?: string;
  timezone: string;
  locale: string;
  remark?: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
