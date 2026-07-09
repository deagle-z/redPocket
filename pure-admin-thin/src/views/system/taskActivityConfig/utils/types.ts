import type {
  TaskActivityLevelCode,
  TaskActivityRewardTarget,
  TaskActivityType
} from "@/api/taskActivityConfig";

export interface FormItemProps {
  title: string;
  id: number;
  configTitle: string;
  subTitle: string;
  levelCode: TaskActivityLevelCode;
  levelName: string;
  sort: number;
  status: number;
  taskType: TaskActivityType;
  requiredInviteCount: number;
  requiredRechargeAmount: number;
  durationMinutes: number;
  rewardAmount: number;
  rewardCurrency: string;
  rewardTarget: TaskActivityRewardTarget;
  startAt: string | null;
  endAt: string | null;
  remark: string;
}

export interface FormProps {
  formInline: FormItemProps;
}
