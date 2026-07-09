import editForm from "../form.vue";
import dayjs from "dayjs";
import { ElTag } from "element-plus";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import { deviceDetection } from "@pureadmin/utils";
import { type Ref, h, onMounted, reactive, ref, toRaw } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import type { FormItemProps } from "./types";
import {
  delTaskActivityConfig,
  getTaskActivityConfigList,
  setTaskActivityConfig,
  type TaskActivityConfig,
  type TaskActivityLevelCode
} from "@/api/taskActivityConfig";
import { withBackstageDisplayCurrency } from "@/utils/currency";

export const levelOptions = [
  { label: "初级任务", value: "primary" },
  { label: "中级任务", value: "middle" },
  { label: "高级任务", value: "advanced" }
] as const;

export const statusOptions = [
  { label: "启用", value: 1 },
  { label: "禁用", value: 0 }
];

const levelNameMap: Record<TaskActivityLevelCode, string> = {
  primary: "初级任务",
  middle: "中级任务",
  advanced: "高级任务"
};

function getLevelName(value: TaskActivityLevelCode) {
  return levelNameMap[value] || "-";
}

function formatMoney(value?: number | null) {
  if (typeof value !== "number" || Number.isNaN(value)) return "0.00";
  return value.toFixed(2);
}

function formatDate(value?: string | null) {
  return value ? dayjs(value).format("YYYY-MM-DD HH:mm:ss") : "-";
}

function formatDuration(minutes: number) {
  if (!minutes) return "-";
  if (minutes < 60) return `${minutes}分钟`;
  if (minutes % 1440 === 0) return `${minutes / 1440}天`;
  if (minutes % 60 === 0) return `${minutes / 60}小时`;
  return `${minutes}分钟`;
}

function normalizeDate(value?: string | null) {
  return value && String(value).trim() ? value : null;
}

function normalizePickerDate(value?: string | null) {
  return value ? dayjs(value).format("YYYY-MM-DDTHH:mm:ssZ") : null;
}

function defaultForm(title: string, row?: TaskActivityConfig): FormItemProps {
  const levelCode = row?.levelCode ?? "primary";
  return {
    title,
    id: row?.id ?? 0,
    configTitle: row?.title ?? "",
    subTitle: row?.subTitle ?? "",
    levelCode,
    levelName: row?.levelName ?? getLevelName(levelCode),
    sort: row?.sort ?? 0,
    status: row?.status ?? 1,
    taskType: row?.taskType ?? "invite_recharge",
    requiredInviteCount: row?.requiredInviteCount ?? 1,
    requiredRechargeAmount: row?.requiredRechargeAmount ?? 50,
    durationMinutes: row?.durationMinutes ?? 1440,
    rewardAmount: row?.rewardAmount ?? 0,
    rewardCurrency: row?.rewardCurrency ?? "USD",
    rewardTarget: row?.rewardTarget ?? "rebate",
    startAt: normalizePickerDate(row?.startAt),
    endAt: normalizePickerDate(row?.endAt),
    remark: row?.remark ?? ""
  };
}

export function useTaskActivityConfig(tableRef: Ref) {
  const form = reactive({
    title: "",
    levelCode: "" as TaskActivityLevelCode | "",
    status: null as number | null
  });
  const formRef = ref();
  const dataList = ref<TaskActivityConfig[]>([]);
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
      label: "邀请人数",
      prop: "requiredInviteCount",
      minWidth: 100
    },
    {
      label: "好友充值要求",
      prop: "requiredRechargeAmount",
      minWidth: 130,
      formatter: ({ requiredRechargeAmount }) =>
        formatMoney(requiredRechargeAmount)
    },
    {
      label: "限制时间",
      prop: "durationMinutes",
      minWidth: 110,
      formatter: ({ durationMinutes }) => formatDuration(durationMinutes)
    },
    {
      label: "奖励",
      prop: "rewardAmount",
      minWidth: 130,
      formatter: ({ rewardAmount }) =>
        withBackstageDisplayCurrency(formatMoney(rewardAmount))
    },
    {
      label: "奖励账户",
      prop: "rewardTarget",
      minWidth: 100,
      formatter: ({ rewardTarget }) =>
        rewardTarget === "rebate" ? "佣金" : "-"
    },
    { label: "排序", prop: "sort", minWidth: 80 },
    {
      label: "状态",
      prop: "status",
      minWidth: 90,
      cellRenderer: scope => (
        <ElTag
          type={scope.row.status === 1 ? "success" : "danger"}
          effect="plain"
        >
          {scope.row.status === 1 ? "启用" : "禁用"}
        </ElTag>
      )
    },
    {
      label: "开始时间",
      prop: "startAt",
      minWidth: 160,
      formatter: ({ startAt }) => formatDate(startAt)
    },
    {
      label: "结束时间",
      prop: "endAt",
      minWidth: 160,
      formatter: ({ endAt }) => formatDate(endAt)
    },
    {
      label: "更新时间",
      prop: "updatedAt",
      minWidth: 160,
      formatter: ({ updatedAt }) => formatDate(updatedAt)
    },
    {
      label: "备注",
      prop: "remark",
      minWidth: 160,
      formatter: ({ remark }) => remark || "-"
    },
    { label: "操作", fixed: "right", width: 150, slot: "operation" }
  ];

  async function fetchList() {
    loading.value = true;
    try {
      const { data } = await getTaskActivityConfigList({
        ...toRaw(form),
        currentPage: pagination.currentPage,
        pageSize: pagination.pageSize
      });
      dataList.value = data?.list || [];
      pagination.total = data?.total || 0;
      pagination.pageSize = data?.pageSize || pagination.pageSize;
      pagination.currentPage = data?.currentPage || 0;
    } catch (error) {
      console.error("获取任务活动配置失败", error);
      message("获取任务活动配置失败", { type: "error" });
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

  async function handleDelete(row: TaskActivityConfig) {
    await delTaskActivityConfig(row.id);
    message(`已删除任务 [${row.title}]`, { type: "success" });
    fetchList();
  }

  function openDialog(title = "新增", row?: TaskActivityConfig) {
    addDialog({
      title: `${title}任务活动配置`,
      props: {
        formInline: defaultForm(title, row)
      },
      width: "720px",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(editForm, { ref: formRef, formInline: null }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as FormItemProps;
        FormRef.validate(async valid => {
          if (!valid) return;
          try {
            await setTaskActivityConfig({
              id: curData.id || undefined,
              title: curData.configTitle.trim(),
              subTitle: curData.subTitle.trim(),
              levelCode: curData.levelCode,
              levelName: curData.levelName.trim(),
              sort: Number(curData.sort || 0),
              status: curData.status,
              taskType: "invite_recharge",
              requiredInviteCount: Number(curData.requiredInviteCount || 0),
              requiredRechargeAmount: Number(
                curData.requiredRechargeAmount || 0
              ),
              durationMinutes: Number(curData.durationMinutes || 0),
              rewardAmount: Number(curData.rewardAmount || 0),
              rewardCurrency: curData.rewardCurrency.trim() || "USD",
              rewardTarget: "rebate",
              startAt: normalizeDate(curData.startAt),
              endAt: normalizeDate(curData.endAt),
              remark: curData.remark.trim()
            });
            message(`${title}成功`, { type: "success" });
            done();
            fetchList();
          } catch (error) {
            console.error("保存任务活动配置失败", error);
            message("保存任务活动配置失败", { type: "error" });
          }
        });
      }
    });
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
    statusOptions,
    onSearch,
    resetForm,
    openDialog,
    handleDelete,
    handleSizeChange,
    handleCurrentChange
  };
}
