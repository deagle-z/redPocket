import { http } from "@/utils/http";

export type AttributionEvent = {
  id: number;
  createdAt: string;
  updatedAt: string;
  tenantId: number;
  userId?: number | null;
  visitorId: string;
  sessionId: string;
  eventName: string;
  thirdPartyEventId?: string | null;
  pixelId?: string | null;
  sourceChannelId?: number | null;
  sourceChannelCode?: string | null;
  pageUrl?: string | null;
  referrer?: string | null;
  ip?: string | null;
  userAgent?: string | null;
  metadata?: string | null;
};

export type AttributionEventSearch = {
  currentPage: number;
  pageSize: number;
  eventName?: string;
  pixelId?: string;
  startTime?: number;
  endTime?: number;
};

export type AttributionEventListResult = {
  code: number;
  message: string;
  success: boolean;
  data: {
    list: AttributionEvent[];
    total: number;
    pageSize: number;
    currentPage: number;
  };
};

export const getAttributionEventListAdmin = (data: AttributionEventSearch) => {
  return http.request<AttributionEventListResult>(
    "post",
    "/api/v1/admin/attributionEvent/list",
    { data }
  );
};
