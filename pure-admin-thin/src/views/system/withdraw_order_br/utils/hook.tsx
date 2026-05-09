import dayjs from "dayjs";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import {
  getWithdrawOrderBrListAdmin,
  setWithdrawOrderBr,
  type WithdrawOrderBr
} from "@/api/withdrawOrderBr";
import { getSysPayChannelList, type SysPayChannel } from "@/api/sysPayChannel";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import {
  ElAlert,
  ElForm,
  ElFormItem,
  ElInput,
  ElOption,
  ElSelect,
  ElTag
} from "element-plus";

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
    userUid: "",
    orderNo: "",
    merchantOrderNo: "",
    providerPayoutNo: "",
    status: undefined as number | undefined,
    countryCode: "",
    channel: "",
    payMethod: ""
  });
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
  const approveForm = reactive({
    channelCode: ""
  });
  const approveFormRef = ref();
  const approveChannelLoading = ref(false);
  const approveChannelOptions = ref<SysPayChannel[]>([]);
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
      label: "国家",
      prop: "countryCode",
      minWidth: 100,
      formatter: row => getOrderCountryCode(row) || "-"
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
        tableRef.value?.setAdaptive?.();
      }, 500);
    }
  }

  const resetForm = formEl => {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  };

  function getOrderCountryCode(row: WithdrawOrderBr) {
    const rowCountryCode = String(row.countryCode || "").trim();
    if (rowCountryCode) return rowCountryCode.toUpperCase();
    const extra = row.extra?.trim();
    if (extra) {
      try {
        const parsed = JSON.parse(extra);
        const countryCode = String(parsed?.countryCode || "").trim();
        if (countryCode) return countryCode.toUpperCase();
      } catch {
        // ignore invalid historical extra data
      }
    }
    const currencyCountryMap: Record<string, string> = {
      BRL: "BR",
      MXN: "MX"
    };
    return currencyCountryMap[String(row.currency || "").toUpperCase()] || "";
  }

  async function loadWithdrawPayChannels(countryCode: string) {
    approveChannelLoading.value = true;
    try {
      const requests = ["withdraw", "both"].map(channelType =>
        getSysPayChannelList({
          currentPage: 0,
          pageSize: 500,
          channelType,
          countryCode,
          status: 1
        })
      );
      const results = await Promise.all(requests);
      const seen = new Set<number>();
      approveChannelOptions.value = results
        .flatMap(item => item.data?.list || [])
        .filter(item => {
          if (seen.has(item.id)) return false;
          seen.add(item.id);
          return true;
        });
    } catch (error) {
      console.error("获取提现支付通道失败", error);
      approveChannelOptions.value = [];
      message("获取提现支付通道失败", { type: "error" });
    } finally {
      approveChannelLoading.value = false;
    }
  }

  async function approveOrder(row: WithdrawOrderBr) {
    const countryCode = getOrderCountryCode(row);
    approveForm.channelCode = "";
    approveChannelOptions.value = [];
    await loadWithdrawPayChannels(countryCode);

    addDialog({
      title: `审核通过订单 ${row.orderNo}`,
      width: "420px",
      draggable: true,
      closeOnClickModal: false,
      contentRenderer: () => (
        <ElForm ref={approveFormRef} model={approveForm} labelWidth="96px">
          <ElAlert
            class="mb-4"
            type="info"
            showIcon={true}
            closable={false}
            title={`订单国家：${countryCode || "未识别"}，请选择对应国家的提现支付通道`}
          />
          <ElFormItem
            label="支付通道"
            prop="channelCode"
            rules={[
              {
                required: true,
                message: "请选择支付通道",
                trigger: "change"
              }
            ]}
          >
            <ElSelect
              v-model={approveForm.channelCode}
              filterable
              class="!w-full"
              loading={approveChannelLoading.value}
              placeholder="请选择提现支付通道"
            >
              {approveChannelOptions.value.map(item => (
                <ElOption
                  key={item.id}
                  label={`${item.channelName} (${item.channelCode})`}
                  value={item.channelCode}
                />
              ))}
            </ElSelect>
          </ElFormItem>
        </ElForm>
      ),
      beforeSure: done => {
        approveFormRef.value.validate(async valid => {
          if (!valid) return;
          const selected = approveChannelOptions.value.find(
            item => item.channelCode === approveForm.channelCode
          );
          if (!selected) {
            message("请选择有效的支付通道", { type: "warning" });
            return;
          }
          try {
            await setWithdrawOrderBr({
              id: row.id,
              status: 1,
              channel: selected.channelCode,
              provider: selected.channelCode,
              payMethod: selected.channelName,
              reviewedAt: dayjs().format("YYYY-MM-DD HH:mm:ss")
            });
            message(`已通过订单 ${row.orderNo}`, { type: "success" });
            done();
            onSearch();
          } catch (error) {
            console.error("审核通过失败", error);
            message("审核通过失败", { type: "error" });
          }
        });
      }
    });
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
