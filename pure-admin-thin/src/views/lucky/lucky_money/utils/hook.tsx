import dayjs from "dayjs";
import { message } from "@/utils/message";
import { addDialog } from "@/components/ReDialog";
import type { PaginationProps } from "@pureadmin/table";
import { deviceDetection } from "@pureadmin/utils";
import {
  getLuckyMoneyListAdmin,
  getLuckyMoneyDetailAdmin,
  type LuckyMoney,
  type LuckyHistory
} from "@/api/luckyMoney";
import { type Ref, reactive, ref, onMounted, h, defineComponent, PropType } from "vue";
import { ElMessageBox, ElTag, ElTable, ElTableColumn } from "element-plus";

export function useLuckyMoney(tableRef: Ref) {
  const form = reactive({
    senderId: undefined as number | undefined,
    chatId: undefined as number | undefined,
    status: undefined as number | undefined
  });
  const curRow = ref<LuckyMoney>();
  const formRef = ref();
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
        const statusInfo = statusMap[status] || { text: "未知", type: "warning" };
        return <ElTag type={statusInfo.type}>{statusInfo.text}</ElTag>;
      }
    },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) => dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
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
          const { luckyMoney, history } = res.data;
          addDialog({
            title: `红包详情 #${luckyMoney.id}`,
            width: "800px",
            contentRenderer: () =>
              h(DetailDialog, {
                luckyMoney,
                history
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
    }
  },
  setup(props) {
    const redList = ref<number[]>([]);
    
    onMounted(() => {
      try {
        if (props.luckyMoney.redList) {
          redList.value = JSON.parse(props.luckyMoney.redList);
        }
      } catch (e) {
        console.error("解析红包列表失败", e);
      }
    });

    return () => (
      <div class="p-4">
        <el-descriptions title="红包信息" border column={2}>
          <el-descriptions-item label="发送者">{props.luckyMoney.senderName}</el-descriptions-item>
          <el-descriptions-item label="红包金额">{props.luckyMoney.amount.toFixed(3)} U</el-descriptions-item>
          <el-descriptions-item label="已领取">{props.luckyMoney.received.toFixed(3)} U</el-descriptions-item>
          <el-descriptions-item label="红包数量">{props.luckyMoney.number}</el-descriptions-item>
          <el-descriptions-item label="雷号">
            <ElTag type="danger">{props.luckyMoney.thunder}</ElTag>
          </el-descriptions-item>
          <el-descriptions-item label="中雷倍数">{props.luckyMoney.loseRate.toFixed(2)}x</el-descriptions-item>
          <el-descriptions-item label="状态">
            {props.luckyMoney.status === 1 ? (
              <ElTag type="success">正常</ElTag>
            ) : (
              <ElTag type="info">已退回</ElTag>
            )}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {dayjs(props.luckyMoney.createdAt).format("YYYY-MM-DD HH:mm:ss")}
          </el-descriptions-item>
        </el-descriptions>
        
        <div class="mt-4">
          <h3 class="mb-2">红包金额列表</h3>
          <div class="flex flex-wrap gap-2">
            {redList.value.map((amount, index) => (
              <ElTag key={index} type={index < props.history.length ? "success" : "info"}>
                {amount.toFixed(3)} U
              </ElTag>
            ))}
          </div>
        </div>

        <div class="mt-4">
          <h3 class="mb-2">领取历史</h3>
          <ElTable data={props.history} border>
            <ElTableColumn label="序号" type="index" width="60" />
            <ElTableColumn label="用户" prop="firstName" minWidth="120" />
            <ElTableColumn label="金额" prop="amount" minWidth="100">
              {{
                default: ({ row }) => `${row.amount.toFixed(3)} U`
              }}
            </ElTableColumn>
            <ElTableColumn label="实际到账" prop="actualAmount" minWidth="100">
              {{
                default: ({ row }) =>
                  `${Number(row.actualAmount || (row.isThunder === 1 ? 0 : row.amount) || 0).toFixed(3)} U`
              }}
            </ElTableColumn>
            <ElTableColumn label="是否中雷" prop="isThunder" width="100">
              {{
                default: ({ row }) =>
                  row.isThunder === 1 ? (
                    <ElTag type="danger">💣 中雷</ElTag>
                  ) : (
                    <ElTag type="success">💵 正常</ElTag>
                  )
              }}
            </ElTableColumn>
            <ElTableColumn label="损失金额" prop="loseMoney" minWidth="100">
              {{
                default: ({ row }) =>
                  row.loseMoney > 0 ? (
                    <span class="text-red-500">-{row.loseMoney.toFixed(3)} U</span>
                  ) : (
                    <span>-</span>
                  )
              }}
            </ElTableColumn>
            <ElTableColumn label="领取时间" prop="createdAt" minWidth="160">
              {{
                default: ({ row }) => dayjs(row.createdAt).format("YYYY-MM-DD HH:mm:ss")
              }}
            </ElTableColumn>
          </ElTable>
        </div>
      </div>
    );
  }
});
