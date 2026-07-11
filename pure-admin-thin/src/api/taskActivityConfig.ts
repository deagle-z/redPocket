import { http } from "@/utils/http";

export type TaskActivityLevelCode = "primary" | "middle" | "advanced" | "custom";
export type TaskActivityType = "invite_recharge";
export type TaskActivityRewardTarget = "rebate";

export type TaskActivityConfig = {
  id: number;
  createdAt: string;
  updatedAt: string;
  title: string;
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
  startAt?: string | null;
  endAt?: string | null;
  remark: string;
};

export type TaskActivityConfigSearch = {
  currentPage: number;
  pageSize: number;
  title?: string;
  levelCode?: TaskActivityLevelCode | "";
  status?: number | null;
};

export type TaskActivityConfigSet = {
  id?: number;
  title: string;
  subTitle?: string;
  levelCode: TaskActivityLevelCode;
  levelName?: string;
  sort: number;
  status: number;
  taskType?: TaskActivityType;
  requiredInviteCount: number;
  requiredRechargeAmount: number;
  durationMinutes: number;
  rewardAmount: number;
  rewardCurrency?: string;
  rewardTarget?: TaskActivityRewardTarget;
  startAt?: string | null;
  endAt?: string | null;
  remark?: string;
};

type TaskActivityConfigListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: TaskActivityConfig[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

type TaskActivityConfigResult = {
  code: number;
  message: string;
  success: boolean;
  data: TaskActivityConfig;
};

export const getTaskActivityConfigList = (data: TaskActivityConfigSearch) => {
  return http.request<TaskActivityConfigListResult>(
    "post",
    "/api/v1/admin/taskActivityConfig/list",
    { data }
  );
};

export const getTaskActivityConfigById = (id: number) => {
  return http.request<TaskActivityConfigResult>(
    "get",
    `/api/v1/admin/taskActivityConfig/${id}`
  );
};

export const setTaskActivityConfig = (data: TaskActivityConfigSet) => {
  return http.request<TaskActivityConfigResult>(
    "post",
    "/api/v1/admin/taskActivityConfig",
    { data }
  );
};

export const delTaskActivityConfig = (id: number) => {
  return http.request<TaskActivityConfigResult>(
    "delete",
    `/api/v1/admin/taskActivityConfig/${id}`
  );
};
