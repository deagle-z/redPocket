interface FormItemProps {
  title: string;
  id: number;
  tenantId: number;
  channelCode: string;
  channelName: string;
  parentId: number | null;
  level: number;
  status: number;
  sort: number;
  remark: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
