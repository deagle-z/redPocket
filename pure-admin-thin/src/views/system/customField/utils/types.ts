interface FormItemProps {
  title: string;
  id: number;
  fieldKey: string;
  fieldLabel: string;
  fieldPlaceholder: string;
  fieldType: string;
  dataType: string;
  defaultValue: string;
  isRequired: number;
  isSensitive: number;
  maxLength: number | null;
  minLength: number | null;
  regexRule: string;
  errorTips: string;
  optionsJson: string;
  status: number;
  remark: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
