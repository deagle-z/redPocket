import dayjs from "dayjs";
import { ElProgress, ElTag } from "element-plus";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { type Ref, onMounted, reactive, ref, toRaw } from "vue";
import {
  getTaskActivityRecordList,
  type TaskActivityRecord
} from "@/api/taskActivityRecord";
import { levelOptions } from "@/views/system/taskActivityConfig/utils/hook";
import type { TaskActivityLevelCode } from "@/api/taskActivityConfig";
import { withBackstageDisplayCurrency } from "@/utils/currency";

export const recordStatusOptions = [
  { label: "进行中", value: 0 },
  { label: "已达标", value: 1 },
  { label: "已发放", value: 2 },
  { label: "已过期", value: 3 },
  { label: "已取消", value: 4 }
];

const recordStatusTypeMap: Record<
  number,
  "success" | "warning" | "info" | "danger"
> = {
  0: "warning",
  1: "success",
  2: "success",
  3: "info",
  4: "danger"
};

const levelNameMap: Record<TaskActivityLevelCode, string> = {
  primary: "初级任务",
  middle: "中级任务",
  advanced: "高级任务",
  custom: "自定义任务"
};

function getStatusLabel(status: number) {
  return recordStatusOptions.find(item => item.value === status)?.label || "-";
}

function getLevelName(levelCode: TaskActivityLevelCode) {
  return levelNameMap[levelCode] || "-";
}

function formatDate(value?: string | null) {
  return value ? dayjs(value).format("YYYY-MM-DD HH:mm:ss") : "-";
}

function formatMoney(value?: number | null) {
  if (typeof value !== "number" || Number.isNaN(value)) return "0.00";
  return value.toFixed(2);
}

function percent(current: number, required: number) {
  if (!required || required <= 0) return 0;
  return Math.min(100, Number(((current / required) * 100).toFixed(2)));
}

export function useTaskActivityRecord(tableRef: Ref) {
  const form = reactive({
    uid: "",
    configId: undefined as number | undefined,
    title: "",
    levelCode: "" as TaskActivityLevelCode | "",
    status: null as number | null
  });
  const dataList = ref<TaskActivityRecord[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });

  const columns: TableColumnList = [
    { label: "ID", prop: "id", minWidth: 80 },
    {
      label: "用户UID",
      prop: "uid",
      minWidth: 120,
      formatter: ({ uid }) => uid || "-"
    },
    {
      label: "任务标题",
      prop: "title",
      minWidth: 180,
      showOverflowTooltip: true
    },
    {
      label: "任务等级",
      prop: "levelCode",
      minWidth: 110,
      cellRenderer: scope => (
        <ElTag effect="plain">{getLevelName(scope.row.levelCode)}</ElTag>
      )
    },
    {
      label: "邀请进度",
      prop: "progressInviteCount",
      minWidth: 170,
      cellRenderer: scope => (
        <div class="task-progress-cell">
          <span>
            {scope.row.progressInviteCount}/{scope.row.requiredInviteCount}
          </span>
          <ElProgress
            percentage={percent(
              scope.row.progressInviteCount,
              scope.row.requiredInviteCount
            )}
            show-text={false}
          />
        </div>
      )
    },
    {
      label: "充值达标人数",
      prop: "progressRechargeCount",
      minWidth: 120,
      formatter: ({ progressRechargeCount, requiredInviteCount }) =>
        `${progressRechargeCount}/${requiredInviteCount}`
    },
    {
      label: "充值达标金额",
      prop: "progressRechargeAmount",
      minWidth: 130,
      formatter: ({ progressRechargeAmount }) =>
        formatMoney(progressRechargeAmount)
    },
    {
      label: "单人充值要求",
      prop: "requiredRechargeAmount",
      minWidth: 130,
      formatter: ({ requiredRechargeAmount }) =>
        formatMoney(requiredRechargeAmount)
    },
    {
      label: "奖励",
      prop: "rewardAmount",
      minWidth: 130,
      formatter: ({ rewardAmount }) =>
        withBackstageDisplayCurrency(formatMoney(rewardAmount))
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 100,
      cellRenderer: scope => (
        <ElTag
          type={recordStatusTypeMap[scope.row.status] || "info"}
          effect="plain"
        >
          {getStatusLabel(scope.row.status)}
        </ElTag>
      )
    },
    {
      label: "领取时间",
      prop: "claimedAt",
      minWidth: 160,
      formatter: ({ claimedAt }) => formatDate(claimedAt)
    },
    {
      label: "截止时间",
      prop: "deadlineAt",
      minWidth: 160,
      formatter: ({ deadlineAt }) => formatDate(deadlineAt)
    },
    {
      label: "达标时间",
      prop: "completedAt",
      minWidth: 160,
      formatter: ({ completedAt }) => formatDate(completedAt)
    },
    {
      label: "发放时间",
      prop: "rewardedAt",
      minWidth: 160,
      formatter: ({ rewardedAt }) => formatDate(rewardedAt)
    },
    {
      label: "返水记录ID",
      prop: "rebateRecordId",
      minWidth: 120,
      formatter: ({ rebateRecordId }) => rebateRecordId || "-"
    },
    {
      label: "备注",
      prop: "remark",
      minWidth: 160,
      formatter: ({ remark }) => remark || "-"
    }
  ];

  async function fetchList() {
    loading.value = true;
    try {
      const { data } = await getTaskActivityRecordList({
        ...toRaw(form),
        currentPage: pagination.currentPage,
        pageSize: pagination.pageSize
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = data?.currentPage || 0;
    } catch (error) {
      console.error("获取任务活动记录失败", error);
      message("获取任务活动记录失败", { type: "error" });
    } finally {
      loading.value = false;
      tableRef.value?.setAdaptive?.();
    }
  }

  function onSearch() {
    pagination.currentPage = 0;
    fetchList();
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  }

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    pagination.currentPage = 0;
    fetchList();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val - 1;
    fetchList();
  }

  onMounted(() => {
    fetchList();
  });

  return {
    form,
    loading,
    columns,
    dataList,
    pagination,
    levelOptions,
    recordStatusOptions,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange
  };
}
