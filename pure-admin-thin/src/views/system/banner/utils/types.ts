interface FormItemProps {
  title: string;
  id: number;
  tenantId: number;
  bannerName: string;
  bannerCode: string;
  position: string;
  platform: string;
  bannerType: string;
  jumpType: string;
  displayType: string;
  openMode: string;
  sort: number;
  status: number;
  startTime: string;
  endTime: string;
  version: number;
  remark: string;
  i18nList: BannerI18nFormItem[];
  countryList: BannerCountryFormItem[];
}

interface BannerI18nFormItem {
  id: number;
  tenantId: number;
  bannerId: number;
  languageCode: string;
  countryCode: string;
  title: string;
  subTitle: string;
  description: string;
  buttonText: string;
  imageUrl: string;
  thumbUrl: string;
  bgImageUrl: string;
  iconUrl: string;
  videoUrl: string;
  jumpValue: string;
  textColor: string;
  buttonColor: string;
  bgColor: string;
  status: number;
  remark: string;
}

interface BannerCountryFormItem {
  id: number;
  tenantId: number;
  bannerId: number;
  countryCode: string;
  status: number;
  remark: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type {
  BannerCountryFormItem,
  BannerI18nFormItem,
  FormItemProps,
  FormProps
};
