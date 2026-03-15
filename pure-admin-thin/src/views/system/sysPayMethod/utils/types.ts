interface FormItemProps {
  title: string;
  id: number;
  methodCode: string;
  methodName: string;
  icon: string;
  sort: number;
  status: number;
  remark: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
