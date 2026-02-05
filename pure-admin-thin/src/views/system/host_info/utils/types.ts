interface FormItemProps {
  title: string;
  id: number;
  hostName: string;
  tablePrefix: string;
  hostMark: string;
  hostDesc: string;
  enabled: boolean;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
