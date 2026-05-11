import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import {
  getAttributionEventListAdmin,
  type AttributionEvent,
  type AttributionEventSearch
} from "@/api/attributionEvent";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import { ElTag } from "element-plus";

function formatNullable(val?: string | number | null) {
  if (val === undefined || val === null || val === "") return "-";
  return String(val);
}

function normalizeTimeRange(timeRange: Date[] | string[] | []) {
  if (!Array.isArray(timeRange) || timeRange.length !== 2) return {};
  const [start, end] = timeRange;
  const startTime = dayjs(start).unix();
  const endTime = dayjs(end).unix() + 1;
  if (Number.isNaN(startTime) || Number.isNaN(endTime)) return {};
  return { startTime, endTime };
}

function buildSearchParams(
  form: {
    eventName: string;
    pixelId: string;
    timeRange: Date[] | string[] | [];
  },
  pagination: PaginationProps
): AttributionEventSearch {
  const params: AttributionEventSearch = {
    currentPage: pagination.currentPage,
    pageSize: pagination.pageSize,
    ...normalizeTimeRange(form.timeRange)
  };
  const eventName = form.eventName.trim();
  const pixelId = form.pixelId.trim();
  if (eventName) params.eventName = eventName;
  if (pixelId) params.pixelId = pixelId;
  return params;
}

export function useAttributionEvent(tableRef: Ref) {
  const form = reactive({
    eventName: "",
    pixelId: "",
    timeRange: [] as Date[] | string[] | []
  });
  const dataList = ref<AttributionEvent[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });
  const columns: TableColumnList = [
    {
      label: "时间",
      prop: "createdAt",
      minWidth: 170,
      formatter: ({ createdAt }) =>
        createdAt ? dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss") : "-"
    },
    {
      label: "事件名称",
      prop: "eventName",
      minWidth: 140,
      cellRenderer: scope => (
        <ElTag type="primary" effect="plain">
          {formatNullable(scope.row.eventName)}
        </ElTag>
      )
    },
    {
      label: "Pixel ID",
      prop: "pixelId",
      minWidth: 170,
      showOverflowTooltip: true,
      formatter: ({ pixelId }) => formatNullable(pixelId)
    },
    {
      label: "三方事件ID",
      prop: "thirdPartyEventId",
      minWidth: 180,
      showOverflowTooltip: true,
      formatter: ({ thirdPartyEventId }) => formatNullable(thirdPartyEventId)
    },
    {
      label: "渠道",
      prop: "sourceChannelCode",
      minWidth: 130,
      formatter: ({ sourceChannelCode }) => formatNullable(sourceChannelCode)
    },
    {
      label: "用户ID",
      prop: "userId",
      minWidth: 110,
      formatter: ({ userId }) => formatNullable(userId)
    },
    {
      label: "访客ID",
      prop: "visitorId",
      minWidth: 150,
      showOverflowTooltip: true,
      formatter: ({ visitorId }) => formatNullable(visitorId)
    },
    {
      label: "会话ID",
      prop: "sessionId",
      minWidth: 150,
      showOverflowTooltip: true,
      formatter: ({ sessionId }) => formatNullable(sessionId)
    },
    {
      label: "IP",
      prop: "ip",
      minWidth: 130,
      formatter: ({ ip }) => formatNullable(ip)
    },
    {
      label: "页面",
      prop: "pageUrl",
      minWidth: 220,
      showOverflowTooltip: true,
      formatter: ({ pageUrl }) => formatNullable(pageUrl)
    },
    {
      label: "元数据",
      prop: "metadata",
      minWidth: 220,
      showOverflowTooltip: true,
      formatter: ({ metadata }) => formatNullable(metadata)
    }
  ];

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val - 1;
    onSearch();
  }

  function handleSelectionChange(val) {
    console.log("handleSelectionChange", val);
  }

  async function onSearch() {
    loading.value = true;
    try {
      const { data } = await getAttributionEventListAdmin(
        buildSearchParams(toRaw(form), toRaw(pagination))
      );
      dataList.value = data.list || [];
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取事件列表失败", error);
      message("获取事件列表失败", { type: "error" });
    } finally {
      setTimeout(() => {
        loading.value = false;
        tableRef.value?.setAdaptive?.();
      }, 500);
    }
  }

  const resetForm = formEl => {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  };

  onMounted(() => {
    onSearch();
  });

  return {
    form,
    loading,
    columns,
    dataList,
    pagination,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
