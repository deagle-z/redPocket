import dayjs from "dayjs";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import {
  getLuckyMoneyListAdmin,
  getLuckyMoneyDetailAdmin,
  manualGrabLuckyMoneyAdmin,
  type LuckyMoney,
  type LuckyHistory,
  type LuckyMoneyItem
} from "@/api/luckyMoney";
import { getAdminBotUserList, type TgUser } from "@/api/tgUser";
import {
  type Ref,
  reactive,
  ref,
  onMounted,
  h,
  defineComponent,
  type PropType,
  computed
} from "vue";
import {
  ElAlert,
  ElButton,
  ElForm,
  ElFormItem,
  ElOption,
  ElRadio,
  ElRadioGroup,
  ElSelect,
  ElTag,
  ElTable,
  ElTableColumn
} from "element-plus";

type LuckyMoneyDetailRow = {
  index: number;
  amount: number;
  status: "grabbed" | "ungrabbed";
  item?: LuckyMoneyItem;
  history?: LuckyHistory;
};

export function useLuckyMoney(_tableRef: Ref) {
  const form = reactive({
    senderId: undefined as number | undefined,
    chatId: undefined as number | undefined,
    status: undefined as number | undefined
  });
  const curRow = ref<LuckyMoney>();
  const dataList = ref<LuckyMoney[]>([]);
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
      label: "发送者",
      prop: "senderName",
      minWidth: 120
    },
    {
      label: "红包金额",
      prop: "amount",
      minWidth: 100,
      formatter: ({ amount }) => `${amount.toFixed(3)} U`
    },
    {
      label: "已领取",
      prop: "received",
      minWidth: 100,
      formatter: ({ received }) => `${received.toFixed(3)} U`
    },
    {
      label: "红包数量",
      prop: "number",
      minWidth: 100
    },
    {
      label: "雷号",
      prop: "thunder",
      minWidth: 80,
      cellRenderer: scope => (
        <ElTag type="danger" size="small">
          {scope.row.thunder}
        </ElTag>
      )
    },
    {
      label: "中雷倍数",
      prop: "loseRate",
      minWidth: 100,
      formatter: ({ loseRate }) => `${loseRate.toFixed(2)}x`
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 90,
      cellRenderer: scope => {
        const status = scope.row.status;
        const statusMap = {
          1: { text: "正常", type: "success" },
          2: { text: "已退回", type: "info" }
        };
        const statusInfo = statusMap[status] || {
          text: "未知",
          type: "warning"
        };
        return <ElTag type={statusInfo.type}>{statusInfo.text}</ElTag>;
      }
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
      width: 120,
      slot: "operation"
    }
  ];

  function handleDetail(row: LuckyMoney) {
    curRow.value = row;
    getLuckyMoneyDetailAdmin(row.id)
      .then(res => {
        if (res.success && res.data) {
          const { luckyMoney, history, items } = res.data;
          addDialog({
            title: `红包详情 #${luckyMoney.id}`,
            width: "800px",
            contentRenderer: () =>
              h(DetailDialog, {
                luckyMoney,
                history,
                items,
                onChanged: onSearch
              })
          });
        } else {
          message("获取红包详情失败", { type: "error" });
        }
      })
      .catch(() => {
        message("获取红包详情失败", { type: "error" });
      });
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
      const { data } = await getLuckyMoneyListAdmin({
        ...form,
        ...pagination
      });
      dataList.value = data.list;
      pagination.total = data.total;
      pagination.pageSize = data.pageSize;
      pagination.currentPage = data.currentPage;
    } catch (error) {
      console.error("获取红包列表失败", error);
      message("获取红包列表失败", { type: "error" });
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
    handleDetail,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}

