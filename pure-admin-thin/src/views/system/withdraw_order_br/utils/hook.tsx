import dayjs from "dayjs";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import {
  getTgUserWithdrawActivityFlowAdmin,
  getTgUserWithdrawFlowBatchOverviewAdmin,
  getWithdrawOrderBrListAdmin,
  setWithdrawOrderBr,
  type WithdrawActivityFlow,
  type WithdrawActivityFlowCycle,
  type WithdrawFlowBatch,
  type WithdrawFlowBatchOverview,
  type WithdrawOrderBr
} from "@/api/withdrawOrderBr";
import { getSysPayChannelList, type SysPayChannel } from "@/api/sysPayChannel";
import { type Ref, reactive, ref, onMounted, toRaw } from "vue";
import {
  ElAlert,
  ElDescriptions,
  ElDescriptionsItem,
  ElEmpty,
  ElForm,
  ElFormItem,
  ElInput,
  ElOption,
  ElProgress,
  ElSelect,
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

function getFlowBatchSourceLabel(sourceType: string) {
  if (sourceType === "recharge_v2") return "V2充值";
  if (sourceType === "vip_weekly_salary") return "VIP周薪";
  return sourceType || "-";
}

function getFlowBatchStatusLabel(status: number) {
  if (status === 1) return "进行中";
  if (status === 2) return "已完成";
  if (status === 3) return "已关闭";
  return "-";
}

function getFlowBatchStatusType(status: number) {
  if (status === 2) return "success";
  if (status === 1) return "warning";
  if (status === 3) return "info";
  return "info";
}

function getFlowBatchClosedReasonLabel(reason: string) {
  if (reason === "balance_below_threshold") return "余额低于阈值";
  return reason || "-";
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
    statusText: getActivityStatusLabel(item.status),
    startedAtText: formatDateTime(item.startedAt),
    endedAtText: formatDateTime(item.endedAt)
  }));
}

function flowBatchTableRows(data: WithdrawFlowBatchOverview) {
  return (data.batches || []).map(item => ({
    ...item,
    sourceTypeText: getFlowBatchSourceLabel(item.sourceType),
    statusText: getFlowBatchStatusLabel(item.status),
    closedReasonText: getFlowBatchClosedReasonLabel(item.closedReason),
    createdAtText: formatDateTime(item.createdAt),
    completedAtText: formatDateTime(item.completedAt),
    closedAtText: formatDateTime(item.closedAt),
    lastFlowAtText: formatDateTime(item.lastFlowAt)
  }));
}

