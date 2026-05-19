import { http } from "@/utils/http";

export type CheckInRecord = {
  id: number;
  userId: number;
  userUid: string;
  checkInDate: string;
  checkInSeq: number;
  rewardAmount: number;
  beforeBalance: number;
  afterBalance: number;
  createdAt: string;
};

export type CheckInRecordSearch = {
  currentPage: number;
  pageSize: number;
  userId?: number;
  userUid?: string;
  startDate?: string;
  endDate?: string;
};

type CheckInRecordListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: CheckInRecord[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export const getCheckInRecordsAdmin = (data: CheckInRecordSearch) =>
  http.request<CheckInRecordListResult>(
    "post",
    "/api/v1/admin/checkInRecord/list",
    { data }
  );