// 详情对话框组件
const DetailDialog = defineComponent({
  props: {
    luckyMoney: {
      type: Object as PropType<LuckyMoney>,
      required: true
    },
    history: {
      type: Array as PropType<LuckyHistory[]>,
      default: () => []
    },
    items: {
      type: Array as PropType<LuckyMoneyItem[]>,
      default: () => []
    },
    onChanged: {
      type: Function as PropType<() => void>,
      default: undefined
    }
  },
  setup(props) {
    const luckyMoney = ref<LuckyMoney>(props.luckyMoney);
    const history = ref<LuckyHistory[]>([...props.history]);
    const items = ref<LuckyMoneyItem[]>([...props.items]);
    const redList = ref<Array<number | string>>([]);
    const botOptions = ref<TgUser[]>([]);
    const botLoading = ref(false);
    const manualGrabFormRef = ref();
    const manualGrabForm = reactive({
      botUserId: undefined as number | undefined,
      oddEvenGuess: undefined as number | undefined
    });

    async function refreshDetail() {
      const res = await getLuckyMoneyDetailAdmin(luckyMoney.value.id);
      if (res.success && res.data) {
        luckyMoney.value = res.data.luckyMoney;
        history.value = res.data.history || [];
        items.value = res.data.items || [];
        if (luckyMoney.value.redList) {
          redList.value = JSON.parse(luckyMoney.value.redList);
        } else {
          redList.value = [];
        }
      }
    }

    async function loadBotOptions() {
      botLoading.value = true;
      try {
        const { data } = await getAdminBotUserList({
          currentPage: 0,
          pageSize: 100,
          status: 1
        });
        botOptions.value = data.list || [];
      } catch (error) {
        console.error("获取机器人列表失败", error);
        message("获取机器人列表失败", { type: "error" });
      } finally {
        botLoading.value = false;
      }
    }

    function openManualGrabDialog(row: LuckyMoneyDetailRow) {
      manualGrabForm.botUserId = undefined;
      manualGrabForm.oddEvenGuess =
        luckyMoney.value.gameMode === 1 ? 0 : undefined;
      loadBotOptions();

      addDialog({
        title: `手动抢红包 #${luckyMoney.value.id} - 第${row.index}包`,
        width: "460px",
        draggable: true,
        closeOnClickModal: false,
        contentRenderer: () => (
          <ElForm
            ref={manualGrabFormRef}
            model={manualGrabForm}
            labelWidth="96px"
          >
            <ElAlert
              class="mb-4"
              type="info"
              showIcon={true}
              closable={false}
              title={`第${row.index}包，金额 ${row.amount.toFixed(3)} U`}
            />
            <ElFormItem
              label="机器人"
              prop="botUserId"
              rules={[
                {
                  required: true,
                  message: "请选择机器人",
                  trigger: "change"
                }
              ]}
            >
              <ElSelect
                v-model={manualGrabForm.botUserId}
                filterable
                class="!w-full"
                loading={botLoading.value}
                placeholder="请选择机器人"
              >
                {botOptions.value.map(bot => (
                  <ElOption
                    key={bot.id}
                    label={`${bot.firstName || bot.username || bot.uid || bot.id} / 余额 ${Number(bot.balance || 0).toFixed(2)} U`}
                    value={bot.id}
                  />
                ))}
              </ElSelect>
            </ElFormItem>
            {luckyMoney.value.gameMode === 1 ? (
              <ElFormItem
                label="奇偶猜测"
                prop="oddEvenGuess"
                rules={[
                  {
                    required: true,
                    message: "请选择奇偶猜测",
                    trigger: "change"
                  }
                ]}
              >
                <ElRadioGroup v-model={manualGrabForm.oddEvenGuess}>
                  <ElRadio value={0}>偶</ElRadio>
                  <ElRadio value={1}>奇</ElRadio>
                </ElRadioGroup>
              </ElFormItem>
            ) : null}
          </ElForm>
        ),
        beforeSure: done => {
          manualGrabFormRef.value.validate(async valid => {
            if (!valid || !manualGrabForm.botUserId) return;
            try {
              await manualGrabLuckyMoneyAdmin({
                luckyId: luckyMoney.value.id,
                seqNo: row.index,
                botUserId: manualGrabForm.botUserId,
                oddEvenGuess: manualGrabForm.oddEvenGuess
              });
              message("手动抢红包成功", { type: "success" });
              await refreshDetail();
              props.onChanged?.();
              done();
            } catch (error) {
              console.error("手动抢红包失败", error);
              message("手动抢红包失败", { type: "error" });
            }
          });
        }
      });
    }

    const detailRows = computed<LuckyMoneyDetailRow[]>(() => {
      if (items.value.length > 0) {
        const usedHistoryIndexes = new Set<number>();
        const takeHistoryByItem = (item: LuckyMoneyItem) => {
          if (item.isGrabbed !== 1) return undefined;

          const matchedIndex = history.value.findIndex((history, index) => {
            if (usedHistoryIndexes.has(index)) return false;

            return (
              history.userId === Number(item.grabbedUid) &&
              Math.abs(Number(history.amount) - Number(item.amount)) < 0.001
            );
          });

          if (matchedIndex >= 0) {
            usedHistoryIndexes.add(matchedIndex);
            return history.value[matchedIndex];
          }

          const fallbackIndex = history.value.findIndex(
            (_history, index) => !usedHistoryIndexes.has(index)
          );
          if (fallbackIndex >= 0) {
            usedHistoryIndexes.add(fallbackIndex);
            return history.value[fallbackIndex];
          }

          return undefined;
        };

        return items.value.map(item => {
          const history = takeHistoryByItem(item);

          return {
            index: Number(item.seqNo),
            amount: Number(item.amount ?? history?.amount ?? 0),
            status: item.isGrabbed === 1 ? "grabbed" : "ungrabbed",
            item,
            history
          };
        });
      }

      const rows: LuckyMoneyDetailRow[] = redList.value.map((amount, index) => {
        const historyItem = history.value[index];

        return {
          index: index + 1,
          amount: Number(amount ?? historyItem?.amount ?? 0),
          status: historyItem ? "grabbed" : "ungrabbed",
          history: historyItem
        };
      });

      if (history.value.length > rows.length) {
        const startIndex = rows.length;
        history.value.slice(startIndex).forEach((history, index) => {
          rows.push({
            index: startIndex + index + 1,
            amount: Number(history.amount ?? 0),
            status: "grabbed",
            history
          });
        });
      }

      return rows;
    });
    const ungrabbedCount = computed(
      () => detailRows.value.filter(row => row.status === "ungrabbed").length
    );
    const ungrabbedAmount = computed(() =>
      detailRows.value
        .filter(row => row.status === "ungrabbed")
        .reduce((total, row) => total + row.amount, 0)
    );

    onMounted(() => {
      try {
        if (luckyMoney.value.redList) {
          redList.value = JSON.parse(luckyMoney.value.redList);
        }
      } catch (e) {
        console.error("解析红包列表失败", e);
      }
    });

    return () => (
      <div class="p-4">
        <el-descriptions title="红包信息" border column={2}>
          <el-descriptions-item label="发送者">
            {luckyMoney.value.senderName}
          </el-descriptions-item>
          <el-descriptions-item label="红包金额">
            {luckyMoney.value.amount.toFixed(3)} U
          </el-descriptions-item>
          <el-descriptions-item label="已领取">
            {luckyMoney.value.received.toFixed(3)} U
          </el-descriptions-item>
          <el-descriptions-item label="红包数量">
            {luckyMoney.value.number}
          </el-descriptions-item>
          <el-descriptions-item label="未抢数量">
            {ungrabbedCount.value}
          </el-descriptions-item>
          <el-descriptions-item label="未抢金额">
            {ungrabbedAmount.value.toFixed(3)} U
          </el-descriptions-item>
          <el-descriptions-item label="雷号">
            <ElTag type="danger">{luckyMoney.value.thunder}</ElTag>
          </el-descriptions-item>
          <el-descriptions-item label="中雷倍数">
            {luckyMoney.value.loseRate.toFixed(2)}x
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            {luckyMoney.value.status === 1 ? (
              <ElTag type="success">正常</ElTag>
            ) : (
              <ElTag type="info">已退回</ElTag>
            )}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {dayjs(luckyMoney.value.createdAt).format("YYYY-MM-DD HH:mm:ss")}
          </el-descriptions-item>
        </el-descriptions>

        <div class="mt-4">
          <h3 class="mb-2">红包金额列表</h3>
          <div class="flex flex-wrap gap-2">
            {redList.value.map((amount, index) => (
              <ElTag
                key={index}
                type={index < history.value.length ? "success" : "info"}
              >
                {Number(amount || 0).toFixed(3)} U
              </ElTag>
            ))}
          </div>
        </div>

        <div class="mt-4">
          <h3 class="mb-2">红包明细</h3>
          <ElTable data={detailRows.value} border>
            <ElTableColumn label="序号" prop="index" width="60" />
            <ElTableColumn label="状态" prop="status" width="90">
              {{
                default: ({ row }) =>
                  row.status === "grabbed" ? (
                    <ElTag type="success">已抢</ElTag>
                  ) : (
                    <ElTag type="info">未抢</ElTag>
                  )
              }}
            </ElTableColumn>
            <ElTableColumn label="用户" minWidth="120">
              {{
                default: ({ row }) => row.history?.firstName || "-"
              }}
            </ElTableColumn>
            <ElTableColumn label="金额" minWidth="100">
              {{
                default: ({ row }) => `${row.amount.toFixed(3)} U`
              }}
            </ElTableColumn>
            <ElTableColumn label="实际到账" minWidth="100">
              {{
                default: ({ row }) =>
                  row.history
                    ? `${Number(
                        row.history.actualAmount ||
                          (row.history.isThunder === 1
                            ? 0
                            : row.history.amount) ||
                          0
                      ).toFixed(3)} U`
                    : "-"
              }}
            </ElTableColumn>
            <ElTableColumn label="是否中雷" width="100">
              {{
                default: ({ row }) =>
                  row.history ? (
                    row.history.isThunder === 1 ? (
                      <ElTag type="danger">中雷</ElTag>
                    ) : (
                      <ElTag type="success">正常</ElTag>
                    )
                  ) : (
                    "-"
                  )
              }}
            </ElTableColumn>
            <ElTableColumn label="损失金额" minWidth="100">
              {{
                default: ({ row }) =>
                  Number(row.history?.loseMoney || 0) > 0 ? (
                    <span class="text-red-500">
                      -{row.history.loseMoney.toFixed(3)} U
                    </span>
                  ) : (
                    <span>-</span>
                  )
              }}
            </ElTableColumn>
            <ElTableColumn label="领取时间" minWidth="160">
              {{
                default: ({ row }) =>
                  row.history
                    ? dayjs(row.history.createdAt).format("YYYY-MM-DD HH:mm:ss")
                    : "-"
              }}
            </ElTableColumn>
            <ElTableColumn label="操作" fixed="right" width="100">
              {{
                default: ({ row }) =>
                  row.status === "ungrabbed" &&
                  luckyMoney.value.status === 1 &&
                  row.item ? (
                    <ElButton
                      link
                      type="primary"
                      onClick={() => openManualGrabDialog(row)}
                    >
                      手动抢
                    </ElButton>
                  ) : (
                    "-"
                  )
              }}
            </ElTableColumn>
          </ElTable>
        </div>
      </div>
    );
  }
});
