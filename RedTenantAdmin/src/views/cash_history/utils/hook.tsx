import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { getCashHistoryListAdmin, type CashHistory } from "@/api/cashHistory";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import { ElTag } from "element-plus";

export function useCashHistory(tableRef: Ref) {
  const form = reactive({
    userId: undefined as number | undefined,
    cashMark: undefined as string | undefined
  });
  const formRef = ref();
  const dataList = ref<CashHistory[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });
  const columns: TableColumnList = [
    {
      label: "用户ID",
      prop: "userId",
      minWidth: 100
    },
    {
      label: "变动金额",
      prop: "amount",
      minWidth: 120,
      cellRenderer: scope => {
        const amount = scope.row.amount;
        const isPositive = amount >= 0;
        return (
          <span class={isPositive ? "text-green-500" : "text-red-500"}>
            {isPositive ? "+" : ""}{amount.toFixed(3)} U
          </span>
        );
      }
    },
    {
      label: "变动前余额",
      prop: "startAmount",
      minWidth: 120,
      formatter: ({ startAmount }) => `${startAmount.toFixed(3)} U`
    },
    {
      label: "变动后余额",
      prop: "endAmount",
      minWidth: 120,
      formatter: ({ endAmount }) => `${endAmount.toFixed(3)} U`
    },
    {
      label: "余额备注",
      prop: "cashMark",
      minWidth: 120,
      showOverflowTooltip: true
    },
    {
      label: "余额描述",
      prop: "cashDesc",
      minWidth: 200,
      showOverflowTooltip: true
    },
    {
      label: "来源用户ID",
      prop: "fromUserId",
      minWidth: 120,
      formatter: ({ fromUserId }) => (fromUserId > 0 ? fromUserId : "-")
    },
    {
      label: "变动时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) => dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
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
      const { data } = await getCashHistoryListAdmin({
        ...toRaw(form),
        ...toRaw(pagination)
      });
      dataList.value = data.list;
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取余额变动记录失败", error);
      message("获取余额变动记录失败", { type: "error" });
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
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
