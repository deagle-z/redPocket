import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import { getCashHistoryListAdmin, type CashHistory } from "@/api/cashHistory";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";

export const cashHistoryTypeOptions = [
  { label: "未知", value: 0 },
  { label: "发送红包", value: 1 },
  { label: "抢红包未中雷收益", value: 2 },
  { label: "抢红包中雷损失", value: 3 },
  { label: "发包者中雷收益", value: 4 },
  { label: "红包相关抽成", value: 5 },
  { label: "充值到账", value: 6 },
  { label: "后台手工加款", value: 7 },
  { label: "后台手工扣款", value: 8 },
  { label: "提现申请扣款", value: 9 },
  { label: "提现失败/取消/退回返还", value: 10 },
  { label: "佣金(返水)转余额", value: 11 },
  { label: "红包过期退回", value: 12 },
  { label: "抽奖消耗", value: 13 },
  { label: "抽奖中奖", value: 14 },
  { label: "VIP升级奖励", value: 15 },
  { label: "首充赠送", value: 16 },
  { label: "今日首充赠送", value: 17 },
  { label: "幸运数字赠送", value: 18 },
  { label: "后台手工加佣金", value: 19 },
  { label: "签到活动赠送", value: 20 },
  { label: "三方游戏下注扣款", value: 21 },
  { label: "三方游戏派奖", value: 22 },
  { label: "三方游戏退回下注", value: 23 },
  { label: "兑换码赠送", value: 24 },
  { label: "VIP周薪", value: 25 },
  { label: "邀请返佣阶梯奖励", value: 26 },
  { label: "注册赠送", value: 27 }
];

function getCashHistoryTypeLabel(type: number) {
  const match = cashHistoryTypeOptions.find(item => item.value === type);
  return match ? match.label : `未知(${type})`;
}

export function useCashHistory(_tableRef: Ref) {
  const form = reactive({
    userId: undefined as number | undefined,
    uid: undefined as string | undefined,
    type: undefined as number | undefined,
    cashMark: undefined as string | undefined
  });
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
      label: "用户UID",
      prop: "uid",
      minWidth: 120,
      formatter: ({ uid }) => uid || "-"
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
            {isPositive ? "+" : ""}
            {amount.toFixed(3)} U
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
      label: "类型",
      prop: "type",
      minWidth: 150,
      formatter: ({ type }) => getCashHistoryTypeLabel(type)
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
      formatter: ({ createdAt }) =>
        dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
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
