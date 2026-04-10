export interface FormItemProps {
  title: string;
  id: number;
  configKey: string;
  configValue: string;
  configDesc: string;
}

export interface FormProps {
  formInline: FormItemProps;
}
