interface FormItemProps {
  title: string;
  id: number;
  tenantId: number;
  bannerName: string;
  position: string;
  platform: string;
  imageUrl: string;
  thumbUrl: string;
  jumpType: string;
  jumpValue: string;
  sort: number;
  status: number;
  startTime: string;
  endTime: string;
  remark: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
