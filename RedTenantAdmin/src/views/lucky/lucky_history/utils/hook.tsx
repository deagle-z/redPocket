import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { getLuckyHistoryListAdmin, type LuckyHistory } from "@/api/luckyMoney";
import { type Ref, reactive, ref, onMounted } from "vue";
import { ElTag } from "element-plus";

export function useLuckyHistory(tableRef: Ref) {
  const form = reactive({
    luckyId: undefined as number | undefined,
    userId: undefined as number | undefined
  });
  const formRef = ref();
  const dataList = ref<LuckyHistory[]>([]);
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
      width: 80
    },
    {
      label: "红包ID",
      prop: "luckyId",
      minWidth: 100
    },
    {
      label: "用户ID",
      prop: "userId",
      minWidth: 100
    },
    {
      label: "用户名称",
      prop: "firstName",
      minWidth: 120
    },
    {
      label: "领取金额",
      prop: "amount",
      minWidth: 100,
      formatter: ({ amount }) => `${amount.toFixed(3)} U`
    },
    {
      label: "是否中雷",
      prop: "isThunder",
      minWidth: 100,
      cellRenderer: scope => {
        const isThunder = scope.row.isThunder === 1;
        return isThunder ? (
          <ElTag type="danger">💣 中雷</ElTag>
        ) : (
          <ElTag type="success">💵 正常</ElTag>
        );
      }
    },
    {
      label: "损失金额",
      prop: "loseMoney",
      minWidth: 100,
      cellRenderer: scope => {
        const loseMoney = scope.row.loseMoney;
        return loseMoney > 0 ? (
          <span class="text-red-500">-{loseMoney.toFixed(3)} U</span>
        ) : (
          <span>-</span>
        );
      }
    },
    {
      label: "领取时间",
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
      const { data } = await getLuckyHistoryListAdmin({
        ...form,
        ...pagination
      });
      dataList.value = data.list;
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取领取历史失败", error);
      message("获取领取历史失败", { type: "error" });
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
