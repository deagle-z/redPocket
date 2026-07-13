import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import { ElTag } from "element-plus";
import {
  getTgUserRebateList,
  type TgUserRebateRecord
} from "@/api/tgUserRebate";

const sourceTypeOptions = [
  { label: "下注流水", value: 1 },
  { label: "充值", value: 2 },
  { label: "游戏输赢", value: 3 },
  { label: "手动补单", value: 4 },
  { label: "其他", value: 5 }
];

const statusOptions = [
  { label: "待结算", value: 0 },
  { label: "已结算", value: 1 },
  { label: "作废/冲正", value: 2 }
];

function getSourceTypeLabel(value: number) {
  const match = sourceTypeOptions.find(item => item.value === value);
  return match ? match.label : "-";
}

function getStatusLabel(status: number) {
  const match = statusOptions.find(item => item.value === status);
  return match ? match.label : "-";
}

function getStatusType(status: number) {
  if (status === 1) return "success";
  if (status === 0) return "warning";
  return "info";
}

function formatMoney(val?: number | null) {
  if (typeof val !== "number" || Number.isNaN(val)) return "0.000";
  return val.toFixed(3);
}

export function useRebateRecord(tableRef: Ref) {
  const form = reactive({
    subUserId: undefined as number | undefined,
    parentUserId: undefined as number | undefined,
    sourceType: undefined as number | undefined,
    sourceOrderId: "",
    status: undefined as number | undefined,
    idempotencyKey: ""
  });
  const formRef = ref();
  const dataList = ref<TgUserRebateRecord[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });
  const columns: TableColumnList = [
    {
      label: "ID",
      prop: "id",
      minWidth: 90
    },
    {
      label: "下级ID",
      prop: "subUserId",
      minWidth: 100
    },
    {
      label: "上级ID",
      prop: "parentUserId",
      minWidth: 100
    },
    {
      label: "来源类型",
      prop: "sourceType",
      minWidth: 120,
      formatter: ({ sourceType }) => getSourceTypeLabel(sourceType)
    },
    {
      label: "来源单号",
      prop: "sourceOrderId",
      minWidth: 180
    },
    {
      label: "来源金额",
      prop: "sourceAmount",
      minWidth: 120,
      formatter: ({ sourceAmount, currency }) =>
        `${formatMoney(sourceAmount)} ${currency}`
    },
    {
      label: "反水比例(%)",
      prop: "rebateRate",
      minWidth: 120,
      formatter: ({ rebateRate }) => rebateRate.toFixed(6)
    },
    {
      label: "反水金额",
      prop: "rebateAmount",
      minWidth: 120,
      formatter: ({ rebateAmount, currency }) =>
        `${formatMoney(rebateAmount)} ${currency}`
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 100,
      cellRenderer: scope => (
        <ElTag type={getStatusType(scope.row.status)} effect="plain">
          {getStatusLabel(scope.row.status)}
        </ElTag>
      )
    },
    {
      label: "结算时间",
      prop: "settledAt",
      minWidth: 160,
      formatter: ({ settledAt }) =>
        settledAt ? dayjs(settledAt).format("YYYY-MM-DD HH:mm:ss") : "-"
    },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) =>
        dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
    },
    {
      label: "幂等键",
      prop: "idempotencyKey",
      minWidth: 200
    },
    {
      label: "备注",
      prop: "remark",
      minWidth: 160,
      formatter: ({ remark }) => (remark ? remark : "-")
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
      const { data } = await getTgUserRebateList({
        ...toRaw(form),
        ...toRaw(pagination)
      });
      dataList.value = data.list || [];
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取返水记录失败", error);
      message("获取返水记录失败", { type: "error" });
    } finally {
      setTimeout(() => {
        loading.value = false;
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
    sourceTypeOptions,
    statusOptions,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
