import { http } from "@/utils/http";
import type { TaskActivityLevelCode } from "@/api/taskActivityConfig";

export type TaskActivityRecordStatus = 0 | 1 | 2 | 3 | 4;

export type TaskActivityRecord = {
  id: number;
  createdAt: string;
  updatedAt: string;
  configId: number;
  userId: number;
  uid: string;
  title: string;
  subTitle: string;
  levelCode: TaskActivityLevelCode;
  levelName: string;
  taskType: string;
  requiredInviteCount: number;
  requiredRechargeAmount: number;
  durationMinutes: number;
  rewardAmount: number;
  rewardCurrency: string;
  rewardTarget: string;
  progressInviteCount: number;
  progressRechargeCount: number;
  progressRechargeAmount: number;
  status: TaskActivityRecordStatus;
  claimedAt: string;
  deadlineAt: string;
  completedAt?: string | null;
  rewardedAt?: string | null;
  rebateRecordId?: number | null;
  remark: string;
};

export type TaskActivityRecordSearch = {
  currentPage: number;
  pageSize: number;
  userId?: number;
  uid?: string;
  configId?: number;
  title?: string;
  levelCode?: TaskActivityLevelCode | "";
  status?: number | null;
};

type TaskActivityRecordListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: TaskActivityRecord[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

type TaskActivityRecordResult = {
  code: number;
  message: string;
  success: boolean;
  data: TaskActivityRecord;
};

export const getTaskActivityRecordList = (data: TaskActivityRecordSearch) => {
  return http.request<TaskActivityRecordListResult>(
    "post",
    "/api/v1/admin/taskActivityRecord/list",
    { data }
  );
};

export const getTaskActivityRecordById = (id: number) => {
  return http.request<TaskActivityRecordResult>(
    "get",
    `/api/v1/admin/taskActivityRecord/${id}`
  );
};
