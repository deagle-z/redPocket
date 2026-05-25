interface FormItemProps {
  title: string;
  gameId: number;
  gameName: string;
  categoryCode: string;
  showIndex: number;
  hot: number;
  homeShow: number;
  type: number | null;
  parentId: number | null;
  platformCode: string;
  thirdGameId: string;
  thirdGameName: string;
  thirdGameCategory: string;
  horizontalImage: string;
  gameIcon: string;
  typeIcon: string;
  typeActiveIcon: string;
  tenantId: number | null;
  sort: number;
  disabledFlag: number;
  deletedFlag: number;
  remark: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
