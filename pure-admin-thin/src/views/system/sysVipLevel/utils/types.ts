interface FormItemProps {
  title: string;
  id: number;
  tenantId: number;
  level: number;
  levelName: string;
  agentTag: string;
  totalRechargeCount: number | null;
  totalRechargeAmount: number | null;
  totalValidBet: number | null;
  monthRechargeAmount: number | null;
  monthValidBet: number | null;
  upgradeBonusAmount: number;
  upgradeType: number;
  keepLevelCondition: number;
  sort: number;
  status: number;
  remark: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
