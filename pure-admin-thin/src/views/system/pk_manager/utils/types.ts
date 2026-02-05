interface FormItemProps {
  title: string;
  id: number;
  apkPackage: string;
  name: string;
  url: string;
  isActive: number;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };


