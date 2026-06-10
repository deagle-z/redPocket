export interface FormItemProps {
  title: string;
  id: number;
  code: string;
  amount: number;
  maxRedeemCount: number;
  generateCount: number;
  status: 1 | 0;
  remark: string;
}

export interface FormProps {
  formInline: FormItemProps;
}