function renderFlowBatchOverview(data: WithdrawFlowBatchOverview) {
  const unfinishedPercent =
    data.unfinishedRequiredFlow > 0
      ? Math.min(
          100,
          Number(
            (
              (data.unfinishedCompletedFlow / data.unfinishedRequiredFlow) *
              100
            ).toFixed(2)
          )
        )
      : 100;

  return (
    <div class="space-y-4">
      <ElDescriptions border column={3} size="small">
        <ElDescriptionsItem label="用户ID">{data.userId}</ElDescriptionsItem>
        <ElDescriptionsItem label="账户余额">
          {formatMoney(data.balance)}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="总流水">
          {formatMoney(data.totalFlow)}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="批次数">
          {formatMoney(data.batchCount)}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="总要求流水">
          {formatMoney(data.totalRequiredFlow)}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="总完成流水">
          {formatMoney(data.totalCompletedFlow)}
        </ElDescriptionsItem>
      </ElDescriptions>

      <div class="rounded border border-[var(--el-border-color)] p-4">
        <div class="mb-3 flex items-center justify-between">
          <span class="text-sm font-medium text-[var(--el-text-color-primary)]">
            未完成批次流水进度
          </span>
          {data.hasUnfinishedBatch ? (
            <ElTag type="warning" effect="plain">
              {data.unfinishedBatchCount} 个未完成
            </ElTag>
          ) : (
            <ElTag type="success" effect="plain">
              全部完成
            </ElTag>
          )}
        </div>
        <ElProgress percentage={unfinishedPercent} strokeWidth={10} />
        <ElDescriptions class="mt-4" border column={3} size="small">
          <ElDescriptionsItem label="未完成要求">
            {formatMoney(data.unfinishedRequiredFlow)}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="未完成已做">
            {formatMoney(data.unfinishedCompletedFlow)}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="未完成剩余">
            {formatMoney(data.unfinishedRemainingFlow)}
          </ElDescriptionsItem>
        </ElDescriptions>
      </div>
    </div>
  );
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
      label: "国家",
      prop: "countryCode",
      minWidth: 100,
      formatter: row => getOrderCountryCode(row) || "-"
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
      label: "备注",
      prop: "remark",
      minWidth: 160,
      showOverflowTooltip: true,
      formatter: ({ remark }) => remark || "-"
    },
    {
      label: "失败原因",
      prop: "failMsg",
      minWidth: 220,
      showOverflowTooltip: true,
      formatter: ({ failMsg }) => failMsg || "-"
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
      label: "自动审核",
      prop: "autoReviewed",
      minWidth: 100,
      cellRenderer: scope => (
        <ElTag
          type={Number(scope.row.autoReviewed) === 1 ? "success" : "info"}
          effect="plain"
        >
          {Number(scope.row.autoReviewed) === 1 ? "是" : "否"}
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
      width: 340,
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
            onSearch();
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

  async function showWithdrawActivityFlow(row: WithdrawOrderBr) {
    if (!row.userId) {
      message("该订单缺少用户ID", { type: "warning" });
      return;
    }

    try {
      const { data } = await getTgUserWithdrawActivityFlowAdmin(row.userId);
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

  async function showWithdrawFlowBatchV2(row: WithdrawOrderBr) {
    if (!row.userId) {
      message("该订单缺少用户ID", { type: "warning" });
      return;
    }

    try {
      const { data } = await getTgUserWithdrawFlowBatchOverviewAdmin(
        row.userId
      );
      const rows = flowBatchTableRows(data);

      addDialog({
        title: `用户 ${row.userUid || row.userId} V2提现流水批次`,
        width: "1080px",
        draggable: true,
        hideFooter: true,
        contentRenderer: () => (
          <div class="max-h-[72vh] overflow-auto pr-1">
            {renderFlowBatchOverview(data)}
            <div class="mt-4">
              {rows.length ? (
                <ElTable data={rows} border size="small">
                  <ElTableColumn
                    label="批次来源"
                    minWidth={120}
                    v-slots={{
                      default: ({ row: item }: { row: WithdrawFlowBatch }) => (
                        <ElTag effect="plain">
                          {getFlowBatchSourceLabel(item.sourceType)}
                        </ElTag>
                      )
                    }}
                  />
                  <ElTableColumn
                    prop="sourceOrderNo"
                    label="来源订单"
                    minWidth={170}
                    showOverflowTooltip={true}
                  />
                  <ElTableColumn
                    prop="activityCode"
                    label="活动"
                    minWidth={140}
                    showOverflowTooltip={true}
                    v-slots={{
                      default: ({ row: item }: { row: WithdrawFlowBatch }) =>
                        item.activityCode || "-"
                    }}
                  />
                  <ElTableColumn
                    prop="baseAmount"
                    label="基准金额"
                    width={110}
                    v-slots={{
                      default: ({ row: item }: { row: WithdrawFlowBatch }) =>
                        formatMoney(item.baseAmount)
                    }}
                  />
                  <ElTableColumn
                    prop="requiredFlow"
                    label="要求流水"
                    width={110}
                    v-slots={{
                      default: ({ row: item }: { row: WithdrawFlowBatch }) =>
                        formatMoney(item.requiredFlow)
                    }}
                  />
                  <ElTableColumn
                    prop="completedFlow"
                    label="完成流水"
                    width={110}
                    v-slots={{
                      default: ({ row: item }: { row: WithdrawFlowBatch }) =>
                        formatMoney(item.completedFlow)
                    }}
                  />
                  <ElTableColumn
                    prop="remainingFlow"
                    label="剩余流水"
                    width={110}
                    v-slots={{
                      default: ({ row: item }: { row: WithdrawFlowBatch }) =>
                        formatMoney(item.remainingFlow)
                    }}
                  />
                  <ElTableColumn
                    prop="progressPercent"
                    label="进度"
                    width={170}
                    v-slots={{
                      default: ({ row: item }: { row: WithdrawFlowBatch }) => (
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
                    label="状态"
                    width={90}
                    v-slots={{
                      default: ({ row: item }: { row: WithdrawFlowBatch }) => (
                        <ElTag
                          type={getFlowBatchStatusType(item.status)}
                          effect="plain"
                        >
                          {getFlowBatchStatusLabel(item.status)}
                        </ElTag>
                      )
                    }}
                  />
                  <ElTableColumn
                    prop="lastFlowAtText"
                    label="最后流水"
                    minWidth={160}
                  />
                  <ElTableColumn
                    prop="completedAtText"
                    label="完成时间"
                    minWidth={160}
                  />
                  <ElTableColumn
                    label="关闭原因"
                    minWidth={130}
                    showOverflowTooltip={true}
                    v-slots={{
                      default: ({ row: item }: { row: WithdrawFlowBatch }) =>
                        getFlowBatchClosedReasonLabel(item.closedReason)
                    }}
                  />
                </ElTable>
              ) : (
                <ElEmpty description="暂无V2提现流水批次" />
              )}
            </div>
          </div>
        )
      });
    } catch (error) {
      console.error("获取V2提现流水批次失败", error);
      message("获取V2提现流水批次失败", { type: "error" });
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
    showWithdrawActivityFlow,
    showWithdrawFlowBatchV2
  };
}
