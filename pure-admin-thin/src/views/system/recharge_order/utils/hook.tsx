import dayjs from "dayjs";
import { message } from "@/utils/message";
import type { PaginationProps } from "@pureadmin/table";
import {
  getRechargeOrderListAdmin,
  adminRechargeOrderCallback,
  type RechargeOrder
} from "@/api/rechargeOrder";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import { ElTag, ElButton, ElMessageBox } from "element-plus";

const statusOptions = [
  { label: "待支付", value: 0 },
  { label: "支付成功", value: 1 },
  { label: "状态2", value: 2 },
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
  if (status === 1) return "success";
  if (status === 6) return "warning";
  if (status === 3) return "danger";
  if (status === 4 || status === 5 || status === 7) return "info";
  return "info";
}

function getActivityTypeLabel(activityType?: number | null) {
  if (activityType === 1) return "首充活动";
  if (activityType === 2) return "今日首充";
  return "无";
}

function getActivityTypeTagType(activityType?: number | null) {
  if (activityType === 1) return "success";
  if (activityType === 2) return "warning";
  return "info";
}

const callbackLoadingIds = ref<Set<number>>(new Set());

function formatMoney(value: number | string | null | undefined) {
  const amount = Number(value ?? 0);
  return Number.isFinite(amount)
    ? amount.toLocaleString("en-US", { maximumFractionDigits: 2 })
    : "0";
}

function formatMoneyWithCurrency(
  value: number | string | null | undefined,
  currency?: string | null
) {
  return `${formatMoney(value)} ${currency || ""}`.trim();
}

export function useRechargeOrder(tableRef: Ref) {
  const form = reactive({
    userUid: "",
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
      label: "用户UID",
      prop: "userUid",
      minWidth: 120,
      formatter: ({ userUid }) => userUid || "-"
    },
    {
      label: "金额",
      prop: "amount",
      minWidth: 140,
      formatter: ({ amount }) => formatMoney(amount)
    },
    {
      label: "手续费",
      prop: "fee",
      minWidth: 120,
      formatter: ({ fee }) => formatMoney(fee)
    },
    {
      label: "净入账",
      prop: "netAmount",
      minWidth: 120,
      formatter: ({ netAmount, currency }) =>
        formatMoneyWithCurrency(netAmount, currency)
    },
    {
      label: "赠送金额",
      prop: "bonusAmount",
      minWidth: 120,
      formatter: ({ bonusAmount }) => formatMoney(bonusAmount)
    },
    {
      label: "活动类型",
      prop: "activityType",
      minWidth: 120,
      cellRenderer: scope => (
        <ElTag
          type={getActivityTypeTagType(scope.row.activityType)}
          effect="plain"
        >
          {getActivityTypeLabel(scope.row.activityType)}
        </ElTag>
      )
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
    },
    {
      label: "操作",
      fixed: "right",
      width: 130,
      cellRenderer: scope => {
        if (scope.row.status !== 0) return null;
        const id: number = scope.row.id;
        const isLoading = callbackLoadingIds.value.has(id);
        return (
          <ElButton
            type="warning"
            size="small"
            loading={isLoading}
            onClick={() => handleCallback(scope.row)}
          >
            手动回调
          </ElButton>
        );
      }
    }
  ];

  async function handleCallback(row: RechargeOrder) {
    try {
      await ElMessageBox.confirm(
        `确认对订单 ${row.orderNo} 执行手动回调入账？`,
        "手动回调确认",
        { type: "warning", confirmButtonText: "确认", cancelButtonText: "取消" }
      );
    } catch {
      return;
    }
    callbackLoadingIds.value.add(row.id);
    try {
      await adminRechargeOrderCallback(row.id);
      message("回调成功，订单已入账", { type: "success" });
      onSearch();
    } catch {
      message("回调失败，请重试", { type: "error" });
    } finally {
      callbackLoadingIds.value.delete(row.id);
    }
  }

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
