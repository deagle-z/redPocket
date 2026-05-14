import dayjs from "dayjs";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import {
  getTgUserWithdrawActivityFlowTenant,
  getWithdrawOrderBrListAdmin,
  setWithdrawOrderBr,
  type WithdrawActivityFlow,
  type WithdrawActivityFlowCycle,
  type WithdrawOrderBr
} from "@/api/withdrawOrderBr";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import {
  ElDescriptions,
  ElDescriptionsItem,
  ElEmpty,
  ElForm,
  ElFormItem,
  ElInput,
  ElProgress,
  ElTable,
  ElTableColumn,
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

function formatDateTime(value?: string | null) {
  return value ? dayjs(value).format("YYYY-MM-DD HH:mm:ss") : "-";
}

function getActivityTypeLabel(type: number) {
  if (type === 1) return "首充";
  if (type === 2) return "今日首充";
  return "普通";
}

function getActivityStatusLabel(status: number) {
  if (status === 1) return "进行中";
  if (status === 2) return "已结束";
  return "-";
}

function getActivityStatusType(status: number) {
  return status === 1 ? "success" : "info";
}

function renderActivityOverview(data: WithdrawActivityFlow) {
  const active = data.activeActivity;

  return (
    <div class="space-y-4">
      <ElDescriptions border column={2} size="small">
        <ElDescriptionsItem label="用户ID">{data.userId}</ElDescriptionsItem>
        <ElDescriptionsItem label="账户余额">
          {formatMoney(data.balance)}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="累计流水">
          {formatMoney(data.totalFlow)}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="当前活动">
          {data.hasActivity && active ? (
            <ElTag type="success" effect="plain">
              {active.activityCode || getActivityTypeLabel(active.activityType)}
            </ElTag>
          ) : (
            <ElTag type="info" effect="plain">
              无进行中活动
            </ElTag>
          )}
        </ElDescriptionsItem>
      </ElDescriptions>

      {active ? (
        <div class="rounded border border-[var(--el-border-color)] p-4">
          <div class="mb-3 flex items-center justify-between">
            <span class="text-sm font-medium text-[var(--el-text-color-primary)]">
              当前活动流水进度
            </span>
            <ElTag type={getActivityStatusType(active.status)} effect="plain">
              {getActivityStatusLabel(active.status)}
            </ElTag>
          </div>
          <ElProgress
            percentage={Math.min(100, Number(active.progressPercent || 0))}
            strokeWidth={10}
          />
          <ElDescriptions class="mt-4" border column={3} size="small">
            <ElDescriptionsItem label="活动类型">
              {getActivityTypeLabel(active.activityType)}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="流水倍数">
              {formatMoney(active.multiplier)}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="基础金额">
              {formatMoney(active.baseAmount)}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="要求流水">
              {formatMoney(active.requiredFlow)}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="当前流水">
              {formatMoney(active.currentFlow)}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="剩余流水">
              {formatMoney(active.remainingFlow)}
            </ElDescriptionsItem>
          </ElDescriptions>
        </div>
      ) : null}
    </div>
  );
}

function activityTableRows(data: WithdrawActivityFlow) {
  return (data.activities || []).map(item => ({
    ...item,
    activityName: item.activityCode || getActivityTypeLabel(item.activityType),
    startedAtText: formatDateTime(item.startedAt),
    endedAtText: formatDateTime(item.endedAt)
  }));
}

export function useWithdrawOrderBr(_tableRef: Ref) {
  const form = reactive({
    userUid: "",
    orderNo: "",
    merchantOrderNo: "",
    providerPayoutNo: "",
    status: undefined as number | undefined,
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
  const columns: TableColumnList = [
    {
      label: "订单号",
      prop: "merchantOrderNo",
      minWidth: 180,
      showOverflowTooltip: true,
      formatter: ({ merchantOrderNo, orderNo }) => merchantOrderNo || orderNo
    },
    {
      label: "用户UID",
      prop: "userUid",
      minWidth: 120,
      formatter: ({ userUid }) => userUid || "-"
    },
    {
      label: "提现金额",
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
      label: "净打款",
      prop: "netAmount",
      minWidth: 120,
      formatter: ({ netAmount, currency }) =>
        formatMoneyWithCurrency(netAmount, currency)
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
      width: 260,
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

  async function showWithdrawActivityFlow(row: WithdrawOrderBr) {
    if (!row.userId) {
      message("该订单缺少用户ID", { type: "warning" });
      return;
    }

    try {
      const { data } = await getTgUserWithdrawActivityFlowTenant(row.userId);
      const rows = activityTableRows(data);

      addDialog({
        title: `用户 ${row.userUid || row.userId} 活动流水`,
        width: "860px",
        draggable: true,
        hideFooter: true,
        contentRenderer: () => (
          <div class="max-h-[70vh] overflow-auto pr-1">
            {renderActivityOverview(data)}
            <div class="mt-4">
              {rows.length ? (
                <ElTable data={rows} border size="small">
                  <ElTableColumn
                    prop="activityName"
                    label="活动"
                    minWidth={150}
                    showOverflowTooltip={true}
                  />
                  <ElTableColumn
                    label="状态"
                    width={90}
                    v-slots={{
                      default: ({
                        row: item
                      }: {
                        row: WithdrawActivityFlowCycle;
                      }) => (
                        <ElTag
                          type={getActivityStatusType(item.status)}
                          effect="plain"
                        >
                          {getActivityStatusLabel(item.status)}
                        </ElTag>
                      )
                    }}
                  />
                  <ElTableColumn
                    prop="requiredFlow"
                    label="要求流水"
                    width={110}
                    v-slots={{
                      default: ({
                        row: item
                      }: {
                        row: WithdrawActivityFlowCycle;
                      }) => formatMoney(item.requiredFlow)
                    }}
                  />
                  <ElTableColumn
                    prop="currentFlow"
                    label="当前流水"
                    width={110}
                    v-slots={{
                      default: ({
                        row: item
                      }: {
                        row: WithdrawActivityFlowCycle;
                      }) => formatMoney(item.currentFlow)
                    }}
                  />
                  <ElTableColumn
                    prop="remainingFlow"
                    label="剩余流水"
                    width={110}
                    v-slots={{
                      default: ({
                        row: item
                      }: {
                        row: WithdrawActivityFlowCycle;
                      }) => formatMoney(item.remainingFlow)
                    }}
                  />
                  <ElTableColumn
                    prop="progressPercent"
                    label="进度"
                    width={160}
                    v-slots={{
                      default: ({
                        row: item
                      }: {
                        row: WithdrawActivityFlowCycle;
                      }) => (
                        <ElProgress
                          percentage={Math.min(
                            100,
                            Number(item.progressPercent || 0)
                          )}
                        />
                      )
                    }}
                  />
                  <ElTableColumn
                    prop="startedAtText"
                    label="开始时间"
                    minWidth={160}
                  />
                  <ElTableColumn
                    prop="endedAtText"
                    label="结束时间"
                    minWidth={160}
                  />
                </ElTable>
              ) : (
                <ElEmpty description="暂无活动记录" />
              )}
            </div>
          </div>
        )
      });
    } catch (error) {
      console.error("获取活动流水失败", error);
      message("获取活动流水失败", { type: "error" });
    }
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
    rejectOrder,
    showWithdrawActivityFlow
  };
}
