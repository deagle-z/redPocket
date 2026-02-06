import type { AuthGroup } from "@/api/authGroup";

interface FormItemProps {
  id: number;
  groupId: number;
  groupName: string;
  status: number;
  serviceUrl: string;
  rechargeUrl: string;
  channelUrl: string;
  sendPacketImage?: string;
  loseRate?: number;
  numConfig?: string;
  sendCommission?: number;
  grabbingCommission?: number;
  deleteMsg?: number;
  whiteIds?: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
export type { AuthGroup };
