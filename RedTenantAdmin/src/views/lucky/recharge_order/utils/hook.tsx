import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import {
  getRechargeOrderListAdmin,
  type RechargeOrder
} from "@/api/rechargeOrder";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import { ElTag } from "element-plus";

const statusOptions = [
  { label: "待支付", value: 0 },
  { label: "支付中", value: 1 },
  { label: "成功", value: 2 },
  { label: "失败", value: 3 },
  { label: "取消", value: 4 },
  { label: "关闭/超时", value: 5 },
  { label: "退款中", value: 6 },
  { label: "已退款", value: 7 }
];

function getStatusLabel(status: number) {
  const match = statusOptions.find(item => item.value === status);
  return match ? match.label : "-";
}

function getStatusType(status: number) {
  if (status === 2) return "success";
  if (status === 1 || status === 6) return "warning";
  if (status === 3) return "danger";
  if (status === 4 || status === 5 || status === 7) return "info";
  return "info";
}

export function useRechargeOrder(tableRef: Ref) {
  const form = reactive({
    userId: undefined as number | undefined,
    orderNo: "",
    merchantOrderNo: "",
    providerTradeNo: "",
    status: undefined as number | undefined,
    channel: "",
    payMethod: ""
  });
  const formRef = ref();
  const dataList = ref<RechargeOrder[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });
  const columns: TableColumnList = [
    {
      label: "订单号",
      prop: "orderNo",
      minWidth: 180,
      showOverflowTooltip: true
    },
    {
      label: "用户ID",
      prop: "userId",
      minWidth: 100
    },
    {
      label: "金额",
      prop: "amount",
      minWidth: 140,
      formatter: ({ amount, currency }) => `${amount.toFixed(6)} ${currency}`
    },
    {
      label: "手续费",
      prop: "fee",
      minWidth: 120,
      formatter: ({ fee, currency }) => `${fee.toFixed(6)} ${currency}`
    },
    {
      label: "净入账",
      prop: "netAmount",
      minWidth: 120,
      formatter: ({ netAmount, currency }) =>
        `${netAmount.toFixed(6)} ${currency}`
    },
    {
      label: "渠道",
      prop: "channel",
      minWidth: 120
    },
    {
      label: "子渠道",
      prop: "payMethod",
      minWidth: 120,
      formatter: ({ payMethod }) => payMethod || "-"
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 120,
      cellRenderer: scope => (
        <ElTag type={getStatusType(scope.row.status)} effect="plain">
          {getStatusLabel(scope.row.status)}
        </ElTag>
      )
    },
    {
      label: "支付时间",
      prop: "payTime",
      minWidth: 160,
      formatter: ({ payTime }) =>
        payTime ? dayjs(payTime).format("YYYY-MM-DD HH:mm:ss") : "-"
    },
    {
      label: "创建时间",
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
      const { data } = await getRechargeOrderListAdmin({
        ...toRaw(form),
        ...toRaw(pagination)
      });
      dataList.value = data.list || [];
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取充值订单失败", error);
      message("获取充值订单失败", { type: "error" });
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
    statusOptions,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
