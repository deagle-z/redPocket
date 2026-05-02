import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import {
  getPlatformProfitLedgerList,
  type PlatformProfitLedger
} from "@/api/platformProfitLedger";

const sourceTypeOptions = [
  { label: "抢包抽水", value: "lucky_grab_commission" },
  { label: "中雷抽水", value: "lucky_thunder_commission" },
  { label: "注册赠送", value: "register_gift" },
  { label: "充值赠送", value: "recharge_gift" }
];

function getSourceTypeLabel(value: string) {
  const match = sourceTypeOptions.find(item => item.value === value);
  return match ? match.label : value || "-";
}

function formatMoney(val?: number | null) {
  if (typeof val !== "number" || Number.isNaN(val)) return "0.000";
  return val.toFixed(3);
}

export function usePlatformProfitLedger(tableRef: Ref) {
  const form = reactive({
    userId: undefined as number | undefined,
    sourceType: "",
    sourceId: "",
    minNet: undefined as number | undefined,
    maxNet: undefined as number | undefined
  });
  const dataList = ref<PlatformProfitLedger[]>([]);
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
      label: "用户ID",
      prop: "userId",
      minWidth: 100
    },
    {
      label: "用户UID",
      prop: "userUid",
      minWidth: 120,
      formatter: ({ userUid }) => userUid || "-"
    },
    {
      label: "用户名称",
      prop: "userName",
      minWidth: 140,
      formatter: ({ userName }) => userName || "-"
    },
    {
      label: "来源类型",
      prop: "sourceType",
      minWidth: 130,
      formatter: ({ sourceType }) => getSourceTypeLabel(sourceType)
    },
    {
      label: "来源单号",
      prop: "sourceId",
      minWidth: 220
    },
    {
      label: "总抽水",
      prop: "incomeAmount",
      minWidth: 120,
      formatter: ({ incomeAmount }) => formatMoney(incomeAmount)
    },
    {
      label: "返佣",
      prop: "rebateAmount",
      minWidth: 120,
      formatter: ({ rebateAmount }) => formatMoney(rebateAmount)
    },
    {
      label: "实际抽水",
      prop: "actualIncomeAmount",
      minWidth: 120,
      formatter: ({ actualIncomeAmount }) => formatMoney(actualIncomeAmount)
    },
    {
      label: "亏损金额",
      prop: "expenseAmount",
      minWidth: 120,
      formatter: ({ expenseAmount }) => formatMoney(expenseAmount)
    },
    {
      label: "净额",
      prop: "netAmount",
      minWidth: 120,
      formatter: ({ netAmount }) => formatMoney(netAmount)
    },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) =>
        dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
    },
    {
      label: "备注",
      prop: "remark",
      minWidth: 180,
      formatter: ({ remark }) => remark || "-"
    }
  ];

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    pagination.currentPage = 0;
    fetchList();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val - 1;
    fetchList();
  }

  function handleSelectionChange(val) {
    console.log("handleSelectionChange", val);
  }

  function onSearch() {
    pagination.currentPage = 0;
    fetchList();
  }

  async function fetchList() {
    loading.value = true;
    try {
      const { data } = await getPlatformProfitLedgerList({
        ...toRaw(form),
        currentPage: pagination.currentPage,
        pageSize: pagination.pageSize
      });
      dataList.value = data.list || [];
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取平台抽水记录失败", error);
      message("获取平台抽水记录失败", { type: "error" });
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
    fetchList();
  });

  return {
    form,
    loading,
    columns,
    dataList,
    pagination,
    sourceTypeOptions,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
