export interface FormItemProps {
  title: string;
  id: number;
  code: string;
  amount: number;
  maxRedeemCount: number;
  generateCount: number;
  tag: string;
  exportAfterCreate: boolean;
  status: 1 | 0;
  remark: string;
}

export interface FormProps {
  formInline: FormItemProps;
}
