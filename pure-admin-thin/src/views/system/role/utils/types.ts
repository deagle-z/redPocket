interface FormItemProps {
  id: number;
  name: string;
  code: string;
  description: string;
  menuIds: number[];
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
