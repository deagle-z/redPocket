interface FormItemProps {
  title: string;
  id: number;
  channelCode: string;
  channelName: string;
  channelType: string;
  providerType: string;
  countryCode: string;
  icon: string;
  remark: string;
  sort: number;
  status: number;
  methodIds: number[];
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
