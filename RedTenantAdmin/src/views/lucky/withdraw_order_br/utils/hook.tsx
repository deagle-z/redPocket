import dayjs from "dayjs";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import {
  getWithdrawOrderBrListAdmin,
  setWithdrawOrderBr,
  type WithdrawOrderBr
} from "@/api/withdrawOrderBr";
import { setBalance } from "@/api/system";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import { ElForm, ElFormItem, ElInput, ElTag } from "element-plus";

const statusOptions = [
  { label: "待审核", value: 0 },
  { label: "待打款", value: 1 },
  { label: "打款中", value: 2 },
  { label: "成功", value: 3 },
  { label: "失败", value: 4 },
  { label: "取消", value: 5 },
  { label: "退回", value: 6 }
];

function getStatusLabel(status: number) {
  const match = statusOptions.find(item => item.value === status);
  return match ? match.label : "-";
}

function getStatusType(status: number) {
  if (status === 3) return "success";
  if (status === 1 || status === 2) return "warning";
  if (status === 4) return "danger";
  if (status === 5 || status === 6) return "info";
  return "info";
}

export function useWithdrawOrderBr(tableRef: Ref) {
  const form = reactive({
    userId: undefined as number | undefined,
    orderNo: "",
    merchantOrderNo: "",
    providerPayoutNo: "",
    status: undefined as number | undefined,
    channel: "",
    payMethod: ""
  });
  const formRef = ref();
  const dataList = ref<WithdrawOrderBr[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 0,
    background: true
  });
  const rejectForm = reactive({
    failMsg: ""
  });
  const rejectFormRef = ref();
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
      label: "净打款",
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
      label: "审核时间",
      prop: "reviewedAt",
      minWidth: 160,
      formatter: ({ reviewedAt }) =>
        reviewedAt ? dayjs(reviewedAt).format("YYYY-MM-DD HH:mm:ss") : "-"
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
      width: 180,
      slot: "operation"
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
      const { data } = await getWithdrawOrderBrListAdmin({
        ...toRaw(form),
        ...toRaw(pagination)
      });
      dataList.value = data.list || [];
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取提现订单失败", error);
      message("获取提现订单失败", { type: "error" });
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

  async function approveOrder(row: WithdrawOrderBr) {
    try {
      await setWithdrawOrderBr({
        id: row.id,
        status: 1,
        reviewedAt: dayjs().format("YYYY-MM-DD HH:mm:ss")
      });
      message(`已通过订单 ${row.orderNo}`, { type: "success" });
      onSearch();
    } catch (error) {
      console.error("审核通过失败", error);
      message("审核通过失败", { type: "error" });
    }
  }

  function rejectOrder(row: WithdrawOrderBr) {
    rejectForm.failMsg = "";
    addDialog({
      title: `驳回订单 ${row.orderNo}`,
      width: "30%",
      draggable: true,
      closeOnClickModal: false,
      contentRenderer: () => (
        <ElForm ref={rejectFormRef} model={rejectForm}>
          <ElFormItem
            prop="failMsg"
            rules={[
              {
                required: true,
                message: "请输入驳回原因",
                trigger: "blur"
              }
            ]}
          >
            <ElInput
              type="textarea"
              rows={4}
              v-model={rejectForm.failMsg}
              placeholder="请输入驳回原因"
            />
          </ElFormItem>
        </ElForm>
      ),
      beforeSure: done => {
        rejectFormRef.value.validate(async valid => {
          if (!valid) return;
          try {
            await setWithdrawOrderBr({
              id: row.id,
              status: 6,
              failMsg: rejectForm.failMsg,
              reviewedAt: dayjs().format("YYYY-MM-DD HH:mm:ss")
            });
            await setBalance({
              userId: row.userId,
              amount: row.amount,
              cashMark: `提现驳回返还（订单号：${row.orderNo}）`
            });
            message("驳回成功，已返还余额", { type: "success" });
            done();
            onSearch();
          } catch (error) {
            console.error("驳回失败", error);
            message("驳回失败", { type: "error" });
          }
        });
      }
    });
  }

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
    handleSelectionChange,
    approveOrder,
    rejectOrder
  };
}
